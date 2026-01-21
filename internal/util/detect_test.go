package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wameedh/ccflow/internal/config"
)

func TestDetectRepoKind(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected config.RepoKind
	}{
		{
			name:     "Node.js project",
			files:    []string{"package.json"},
			expected: config.RepoKindNode,
		},
		{
			name:     "Go project",
			files:    []string{"go.mod"},
			expected: config.RepoKindGo,
		},
		{
			name:     "Java Maven project",
			files:    []string{"pom.xml"},
			expected: config.RepoKindJava,
		},
		{
			name:     "Java Gradle project",
			files:    []string{"build.gradle"},
			expected: config.RepoKindJava,
		},
		{
			name:     "Python project with requirements",
			files:    []string{"requirements.txt"},
			expected: config.RepoKindPython,
		},
		{
			name:     "Python project with pyproject",
			files:    []string{"pyproject.toml"},
			expected: config.RepoKindPython,
		},
		{
			name:     "Swift package",
			files:    []string{"Package.swift"},
			expected: config.RepoKindSwift,
		},
		{
			name:     "Terraform project",
			files:    []string{"main.tf"},
			expected: config.RepoKindTerraform,
		},
		{
			name:     "Unknown project",
			files:    []string{"random.file"},
			expected: config.RepoKindUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				filePath := filepath.Join(tmpDir, file)
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
			}

			kind := DetectRepoKind(tmpDir)
			if kind != tt.expected {
				t.Errorf("DetectRepoKind() = %v, want %v", kind, tt.expected)
			}
		})
	}
}

func TestFindGitRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some git repos
	repos := []string{"repo1", "repo2", "not-a-repo"}
	for _, name := range repos {
		repoPath := filepath.Join(tmpDir, name)
		os.MkdirAll(repoPath, 0755)

		if name != "not-a-repo" {
			gitDir := filepath.Join(repoPath, ".git")
			os.MkdirAll(gitDir, 0755)
		}
	}

	found, err := FindGitRepos(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepos failed: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 git repos, found %d", len(found))
	}
}

func TestIsGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a git repo
	gitRepo := filepath.Join(tmpDir, "git-repo")
	os.MkdirAll(filepath.Join(gitRepo, ".git"), 0755)

	// Create a non-git directory
	nonGit := filepath.Join(tmpDir, "non-git")
	os.MkdirAll(nonGit, 0755)

	if !IsGitRepo(gitRepo) {
		t.Error("IsGitRepo returned false for git repo")
	}

	if IsGitRepo(nonGit) {
		t.Error("IsGitRepo returned true for non-git directory")
	}
}
