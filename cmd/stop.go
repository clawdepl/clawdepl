package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop an OpenClaw instance",
	Long: `Stop a running OpenClaw instance.

Examples:
  clawdpl stop my-agent`,
	Args: cobra.ExactArgs(1),
	RunE: runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdpl login' first.")
		return nil
	}

	name := args[0]

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	instance, err := client.StopInstance(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	fmt.Printf("✓ Stopped '%s' (status: %s)\n", instance.Name, instance.Status)
	return nil
}
