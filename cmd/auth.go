package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clawdepl/clawdepl/internal/auth"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication (login, logout, status)",
	Long: `Manage authentication with clawdepl.dev.

The auth command provides subcommands for managing your authentication
with the clawdepl.dev service.

Available Commands:
  login       Authenticate with clawdepl.dev
  logout      Log out from clawdepl.dev
  status      Show current account information

Examples:
  clawdepl auth login              # OAuth via browser (default)
  clawdepl auth login --no-browser # OAuth without browser (manual token entry)
  clawdepl auth login --token XXX  # Token-based authentication
  clawdepl auth login --api-key    # API key authentication
  clawdepl auth logout             # Sign out and clear credentials
  clawdepl auth status             # Show current account info
  clawdepl auth status --json      # Output account info as JSON`,
}

// Subcommand flags (separate from top-level command flags to avoid conflicts)
var (
	authLoginInfo      bool
	authLoginNoBrowser bool
	authLoginAPIKey    bool
	authLoginToken     string
	authStatusJSON     bool
)

// auth login subcommand
var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with clawdepl.dev",
	Long: `Authenticate with clawdepl.dev to manage your OpenClaw instances.

By default, opens a browser for OAuth authentication. Use flags to change
the authentication method.

Examples:
  clawdepl auth login                    # OAuth via browser (default)
  clawdepl auth login --no-browser       # OAuth without browser (manual token entry)
  clawdepl auth login --token YOUR_TOKEN # Authenticate with a pre-obtained token
  clawdepl auth login --api-key          # Authenticate with an API key
  clawdepl auth login --info             # Show current login status`,
	RunE: runAuthLogin,
}

// auth logout subcommand
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from clawdepl.dev",
	Long: `Log out from clawdepl.dev and clear stored credentials.

This removes the credentials file from ~/.clawdepl/credentials.json.

Example:
  clawdepl auth logout`,
	RunE: runAuthLogout,
}

// auth status subcommand
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current account information",
	Long: `Display information about the currently logged-in user.

Shows your user ID, email, name, authentication method, and token expiry.

Examples:
  clawdepl auth status          # Show account info
  clawdepl auth status --json   # Output as JSON (for scripting)`,
	RunE: runAuthStatus,
}

func init() {
	// Register auth login flags
	authLoginCmd.Flags().BoolVar(&authLoginInfo, "info", false, "Show current user info")
	authLoginCmd.Flags().BoolVar(&authLoginNoBrowser, "no-browser", false, "Don't open browser, print URL instead")
	authLoginCmd.Flags().StringVar(&authLoginToken, "token", "", "Authenticate with a pre-obtained token")
	authLoginCmd.Flags().BoolVar(&authLoginAPIKey, "api-key", false, "Authenticate with an API key")

	// Register auth status flags
	authStatusCmd.Flags().BoolVar(&authStatusJSON, "json", false, "Output as JSON")

	// Add subcommands to auth
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	// Add auth to root
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	// Handle --info flag
	if authLoginInfo {
		return showAuthLoginInfo()
	}

	// Handle --token flag
	if authLoginToken != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		creds, err := auth.LoginWithToken(ctx, authLoginToken)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		fmt.Printf("\n✓ Authenticated successfully\n")
		if creds.User != nil && creds.User.Email != "unknown" {
			fmt.Printf("  Logged in as: %s\n", creds.User.Email)
		}
		return nil
	}

	// Check if already logged in (only for interactive flows)
	if config.IsLoggedIn() {
		creds, _ := config.LoadCredentials()
		if creds != nil && creds.User != nil {
			fmt.Printf("Already logged in as %s (%s)\n", creds.User.Name, creds.User.Email)
			fmt.Printf("Use 'clawdepl auth logout' to sign out first.\n")
			return nil
		}
	}

	// Perform login
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	creds, err := auth.Login(ctx, authLoginNoBrowser, authLoginAPIKey)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Printf("\n✓ Logged in as %s (%s)\n", creds.User.Name, creds.User.Email)
	return nil
}

func showAuthLoginInfo() error {
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil || !creds.IsValid() {
		fmt.Println("Not logged in.")
		fmt.Println("\nRun 'clawdepl auth login' to authenticate.")
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

func runAuthLogout(cmd *cobra.Command, args []string) error {
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

// authStatusInfo represents the JSON output structure for auth status
type authStatusInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AuthType  string `json:"auth_type"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil || !creds.IsValid() {
		if authStatusJSON {
			fmt.Println("{}")
			return nil
		}
		fmt.Println("Not logged in.")
		fmt.Println("\nRun 'clawdepl auth login' to authenticate.")
		return nil
	}

	// Determine auth type
	authType := "oauth"
	if creds.APIKey != "" {
		authType = "api_key"
	}

	// Build account info
	info := authStatusInfo{
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
	if authStatusJSON {
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
				fmt.Printf("  Expires: %s (in %s)\n", creds.ExpiresAt.Format(time.RFC3339), formatAuthDuration(remaining))
			} else {
				fmt.Printf("  Expires: %s (EXPIRED)\n", creds.ExpiresAt.Format(time.RFC3339))
			}
		}
	}

	return nil
}

// formatAuthDuration formats a duration in a human-readable way
func formatAuthDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
