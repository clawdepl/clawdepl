package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <name>",
	Short: "Show status of an OpenClaw instance",
	Long: `Show detailed status information for an OpenClaw instance.

Displays name, status, uptime, region, and creation time.

Examples:
  clawdpl status my-agent`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	instance, err := client.GetInstance(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Display status
	fmt.Printf("Instance: %s\n", instance.Name)
	fmt.Printf("  ID:      %s\n", instance.ID)
	fmt.Printf("  Status:  %s\n", formatStatusDetailed(instance.Status))
	fmt.Printf("  Region:  %s\n", instance.Region)

	if instance.Purpose != "" {
		fmt.Printf("  Purpose: %s\n", instance.Purpose)
	}

	fmt.Printf("  Created: %s\n", formatTimeDetailed(instance.CreatedAt))

	if instance.Status == "running" && !instance.StartedAt.IsZero() {
		uptime := time.Since(instance.StartedAt)
		fmt.Printf("  Uptime:  %s\n", formatDuration(uptime))
	}

	fmt.Printf("  Updated: %s\n", formatTimeDetailed(instance.UpdatedAt))

	return nil
}

func formatStatusDetailed(status string) string {
	switch status {
	case "running":
		return "● Running"
	case "stopped":
		return "○ Stopped"
	case "provisioning":
		return "◐ Provisioning"
	case "error":
		return "✗ Error"
	default:
		return status
	}
}

func formatTimeDetailed(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%s (%s)", t.Format("Jan 2, 2006 15:04:05"), formatTimeAgo(t))
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%d days", days)
	}
	return fmt.Sprintf("%dd %dh", days, hours)
}
