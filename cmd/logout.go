package cmd

import (
	"fmt"

	"github.com/clawdepl/clawdepl/internal/auth"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from clawdepl.dev (alias for 'auth logout')",
	Long: `Log out from clawdepl.dev and clear stored credentials.

This command is an alias for 'clawdepl auth logout'.

This removes the credentials file from ~/.clawdepl/credentials.json.

Example:
  clawdepl logout

See also: clawdepl auth logout`,
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil {
		fmt.Println("Not currently logged in.")
		return nil
	}

	userName := "user"
	if creds.User != nil {
		userName = creds.User.Email
	}

	// Clear credentials
	if err := auth.Logout(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Printf("✓ Logged out from %s\n", userName)
	return nil
}
