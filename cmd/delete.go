package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/spf13/cobra"
)

var (
	deleteYes bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <sandbox-id|name> [sandbox-id|name2...]",
	Short: "Delete one or more Molty instances",
	Long: `Delete one or more Molty instances by sandbox ID or instance name.

By default, prompts for confirmation before deleting. Use -y to skip.

Examples:
  clawdepl delete sandbox_abc123           # Delete with confirmation
  clawdepl delete wifey -y                 # Delete by name without confirmation
  clawdepl delete sandbox_1 wifey          # Delete multiple instances`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	if err := requireLogin(); err != nil {
		return err
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Resolve and verify instances exist.
	var toDelete []string
	var skipped int
	for _, nameOrID := range args {
		sandboxID, err := resolveSandboxID(ctx, client, nameOrID)
		if err != nil {
			fmt.Printf("Warning: %v, skipping\n", err)
			skipped++
			continue
		}
		_, err = client.CheckSandboxStatus(ctx, sandboxID)
		if err != nil {
			fmt.Printf("Warning: instance '%s' not found or inaccessible, skipping\n", nameOrID)
			skipped++
			continue
		}
		toDelete = append(toDelete, sandboxID)
	}

	if len(toDelete) == 0 {
		return fmt.Errorf("no valid instances to delete")
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
		result, err := client.DeleteSandbox(ctx, sandboxID)
		if err != nil {
			fmt.Printf("✗ Failed to delete '%s': %v\n", sandboxID, err)
			failed++
		} else if !result.Success {
			msg := result.Message
			if msg == "" {
				msg = result.Error
			}
			fmt.Printf("✗ Failed to delete '%s': %s\n", sandboxID, msg)
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
