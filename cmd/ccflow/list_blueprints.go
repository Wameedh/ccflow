package ccflow

import (
	"fmt"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/spf13/cobra"
)

var listBlueprintsCmd = &cobra.Command{
	Use:   "list-blueprints",
	Short: "List available blueprints",
	Long:  `List all available workflow blueprints with their descriptions.`,
	Run:   listBlueprints,
}

func listBlueprints(cmd *cobra.Command, args []string) {
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	blueprints := bpManager.List()

	fmt.Println("Available blueprints:")
	fmt.Println()

	for _, bp := range blueprints {
		fmt.Printf("  %s\n", bp.ID)
		fmt.Printf("    %s\n", bp.DisplayName)
		fmt.Printf("    %s\n", bp.Description)
		fmt.Printf("    Default topology: %s\n", bp.DefaultTopology)
		fmt.Println()
	}

	fmt.Println("Usage: ccflow run <blueprint>")
}
