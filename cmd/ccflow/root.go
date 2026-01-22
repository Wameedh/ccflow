package ccflow

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/config"
)

var (
	workspaceFlag string
	forceFlag     bool
)

var rootCmd = &cobra.Command{
	Use:   "ccflow",
	Short: "Claude Code Flow Wizard - manage Claude Code workflows",
	Long: `ccflow is a CLI tool for creating and managing Claude Code workflows.

It provides an interactive wizard to scaffold workflows with agents, commands,
hooks, and settings. ccflow supports both single-repo and multi-repo topologies.

Quick start:
  ccflow run web-dev    Create a new web-dev workflow
  ccflow status         Check workflow health
  ccflow doctor         Run detailed diagnostics

Management:
  ccflow add-agent      Add a new agent
  ccflow add-command    Add a new command
  ccflow add-hook       Add a new hook

For more information, visit: https://github.com/Wameedh/ccflow`,
	Version: config.Version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&workspaceFlag, "workspace", "w", "", "workspace path (overrides auto-detection)")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "force overwrite existing files")

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listBlueprintsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(addAgentCmd)
	rootCmd.AddCommand(addCommandCmd)
	rootCmd.AddCommand(addHookCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(expandCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(permissionsCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ccflow version %s\n", config.Version)
		fmt.Printf("Build time: %s\n", config.BuildTime)
	},
}

// exitWithError prints an error and exits
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// printSuccess prints a success message
func printSuccess(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// printWarning prints a warning message
func printWarning(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}

// printInfo prints an info message
func printInfo(format string, args ...interface{}) {
	fmt.Printf("→ "+format+"\n", args...)
}
