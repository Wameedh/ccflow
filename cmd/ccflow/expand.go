package ccflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var expandCmd = &cobra.Command{
	Use:   "expand",
	Short: "Expand single-repo to multi-repo",
	Long: `Expand a single-repo workflow to a multi-repo topology.

This command:
1. Creates a workflow-hub directory
2. Moves the existing .claude directory to the hub
3. Creates a symlink from the original location to the hub
4. Updates workflow.yaml with the new topology

This is useful when a project grows and needs to be split into
multiple repositories sharing a common workflow configuration.`,
	Run: runExpand,
}

func runExpand(cmd *cobra.Command, args []string) {
	// Discover workspace
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	// Check if already multi-repo
	if ws.Topology == config.TopologyMultiRepo {
		exitWithError("workflow is already multi-repo topology")
	}

	// Confirm with user
	fmt.Println("Expand to Multi-Repo Topology")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  1. Create a workflow-hub directory")
	fmt.Println("  2. Move .claude to workflow-hub/.claude")
	fmt.Println("  3. Create symlink: .claude -> workflow-hub/.claude")
	fmt.Println("  4. Update workflow.yaml")
	fmt.Println()

	var confirm bool
	survey.AskOne(&survey.Confirm{
		Message: "Proceed with expansion?",
		Default: false,
	}, &confirm)

	if !confirm {
		fmt.Println("Expansion canceled.")
		return
	}

	// Perform expansion
	if err := expandWorkflow(ws); err != nil {
		exitWithError("expansion failed: %v", err)
	}

	printSuccess("Workflow expanded to multi-repo topology")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Move other repositories into the workspace")
	fmt.Println("  2. Create symlinks: cd <repo> && ln -s ../workflow-hub/.claude .claude")
	fmt.Println("  3. Update workflow.yaml with the new repos")
	fmt.Println("  4. Run 'ccflow doctor' to verify the setup")
}

func expandWorkflow(ws *workspace.Workspace) error {
	hubName := "workflow-hub"
	hubPath := filepath.Join(ws.Root, hubName)
	oldClaudePath := filepath.Join(ws.Root, ".claude")
	newClaudePath := filepath.Join(hubPath, ".claude")

	// Create workflow-hub directory
	if err := util.EnsureDir(hubPath); err != nil {
		return fmt.Errorf("failed to create hub directory: %w", err)
	}

	// Move .claude to hub
	if err := moveDirectory(oldClaudePath, newClaudePath); err != nil {
		return fmt.Errorf("failed to move .claude: %w", err)
	}

	// Create symlink
	if err := util.CreateRelativeSymlink(newClaudePath, oldClaudePath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Update configuration
	ws.Config.Topology = config.TopologyMultiRepo
	ws.Config.Paths.Hub = hubName
	ws.Config.Paths.Docs = "docs"

	// Add current repo to repos list if empty
	if len(ws.Config.Repos) == 0 {
		repoName := filepath.Base(ws.Root)
		ws.Config.Repos = []config.RepoConfig{
			{
				Name: repoName,
				Path: ".",
				Kind: util.DetectRepoKind(ws.Root),
			},
		}
	}

	// Write new config to hub
	newConfigPath := filepath.Join(hubPath, "workflow.yaml")
	if err := workspace.SaveConfig(newConfigPath, ws.Config); err != nil {
		return fmt.Errorf("failed to write new config: %w", err)
	}

	// Remove old .ccflow directory
	oldCcflowPath := filepath.Join(ws.Root, ".ccflow")
	if util.DirExists(oldCcflowPath) {
		if err := util.RemoveAll(oldCcflowPath); err != nil {
			return fmt.Errorf("failed to remove old .ccflow directory: %w", err)
		}
	}

	return nil
}

func moveDirectory(src, dst string) error {
	// Ensure destination parent exists
	if err := util.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	// Try os.Rename first (fast, atomic on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fall back to copy + delete
	if err := copyDir(src, dst); err != nil {
		return err
	}

	return util.RemoveAll(src)
}

func copyDir(src, dst string) error {
	// Walk source directory and copy files
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return util.EnsureDir(dstPath)
		}

		return util.CopyFile(path, dstPath)
	})
}
