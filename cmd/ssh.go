package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
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
	if err := requireLogin(); err != nil {
		return err
	}

	nameOrID := args[0]

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sandboxID, err := resolveSandboxID(ctx, client, nameOrID)
	if err != nil {
		return err
	}

	// Check sandbox status - must be running
	status, err := client.CheckSandboxStatus(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if !isRunningState(status.State) {
		return fmt.Errorf("cannot SSH to '%s': sandbox is %s (start it first with: clawdepl start %s)", nameOrID, status.State, nameOrID)
	}

	// Check if SSH client is available
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("SSH client not found (install OpenSSH client and retry)")
	}

	// Provision SSH access
	fmt.Printf("Provisioning SSH access to '%s'...\n", nameOrID)
	sshResp, err := client.CreateSSH(ctx, sandboxID, 60)
	if err != nil {
		return fmt.Errorf("failed to provision SSH access: %w", err)
	}

	fmt.Printf("✓ SSH access provisioned (expires in %d minutes)\n", sshResp.ExpiresInMinutes)

	// Setup cleanup handling
	var revokeOnce sync.Once
	revokeSSH := func() {
		revokeOnce.Do(func() {
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
		})
	}

	// Ensure cleanup always happens
	defer revokeSSH()

	// Setup signal handler for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		revokeSSH()
		os.Exit(130)
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
	s := strings.ToLower(strings.TrimSpace(state))
	switch s {
	case "running", "ready", "active", "started":
		return true
	default:
		return false
	}
}
