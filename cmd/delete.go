package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/moltyverse/clawdepl/internal/api"
	"github.com/moltyverse/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var (
	deleteYes bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <sandbox-id> [sandbox-id2...]",
	Short: "Delete one or more Molty instances",
	Long: `Delete one or more Molty instances by sandbox ID.

By default, prompts for confirmation before deleting. Use -y to skip.

Examples:
  clawdepl delete sandbox_abc123           # Delete with confirmation
  clawdepl delete sandbox_abc123 -y        # Delete without confirmation
  clawdepl delete sandbox_1 sandbox_2      # Delete multiple instances`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Check if logged in (or using unsafe token in debug builds)
	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdepl login' first.")
		return nil
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Verify instances exist first by checking status
	var toDelete []string
	for _, sandboxID := range args {
		_, err := client.GetStatus(ctx, sandboxID)
		if err != nil {
			fmt.Printf("Warning: sandbox '%s' not found or inaccessible, skipping\n", sandboxID)
			continue
		}
		toDelete = append(toDelete, sandboxID)
	}

	if len(toDelete) == 0 {
		fmt.Println("No valid instances to delete.")
		return nil
	}

	// Confirm deletion
	if !deleteYes {
		var prompt string
		if len(toDelete) == 1 {
			prompt = fmt.Sprintf("Are you sure you want to delete '%s'? [y/N] ", toDelete[0])
		} else {
			prompt = fmt.Sprintf("Are you sure you want to delete %d instances (%s)? [y/N] ",
				len(toDelete), strings.Join(toDelete, ", "))
		}

		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete instances
	var deleted, failed int
	for _, sandboxID := range toDelete {
		result, err := client.Deprovision(ctx, sandboxID)
		if err != nil {
			fmt.Printf("✗ Failed to delete '%s': %v\n", sandboxID, err)
			failed++
		} else if !result.Success {
			fmt.Printf("✗ Failed to delete '%s': %s\n", sandboxID, result.Message)
			failed++
		} else {
			fmt.Printf("✓ Deleted '%s'\n", sandboxID)
			deleted++
		}
	}

	if failed > 0 {
		return fmt.Errorf("failed to delete %d instance(s)", failed)
	}

	return nil
}
