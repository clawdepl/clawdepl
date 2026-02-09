package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/moltyverse/clawdepl/internal/api"
	"github.com/moltyverse/clawdepl/internal/config"
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

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	moltys, err := client.ListMoltys(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(moltys) == 0 {
		fmt.Println("No instances found.")
		fmt.Println("\nCreate your first instance with 'clawdepl new'")
		return nil
	}

	// Sort by name
	sort.Slice(moltys, func(i, j int) bool {
		return moltys[i].Name < moltys[j].Name
	})

	// Create table using tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSANDBOX ID\tCREATED")

	for _, molty := range moltys {
		status := formatStatus(molty.Status)
		created := formatTimestamp(molty.CreatedAt)
		sandboxID := molty.SandboxID
		if sandboxID == "" {
			sandboxID = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", molty.Name, status, sandboxID, created)
	}

	w.Flush()
	return nil
}

func formatStatus(status string) string {
	switch status {
	case "running", "ready":
		return "● running"
	case "stopped":
		return "○ stopped"
	case "provisioning":
		return "◐ provisioning"
	case "error", "failed":
		return "✗ error"
	default:
		return status
	}
}

func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}

	t := time.UnixMilli(ts)
	duration := time.Since(t)

	if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	}
	if duration < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}

	return t.Format("Jan 2, 2006")
}
