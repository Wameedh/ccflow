package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
)

const (
	// MultiRepoMarker is the path to workflow.yaml in multi-repo mode
	MultiRepoMarker = "workflow-hub/workflow.yaml"
	// SingleRepoMarker is the path to workflow.yaml in single-repo mode
	SingleRepoMarker = ".ccflow/workflow.yaml"
	// EnvWorkspace is the environment variable to override workspace location
	EnvWorkspace = "CCFLOW_WORKSPACE"
)

// Workspace represents a discovered workflow workspace
type Workspace struct {
	Root       string                 // Absolute path to workspace root
	ConfigPath string                 // Absolute path to workflow.yaml
	Topology   config.Topology        // multi-repo or single-repo
	Config     *config.WorkflowConfig // Parsed configuration
}

// Discover finds the nearest workflow workspace from the current directory
// Resolution order:
// 1. --workspace flag (passed as override)
// 2. CCFLOW_WORKSPACE environment variable
// 3. Walk up from current directory looking for markers
func Discover(override string) (*Workspace, error) {
	// Check override first
	if override != "" {
		return loadWorkspace(override)
	}

	// Check environment variable
	if envPath := os.Getenv(EnvWorkspace); envPath != "" {
		return loadWorkspace(envPath)
	}

	// Walk up from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return discoverFromPath(cwd)
}

// discoverFromPath walks up the directory tree looking for workflow markers
func discoverFromPath(startPath string) (*Workspace, error) {
	current := startPath

	for {
		// Check for multi-repo marker first (more specific)
		multiRepoPath := filepath.Join(current, MultiRepoMarker)
		if util.FileExists(multiRepoPath) {
			return loadWorkspaceFromMarker(current, multiRepoPath, config.TopologyMultiRepo)
		}

		// Check for single-repo marker
		singleRepoPath := filepath.Join(current, SingleRepoMarker)
		if util.FileExists(singleRepoPath) {
			return loadWorkspaceFromMarker(current, singleRepoPath, config.TopologySingleRepo)
		}

		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			break
		}
		current = parent
	}

	return nil, fmt.Errorf("no workflow found. Run 'ccflow run' to create one")
}

// loadWorkspace loads a workspace from a given path (which can be a directory or marker file)
func loadWorkspace(path string) (*Workspace, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if it's a direct path to workflow.yaml
	if filepath.Base(absPath) == "workflow.yaml" && util.FileExists(absPath) {
		// Determine topology from path
		dir := filepath.Dir(absPath)
		if filepath.Base(dir) == "workflow-hub" {
			return loadWorkspaceFromMarker(filepath.Dir(dir), absPath, config.TopologyMultiRepo)
		} else if filepath.Base(dir) == ".ccflow" {
			return loadWorkspaceFromMarker(filepath.Dir(dir), absPath, config.TopologySingleRepo)
		}
	}

	// It's a directory - look for markers
	multiRepoPath := filepath.Join(absPath, MultiRepoMarker)
	if util.FileExists(multiRepoPath) {
		return loadWorkspaceFromMarker(absPath, multiRepoPath, config.TopologyMultiRepo)
	}

	singleRepoPath := filepath.Join(absPath, SingleRepoMarker)
	if util.FileExists(singleRepoPath) {
		return loadWorkspaceFromMarker(absPath, singleRepoPath, config.TopologySingleRepo)
	}

	return nil, fmt.Errorf("no workflow.yaml found at %s", absPath)
}

// loadWorkspaceFromMarker loads the workspace configuration from a marker file
func loadWorkspaceFromMarker(root, markerPath string, topology config.Topology) (*Workspace, error) {
	cfg, err := LoadConfig(markerPath)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Root:       root,
		ConfigPath: markerPath,
		Topology:   topology,
		Config:     cfg,
	}, nil
}

// LoadConfig reads and parses a workflow.yaml file
func LoadConfig(path string) (*config.WorkflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow.yaml: %w", err)
	}

	var cfg config.WorkflowConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse workflow.yaml: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes the workflow configuration to the given path
func SaveConfig(path string, cfg *config.WorkflowConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow.yaml: %w", err)
	}

	if err := util.EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetHubPath returns the absolute path to the .claude hub directory
func (w *Workspace) GetHubPath() string {
	if w.Topology == config.TopologyMultiRepo {
		return filepath.Join(w.Root, w.Config.Paths.Hub, ".claude")
	}
	return filepath.Join(w.Root, ".claude")
}

// GetDocsPath returns the absolute path to the docs directory
func (w *Workspace) GetDocsPath() string {
	return filepath.Join(w.Root, w.Config.Paths.Docs)
}

// GetStatePath returns the absolute path to the workflow state directory
func (w *Workspace) GetStatePath() string {
	return filepath.Join(w.Root, w.Config.State.StateDir)
}

// GetDesignsPath returns the absolute path to the workflow designs directory
func (w *Workspace) GetDesignsPath() string {
	return filepath.Join(w.Root, w.Config.State.DesignsDir)
}
