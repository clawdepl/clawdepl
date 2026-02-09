package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/clawdepl/clawdepl/internal/api"
	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionJSON bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build information",
	Long: `Print detailed version and build information for clawdepl.

Displays the version, git commit, build date, build mode (debug/prod),
platform, and Go version.

Examples:
  clawdepl version          # Human-readable output
  clawdepl version --json   # JSON output for scripting`,
	Run: runVersion,
}

func init() {
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Output version info as JSON")
	rootCmd.AddCommand(versionCmd)
}

// VersionInfo holds all version information for JSON output
type VersionInfo struct {
	Name                string `json:"name"`
	Version             string `json:"version"`
	Commit              string `json:"commit"`
	CommitFull          string `json:"commit_full,omitempty"`
	Date                string `json:"date"`
	Mode                string `json:"mode"`
	GOOS                string `json:"goos"`
	GOARCH              string `json:"goarch"`
	GoVersion           string `json:"go_version"`
	AuthEndpoint        string `json:"auth_endpoint,omitempty"`
	ProvisionerEndpoint string `json:"provisioner_endpoint,omitempty"`
	ConvexEndpoint      string `json:"convex_endpoint,omitempty"`
}

func runVersion(cmd *cobra.Command, args []string) {
	info := VersionInfo{
		Name:       buildinfo.Name,
		Version:    buildinfo.Version,
		Commit:     buildinfo.Commit,
		CommitFull: buildinfo.CommitFull,
		Date:       buildinfo.Date,
		Mode:       buildinfo.Mode(),
		GOOS:       buildinfo.GOOS,
		GOARCH:     buildinfo.GOARCH,
		GoVersion:  buildinfo.GoVersion,
	}

	// Include endpoint info in debug mode
	if buildinfo.Debug {
		info.AuthEndpoint = api.GetEffectiveAuthEndpoint()
		info.ProvisionerEndpoint = api.GetEffectiveProvisionerEndpoint()
		info.ConvexEndpoint = api.GetEffectiveConvexEndpoint()
	}

	if versionJSON {
		printVersionJSON(info)
	} else {
		printVersionHuman(info)
	}
}

func printVersionJSON(info VersionInfo) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(info)
}

func printVersionHuman(info VersionInfo) {
	// Version line
	versionStr := info.Version
	if versionStr != "dev" {
		versionStr = "v" + versionStr
	}
	fmt.Printf("%s %s\n", info.Name, versionStr)

	// Commit
	commitStr := info.Commit
	if info.CommitFull != "" && info.CommitFull != "unknown" && info.CommitFull != info.Commit {
		commitStr = fmt.Sprintf("%s (%s)", info.Commit, info.CommitFull)
	}
	fmt.Printf("  commit:   %s\n", commitStr)

	// Build date
	fmt.Printf("  built:    %s\n", info.Date)

	// Build mode
	fmt.Printf("  mode:     %s\n", info.Mode)

	// Platform
	fmt.Printf("  platform: %s/%s\n", info.GOOS, info.GOARCH)

	// Go version
	fmt.Printf("  go:       %s\n", info.GoVersion)

	// Endpoints (debug mode only)
	if info.AuthEndpoint != "" {
		fmt.Printf("\n  endpoints:\n")
		fmt.Printf("    auth:        %s\n", info.AuthEndpoint)
		fmt.Printf("    provisioner: %s\n", info.ProvisionerEndpoint)
		fmt.Printf("    convex:      %s\n", info.ConvexEndpoint)
	}
}
