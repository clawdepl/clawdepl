package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/clawdepl/clawdepl/internal/api"
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
	if err := requireLogin(); err != nil {
		return err
	}

	client, err := api.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bots, err := client.ListBots(ctx)
	if err != nil {
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(bots) == 0 {
		fmt.Println("No instances found.")
		fmt.Println("\nCreate your first instance with 'clawdepl new'")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATE\tSANDBOX ID\tCREATED")

	for _, bot := range bots {
		state := formatState(bot.State)
		created := formatCreatedAt(bot.CreatedAt)
		sandboxID := bot.SandboxID
		if sandboxID == "" {
			sandboxID = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", bot.Name, state, sandboxID, created)
	}

	w.Flush()
	return nil
}

func formatState(state string) string {
	switch state {
	case "running", "ready", "active":
		return "● running"
	case "stopped":
		return "○ stopped"
	case "provisioning":
		return "◐ provisioning"
	case "error", "failed":
		return "✗ error"
	default:
		return state
	}
}

func formatCreatedAt(ts string) string {
	if ts == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
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
