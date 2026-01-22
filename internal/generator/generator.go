package generator

import (
	"fmt"
	"path/filepath"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
)

// Generator handles creating workflow file structures
type Generator struct {
	bpManager *blueprint.Manager
}

// New creates a new Generator
func New(bpManager *blueprint.Manager) *Generator {
	return &Generator{bpManager: bpManager}
}

// GenerateOptions contains options for workflow generation
type GenerateOptions struct {
	WorkspacePath string
	WorkflowName  string
	Blueprint     string
	Topology      config.Topology
	Repos         []config.RepoConfig
	HooksEnabled  bool
	GatesEnabled  bool
	VCS           config.VCSProvider
	Tracker       config.TrackerProvider
	Force         bool
}

// Generate creates a new workflow structure
func (g *Generator) Generate(opts GenerateOptions) (*config.WorkflowConfig, error) {
	bp, err := g.bpManager.Get(opts.Blueprint)
	if err != nil {
		return nil, err
	}

	// Create the workflow configuration
	cfg := config.NewDefaultWorkflowConfig(opts.WorkflowName)
	cfg.Blueprint = opts.Blueprint
	cfg.Topology = opts.Topology
	cfg.Repos = opts.Repos
	cfg.Hooks.Enabled = opts.HooksEnabled
	cfg.Gates.Enabled = opts.GatesEnabled
	cfg.MCP.VCS = opts.VCS
	cfg.MCP.Tracker = opts.Tracker

	// Adjust paths based on topology
	if opts.Topology == config.TopologySingleRepo {
		cfg.Paths.Hub = ""
		cfg.Paths.Docs = "docs"
		cfg.State.Root = "docs/workflow"
		cfg.State.StateDir = "docs/workflow/state"
		cfg.State.DesignsDir = "docs/workflow/designs"
	}

	// Create template data
	templateData := &blueprint.TemplateData{
		WorkflowName:    opts.WorkflowName,
		DocsRoot:        cfg.State.Root,
		DocsStateDir:    cfg.State.StateDir,
		DocsDesignDir:   cfg.State.DesignsDir,
		TrackerProvider: string(opts.Tracker),
		VCSProvider:     string(opts.VCS),
		HooksEnabled:    opts.HooksEnabled,
		GatesEnabled:    opts.GatesEnabled,
	}

	// Generate the structure based on topology
	if opts.Topology == config.TopologyMultiRepo {
		hubPath := filepath.Join(opts.WorkspacePath, cfg.Paths.Hub)
		if err := g.generateMultiRepoStructure(opts.WorkspacePath, hubPath, cfg, bp, templateData, opts.Force); err != nil {
			return nil, err
		}
	} else {
		if err := g.generateSingleRepoStructure(opts.WorkspacePath, cfg, bp, templateData, opts.Force); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// generateMultiRepoStructure creates the multi-repo workflow structure
func (g *Generator) generateMultiRepoStructure(workspacePath, hubPath string, cfg *config.WorkflowConfig, bp *blueprint.Blueprint, data *blueprint.TemplateData, force bool) error {
	// Create workflow-hub directory
	if err := util.EnsureDir(hubPath); err != nil {
		return fmt.Errorf("failed to create hub directory: %w", err)
	}

	// Create .claude directory structure in hub
	claudePath := filepath.Join(hubPath, ".claude")
	if err := g.generateClaudeDirectory(claudePath, bp, data, force); err != nil {
		return err
	}

	// Create docs/workflow directories
	docsPath := filepath.Join(workspacePath, cfg.Paths.Docs)
	if err := g.generateDocsStructure(docsPath, cfg); err != nil {
		return err
	}

	// Write workflow.yaml marker
	markerPath := filepath.Join(hubPath, "workflow.yaml")
	if err := g.writeWorkflowMarker(markerPath, cfg, force); err != nil {
		return err
	}

	return nil
}

// generateSingleRepoStructure creates the single-repo workflow structure
func (g *Generator) generateSingleRepoStructure(repoPath string, cfg *config.WorkflowConfig, bp *blueprint.Blueprint, data *blueprint.TemplateData, force bool) error {
	// Create .claude directory structure
	claudePath := filepath.Join(repoPath, ".claude")
	if err := g.generateClaudeDirectory(claudePath, bp, data, force); err != nil {
		return err
	}

	// Create .ccflow directory for marker
	ccflowPath := filepath.Join(repoPath, ".ccflow")
	if err := util.EnsureDir(ccflowPath); err != nil {
		return fmt.Errorf("failed to create .ccflow directory: %w", err)
	}

	// Create docs/workflow directories
	if err := g.generateDocsStructure(repoPath, cfg); err != nil {
		return err
	}

	// Write workflow.yaml marker
	markerPath := filepath.Join(ccflowPath, "workflow.yaml")
	if err := g.writeWorkflowMarker(markerPath, cfg, force); err != nil {
		return err
	}

	return nil
}

// generateClaudeDirectory creates the .claude directory with all assets
func (g *Generator) generateClaudeDirectory(claudePath string, bp *blueprint.Blueprint, data *blueprint.TemplateData, force bool) error {
	// Create subdirectories
	dirs := []string{"agents", "commands", "hooks"}
	for _, dir := range dirs {
		if err := util.EnsureDir(filepath.Join(claudePath, dir)); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Write agents
	for _, agentName := range bp.Agents.Defaults {
		content, err := g.bpManager.GetAgentContent(bp.ID, agentName, data)
		if err != nil {
			return fmt.Errorf("failed to get agent %s: %w", agentName, err)
		}
		agentPath := filepath.Join(claudePath, "agents", agentName+".md")
		if err := util.SafeWriteFile(agentPath, content, force); err != nil {
			return fmt.Errorf("failed to write agent %s: %w", agentName, err)
		}
	}

	// Write commands
	for _, cmdName := range bp.Commands.Defaults {
		content, err := g.bpManager.GetCommandContent(bp.ID, cmdName, data)
		if err != nil {
			return fmt.Errorf("failed to get command %s: %w", cmdName, err)
		}
		cmdPath := filepath.Join(claudePath, "commands", cmdName+".md")
		if err := util.SafeWriteFile(cmdPath, content, force); err != nil {
			return fmt.Errorf("failed to write command %s: %w", cmdName, err)
		}
	}

	// Write hooks
	for _, hookName := range bp.Hooks.Defaults {
		content, err := g.bpManager.GetHookContent(bp.ID, hookName, data)
		if err != nil {
			return fmt.Errorf("failed to get hook %s: %w", hookName, err)
		}
		hookPath := filepath.Join(claudePath, "hooks", hookName+".sh")
		if err := util.SafeWriteExecutable(hookPath, content, force); err != nil {
			return fmt.Errorf("failed to write hook %s: %w", hookName, err)
		}
	}

	// Write settings.json
	settingsContent, err := g.bpManager.GetAsset(bp.ID, ".claude/settings.json")
	if err != nil {
		return fmt.Errorf("failed to get settings.json: %w", err)
	}
	settingsPath := filepath.Join(claudePath, "settings.json")
	if err := util.SafeWriteFile(settingsPath, settingsContent, force); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}

// generateDocsStructure creates the docs/workflow directories
func (g *Generator) generateDocsStructure(basePath string, cfg *config.WorkflowConfig) error {
	statePath := filepath.Join(basePath, cfg.State.StateDir)
	designsPath := filepath.Join(basePath, cfg.State.DesignsDir)

	if err := util.EnsureDir(statePath); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	if err := util.EnsureDir(designsPath); err != nil {
		return fmt.Errorf("failed to create designs directory: %w", err)
	}

	// Create .gitkeep files to preserve empty directories
	gitkeepState := filepath.Join(statePath, ".gitkeep")
	gitkeepDesigns := filepath.Join(designsPath, ".gitkeep")
	_ = util.SafeWriteFile(gitkeepState, []byte(""), false)
	_ = util.SafeWriteFile(gitkeepDesigns, []byte(""), false)

	return nil
}

// writeWorkflowMarker writes the workflow.yaml configuration file
func (g *Generator) writeWorkflowMarker(path string, cfg *config.WorkflowConfig, force bool) error {
	if util.FileExists(path) && !force {
		return fmt.Errorf("workflow.yaml already exists: %s (use --force to overwrite)", path)
	}

	return g.SaveWorkflowConfig(path, cfg)
}

// SaveWorkflowConfig writes workflow config to file
func (g *Generator) SaveWorkflowConfig(path string, cfg *config.WorkflowConfig) error {
	if err := util.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	content := g.formatWorkflowYAML(cfg)
	return util.SafeWriteFile(path, []byte(content), true)
}

// formatWorkflowYAML creates a nicely formatted YAML string
func (g *Generator) formatWorkflowYAML(cfg *config.WorkflowConfig) string {
	// Build YAML manually for better control over formatting
	yaml := fmt.Sprintf(`version: %d
name: %s
topology: %s
blueprint: %s
paths:
  hub: %s
  docs: %s
state:
  root: %s
  state_dir: %s
  designs_dir: %s
`, cfg.Version, cfg.Name, cfg.Topology, cfg.Blueprint,
		cfg.Paths.Hub, cfg.Paths.Docs,
		cfg.State.Root, cfg.State.StateDir, cfg.State.DesignsDir)

	// Add repos
	yaml += "repos:\n"
	for _, repo := range cfg.Repos {
		yaml += fmt.Sprintf("  - name: %s\n    path: %s\n    kind: %s\n", repo.Name, repo.Path, repo.Kind)
	}

	yaml += fmt.Sprintf(`hooks:
  enabled: %t
gates:
  enabled: %t
mcp:
  vcs: %s
  tracker: %s
  deploy: %s
`, cfg.Hooks.Enabled, cfg.Gates.Enabled,
		cfg.MCP.VCS, cfg.MCP.Tracker, cfg.MCP.Deploy)

	// Add agent permissions if configured
	if len(cfg.AgentPermissions) > 0 {
		yaml += "agent_permissions:\n"
		for agentName, perm := range cfg.AgentPermissions {
			yaml += fmt.Sprintf("  %s:\n", agentName)
			if len(perm.Write) > 0 {
				yaml += "    write:\n"
				for _, repo := range perm.Write {
					yaml += fmt.Sprintf("      - %s\n", repo)
				}
			}
			if len(perm.Read) > 0 {
				yaml += "    read:\n"
				for _, repo := range perm.Read {
					yaml += fmt.Sprintf("      - %s\n", repo)
				}
			}
		}
	}

	return yaml
}
