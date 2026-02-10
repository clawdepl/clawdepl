package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <sandbox-id|name>",
	Short: "Start a Molty instance",
	Long: `Start a stopped Molty instance by sandbox ID or instance name.

Examples:
  clawdepl start sandbox_abc123
  clawdepl start wifey`,
	Args: cobra.ExactArgs(1),
	RunE: runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("failed to start '%s': %s", sandboxID, msg)
	}
	return nil
}
