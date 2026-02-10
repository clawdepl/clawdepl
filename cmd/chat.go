package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var (
	chatRunnerCommand string
	chatTimeout       int
)

var chatCmd = &cobra.Command{
	Use:   "chat <sandbox-name> <message>",
	Short: "Send a message to a running instance without SSH",
	Long: `Send a message to a running instance directly from the CLI.

This command creates a temporary sandbox session, runs a prompt command,
prints the output, then closes the session.

Examples:
  clawdepl chat wifey "Hi, introduce yourself"
  clawdepl chat wifey "What can you do?" --runner "openclaw agent --session-id {{session}} -m"
  clawdepl chat wifey "Status report" --runner "python /app/agent.py prompt"`,
	Args: cobra.ExactArgs(2),
	RunE: runChat,
}

func init() {
	chatCmd.Flags().StringVar(&chatRunnerCommand, "runner", `bash -lc "if command -v openclaw >/dev/null 2>&1; then openclaw agent --session-id {{session}} -m {{message}}; else node /app/openclaw.mjs agent --session-id {{session}} -m {{message}}; fi"`, "Base command used inside sandbox to send the message. Use {{session}} and {{message}} placeholders.")
	chatCmd.Flags().IntVar(&chatTimeout, "timeout", 120, "Execution timeout in seconds")
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	if err := requireLogin(); err != nil {
		return err
	}

	nameOrID := args[0]
	message := strings.TrimSpace(args[1])
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	runner := strings.TrimSpace(chatRunnerCommand)
	if runner == "" {
		return fmt.Errorf("--runner cannot be empty")
	}
	if chatTimeout <= 0 {
		return fmt.Errorf("--timeout must be greater than 0")
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Use a short context for discovery and session setup.
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer setupCancel()

	sandboxID, err := resolveSandboxID(setupCtx, client, nameOrID)
	if err != nil {
		return err
	}

	status, err := client.CheckSandboxStatus(setupCtx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}
	if !isRunningState(status.State) {
		return fmt.Errorf("cannot chat with '%s': sandbox is %s (start it first with: clawdepl start %s)", nameOrID, status.State, nameOrID)
	}

	sessionID := fmt.Sprintf("chat-%d", time.Now().UnixNano())
	createResp, err := client.CreateSession(setupCtx, sandboxID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	if createResp != nil && createResp.SessionID != "" {
		sessionID = createResp.SessionID
	}

	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = client.DeleteSession(cleanupCtx, sandboxID, sessionID)
	}()

	runner = strings.ReplaceAll(runner, "{{session}}", sessionID)
	if strings.Contains(runner, "{{message}}") {
		// Prefer in-script substitution so bash -lc doesn't treat the message as $0.
		runner = strings.ReplaceAll(runner, "{{message}}", shellQuote(message))
	} else {
		// Backwards-compatible behavior for custom runners without {{message}}.
		runner = fmt.Sprintf("%s %s", runner, shellQuote(message))
	}
	fullCommand := runner

	// Exec can legitimately take longer than 30s; give it its own context.
	execCtx, execCancel := context.WithTimeout(context.Background(), time.Duration(chatTimeout+30)*time.Second)
	defer execCancel()

	execResp, err := client.ExecCommand(execCtx, sandboxID, sessionID, fullCommand, chatTimeout)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	if execResp.Output != "" {
		fmt.Print(execResp.Output)
	}
	if execResp.ExitCode != 0 {
		if strings.Contains(execResp.Output, "No API key found for provider") {
			return fmt.Errorf("instance is missing OpenClaw credentials. Run: clawdepl configure %s --claude-token <credential> --auth-choice api-key|setup-token", nameOrID)
		}
		if execResp.ExitCode == 127 {
			return fmt.Errorf("chat runner not found inside sandbox (exit 127). Try --runner \"bash -lc\" and run a command like: openclaw agent ...")
		}
		return fmt.Errorf("chat command failed with exit code %d", execResp.ExitCode)
	}

	return nil
}

func resolveSandboxID(ctx context.Context, client *api.Client, nameOrID string) (string, error) {
	// Allow passing the raw Daytona sandbox ID directly.
	// Historically IDs were "sandbox_..."; newer backends may return UUIDs.
	if strings.HasPrefix(nameOrID, "sandbox_") || looksLikeUUID(nameOrID) {
		return nameOrID, nil
	}

	bots, err := client.ListBots(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list instances: %w", err)
	}

	for _, bot := range bots {
		if strings.EqualFold(bot.Name, nameOrID) {
			if bot.SandboxID == "" {
				return "", fmt.Errorf("instance '%s' does not have a sandbox ID yet", nameOrID)
			}
			return bot.SandboxID, nil
		}
	}

	return "", fmt.Errorf("instance '%s' not found (run 'clawdepl list' to see available instances)", nameOrID)
}

func looksLikeUUID(s string) bool {
	// Accept canonical UUID string forms (case-insensitive):
	//   xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	t := strings.TrimSpace(s)
	if len(t) != 36 {
		return false
	}
	for i := 0; i < len(t); i++ {
		c := t[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			isHex := (c >= '0' && c <= '9') ||
				(c >= 'a' && c <= 'f') ||
				(c >= 'A' && c <= 'F')
			if !isHex {
				return false
			}
		}
	}
	return true
}

func shellQuote(value string) string {
	// Single-quote for POSIX shell and escape embedded single-quotes.
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
