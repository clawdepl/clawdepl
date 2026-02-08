package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/moltyverse/clawdpl/internal/auth"
	"github.com/moltyverse/clawdpl/internal/config"
	"github.com/spf13/cobra"
)

var (
	loginInfo      bool
	loginNoBrowser bool
	loginAPIKey    bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with clawdpl.dev",
	Long: `Authenticate with clawdpl.dev to manage your OpenClaw instances.

By default, opens a browser for OAuth authentication. Use flags to change
the authentication method.

Examples:
  clawdpl login              # OAuth via browser (default)
  clawdpl login --no-browser # OAuth without browser (manual token entry)
  clawdpl login --api-key    # Authenticate with an API key
  clawdpl login --info       # Show current login status`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().BoolVar(&loginInfo, "info", false, "Show current user info")
	loginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Don't open browser, print URL instead")
	loginCmd.Flags().BoolVar(&loginAPIKey, "api-key", false, "Authenticate with an API key")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Handle --info flag
	if loginInfo {
		return showLoginInfo()
	}

	// Check if already logged in
	if config.IsLoggedIn() {
		creds, _ := config.LoadCredentials()
		if creds != nil && creds.User != nil {
			fmt.Printf("Already logged in as %s (%s)\n", creds.User.Name, creds.User.Email)
			fmt.Printf("Use 'clawdpl logout' to sign out first.\n")
			return nil
		}
	}

	// Perform login
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	creds, err := auth.Login(ctx, loginNoBrowser, loginAPIKey)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Printf("\n✓ Logged in as %s (%s)\n", creds.User.Name, creds.User.Email)
	return nil
}

func showLoginInfo() error {
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil || !creds.IsValid() {
		fmt.Println("Not logged in.")
		fmt.Println("\nRun 'clawdpl login' to authenticate.")
		return nil
	}

	fmt.Println("Current login status:")
	fmt.Println()

	if creds.User != nil {
		fmt.Printf("  Name:  %s\n", creds.User.Name)
		fmt.Printf("  Email: %s\n", creds.User.Email)
		fmt.Printf("  ID:    %s\n", creds.User.ID)
	}

	if creds.APIKey != "" {
		fmt.Printf("  Auth:  API Key\n")
	} else {
		fmt.Printf("  Auth:  OAuth Token\n")
		if !creds.ExpiresAt.IsZero() {
			remaining := time.Until(creds.ExpiresAt)
			if remaining > 0 {
				fmt.Printf("  Expires: %s (in %s)\n", creds.ExpiresAt.Format(time.RFC3339), remaining.Round(time.Minute))
			} else {
				fmt.Printf("  Expires: %s (EXPIRED)\n", creds.ExpiresAt.Format(time.RFC3339))
			}
		}
	}

	return nil
}

// RunLoginFlow runs the login flow and returns true if successful
// This is exported for use by other commands (e.g., root with no args)
func RunLoginFlow(noBrowser bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	creds, err := auth.Login(ctx, noBrowser, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		return false
	}

	fmt.Printf("\n✓ Logged in as %s (%s)\n\n", creds.User.Name, creds.User.Email)
	return true
}
