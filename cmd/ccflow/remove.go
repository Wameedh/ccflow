package ccflow

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var (
	removeDryRunFlag   bool
	removeKeepDocsFlag bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [workflow-name]",
	Short: "Remove workflow and all ccflow artifacts",
	Long: `Remove a ccflow workflow and all associated artifacts.

This command removes:
- .claude/ directory (agents, commands, hooks, settings)
- .ccflow/ directory (single-repo) or workflow-hub/ (multi-repo)
- Symlinks to .claude in repositories (multi-repo)
- docs/workflow/ directory (unless --keep-docs)
- Registry entry from ~/.ccflow/registry.json

If run without arguments, removes the workflow in the current directory.
If a workflow name is provided, removes that workflow from anywhere.

Examples:
  ccflow remove                  # Remove workflow in current directory
  ccflow remove my-project       # Remove workflow by name (from anywhere)
  ccflow remove --dry-run        # Preview what would be removed
  ccflow remove --force          # Remove without confirmation
  ccflow remove --keep-docs      # Preserve docs/workflow directory`,
	Args: cobra.MaximumNArgs(1),
	Run:  runRemove,
}

func init() {
	removeCmd.Flags().BoolVar(&removeDryRunFlag, "dry-run", false,
		"show what would be removed without making changes")
	removeCmd.Flags().BoolVar(&removeKeepDocsFlag, "keep-docs", false,
		"preserve the docs/workflow directory")
}

// removalItem represents an item to be removed
type removalItem struct {
	path        string
	description string
	isSymlink   bool
}

func runRemove(cmd *cobra.Command, args []string) {
	var ws *workspace.Workspace
	var err error

	if len(args) > 0 {
		// Workflow name provided - look up in registry
		workflowName := args[0]
		ws, err = findWorkflowByName(workflowName)
		if err != nil {
			exitWithError("workflow '%s' not found in registry\nRun 'ccflow list' to see registered workflows", workflowName)
		}
	} else {
		// No name - discover from current directory
		ws, err = workspace.Discover(workspaceFlag)
		if err != nil {
			exitWithError("%v", err)
		}
	}

	// Build the list of items to remove
	items := buildRemovalList(ws)

	// Print removal plan
	printRemovalPlan(ws, items)

	if removeDryRunFlag {
		fmt.Println("\nThis was a dry run. No files were modified.")
		return
	}

	// Confirm unless --force
	if !forceFlag {
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Remove workflow '%s'?", ws.Config.Name),
			Default: false,
		}
		if surveyErr := survey.AskOne(prompt, &confirm); surveyErr != nil {
			exitWithError("prompt failed: %v", surveyErr)
		}
		if !confirm {
			fmt.Println("Removal cancelled.")
			return
		}
	}

	// Execute removal
	results := executeRemoval(items)

	// Clean up registry
	registryRemoved := removeFromRegistry(ws.Root)

	// Print results
	printRemovalResults(ws, results, registryRemoved)
}

func findWorkflowByName(name string) (*workspace.Workspace, error) {
	reg, err := workspace.LoadRegistry()
	if err != nil {
		return nil, err
	}

	entry := reg.FindByName(name)
	if entry == nil {
		return nil, fmt.Errorf("not found")
	}

	// Discover workspace at the registered path
	return workspace.Discover(entry.Path)
}

func buildRemovalList(ws *workspace.Workspace) []removalItem {
	var items []removalItem

	if ws.Topology == config.TopologyMultiRepo {
		// Multi-repo: remove symlinks first, then hub
		for _, repo := range ws.Config.Repos {
			symlinkPath := filepath.Join(ws.Root, repo.Path, ".claude")
			if util.IsSymlink(symlinkPath) {
				items = append(items, removalItem{
					path:        symlinkPath,
					description: fmt.Sprintf("%s/.claude (symlink)", repo.Path),
					isSymlink:   true,
				})
			}
		}

		// workflow-hub directory
		hubPath := filepath.Join(ws.Root, ws.Config.Paths.Hub)
		if util.DirExists(hubPath) {
			items = append(items, removalItem{
				path:        hubPath,
				description: ws.Config.Paths.Hub + "/",
				isSymlink:   false,
			})
		}
	} else {
		// Single-repo: remove .claude and .ccflow directories
		claudePath := filepath.Join(ws.Root, ".claude")
		if util.DirExists(claudePath) {
			items = append(items, removalItem{
				path:        claudePath,
				description: ".claude/",
				isSymlink:   false,
			})
		}

		ccflowPath := filepath.Join(ws.Root, ".ccflow")
		if util.DirExists(ccflowPath) {
			items = append(items, removalItem{
				path:        ccflowPath,
				description: ".ccflow/",
				isSymlink:   false,
			})
		}
	}

	// docs/workflow directory (unless --keep-docs)
	if !removeKeepDocsFlag {
		workflowDocsPath := filepath.Join(ws.Root, ws.Config.State.Root)
		if util.DirExists(workflowDocsPath) {
			items = append(items, removalItem{
				path:        workflowDocsPath,
				description: ws.Config.State.Root + "/",
				isSymlink:   false,
			})
		}
	}

	return items
}

func printRemovalPlan(ws *workspace.Workspace, items []removalItem) {
	fmt.Println("Removal Plan")
	fmt.Println("============")
	fmt.Printf("Workflow: %s\n", ws.Config.Name)
	fmt.Printf("Topology: %s\n", ws.Topology)
	fmt.Printf("Path: %s\n", ws.Root)
	fmt.Println()

	if len(items) == 0 {
		fmt.Println("No items to remove.")
		return
	}

	fmt.Println("The following will be removed:")
	for _, item := range items {
		fmt.Printf("  -> %s\n", item.description)
	}
}

type removalResult struct {
	item    removalItem
	success bool
	err     error
}

func executeRemoval(items []removalItem) []removalResult {
	var results []removalResult

	for _, item := range items {
		var err error
		if item.isSymlink {
			err = util.RemoveSymlink(item.path)
		} else {
			err = util.RemoveAll(item.path)
		}

		results = append(results, removalResult{
			item:    item,
			success: err == nil,
			err:     err,
		})
	}

	return results
}

func removeFromRegistry(workspacePath string) bool {
	reg, err := workspace.LoadRegistry()
	if err != nil {
		return false
	}

	removed := reg.RemoveWorkflow(workspacePath)
	if removed {
		_ = workspace.SaveRegistry(reg) // Best effort
	}

	return removed
}

func printRemovalResults(ws *workspace.Workspace, results []removalResult, registryRemoved bool) {
	fmt.Println()
	fmt.Println("Removal Results")
	fmt.Println("===============")

	allSuccess := true
	for _, r := range results {
		if r.success {
			printSuccess("Removed %s", r.item.description)
		} else {
			printWarning("Failed to remove %s: %v", r.item.description, r.err)
			allSuccess = false
		}
	}

	if registryRemoved {
		printSuccess("Removed from registry")
	}

	fmt.Println()
	if allSuccess {
		fmt.Printf("Workflow '%s' removed successfully.\n", ws.Config.Name)
	} else {
		fmt.Printf("Workflow '%s' partially removed (some errors occurred).\n", ws.Config.Name)
	}
}
