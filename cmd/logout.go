package cmd

import (
	"fmt"

	"github.com/moltyverse/clawdpl/internal/auth"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from clawdpl.dev",
	Long: `Log out from clawdpl.dev and clear stored credentials.

This removes the credentials file from ~/.clawdpl/credentials.json.

Example:
  clawdpl logout`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Check if logged in first
	if !config.IsLoggedIn() {
		fmt.Println("Not currently logged in.")
		return nil
	}

	// Get user info before logging out
	creds, _ := config.LoadCredentials()
	userName := "user"
	if creds != nil && creds.User != nil {
		userName = creds.User.Email
	}

	// Clear credentials
	if err := auth.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Printf("✓ Logged out from %s\n", userName)
	return nil
}
