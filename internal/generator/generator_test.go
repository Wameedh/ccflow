package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
)

func setupTestGenerator(t *testing.T) (*Generator, *blueprint.Manager) {
	t.Helper()
	mgr, err := blueprint.NewManager()
	if err != nil {
		t.Fatalf("Failed to create blueprint manager: %v", err)
	}
	return New(mgr), mgr
}

func TestGenerate_SingleRepo(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "test-workflow",
		Blueprint:     "go-cli-dev",
		Topology:      config.TopologySingleRepo,
		Force:         false,
		HooksEnabled:  true,
		GatesEnabled:  true,
		VCS:           config.VCSGitHub,
		Tracker:       config.TrackerNone,
	}

	cfg, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify config values
	if cfg.Name != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got '%s'", cfg.Name)
	}
	if cfg.Blueprint != "go-cli-dev" {
		t.Errorf("Expected blueprint 'go-cli-dev', got '%s'", cfg.Blueprint)
	}
	if cfg.Topology != config.TopologySingleRepo {
		t.Errorf("Expected single-repo topology, got '%s'", cfg.Topology)
	}

	// Verify .claude directory was created
	claudePath := filepath.Join(tmpDir, ".claude")
	if !util.DirExists(claudePath) {
		t.Error(".claude directory not created")
	}

	// Verify agents directory has files
	agentsDir := filepath.Join(claudePath, "agents")
	if !util.DirExists(agentsDir) {
		t.Error("agents directory not created")
	}

	// Verify commands directory has files
	commandsDir := filepath.Join(claudePath, "commands")
	if !util.DirExists(commandsDir) {
		t.Error("commands directory not created")
	}

	// Verify settings.json exists
	settingsPath := filepath.Join(claudePath, "settings.json")
	if !util.FileExists(settingsPath) {
		t.Error("settings.json not created")
	}

	// Verify workflow.yaml marker exists
	markerPath := filepath.Join(tmpDir, ".ccflow", "workflow.yaml")
	if !util.FileExists(markerPath) {
		t.Error("workflow.yaml marker not created")
	}

	// Verify docs structure
	docsPath := filepath.Join(tmpDir, "docs", "workflow")
	if !util.DirExists(docsPath) {
		t.Error("docs/workflow directory not created")
	}
}

func TestGenerate_MultiRepo(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "multi-workflow",
		Blueprint:     "web-dev",
		Topology:      config.TopologyMultiRepo,
		Force:         false,
		HooksEnabled:  true,
		GatesEnabled:  true,
		VCS:           config.VCSGitHub,
		Tracker:       config.TrackerLinear,
		Repos: []config.RepoConfig{
			{Name: "frontend", Path: "frontend", Kind: config.RepoKindNode},
		},
	}

	cfg, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify topology
	if cfg.Topology != config.TopologyMultiRepo {
		t.Errorf("Expected multi-repo topology, got '%s'", cfg.Topology)
	}

	// Verify hub directory was created
	hubPath := filepath.Join(tmpDir, "workflow-hub")
	if !util.DirExists(hubPath) {
		t.Error("workflow-hub directory not created")
	}

	// Verify .claude directory in hub
	claudePath := filepath.Join(hubPath, ".claude")
	if !util.DirExists(claudePath) {
		t.Error(".claude directory not created in hub")
	}

	// Verify workflow.yaml marker in hub
	markerPath := filepath.Join(hubPath, "workflow.yaml")
	if !util.FileExists(markerPath) {
		t.Error("workflow.yaml not created in hub")
	}
}

func TestCheckExistingFiles_NoExisting(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		Blueprint:     "go-cli-dev",
		Topology:      config.TopologySingleRepo,
	}

	existing, err := gen.CheckExistingFiles(opts)
	if err != nil {
		t.Fatalf("CheckExistingFiles failed: %v", err)
	}

	if len(existing) != 0 {
		t.Errorf("Expected no existing files, got %v", existing)
	}
}

func TestCheckExistingFiles_DetectsExisting(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	// First, generate the workflow
	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "test",
		Blueprint:     "go-cli-dev",
		Topology:      config.TopologySingleRepo,
		Force:         true,
	}

	_, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Initial generate failed: %v", err)
	}

	// Now check for existing files
	opts.Force = false
	existing, err := gen.CheckExistingFiles(opts)
	if err != nil {
		t.Fatalf("CheckExistingFiles failed: %v", err)
	}

	if len(existing) == 0 {
		t.Error("Expected to detect existing files, but got none")
	}

	// Should detect settings.json and workflow.yaml at minimum
	foundSettings := false
	foundMarker := false
	for _, path := range existing {
		if filepath.Base(path) == "settings.json" {
			foundSettings = true
		}
		if filepath.Base(path) == "workflow.yaml" {
			foundMarker = true
		}
	}

	if !foundSettings {
		t.Error("Failed to detect existing settings.json")
	}
	if !foundMarker {
		t.Error("Failed to detect existing workflow.yaml")
	}
}

func TestGenerate_WithExistingFiles_NoForce(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "test",
		Blueprint:     "go-cli-dev",
		Topology:      config.TopologySingleRepo,
		Force:         true,
	}

	// First generate
	_, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Initial generate failed: %v", err)
	}

	// Try to generate again without force
	opts.Force = false
	_, err = gen.Generate(opts)
	if err == nil {
		t.Error("Expected error when generating over existing files without force")
	}
}

func TestGenerate_WithExistingFiles_Force(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "test",
		Blueprint:     "go-cli-dev",
		Topology:      config.TopologySingleRepo,
		Force:         true,
	}

	// First generate
	_, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Initial generate failed: %v", err)
	}

	// Modify a file to verify it gets overwritten
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	os.WriteFile(settingsPath, []byte("modified"), 0644)

	// Generate again with force
	_, err = gen.Generate(opts)
	if err != nil {
		t.Fatalf("Regenerate with force failed: %v", err)
	}

	// Verify file was overwritten (not "modified" anymore)
	content, _ := os.ReadFile(settingsPath)
	if string(content) == "modified" {
		t.Error("File was not overwritten with force=true")
	}
}

func TestGenerate_InvalidBlueprint(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "test",
		Blueprint:     "nonexistent-blueprint",
		Topology:      config.TopologySingleRepo,
	}

	_, err := gen.Generate(opts)
	if err == nil {
		t.Error("Expected error for invalid blueprint")
	}
}

func TestCheckExistingFiles_MultiRepo(t *testing.T) {
	gen, _ := setupTestGenerator(t)
	tmpDir := t.TempDir()

	// First, generate multi-repo workflow
	opts := GenerateOptions{
		WorkspacePath: tmpDir,
		WorkflowName:  "multi",
		Blueprint:     "web-dev",
		Topology:      config.TopologyMultiRepo,
		Force:         true,
	}

	_, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Initial generate failed: %v", err)
	}

	// Check for existing files
	existing, err := gen.CheckExistingFiles(opts)
	if err != nil {
		t.Fatalf("CheckExistingFiles failed: %v", err)
	}

	if len(existing) == 0 {
		t.Error("Expected to detect existing files in multi-repo setup")
	}

	// Should detect workflow.yaml in hub
	foundHubMarker := false
	for _, path := range existing {
		if filepath.Base(path) == "workflow.yaml" {
			foundHubMarker = true
			break
		}
	}
	if !foundHubMarker {
		t.Error("Failed to detect existing workflow.yaml in hub")
	}
}
