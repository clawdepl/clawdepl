package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var (
	deleteYes bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name> [name2...]",
	Short: "Delete one or more OpenClaw instances",
	Long: `Delete one or more OpenClaw instances by name.

By default, prompts for confirmation before deleting. Use -y to skip.

Examples:
  clawdpl delete my-agent           # Delete with confirmation
  clawdpl delete my-agent -y        # Delete without confirmation
  clawdpl delete bot1 bot2 bot3     # Delete multiple instances`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdpl login' first.")
		return nil
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Verify instances exist first
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var toDelete []string
	for _, name := range args {
		_, err := client.GetInstance(ctx, name)
		if err != nil {
			fmt.Printf("Warning: instance '%s' not found, skipping\n", name)
			continue
		}
		toDelete = append(toDelete, name)
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
	for _, name := range toDelete {
		if err := client.DeleteInstance(ctx, name); err != nil {
			fmt.Printf("✗ Failed to delete '%s': %v\n", name, err)
			failed++
		} else {
			fmt.Printf("✓ Deleted '%s'\n", name)
			deleted++
		}
	}

	if failed > 0 {
		return fmt.Errorf("failed to delete %d instance(s)", failed)
	}

	return nil
}
