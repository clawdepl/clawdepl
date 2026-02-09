//go:build debug

package cmd

import (
	"github.com/moltyverse/clawdpl/internal/api"
	"github.com/moltyverse/clawdpl/internal/buildinfo"
	"github.com/spf13/cobra"
)

// unsafeEndpoint is the debug-only endpoint override flag
var unsafeEndpoint string

func init() {
	// Register debug-only unsafe flags
	// These flags only exist in debug builds and are completely absent in production
	rootCmd.PersistentFlags().StringVar(&unsafeEndpoint, "unsafe-endpoint", "",
		"[DEBUG ONLY] Override the API endpoint URL")

	// Hook into PreRun to set the endpoint override before any command runs
	originalPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if unsafeEndpoint != "" {
			api.SetEndpointOverride(unsafeEndpoint)
		}
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}
}

// GetEffectiveEndpoint returns the API endpoint to use.
// In debug builds, this respects the --unsafe-endpoint flag if set.
func GetEffectiveEndpoint() string {
	if unsafeEndpoint != "" {
		return unsafeEndpoint
	}
	return buildinfo.DefaultEndpoint
}

// HasUnsafeEndpoint returns true if an unsafe endpoint override is set
func HasUnsafeEndpoint() bool {
	return unsafeEndpoint != ""
}
