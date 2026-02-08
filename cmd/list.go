package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all OpenClaw instances",
	Long: `List all OpenClaw instances associated with your account.

Displays instance name, status, region, and creation time in a table format.

Examples:
  clawdpl list
  clawdpl ls`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Check if logged in
	if !config.IsLoggedIn() {
		fmt.Println("Not logged in. Run 'clawdpl login' first.")
		return nil
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println("No instances found.")
		fmt.Println("\nCreate your first instance with 'clawdpl new'")
		return nil
	}

	// Sort by name
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})

	// Create table using tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tREGION\tCREATED")

	for _, inst := range instances {
		status := formatStatus(inst.Status)
		created := formatTime(inst.CreatedAt)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", inst.Name, status, inst.Region, created)
	}

	w.Flush()
	return nil
}

func formatStatus(status string) string {
	switch status {
	case "running":
		return "● running"
	case "stopped":
		return "○ stopped"
	case "provisioning":
		return "◐ provisioning"
	case "error":
		return "✗ error"
	default:
		return status
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

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
