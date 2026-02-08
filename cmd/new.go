package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/moltyverse/clawdpl/internal/tui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new OpenClaw instance",
	Long: `Create a new OpenClaw instance with an interactive wizard.

If a name is provided, the wizard will skip the name prompt.
The wizard will guide you through:
  1. Instance name (if not provided)
  2. Claude API token
  3. Purpose/description

Examples:
  clawdpl new              # Interactive wizard
  clawdpl new my-agent     # Skip name prompt`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdpl login' first.")
		return nil
	}

	// Get optional name from args
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	// Create API client
	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Run the wizard
	result, err := tui.RunNewInstanceWizard(name, func(name, token, purpose string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Create the instance
		_, err := client.CreateInstance(ctx, &api.CreateInstanceRequest{
			Name:        name,
			ClaudeToken: token,
			Purpose:     purpose,
		})
		if err != nil {
			return err
		}

		// Wait for provisioning (mock: 3 seconds)
		_, err = client.WaitForProvisioning(ctx, name)
		return err
	})

	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	if result.Cancelled {
		fmt.Println("\nCancelled.")
		return nil
	}

	if result.Error != nil {
		return fmt.Errorf("failed to create instance: %w", result.Error)
	}

	return nil
}

// RunNewFlow runs the new instance flow (exported for use by root command)
func RunNewFlow() error {
	return runNew(newCmd, nil)
}
