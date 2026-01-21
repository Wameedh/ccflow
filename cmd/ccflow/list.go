package ccflow

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/workspace"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered workflows",
	Long: `List all workflows registered in the global registry.

The registry is stored at ~/.ccflow/registry.json and is updated
automatically when workflows are created.

Note: This is a convenience feature. All commands work via marker
discovery even without the registry.`,
	Run: listWorkflows,
}

func listWorkflows(cmd *cobra.Command, args []string) {
	reg, err := workspace.LoadRegistry()
	if err != nil {
		exitWithError("failed to load registry: %v", err)
	}

	if len(reg.Workflows) == 0 {
		fmt.Println("No workflows registered.")
		fmt.Println()
		fmt.Println("Run 'ccflow run' to create a new workflow.")
		return
	}

	fmt.Println("Registered Workflows")
	fmt.Println("====================")
	fmt.Println()

	for _, entry := range reg.Workflows {
		fmt.Printf("  %s\n", entry.Name)
		fmt.Printf("    Path:      %s\n", entry.Path)
		fmt.Printf("    Blueprint: %s\n", entry.Blueprint)
		fmt.Printf("    Created:   %s\n", formatTime(entry.CreatedAt))
		fmt.Printf("    Last used: %s\n", formatTime(entry.LastUsedAt))
		fmt.Println()
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05")
}
