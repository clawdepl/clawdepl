package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <sandbox-id>",
	Short: "Stop a Molty instance",
	Long: `Stop a running Molty instance by sandbox ID.

Examples:
  clawdepl stop sandbox_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
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

	result, err := client.StopSandbox(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	if result.Success {
		fmt.Printf("✓ Stopped '%s'\n", sandboxID)
	} else {
		msg := result.Message
		if msg == "" {
			msg = result.Error
		}
		fmt.Printf("✗ Failed to stop '%s': %s\n", sandboxID, msg)
	}
	return nil
}
