package cmd

import (
	"fmt"

	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all Molty instances",
	Long: `List all Molty instances associated with your account.

Displays instance name, status, sandbox ID, and creation time in a table format.

Examples:
  clawdepl list
  clawdepl ls`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Check if logged in (or using unsafe token in debug builds)
	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdepl login' first.")
		return nil
	}

	fmt.Println("List is not yet available in this version.")
	fmt.Println("Use the web dashboard to view your instances.")
	fmt.Println("\nIf you know a sandbox ID, you can check its status with:")
	fmt.Println("  clawdepl status <sandbox-id>")

	return nil
}
