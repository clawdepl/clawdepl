package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure <sandbox-name>",
	Short: "Configure OpenClaw credentials inside an existing instance",
	Long: `Configure OpenClaw auth inside a running instance so 'chat' works immediately.

This writes OpenClaw config and auth profiles inside the sandbox (no interactive onboarding).`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigure,
}

func init() {
	configureCmd.Flags().StringVar(&newClaudeToken, "claude-token", "", "Claude credential (API key or setup-token)")
	configureCmd.Flags().StringVar(&newAuthChoice, "auth-choice", "api-key", "Anthropic auth mode: api-key or setup-token")
	configureCmd.Flags().StringVar(&newModel, "model", api.DefaultOpenClawModel, "Primary model (provider/model)")
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	if err := requireLogin(); err != nil {
		return err
	}

	nameOrID := args[0]
	credential := strings.TrimSpace(newClaudeToken)
	if credential == "" {
		return fmt.Errorf("--claude-token is required")
	}

	authMethod, err := api.CanonicalAnthropicAuthChoice(newAuthChoice)
	if err != nil {
		return fmt.Errorf("invalid auth choice: %w", err)
	}

	if err := api.ValidateASCIICredential(credential); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Claude Code OAuth tokens from `claude setup-token` commonly start with "sk-ant-oat".
	if authMethod == "anthropic-api-key" && strings.HasPrefix(credential, "sk-ant-oat") {
		return fmt.Errorf("this is a Claude Code OAuth token (sk-ant-oat...), not an Anthropic API key. Use --auth-choice setup-token")
	}
	if authMethod == "anthropic-setup-token" && strings.HasPrefix(credential, "sk-ant-api") {
		return fmt.Errorf("this looks like an Anthropic API key (sk-ant-api...), but you selected setup-token. Use --auth-choice api-key")
	}
	if authMethod == "anthropic-api-key" {
		validateCtx, validateCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer validateCancel()
		if err := api.ValidateAnthropicAPIKey(validateCtx, credential); err != nil {
			return fmt.Errorf("token validation failed: %w", err)
		}
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	sandboxID, err := resolveSandboxID(ctx, client, nameOrID)
	if err != nil {
		return err
	}

	if err := configureOpenClawInSandbox(ctx, client, sandboxID, credential, authMethod, newModel); err != nil {
		return err
	}

	fmt.Println("✓ OpenClaw configured")
	return nil
}
