package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
)

func TestInstall_Symlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude directory with content
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Create a test file in hub
	testFile := filepath.Join(hubPath, "settings.json")
	if err := os.WriteFile(testFile, []byte(`{"test": true}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create repo directory
	repoPath := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	// Install
	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         []config.RepoConfig{{Name: "repo", Path: "repo"}},
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         false,
	})

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Install failed: %v", result.Error)
	}
	if result.Message != "symlinked to hub" {
		t.Errorf("Expected 'symlinked to hub', got '%s'", result.Message)
	}

	// Verify symlink was created
	claudePath := filepath.Join(repoPath, ".claude")
	if !util.IsSymlink(claudePath) {
		t.Error(".claude is not a symlink")
	}

	// Verify symlink works (can access files through it)
	linkedFile := filepath.Join(claudePath, "settings.json")
	content, err := os.ReadFile(linkedFile)
	if err != nil {
		t.Errorf("Failed to read through symlink: %v", err)
	}
	if string(content) != `{"test": true}` {
		t.Errorf("Content mismatch through symlink")
	}
}

func TestInstall_ExistingSymlink_SameTarget(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Create repo directory
	repoPath := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	// Manually create symlink pointing to hub
	claudePath := filepath.Join(repoPath, ".claude")
	relTarget, _ := filepath.Rel(repoPath, hubPath)
	if err := os.Symlink(relTarget, claudePath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Install (should skip since already linked)
	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         []config.RepoConfig{{Name: "repo", Path: "repo"}},
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         false,
	})

	result := results[0]
	if !result.Success {
		t.Errorf("Install failed: %v", result.Error)
	}
	if !result.Skipped {
		t.Error("Expected install to be skipped (already linked)")
	}
	if result.Message != "already linked to hub" {
		t.Errorf("Expected 'already linked to hub', got '%s'", result.Message)
	}
}

func TestInstall_ExistingClaudeDir_NoForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Create repo with existing .claude directory (not symlink)
	repoPath := filepath.Join(tmpDir, "repo")
	claudePath := filepath.Join(repoPath, ".claude")
	if err := os.MkdirAll(claudePath, 0755); err != nil {
		t.Fatalf("Failed to create existing .claude: %v", err)
	}

	// Install without force
	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         []config.RepoConfig{{Name: "repo", Path: "repo"}},
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         false,
	})

	result := results[0]
	if result.Success {
		t.Error("Expected failure when .claude exists without force")
	}
	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestInstall_ExistingClaudeDir_Force(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude with content
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}
	hubFile := filepath.Join(hubPath, "settings.json")
	if err := os.WriteFile(hubFile, []byte(`{"hub": true}`), 0644); err != nil {
		t.Fatalf("Failed to create hub file: %v", err)
	}

	// Create repo with existing .claude directory
	repoPath := filepath.Join(tmpDir, "repo")
	claudePath := filepath.Join(repoPath, ".claude")
	if err := os.MkdirAll(claudePath, 0755); err != nil {
		t.Fatalf("Failed to create existing .claude: %v", err)
	}
	// Put a file in the existing .claude to verify it gets replaced
	oldFile := filepath.Join(claudePath, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	// Install with force
	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         []config.RepoConfig{{Name: "repo", Path: "repo"}},
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         true,
	})

	result := results[0]
	if !result.Success {
		t.Errorf("Install with force failed: %v", result.Error)
	}

	// Verify it's now a symlink
	if !util.IsSymlink(claudePath) {
		t.Error(".claude should be a symlink after force")
	}

	// Verify old file is gone (now it's a symlink to hub)
	if util.FileExists(oldFile) {
		t.Error("Old file should not exist after force replace")
	}

	// Verify hub content is accessible
	linkedSettings := filepath.Join(claudePath, "settings.json")
	content, err := os.ReadFile(linkedSettings)
	if err != nil {
		t.Errorf("Failed to read hub content through symlink: %v", err)
	}
	if string(content) != `{"hub": true}` {
		t.Errorf("Expected hub content, got '%s'", string(content))
	}
}

func TestInstall_RepoDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Don't create repo directory

	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         []config.RepoConfig{{Name: "nonexistent", Path: "nonexistent"}},
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         false,
	})

	result := results[0]
	if result.Success {
		t.Error("Expected failure when repo doesn't exist")
	}
	if result.Error == nil {
		t.Error("Expected error for nonexistent repo")
	}
}

func TestInstall_MultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub/.claude
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Create multiple repo directories
	repos := []config.RepoConfig{
		{Name: "frontend", Path: "frontend"},
		{Name: "backend", Path: "backend"},
		{Name: "shared", Path: "libs/shared"},
	}

	for _, repo := range repos {
		repoPath := filepath.Join(tmpDir, repo.Path)
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("Failed to create repo %s: %v", repo.Name, err)
		}
	}

	// Install to all repos
	inst := New()
	results := inst.Install(InstallOptions{
		HubPath:       hubPath,
		Repos:         repos,
		WorkspacePath: tmpDir,
		Mode:          InstallModeSymlink,
		Force:         false,
	})

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if !result.Success {
			t.Errorf("Repo %s: install failed: %v", repos[i].Name, result.Error)
		}

		claudePath := filepath.Join(tmpDir, repos[i].Path, ".claude")
		if !util.IsSymlink(claudePath) {
			t.Errorf("Repo %s: .claude is not a symlink", repos[i].Name)
		}
	}
}

func TestVerifyInstallation_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	repoPath := filepath.Join(tmpDir, "repo")

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	inst := New()
	ok, status := inst.VerifyInstallation(repoPath, hubPath)

	if ok {
		t.Error("Expected verification to fail for missing .claude")
	}
	if status != "missing" {
		t.Errorf("Expected status 'missing', got '%s'", status)
	}
}

func TestVerifyInstallation_ValidSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create hub
	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	if err := os.MkdirAll(hubPath, 0755); err != nil {
		t.Fatalf("Failed to create hub: %v", err)
	}

	// Create repo with symlink
	repoPath := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	claudePath := filepath.Join(repoPath, ".claude")
	relTarget, _ := filepath.Rel(repoPath, hubPath)
	if err := os.Symlink(relTarget, claudePath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	inst := New()
	ok, status := inst.VerifyInstallation(repoPath, hubPath)

	if !ok {
		t.Errorf("Expected verification to pass, got status: %s", status)
	}
	if status != "ok (symlink)" {
		t.Errorf("Expected status 'ok (symlink)', got '%s'", status)
	}
}

func TestVerifyInstallation_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repo with symlink to nonexistent target
	repoPath := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	claudePath := filepath.Join(repoPath, ".claude")
	if err := os.Symlink("../nonexistent", claudePath); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	inst := New()
	ok, status := inst.VerifyInstallation(repoPath, hubPath)

	if ok {
		t.Error("Expected verification to fail for broken symlink")
	}
	if status != "broken symlink (target missing: ../nonexistent)" {
		t.Errorf("Unexpected status: %s", status)
	}
}

func TestVerifyInstallation_LocalDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repo with local .claude directory (not symlink)
	repoPath := filepath.Join(tmpDir, "repo")
	claudePath := filepath.Join(repoPath, ".claude")
	if err := os.MkdirAll(claudePath, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	hubPath := filepath.Join(tmpDir, "hub", ".claude")
	inst := New()
	ok, status := inst.VerifyInstallation(repoPath, hubPath)

	if !ok {
		t.Errorf("Expected verification to pass for local dir, got status: %s", status)
	}
	if status != "ok (local)" {
		t.Errorf("Expected status 'ok (local)', got '%s'", status)
	}
}
