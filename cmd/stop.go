package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <sandbox-id|name>",
	Short: "Stop a Molty instance",
	Long: `Stop a running Molty instance by sandbox ID or instance name.

Examples:
  clawdepl stop sandbox_abc123
  clawdepl stop wifey`,
	Args: cobra.ExactArgs(1),
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("failed to stop '%s': %s", sandboxID, msg)
	}
	return nil
}
