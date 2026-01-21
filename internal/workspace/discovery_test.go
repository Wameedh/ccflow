package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wameedh/ccflow/internal/config"
)

func TestDiscover_MultiRepo(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Resolve symlinks for path comparison (macOS /var -> /private/var)
	tmpDir, _ = filepath.EvalSymlinks(tmpDir)

	// Create multi-repo marker
	hubDir := filepath.Join(tmpDir, "workflow-hub")
	if err := os.MkdirAll(hubDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewDefaultWorkflowConfig("test-workflow")
	markerPath := filepath.Join(hubDir, "workflow.yaml")
	if err := SaveConfig(markerPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory to discover from
	subDir := filepath.Join(tmpDir, "web", "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to subdirectory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(subDir)

	// Test discovery
	ws, err := Discover("")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Resolve symlinks in the result for comparison
	wsRoot, _ := filepath.EvalSymlinks(ws.Root)
	if wsRoot != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, wsRoot)
	}

	if ws.Topology != config.TopologyMultiRepo {
		t.Errorf("Expected multi-repo topology, got %s", ws.Topology)
	}

	if ws.Config.Name != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got %s", ws.Config.Name)
	}
}

func TestDiscover_SingleRepo(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create single-repo marker
	ccflowDir := filepath.Join(tmpDir, ".ccflow")
	if err := os.MkdirAll(ccflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewDefaultWorkflowConfig("single-repo-test")
	cfg.Topology = config.TopologySingleRepo
	markerPath := filepath.Join(ccflowDir, "workflow.yaml")
	if err := SaveConfig(markerPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Change to directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	// Test discovery
	ws, err := Discover("")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if ws.Topology != config.TopologySingleRepo {
		t.Errorf("Expected single-repo topology, got %s", ws.Topology)
	}
}

func TestDiscover_WithOverride(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create multi-repo marker
	hubDir := filepath.Join(tmpDir, "workflow-hub")
	if err := os.MkdirAll(hubDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewDefaultWorkflowConfig("override-test")
	markerPath := filepath.Join(hubDir, "workflow.yaml")
	if err := SaveConfig(markerPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Test discovery with override
	ws, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover with override failed: %v", err)
	}

	if ws.Config.Name != "override-test" {
		t.Errorf("Expected workflow name 'override-test', got %s", ws.Config.Name)
	}
}

func TestDiscover_WithEnvVar(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create multi-repo marker
	hubDir := filepath.Join(tmpDir, "workflow-hub")
	if err := os.MkdirAll(hubDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.NewDefaultWorkflowConfig("env-test")
	markerPath := filepath.Join(hubDir, "workflow.yaml")
	if err := SaveConfig(markerPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Set env var
	os.Setenv(EnvWorkspace, tmpDir)
	defer os.Unsetenv(EnvWorkspace)

	// Test discovery from different directory
	otherDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(otherDir)

	ws, err := Discover("")
	if err != nil {
		t.Fatalf("Discover with env var failed: %v", err)
	}

	if ws.Config.Name != "env-test" {
		t.Errorf("Expected workflow name 'env-test', got %s", ws.Config.Name)
	}
}

func TestDiscover_NotFound(t *testing.T) {
	// Create empty temp directory
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	_, err := Discover("")
	if err == nil {
		t.Error("Expected error when no workflow found")
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workflow.yaml")

	original := &config.WorkflowConfig{
		Version:   1,
		Name:      "test",
		Topology:  config.TopologyMultiRepo,
		Blueprint: "web-dev",
		Paths: config.PathsConfig{
			Hub:  "workflow-hub",
			Docs: "docs",
		},
	}

	if err := SaveConfig(configPath, original); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", loaded.Name, original.Name)
	}

	if loaded.Blueprint != original.Blueprint {
		t.Errorf("Blueprint mismatch: got %s, want %s", loaded.Blueprint, original.Blueprint)
	}
}
