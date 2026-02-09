package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <sandbox-id>",
	Short: "Start a Molty instance",
	Long: `Start a stopped Molty instance by sandbox ID.

Examples:
  clawdepl start sandbox_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
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

	result, err := client.StartSandbox(ctx, sandboxID)
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	if result.Success {
		fmt.Printf("✓ Started '%s'\n", sandboxID)
	} else {
		msg := result.Message
		if msg == "" {
			msg = result.Error
		}
		fmt.Printf("✗ Failed to start '%s': %s\n", sandboxID, msg)
	}
	return nil
}
