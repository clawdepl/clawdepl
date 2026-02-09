//go:build debug

package cmd

import (
	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/spf13/cobra"
)

// Debug-only unsafe flags
var (
	unsafeProvisionerEndpoint string
	unsafeConvexEndpoint      string
	unsafeAuthEndpoint        string
	unsafeToken               string
	unsafeUser                string
)

func init() {
	// Register debug-only unsafe flags
	// These flags only exist in debug builds and are completely absent in production
	rootCmd.PersistentFlags().StringVar(&unsafeProvisionerEndpoint, "unsafe-provisioner-endpoint", "",
		"[DEBUG] Override the provisioner API endpoint URL")
	rootCmd.PersistentFlags().StringVar(&unsafeConvexEndpoint, "unsafe-convex-endpoint", "",
		"[DEBUG] Override the Convex backend endpoint URL")
	rootCmd.PersistentFlags().StringVar(&unsafeAuthEndpoint, "unsafe-auth-endpoint", "",
		"[DEBUG] Override the authentication service endpoint URL")
	rootCmd.PersistentFlags().StringVar(&unsafeToken, "unsafe-token", "",
		"[DEBUG] Skip login check, use this Bearer token for API calls")
	rootCmd.PersistentFlags().StringVar(&unsafeUser, "unsafe-user", "",
		"[DEBUG] Use this userId in request bodies (default: debug-user)")

	// Hook into PreRun to set the overrides before any command runs
	originalPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if unsafeProvisionerEndpoint != "" {
			api.SetProvisionerEndpointOverride(unsafeProvisionerEndpoint)
		}
		if unsafeConvexEndpoint != "" {
			api.SetConvexEndpointOverride(unsafeConvexEndpoint)
		}
		if unsafeAuthEndpoint != "" {
			api.SetAuthEndpointOverride(unsafeAuthEndpoint)
		}
		if unsafeToken != "" {
			api.SetTokenOverride(unsafeToken)
		}
		if unsafeUser != "" {
			api.SetUserIDOverride(unsafeUser)
		}
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}
}

// HasUnsafeToken returns true if an unsafe token override is set
func HasUnsafeToken() bool {
	return unsafeToken != ""
}

// HasUnsafeProvisionerEndpoint returns true if an unsafe provisioner endpoint override is set
func HasUnsafeProvisionerEndpoint() bool {
	return unsafeProvisionerEndpoint != ""
}

// HasUnsafeConvexEndpoint returns true if an unsafe Convex endpoint override is set
func HasUnsafeConvexEndpoint() bool {
	return unsafeConvexEndpoint != ""
}

// HasUnsafeAuthEndpoint returns true if an unsafe auth endpoint override is set
func HasUnsafeAuthEndpoint() bool {
	return unsafeAuthEndpoint != ""
}

// GetUnsafeUser returns the unsafe user ID if set, otherwise empty string
func GetUnsafeUser() string {
	return unsafeUser
}
