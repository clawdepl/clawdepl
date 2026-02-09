// Package buildinfo contains build-time metadata for the clawdepl CLI.
// Variables are set via -ldflags during build. Safe defaults are provided
// for local development with `go run`.
package buildinfo

import "runtime"

var (
	// Name is the application name
	Name = "clawdepl"

	// Version is the semantic version (set via -ldflags for releases, "dev" otherwise)
	Version = "dev"

	// Commit is the short git commit hash
	Commit = "unknown"

	// CommitFull is the full git commit hash
	CommitFull = "unknown"

	// Date is the build date in ISO-8601 format
	Date = "unknown"

	// APIEndpoint is the Supabase edge functions API URL
	APIEndpoint = "https://ghcvpnixedcjpokvbsep.supabase.co/functions/v1"

	// AuthEndpoint is the authentication service URL (Better Auth)
	AuthEndpoint = "https://clawdepl-auth.vercel.app"
)

// Runtime information (not set via ldflags)
var (
	GoVersion = runtime.Version()
	GOOS      = runtime.GOOS
	GOARCH    = runtime.GOARCH
)

// Mode returns "debug" or "prod" based on the Debug build flag
func Mode() string {
	if Debug {
		return "debug"
	}
	return "prod"
}

// IsRelease returns true if this is a release build (version != "dev")
func IsRelease() bool {
	return Version != "dev"
}

// ShortCommit returns the short commit hash (first 7 chars) or the full Commit if shorter
func ShortCommit() string {
	if len(Commit) > 7 {
		return Commit[:7]
	}
	return Commit
}
