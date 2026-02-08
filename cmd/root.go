package cmd

import (
	"fmt"
	"os"

	"github.com/moltyverse/clawdpl/internal/config"
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

Quick start:
  clawdpl                    Run the setup wizard (login + new)
  clawdpl login              Authenticate with clawdpl.dev
  clawdpl new [name]         Create a new instance
  clawdpl list               List all instances

Instance management:
  clawdpl status <name>      Show instance status
  clawdpl start <name>       Start an instance
  clawdpl stop <name>        Stop an instance
  clawdpl delete <name>      Delete an instance

For more information about a command, run:
  clawdpl <command> --help`,
	Version: fmt.Sprintf("%s (commit: %s)", Version, Commit),
	RunE:    runRoot,
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

	// Customize help template for better subcommand display
	rootCmd.SetHelpTemplate(`{{.Long}}

{{if .HasAvailableSubCommands}}Commands:
{{range .Commands}}{{if .IsAvailableCommand}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}
{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasAvailableInheritedFlags}}Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}Use "{{.CommandPath}} [command] --help" for more information about a command.
`)
}

// runRoot handles the case when clawdpl is called without arguments
func runRoot(cmd *cobra.Command, args []string) error {
	// If no subcommand is provided, run the default flow:
	// 1. If not logged in, login first
	// 2. Then run the new instance wizard

	if !config.IsLoggedIn() {
		fmt.Println("Welcome to clawdpl!")
		fmt.Println()
		fmt.Println("Let's get you set up. First, we need to log you in.")
		fmt.Println()

		if !RunLoginFlow(false) {
			return nil // Login failed or cancelled
		}
	}

	// Run the new instance wizard
	return RunNewFlow()
}
