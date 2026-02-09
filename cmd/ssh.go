package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh <sandbox-name>",
	Short: "SSH into a running Molty instance",
	Long: `SSH into a running Molty instance by name or sandbox ID.

Automatically provisions temporary SSH credentials, connects to the sandbox
interactively, and revokes the credentials when the session ends.

The sandbox must be in a running state to establish SSH access.

Examples:
  clawdepl ssh my-bot            # SSH by instance name
  clawdepl ssh sandbox_abc123    # SSH by sandbox ID`,
	Args: cobra.ExactArgs(1),
	RunE: runSSH,
}

func init() {
	rootCmd.AddCommand(sshCmd)
}

func runSSH(cmd *cobra.Command, args []string) error {
	// Check if logged in (or using unsafe token in debug builds)
	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdepl login' first.")
		return nil
	}

	nameOrID := args[0]

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Resolve name to sandbox ID if needed
	sandboxID := nameOrID
	if !strings.HasPrefix(nameOrID, "sandbox_") {
		// Look up bot by name
		bots, err := client.ListBots(ctx)
		if err != nil {
			return fmt.Errorf("failed to list instances: %w", err)
		}

		found := false
		for _, bot := range bots {
			if bot.Name == nameOrID {
				sandboxID = bot.SandboxID
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("✗ Instance '%s' not found\n", nameOrID)
			fmt.Println("\nRun 'clawdepl list' to see available instances")
			return nil
		}
	}

	// Check sandbox status - must be running
	status, err := client.CheckSandboxStatus(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if !isRunningState(status.State) {
		fmt.Printf("✗ Cannot SSH to '%s': sandbox is %s (must be running)\n", nameOrID, status.State)
		fmt.Printf("\nStart it first with: clawdepl start %s\n", nameOrID)
		return nil
	}

	// Check if SSH client is available
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		fmt.Println("✗ SSH client not found")
		fmt.Println("\nPlease install OpenSSH client:")
		fmt.Println("  Ubuntu/Debian: sudo apt install openssh-client")
		fmt.Println("  macOS: SSH is pre-installed")
		fmt.Println("  Windows: Install OpenSSH via Settings → Apps → Optional Features")
		return nil
	}

	// Provision SSH access
	fmt.Printf("Provisioning SSH access to '%s'...\n", nameOrID)
	sshResp, err := client.CreateSSH(ctx, sandboxID, 60)
	if err != nil {
		return fmt.Errorf("failed to provision SSH access: %w", err)
	}

	fmt.Printf("✓ SSH access provisioned (expires in %d minutes)\n", sshResp.ExpiresInMinutes)

	// Setup cleanup handling
	var revoked bool
	revokeSSH := func() {
		if revoked {
			return
		}
		revoked = true

		fmt.Println("\nRevoking SSH access...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := client.RevokeSSH(ctx, sandboxID, sshResp.SSHToken)
		if err != nil {
			fmt.Printf("⚠ Warning: failed to revoke SSH token: %v\n", err)
			fmt.Println("  (Token will auto-expire in 60 minutes)")
		} else {
			fmt.Println("✓ SSH access revoked")
		}
	}

	// Ensure cleanup always happens
	defer revokeSSH()

	// Setup signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		revokeSSH()
		os.Exit(0)
	}()

	// Parse SSH command to extract user and host
	// Expected format: "ssh token@ssh.app.daytona.io"
	sshCmd := strings.TrimSpace(sshResp.SSHCommand)
	parts := strings.Fields(sshCmd)
	if len(parts) < 2 {
		return fmt.Errorf("invalid SSH command format: %s", sshCmd)
	}
	userHost := parts[1] // "token@ssh.app.daytona.io"

	// Launch SSH client
	fmt.Println("Connecting to sandbox...")
	fmt.Println()

	sshProcess := exec.Command(sshPath, userHost)
	sshProcess.Stdin = os.Stdin
	sshProcess.Stdout = os.Stdout
	sshProcess.Stderr = os.Stderr

	if err := sshProcess.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// SSH exited with non-zero status - this is normal for user logout
			if exitErr.ExitCode() == 255 {
				// SSH connection error
				return fmt.Errorf("SSH connection failed: %w", err)
			}
		} else {
			return fmt.Errorf("failed to run SSH: %w", err)
		}
	}

	return nil
}

func isRunningState(state string) bool {
	switch state {
	case "running", "ready", "active":
		return true
	default:
		return false
	}
}
