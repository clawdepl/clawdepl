package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/clawdepl/clawdepl/internal/tui"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new Molty instance",
	Long: `Create a new Molty (AI agent) instance with an interactive wizard.

If a name is provided, the wizard will skip the name prompt.
The wizard will guide you through:
  1. Instance name (if not provided)
  2. Claude API token
  3. Purpose/description (vibe)

Examples:
  clawdepl new              # Interactive wizard
  clawdepl new my-agent     # Skip name prompt`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Check if logged in (or using unsafe token in debug builds)
	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdepl login' first.")
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
	result, err := tui.RunNewInstanceWizard(name, func(name, apiKey, vibe string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Create sandbox via Supabase API
		createResp, err := client.CreateSandbox(ctx, name, apiKey, vibe)
		if err != nil {
			return fmt.Errorf("failed to create sandbox: %w", err)
		}

		if !createResp.Success {
			return fmt.Errorf("sandbox creation failed: %s", createResp.Error)
		}

		// Poll until sandbox is ready
		_, err = client.WaitForReady(ctx, createResp.Sandbox.ID, func(state string, ready bool) {
			// Progress updates are handled by the TUI spinner
		})
		if err != nil {
			return fmt.Errorf("provisioning failed: %w", err)
		}

		return nil
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
