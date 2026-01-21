package ccflow

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Wameedh/ccflow/internal/validator"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workflow status",
	Long: `Display the current status of the workflow, including:
- Workflow name and blueprint
- Topology (multi-repo or single-repo)
- Hub and docs paths
- Repository installation status
- Hook configuration

Exit codes:
  0 - Healthy
  2 - Warnings (non-fatal issues)
  3 - Errors (broken workflow)`,
	Run: showStatus,
}

func showStatus(cmd *cobra.Command, args []string) {
	// Discover workspace
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	// Run status check
	v := validator.New()
	result := v.Status(ws)

	// Print results
	fmt.Println("Workflow Status")
	fmt.Println("===============")
	fmt.Println()

	fmt.Printf("Name:      %s\n", result.WorkflowName)
	fmt.Printf("Blueprint: %s\n", result.Blueprint)
	fmt.Printf("Topology:  %s\n", result.Topology)
	fmt.Println()

	fmt.Println("Paths:")
	fmt.Printf("  Hub:  %s\n", result.HubPath)
	fmt.Printf("  Docs: %s\n", result.DocsPath)
	fmt.Println()

	// Repos
	if len(result.Repos) > 0 {
		fmt.Println("Repositories:")
		for _, repo := range result.Repos {
			statusIcon := "✓"
			if repo.Status == "broken" || repo.Status == "missing" {
				statusIcon = "✗"
			}
			fmt.Printf("  %s %s (%s) - %s\n", statusIcon, repo.Name, repo.Kind, repo.Message)
		}
		fmt.Println()
	}

	// Hooks
	fmt.Printf("Hooks: %s\n", enabledString(result.HooksEnabled))
	if len(result.Hooks) > 0 {
		for _, hook := range result.Hooks {
			status := "✓"
			if !hook.Exists {
				status = "✗ missing"
			} else if !hook.Executable {
				status = "⚠ not executable"
			}
			fmt.Printf("  %s %s (events: %v)\n", status, hook.Name, hook.Events)
		}
	}
	fmt.Println()

	fmt.Printf("Gates: %s\n", enabledString(result.GatesEnabled))
	fmt.Println()

	// Warnings and errors
	exitCode := 0
	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
		exitCode = 2
	}

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s\n", e)
		}
		exitCode = 3
	}

	if exitCode == 0 {
		fmt.Println("Status: Healthy ✓")
	}

	os.Exit(exitCode)
}

func enabledString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
