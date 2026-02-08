package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "0.1.0"
	// Commit is set at build time
	Commit = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "clawdpl",
	Short: "Create and manage OpenClaw instances",
	Long: `clawdpl is a CLI tool for creating and managing OpenClaw 
AI Agent orchestrator instances on hosted infrastructure.

OpenClaw enables you to deploy, configure, and scale AI agent workflows
with a single command.

Get started:
  clawdpl init my-project    Create a new OpenClaw project
  clawdpl deploy             Deploy your project to the cloud
  clawdpl status             Check the status of your instances`,
	Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(`{{.Name}} version {{.Version}}
`)
}
