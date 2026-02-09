package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/moltyverse/clawdepl/internal/api"
	"github.com/moltyverse/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <sandbox-id>",
	Short: "Show status of a Molty instance",
	Long: `Show detailed status information for a Molty instance by sandbox ID.

Displays status, stage, progress, and gateway URL if available.

Examples:
  clawdepl status sandbox_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Check if logged in (or using unsafe token in debug builds)
	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdepl login' first.")
		return nil
	}

	sandboxID := args[0]

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := client.GetStatus(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Display status
	fmt.Printf("Sandbox: %s\n", status.SandboxID)
	if status.MoltyName != "" {
		fmt.Printf("  Name:     %s\n", status.MoltyName)
	}
	fmt.Printf("  Status:   %s\n", formatStatusDetailed(status.Status))
	if status.Stage != "" {
		fmt.Printf("  Stage:    %s\n", status.Stage)
	}
	if status.Progress > 0 {
		fmt.Printf("  Progress: %d%%\n", status.Progress)
	}
	if status.Message != "" {
		fmt.Printf("  Message:  %s\n", status.Message)
	}
	if status.GatewayURL != "" {
		fmt.Printf("  Gateway:  %s\n", status.GatewayURL)
	}

	return nil
}

func formatStatusDetailed(status string) string {
	switch status {
	case "running", "ready":
		return "● Running"
	case "stopped":
		return "○ Stopped"
	case "provisioning":
		return "◐ Provisioning"
	case "error", "failed":
		return "✗ Error"
	default:
		return status
	}
}
