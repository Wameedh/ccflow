package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/workspace"
)

func createTestWorkspace(t *testing.T) (*workspace.Workspace, string) {
	tmpDir := t.TempDir()

	// Create multi-repo structure
	hubDir := filepath.Join(tmpDir, "workflow-hub")
	claudeDir := filepath.Join(hubDir, ".claude")
	os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755)
	os.MkdirAll(filepath.Join(claudeDir, "commands"), 0755)
	os.MkdirAll(filepath.Join(claudeDir, "hooks"), 0755)

	// Create settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	settings := `{
		"hooks": [
			{"event": "Stop", "script": "./hooks/end-of-turn.sh"}
		],
		"permissions": {}
	}`
	os.WriteFile(settingsPath, []byte(settings), 0644)

	// Create hook script
	hookPath := filepath.Join(claudeDir, "hooks", "end-of-turn.sh")
	os.WriteFile(hookPath, []byte("#!/bin/bash\nexit 0"), 0755)

	// Create state directories
	stateDir := filepath.Join(tmpDir, "docs", "workflow", "state")
	designsDir := filepath.Join(tmpDir, "docs", "workflow", "designs")
	os.MkdirAll(stateDir, 0755)
	os.MkdirAll(designsDir, 0755)

	// Create workflow.yaml
	cfg := &config.WorkflowConfig{
		Version:   1,
		Name:      "test-workflow",
		Topology:  config.TopologyMultiRepo,
		Blueprint: "web-dev",
		Paths: config.PathsConfig{
			Hub:  "workflow-hub",
			Docs: "docs",
		},
		State: config.StateConfig{
			Root:       "docs/workflow",
			StateDir:   "docs/workflow/state",
			DesignsDir: "docs/workflow/designs",
		},
		Repos: []config.RepoConfig{
			{Name: "web", Path: "web", Kind: config.RepoKindNode},
		},
	}
	cfg.Hooks.Enabled = true
	cfg.Gates.Enabled = true

	markerPath := filepath.Join(hubDir, "workflow.yaml")
	workspace.SaveConfig(markerPath, cfg)

	// Create repo with symlink
	repoDir := filepath.Join(tmpDir, "web")
	os.MkdirAll(repoDir, 0755)
	os.Symlink(filepath.Join("..", "workflow-hub", ".claude"), filepath.Join(repoDir, ".claude"))

	ws := &workspace.Workspace{
		Root:       tmpDir,
		ConfigPath: markerPath,
		Topology:   config.TopologyMultiRepo,
		Config:     cfg,
	}

	return ws, tmpDir
}

func TestStatus(t *testing.T) {
	ws, _ := createTestWorkspace(t)
	v := New()

	result := v.Status(ws)

	if result.WorkflowName != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got %s", result.WorkflowName)
	}

	if result.Blueprint != "web-dev" {
		t.Errorf("Expected blueprint 'web-dev', got %s", result.Blueprint)
	}

	if !result.HooksEnabled {
		t.Error("Expected hooks to be enabled")
	}

	if len(result.Repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(result.Repos))
	}

	if result.Repos[0].Status != "ok" {
		t.Errorf("Expected repo status 'ok', got %s", result.Repos[0].Status)
	}
}

func TestDoctor(t *testing.T) {
	ws, _ := createTestWorkspace(t)
	v := New()

	result := v.Doctor(ws)

	if result.Failed > 0 {
		t.Errorf("Expected no failures, got %d", result.Failed)
		for _, check := range result.Checks {
			if check.Status == "fail" {
				t.Logf("Failed check: %s - %s", check.Name, check.Message)
			}
		}
	}
}

func TestDoctor_BrokenSymlink(t *testing.T) {
	ws, tmpDir := createTestWorkspace(t)

	// Break the symlink by removing the target
	os.RemoveAll(filepath.Join(tmpDir, "workflow-hub", ".claude"))

	v := New()
	result := v.Doctor(ws)

	// Should have failures due to broken symlink and missing directories
	if result.Failed == 0 {
		t.Error("Expected failures with broken symlink")
	}
}

func TestDoctor_InvalidSettingsJSON(t *testing.T) {
	ws, tmpDir := createTestWorkspace(t)

	// Write invalid JSON
	settingsPath := filepath.Join(tmpDir, "workflow-hub", ".claude", "settings.json")
	os.WriteFile(settingsPath, []byte("invalid json"), 0644)

	v := New()
	result := v.Doctor(ws)

	// Should have failure for invalid JSON
	foundJSONFailure := false
	for _, check := range result.Checks {
		if check.Name == "Settings JSON" && check.Status == "fail" {
			foundJSONFailure = true
			break
		}
	}

	if !foundJSONFailure {
		t.Error("Expected failure for invalid settings.json")
	}
}

func TestDoctor_NonExecutableHook(t *testing.T) {
	ws, tmpDir := createTestWorkspace(t)

	// Make hook non-executable
	hookPath := filepath.Join(tmpDir, "workflow-hub", ".claude", "hooks", "end-of-turn.sh")
	os.Chmod(hookPath, 0644)

	v := New()
	result := v.Doctor(ws)

	// Should have warning for non-executable hook
	foundWarning := false
	for _, check := range result.Checks {
		if check.Status == "warn" && check.Name == "Hook: end-of-turn.sh" {
			foundWarning = true
			break
		}
	}

	if !foundWarning {
		t.Error("Expected warning for non-executable hook")
	}
}
