package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var accountJSON bool

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Show current account information (alias for 'auth status')",
	Long: `Display information about the currently logged-in user.

This command is an alias for 'clawdepl auth status'.

Shows your user ID, email, name, authentication method, and token expiry.

Examples:
  clawdepl account          # Show account info
  clawdepl account --json   # Output as JSON (for scripting)

See also: clawdepl auth status`,
	RunE: runAccount,
}

func init() {
	accountCmd.Flags().BoolVar(&accountJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(accountCmd)
}

// accountInfo represents the JSON output structure
type accountInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AuthType  string `json:"auth_type"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func runAccount(cmd *cobra.Command, args []string) error {
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil || !creds.IsValid() {
		if accountJSON {
			fmt.Println("{}")
			return nil
		}
		fmt.Println("Not logged in.")
		fmt.Println("\nRun 'clawdepl login' to authenticate.")
		return nil
	}

	// Determine auth type
	authType := "oauth"
	if creds.APIKey != "" {
		authType = "api_key"
	}

	// Build account info
	info := accountInfo{
		AuthType: authType,
	}

	if creds.User != nil {
		info.ID = creds.User.ID
		info.Email = creds.User.Email
		info.Name = creds.User.Name
	}

	if authType == "oauth" && !creds.ExpiresAt.IsZero() {
		info.ExpiresAt = creds.ExpiresAt.Format(time.RFC3339)
	}

	// Output as JSON
	if accountJSON {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Output as human-readable text
	if creds.User != nil {
		fmt.Printf("Logged in as %s (%s)\n", creds.User.Name, creds.User.Email)
		fmt.Println()
		fmt.Printf("  ID:    %s\n", creds.User.ID)
	}

	if authType == "api_key" {
		fmt.Printf("  Auth:  API Key\n")
	} else {
		fmt.Printf("  Auth:  OAuth Token\n")
		if !creds.ExpiresAt.IsZero() {
			remaining := time.Until(creds.ExpiresAt)
			if remaining > 0 {
				fmt.Printf("  Expires: %s (in %s)\n", creds.ExpiresAt.Format(time.RFC3339), formatDuration(remaining))
			} else {
				fmt.Printf("  Expires: %s (EXPIRED)\n", creds.ExpiresAt.Format(time.RFC3339))
			}
		}
	}

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
