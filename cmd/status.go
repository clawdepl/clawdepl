package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <sandbox-id|name>",
	Short: "Show status of a Molty instance",
	Long: `Show detailed status information for a Molty instance by sandbox ID or instance name.

Displays sandbox state and readiness.

Examples:
  clawdepl status sandbox_abc123
  clawdepl status wifey`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if err := requireLogin(); err != nil {
		return err
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sandboxID, err := resolveSandboxID(ctx, client, args[0])
	if err != nil {
		return err
	}

	status, err := client.CheckSandboxStatus(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Display status
	fmt.Printf("Sandbox: %s\n", sandboxID)
	fmt.Printf("  State:  %s\n", formatStatusDetailed(status.State))
	fmt.Printf("  Ready:  %v\n", status.Ready)

	return nil
}

func formatStatusDetailed(state string) string {
	switch state {
	case "running", "ready", "active", "started":
		return "● Running"
	case "stopped":
		return "○ Stopped"
	case "provisioning", "starting":
		return "◐ Provisioning"
	case "error", "failed":
		return "✗ Error"
	default:
		return state
	}
}
