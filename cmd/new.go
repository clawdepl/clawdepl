package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/tui"
	"github.com/spf13/cobra"
)

var (
	newClaudeToken string
	newPurpose     string
	newAuthChoice  string
	newModel       string
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new Molty instance",
	Long: `Create a new Molty (AI agent) instance with an interactive wizard.

If a name is provided, the wizard will skip the name prompt.
The wizard will guide you through:
  1. Instance name (if not provided)
  2. Claude credential (API key or setup-token)
  3. Identity prompt (IDENTITY.md content)

For CI/non-interactive usage, provide all required flags:
  clawdepl new my-agent --claude-token <token> --purpose <text> [--auth-choice api-key|setup-token] [--model provider/model]

Examples:
  clawdepl new              # Interactive wizard
  clawdepl new my-agent     # Skip name prompt
  clawdepl new my-agent --claude-token "$ANTHROPIC_API_KEY" --auth-choice api-key --model anthropic/claude-sonnet-4-5 --purpose "CI instance"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	newCmd.Flags().StringVar(&newClaudeToken, "claude-token", "", "Claude credential (API key or setup-token) for non-interactive mode")
	newCmd.Flags().StringVar(&newPurpose, "purpose", "", "Instance identity prompt (IDENTITY.md content) (non-interactive mode)")
	newCmd.Flags().StringVar(&newAuthChoice, "auth-choice", "api-key", "Anthropic auth mode: api-key or setup-token")
	newCmd.Flags().StringVar(&newModel, "model", api.DefaultOpenClawModel, "Primary model (provider/model), e.g. anthropic/claude-opus-4-6")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	if err := requireLogin(); err != nil {
		return err
	}

	// Get optional name from args
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	if isNonInteractiveNewMode() {
		return runNewNonInteractive(client, name)
	}

	var outcome provisionOutcome

	// Run the wizard
	result, err := tui.RunNewInstanceWizard(name, func(name, credential, authChoice, vibe string) error {
		var err error
		outcome, err = provisionSandboxWithOutcome(client, name, credential, authChoice, newModel, vibe)
		return err
	})

	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	if result.Cancelled {
		fmt.Println("\nCancelled.")
		return nil
	}

	if result.Error != nil {
		return fmt.Errorf("failed to create instance: %w", result.Error)
	}

	printProvisionOutcome(result.Name, outcome)
	return nil
}

// RunNewFlow runs the new instance flow (exported for use by root command)
func RunNewFlow() error {
	return runNew(newCmd, nil)
}

func isNonInteractiveNewMode() bool {
	return strings.TrimSpace(newClaudeToken) != "" || strings.TrimSpace(newPurpose) != ""
}

func runNewNonInteractive(client *api.Client, name string) error {
	name = strings.TrimSpace(name)
	token := strings.TrimSpace(newClaudeToken)
	purpose := strings.TrimSpace(newPurpose)

	if name == "" {
		return fmt.Errorf("name is required in non-interactive mode")
	}
	if token == "" {
		return fmt.Errorf("--claude-token is required in non-interactive mode")
	}
	if purpose == "" {
		return fmt.Errorf("--purpose is required in non-interactive mode")
	}

	outcome, err := provisionSandboxWithOutcome(client, name, token, newAuthChoice, newModel, purpose)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Instance '%s' created successfully\n", name)
	printProvisionOutcome(name, outcome)
	return nil
}

type provisionOutcome struct {
	SandboxID      string
}

func printProvisionOutcome(requestedName string, outcome provisionOutcome) {
	if strings.TrimSpace(outcome.SandboxID) != "" {
		// Keep this line stable for scripting.
		fmt.Printf("sandbox_id: %s\n", outcome.SandboxID)
	}
}

func provisionSandboxWithOutcome(client *api.Client, name, anthropicCredential, authChoice, model, vibe string) (provisionOutcome, error) {
	out := provisionOutcome{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	authMethod, err := api.CanonicalAnthropicAuthChoice(authChoice)
	if err != nil {
		return out, fmt.Errorf("invalid auth choice: %w", err)
	}

	if strings.TrimSpace(anthropicCredential) == "" {
		return out, fmt.Errorf("token validation failed: Claude credential is required")
	}

	if err := api.ValidateASCIICredential(anthropicCredential); err != nil {
		return out, fmt.Errorf("token validation failed: %w", err)
	}

	credTrim := strings.TrimSpace(anthropicCredential)

	// Reduce user confusion: Claude Code OAuth tokens are not Anthropic API keys.
	// `claude setup-token` outputs an OAuth token that commonly starts with "sk-ant-oat".
	if authMethod == "anthropic-api-key" && strings.HasPrefix(credTrim, "sk-ant-oat") {
		return out, fmt.Errorf("token validation failed: this is a Claude Code OAuth token (sk-ant-oat...), not an Anthropic API key. Re-run and choose --auth-choice setup-token")
	}
	if authMethod == "anthropic-setup-token" && strings.HasPrefix(credTrim, "sk-ant-api") {
		return out, fmt.Errorf("token validation failed: this looks like an Anthropic API key (sk-ant-api...), but you selected setup-token. Re-run and choose --auth-choice api-key")
	}

	// Anthropic setup-token is not an x-api-key; only validate API keys against Anthropic API.
	if authMethod == "anthropic-api-key" {
		validateCtx, validateCancel := context.WithTimeout(ctx, 20*time.Second)
		defer validateCancel()

		if err := api.ValidateAnthropicAPIKey(validateCtx, anthropicCredential); err != nil {
			return out, fmt.Errorf("token validation failed: %w", err)
		}
	}

	modelRef, err := api.ValidateOpenClawModel(model)
	if err != nil {
		return out, fmt.Errorf("invalid model: %w", err)
	}
	_ = modelRef // current backend schema doesn't accept model; keep validation for future compatibility.

	anthropicCredentialType := "api_key"
	if authMethod == "anthropic-setup-token" {
		anthropicCredentialType = "token"
	}

	createResp, err := client.CreateSandbox(ctx, &api.CreateSandboxRequest{
		MoltyName:               name,
		AnthropicCredentialType: anthropicCredentialType,
		AnthropicCredential:     anthropicCredential,
		MoltyPrompt:             vibe,
	})
	if err != nil {
		return out, fmt.Errorf("failed to create sandbox: %w", err)
	}

	// Support both legacy and new response shapes.
	if createResp.Error != "" {
		return out, fmt.Errorf("sandbox creation failed: %s", createResp.Error)
	}
	if createResp.Success == false && createResp.SandboxID == "" && createResp.Sandbox.ID == "" {
		return out, fmt.Errorf("sandbox creation failed: unexpected response")
	}

	sandboxID := createResp.EffectiveSandboxID()
	if sandboxID == "" {
		return out, fmt.Errorf("sandbox creation failed: missing sandbox_id in response")
	}
	out.SandboxID = sandboxID

	_, err = client.WaitForReady(ctx, sandboxID, func(state string, ready bool) {
		// Progress updates are handled by the TUI spinner.
	})
	if err != nil {
		return out, fmt.Errorf("provisioning failed: %w", err)
	}

	return out, nil
}

func configureOpenClawInSandbox(ctx context.Context, client *api.Client, sandboxID, credential, authMethod, model string) error {
	sessionID := fmt.Sprintf("cfg-%d", time.Now().UnixNano())
	createResp, err := client.CreateSession(ctx, sandboxID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	if createResp != nil && createResp.SessionID != "" {
		sessionID = createResp.SessionID
	}
	defer func() { _, _ = client.DeleteSession(context.Background(), sandboxID, sessionID) }()

	modelRef, err := api.ValidateOpenClawModel(model)
	if err != nil {
		return err
	}
	_ = modelRef // backend-injected config owns model; keep validation only.

	// Avoid network-dependent onboarding; write config + auth profiles directly (legacy support / manual configure).
	type authProfile struct {
		Type     string `json:"type"`
		Provider string `json:"provider"`
		Token    string `json:"token,omitempty"`
		APIKey   string `json:"apiKey,omitempty"`
	}

	profileType := "api-key"
	if authMethod == "anthropic-setup-token" {
		profileType = "token"
	}
	profileName := "anthropic:default"

	authProfiles := map[string]authProfile{
		profileName: {
			Type:     profileType,
			Provider: "anthropic",
		},
	}
	if profileType == "token" {
		authProfiles[profileName] = authProfile{Type: profileType, Provider: "anthropic", Token: credential}
	} else {
		authProfiles[profileName] = authProfile{Type: profileType, Provider: "anthropic", APIKey: credential}
	}

	authDoc := map[string]any{
		"version":  1,
		"profiles": authProfiles,
		"lastGood": map[string]string{
			"anthropic": profileName,
		},
		"usageStats": map[string]any{},
	}

	authJSON, err := json.MarshalIndent(authDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to build auth-profiles.json: %w", err)
	}

	// For setup-token, OpenClaw provider behavior differs across builds. Some treat it as a
	// bearer token, others treat it as an x-api-key. We'll write token-mode first and probe;
	// if we see "Invalid bearer token", we rewrite as api-key with the same credential.
	buildAuthJSON := func(profileType string) (string, error) {
		authProfiles := map[string]authProfile{
			profileName: {
				Type:     profileType,
				Provider: "anthropic",
			},
		}
		if profileType == "token" {
			authProfiles[profileName] = authProfile{Type: profileType, Provider: "anthropic", Token: credential}
		} else {
			authProfiles[profileName] = authProfile{Type: profileType, Provider: "anthropic", APIKey: credential}
		}

		authDoc := map[string]any{
			"version":  1,
			"profiles": authProfiles,
			"lastGood": map[string]string{
				"anthropic": profileName,
			},
			"usageStats": map[string]any{},
		}

		b, err := json.MarshalIndent(authDoc, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	authJSONStr := string(authJSON)
	if authMethod == "anthropic-setup-token" {
		// Start with token-mode.
		if s, err := buildAuthJSON("token"); err == nil {
			authJSONStr = s
		}
	}

	authB64 := base64.StdEncoding.EncodeToString([]byte(authJSONStr))

	script := `
set -euo pipefail

mkdir -p "$HOME/.openclaw/agents/main/agent" "$HOME/.openclaw/workspace"

AUTH_FILE="$HOME/.openclaw/agents/main/agent/auth-profiles.json"
OC_FILE="$HOME/.openclaw/openclaw.json"

echo '` + authB64 + `' | base64 -d > "$AUTH_FILE"

# Patch injected config in-place: preserve existing fields (including gateway auth token),
# but force bind=lan for container networking and ensure chatCompletions endpoint is enabled.
if [ -s "$OC_FILE" ]; then
  node - <<'NODE'
const fs = require('fs');
const p = process.env.HOME + "/.openclaw/openclaw.json";
let c = {};
try { c = JSON.parse(fs.readFileSync(p, 'utf8')); } catch {}
c.gateway = c.gateway || {};
c.gateway.mode = c.gateway.mode || "local";
c.gateway.bind = "lan";
c.gateway.http = c.gateway.http || {};
c.gateway.http.endpoints = c.gateway.http.endpoints || {};
c.gateway.http.endpoints.chatCompletions = c.gateway.http.endpoints.chatCompletions || {};
c.gateway.http.endpoints.chatCompletions.enabled = true;
fs.writeFileSync(p, JSON.stringify(c, null, 2));
NODE
fi

if command -v openclaw >/dev/null 2>&1; then
  pkill -f 'openclaw gateway' >/dev/null 2>&1 || true
  nohup bash -lc 'openclaw gateway --allow-unconfigured --bind lan' > "$HOME/.openclaw/gateway.log" 2>&1 &
elif [ -f /app/openclaw.mjs ]; then
  pkill -f 'openclaw.mjs gateway' >/dev/null 2>&1 || true
  nohup bash -lc 'node /app/openclaw.mjs gateway --allow-unconfigured --bind lan' > "$HOME/.openclaw/gateway.log" 2>&1 &
fi
`

	resp, err := client.ExecCommand(ctx, sandboxID, sessionID, "bash -lc "+shellQuote(script), 180)
	if err != nil {
		return fmt.Errorf("openclaw configure failed: %w", err)
	}
	if resp.ExitCode != 0 {
		if resp.Output != "" {
			return fmt.Errorf("openclaw configure failed: %s", strings.TrimSpace(resp.Output))
		}
		return fmt.Errorf("openclaw configure failed with exit code %d", resp.ExitCode)
	}

	if authMethod == "anthropic-setup-token" {
		probeCmd := `node /app/openclaw.mjs models status --probe 2>&1 || true`
		probe, _ := client.ExecCommand(ctx, sandboxID, sessionID, "bash -lc "+shellQuote(probeCmd), 60)
		if probe != nil && strings.Contains(probe.Output, "Invalid bearer token") {
			if s, err := buildAuthJSON("api-key"); err == nil {
				fixB64 := base64.StdEncoding.EncodeToString([]byte(s))
				fixScript := `
set -euo pipefail
AUTH_FILE="$HOME/.openclaw/agents/main/agent/auth-profiles.json"
echo '` + fixB64 + `' | base64 -d > "$AUTH_FILE"
`
				fix, err := client.ExecCommand(ctx, sandboxID, sessionID, "bash -lc "+shellQuote(fixScript), 60)
				if err != nil {
					return fmt.Errorf("openclaw configure failed: %w", err)
				}
				if fix.ExitCode != 0 {
					return fmt.Errorf("openclaw configure failed: %s", strings.TrimSpace(fix.Output))
				}
			}
		}
	}

	return nil
}
