package ccflow

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Wameedh/ccflow/internal/validator"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks",
	Long: `Run comprehensive diagnostic checks on the workflow.

Checks include:
- Workflow marker file exists
- settings.json is valid JSON
- Hook scripts exist and are executable
- Symlinks point to correct targets (multi-repo)
- Required directories exist

Provides remediation suggestions for any issues found.`,
	Run: runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) {
	// Discover workspace
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	// Run doctor checks
	v := validator.New()
	result := v.Doctor(ws)

	// Print results
	fmt.Println("ccflow Doctor")
	fmt.Println("=============")
	fmt.Println()

	for _, check := range result.Checks {
		var icon string
		switch check.Status {
		case "pass":
			icon = "✓"
		case "fail":
			icon = "✗"
		case "warn":
			icon = "⚠"
		}

		fmt.Printf("%s %s\n", icon, check.Name)
		fmt.Printf("  %s\n", check.Message)
		if check.Remediation != "" {
			fmt.Printf("  → %s\n", check.Remediation)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Printf("  Passed:   %d\n", result.Passed)
	fmt.Printf("  Warnings: %d\n", result.Warnings)
	fmt.Printf("  Failed:   %d\n", result.Failed)
	fmt.Println()

	// Exit code
	if result.Failed > 0 {
		fmt.Println("Some checks failed. Please address the issues above.")
		os.Exit(1)
	} else if result.Warnings > 0 {
		fmt.Println("All checks passed with warnings.")
		os.Exit(0)
	} else {
		fmt.Println("All checks passed!")
		os.Exit(0)
	}
}
