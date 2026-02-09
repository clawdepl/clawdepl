package cmd

import (
	"fmt"
	"os"

	"github.com/clawdepl/clawdepl/internal/buildinfo"
	"github.com/clawdepl/clawdepl/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clawdepl",
	Short: "Create and manage OpenClaw instances",
	Long: `clawdepl is a CLI tool for creating and managing OpenClaw 
AI Agent orchestrator instances on hosted infrastructure.

OpenClaw enables you to deploy, configure, and scale AI agent workflows
with a single command.

Quick start:
  clawdepl                    Run the setup wizard (login + new)
  clawdepl login              Authenticate with clawdepl.dev
  clawdepl new [name]         Create a new instance
  clawdepl list               List all instances

Authentication:
  clawdepl auth login         Authenticate with clawdepl.dev
  clawdepl auth logout        Sign out and clear credentials
  clawdepl auth status        Show current account info

  Shortcuts:
    clawdepl login            Alias for 'auth login'
    clawdepl logout           Alias for 'auth logout'
    clawdepl account          Alias for 'auth status'

Instance management:
  clawdepl status <name>      Show instance status
  clawdepl start <name>       Start an instance
  clawdepl stop <name>        Stop an instance
  clawdepl delete <name>      Delete an instance

For more information about a command, run:
  clawdepl <command> --help`,
	Version: fmt.Sprintf("%s (commit: %s)", buildinfo.Version, buildinfo.Commit),
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

// runRoot handles the case when clawdepl is called without arguments
func runRoot(cmd *cobra.Command, args []string) error {
	// If no subcommand is provided, run the default flow:
	// 1. If not logged in (and no unsafe token), login first
	// 2. Then run the new instance wizard

	if !HasUnsafeToken() && !config.IsLoggedIn() {
		fmt.Println("Welcome to clawdepl!")
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
