package util

import (
	"path/filepath"

	"github.com/wameedh/ccflow/internal/config"
)

// DetectRepoKind attempts to detect the type of repository based on marker files
func DetectRepoKind(repoPath string) config.RepoKind {
	// Check for various language/framework markers
	markers := map[string]config.RepoKind{
		"package.json":      config.RepoKindNode,
		"pom.xml":           config.RepoKindJava,
		"build.gradle":      config.RepoKindJava,
		"build.gradle.kts":  config.RepoKindJava,
		"go.mod":            config.RepoKindGo,
		"requirements.txt":  config.RepoKindPython,
		"setup.py":          config.RepoKindPython,
		"pyproject.toml":    config.RepoKindPython,
		"Package.swift":     config.RepoKindSwift,
		"*.xcodeproj":       config.RepoKindSwift,
		"*.xcworkspace":     config.RepoKindSwift,
		"main.tf":           config.RepoKindTerraform,
		"terraform.tf":      config.RepoKindTerraform,
	}

	for marker, kind := range markers {
		// Handle glob patterns
		if marker[0] == '*' {
			matches, _ := filepath.Glob(filepath.Join(repoPath, marker))
			if len(matches) > 0 {
				return kind
			}
		} else {
			if FileExists(filepath.Join(repoPath, marker)) {
				return kind
			}
		}
	}

	// Check for docs-only repo
	if FileExists(filepath.Join(repoPath, "README.md")) || FileExists(filepath.Join(repoPath, "docs")) {
		// If no code markers but has docs, might be a docs repo
		// But only if explicitly no other markers
		hasCode := false
		codeMarkers := []string{"src", "lib", "app", "pkg", "cmd"}
		for _, dir := range codeMarkers {
			if DirExists(filepath.Join(repoPath, dir)) {
				hasCode = true
				break
			}
		}
		if !hasCode {
			return config.RepoKindDocs
		}
	}

	return config.RepoKindUnknown
}

// FindGitRepos finds all git repositories under a given path (non-recursive beyond first level)
func FindGitRepos(rootPath string) ([]string, error) {
	var repos []string

	entries, err := filepath.Glob(filepath.Join(rootPath, "*", ".git"))
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		repoPath := filepath.Dir(entry)
		repos = append(repos, repoPath)
	}

	return repos, nil
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(path string) bool {
	return DirExists(filepath.Join(path, ".git"))
}
