//go:build !debug

package cmd

import (
	"github.com/moltyverse/clawdpl/internal/buildinfo"
)

// GetEffectiveEndpoint returns the API endpoint to use.
// In production builds, this always returns the default endpoint.
func GetEffectiveEndpoint() string {
	return buildinfo.DefaultEndpoint
}

// HasUnsafeEndpoint always returns false in production builds
func HasUnsafeEndpoint() bool {
	return false
}
