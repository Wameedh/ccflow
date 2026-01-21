package ccflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	upgradeDryRunFlag bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade workflow templates",
	Long: `Upgrade workflow templates to the latest versions.

This command updates managed files without overwriting user modifications:
- Files unchanged from templates are updated automatically
- Modified files are preserved; new versions written as .new files
- User-created files are never touched

Use --dry-run to see what would change without making modifications.`,
	Run: runUpgrade,
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeDryRunFlag, "dry-run", false, "show what would change without modifying files")
}

func runUpgrade(cmd *cobra.Command, args []string) {
	// Discover workspace
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	bp, err := bpManager.Get(ws.Config.Blueprint)
	if err != nil {
		exitWithError("failed to load blueprint: %v", err)
	}

	// Load or create managed files manifest
	manifest := loadManifest(ws)

	// Create template data
	templateData := &blueprint.TemplateData{
		WorkflowName:  ws.Config.Name,
		DocsStateDir:  ws.Config.State.StateDir,
		DocsDesignDir: ws.Config.State.DesignsDir,
		GatesEnabled:  ws.Config.Gates.Enabled,
		HooksEnabled:  ws.Config.Hooks.Enabled,
	}

	// Track changes
	var updated, skipped, newFiles int

	fmt.Println("Checking for updates...")
	fmt.Println()

	// Check agents
	for _, agentName := range bp.Agents.Defaults {
		result := checkAndUpdate(ws, bpManager, bp, "agents", agentName+".md", templateData, manifest, upgradeDryRunFlag)
		switch result {
		case "updated":
			updated++
		case "skipped":
			skipped++
		case "new":
			newFiles++
		}
	}

	// Check commands
	for _, cmdName := range bp.Commands.Defaults {
		result := checkAndUpdate(ws, bpManager, bp, "commands", cmdName+".md", templateData, manifest, upgradeDryRunFlag)
		switch result {
		case "updated":
			updated++
		case "skipped":
			skipped++
		case "new":
			newFiles++
		}
	}

	// Check hooks
	for _, hookName := range bp.Hooks.Defaults {
		result := checkAndUpdate(ws, bpManager, bp, "hooks", hookName+".sh", templateData, manifest, upgradeDryRunFlag)
		switch result {
		case "updated":
			updated++
		case "skipped":
			skipped++
		case "new":
			newFiles++
		}
	}

	// Save manifest if not dry run
	if !upgradeDryRunFlag {
		saveManifest(ws, manifest)
	}

	// Summary
	fmt.Println()
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Printf("  Updated:     %d\n", updated)
	fmt.Printf("  Skipped:     %d (user-modified)\n", skipped)
	fmt.Printf("  New files:   %d\n", newFiles)

	if upgradeDryRunFlag {
		fmt.Println()
		fmt.Println("This was a dry run. No files were modified.")
	}
}

func checkAndUpdate(ws *workspace.Workspace, bpManager *blueprint.Manager, bp *blueprint.Blueprint, dir, filename string, data *blueprint.TemplateData, manifest *config.ManagedFilesManifest, dryRun bool) string {
	hubPath := ws.GetHubPath()
	filePath := filepath.Join(hubPath, dir, filename)
	relPath := filepath.Join(dir, filename)

	// Get new template content
	var newContent []byte
	var err error

	switch dir {
	case "agents":
		name := strings.TrimSuffix(filename, ".md")
		newContent, err = bpManager.GetAgentContent(bp.ID, name, data)
	case "commands":
		name := strings.TrimSuffix(filename, ".md")
		newContent, err = bpManager.GetCommandContent(bp.ID, name, data)
	case "hooks":
		name := strings.TrimSuffix(filename, ".sh")
		newContent, err = bpManager.GetHookContent(bp.ID, name, data)
	}

	if err != nil {
		fmt.Printf("  ⚠ %s: failed to get template\n", relPath)
		return "error"
	}

	newHash := hashContent(newContent)

	// Check if file exists
	if !util.FileExists(filePath) {
		if dryRun {
			fmt.Printf("  + %s (would create)\n", relPath)
		} else {
			var writeErr error
			if dir == "hooks" {
				writeErr = util.SafeWriteExecutable(filePath, newContent, false)
			} else {
				writeErr = util.SafeWriteFile(filePath, newContent, false)
			}
			if writeErr != nil {
				fmt.Printf("  ✗ %s: %v\n", relPath, writeErr)
				return "error"
			}
			fmt.Printf("  + %s (created)\n", relPath)
			manifest.Files[relPath] = config.ManagedFileInfo{
				TemplateID: bp.ID + "/" + relPath,
				Hash:       newHash,
				Version:    config.Version,
			}
		}
		return "new"
	}

	// File exists - check if managed and unchanged
	existingContent, _ := os.ReadFile(filePath)
	existingHash := hashContent(existingContent)

	info, isManaged := manifest.Files[relPath]

	if isManaged && info.Hash == existingHash {
		// File is managed and unchanged - safe to update
		if info.Hash == newHash {
			// Already up to date
			return "unchanged"
		}

		if dryRun {
			fmt.Printf("  ↑ %s (would update)\n", relPath)
		} else {
			var writeErr error
			if dir == "hooks" {
				writeErr = util.SafeWriteExecutable(filePath, newContent, true)
			} else {
				writeErr = util.SafeWriteFile(filePath, newContent, true)
			}
			if writeErr != nil {
				fmt.Printf("  ✗ %s: %v\n", relPath, writeErr)
				return "error"
			}
			fmt.Printf("  ↑ %s (updated)\n", relPath)
			manifest.Files[relPath] = config.ManagedFileInfo{
				TemplateID: bp.ID + "/" + relPath,
				Hash:       newHash,
				Version:    config.Version,
			}
		}
		return "updated"
	}

	// File is modified or not managed - write .new file
	if dryRun {
		fmt.Printf("  ~ %s (user-modified, would write .new)\n", relPath)
	} else {
		newFilePath := filePath + ".new"
		var writeErr error
		if dir == "hooks" {
			writeErr = util.SafeWriteExecutable(newFilePath, newContent, true)
		} else {
			writeErr = util.SafeWriteFile(newFilePath, newContent, true)
		}
		if writeErr != nil {
			fmt.Printf("  ✗ %s: %v\n", relPath, writeErr)
			return "error"
		}
		fmt.Printf("  ~ %s (user-modified, wrote %s.new)\n", relPath, relPath)
	}
	return "skipped"
}

func hashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func loadManifest(ws *workspace.Workspace) *config.ManagedFilesManifest {
	manifestPath := filepath.Join(ws.GetHubPath(), ".ccflow-managed.json")

	manifest := &config.ManagedFilesManifest{
		Version: 1,
		Files:   make(map[string]config.ManagedFileInfo),
	}

	if util.FileExists(manifestPath) {
		data, err := os.ReadFile(manifestPath)
		if err == nil {
			_ = json.Unmarshal(data, manifest) // Ignore errors, use default if invalid
		}
	}

	return manifest
}

func saveManifest(ws *workspace.Workspace, manifest *config.ManagedFilesManifest) {
	manifestPath := filepath.Join(ws.GetHubPath(), ".ccflow-managed.json")

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(manifestPath, data, 0644) // Best effort
}
