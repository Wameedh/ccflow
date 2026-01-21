package installer

import (
	"fmt"
	"path/filepath"

	"github.com/wameedh/ccflow/internal/config"
	"github.com/wameedh/ccflow/internal/util"
)

// Installer handles installing .claude into repositories
type Installer struct{}

// New creates a new Installer
func New() *Installer {
	return &Installer{}
}

// InstallMode represents how .claude should be installed
type InstallMode string

const (
	// InstallModeSymlink creates a symbolic link (default for macOS/Linux)
	InstallModeSymlink InstallMode = "symlink"
	// InstallModeCopy copies the directory (for Windows or when symlinks fail)
	InstallModeCopy InstallMode = "copy"
)

// InstallOptions contains options for installation
type InstallOptions struct {
	HubPath       string           // Path to the .claude directory in the hub
	Repos         []config.RepoConfig // Repositories to install to
	WorkspacePath string           // Root workspace path
	Mode          InstallMode      // Installation mode
	Force         bool             // Force overwrite existing
}

// InstallResult represents the result of installing to a single repo
type InstallResult struct {
	RepoName string
	RepoPath string
	Success  bool
	Error    error
	Skipped  bool
	Message  string
}

// Install installs .claude into all configured repositories
func (i *Installer) Install(opts InstallOptions) []InstallResult {
	var results []InstallResult

	for _, repo := range opts.Repos {
		result := i.installToRepo(opts.HubPath, repo, opts.WorkspacePath, opts.Mode, opts.Force)
		results = append(results, result)
	}

	return results
}

// installToRepo installs .claude to a single repository
func (i *Installer) installToRepo(hubPath string, repo config.RepoConfig, workspacePath string, mode InstallMode, force bool) InstallResult {
	result := InstallResult{
		RepoName: repo.Name,
		RepoPath: filepath.Join(workspacePath, repo.Path),
	}

	// Ensure repo directory exists
	if !util.DirExists(result.RepoPath) {
		result.Error = fmt.Errorf("repository directory does not exist: %s", result.RepoPath)
		return result
	}

	claudePath := filepath.Join(result.RepoPath, ".claude")

	// Check if .claude already exists
	if util.FileExists(claudePath) || util.IsSymlink(claudePath) {
		if util.IsSymlink(claudePath) {
			// Check if it points to our hub
			target, err := util.ReadSymlinkTarget(claudePath)
			if err == nil {
				expectedTarget := i.calculateRelativePath(result.RepoPath, hubPath)
				if target == expectedTarget || target == hubPath {
					result.Success = true
					result.Skipped = true
					result.Message = "already linked to hub"
					return result
				}
			}
		}

		if !force {
			result.Error = fmt.Errorf(".claude already exists in %s (use --force to overwrite)", repo.Name)
			return result
		}

		// Remove existing
		if err := i.removeExisting(claudePath); err != nil {
			result.Error = fmt.Errorf("failed to remove existing .claude: %w", err)
			return result
		}
	}

	// Install based on mode
	switch mode {
	case InstallModeSymlink:
		if err := i.createSymlink(hubPath, claudePath, result.RepoPath); err != nil {
			result.Error = err
			return result
		}
		result.Message = "symlinked to hub"
	case InstallModeCopy:
		if err := i.copyDirectory(hubPath, claudePath); err != nil {
			result.Error = err
			return result
		}
		result.Message = "copied from hub"
	default:
		result.Error = fmt.Errorf("unknown install mode: %s", mode)
		return result
	}

	result.Success = true
	return result
}

// createSymlink creates a relative symbolic link from linkPath to target
func (i *Installer) createSymlink(target, linkPath, repoPath string) error {
	// Calculate relative path and use it for symlink creation
	return util.CreateRelativeSymlink(target, linkPath)
}

// calculateRelativePath calculates the relative path from repoPath to hubPath
func (i *Installer) calculateRelativePath(repoPath, hubPath string) string {
	rel, err := filepath.Rel(repoPath, hubPath)
	if err != nil {
		return hubPath // Fall back to absolute
	}
	return rel
}

// copyDirectory copies a directory recursively
func (i *Installer) copyDirectory(src, dst string) error {
	// This would be implemented for Windows support
	// For now, return an error suggesting symlink mode
	return fmt.Errorf("copy mode not yet implemented; use symlink mode on macOS/Linux")
}

// removeExisting removes an existing .claude (file, directory, or symlink)
func (i *Installer) removeExisting(path string) error {
	if util.IsSymlink(path) {
		return util.RemoveSymlink(path)
	}
	return util.RemoveAll(path)
}

// VerifyInstallation checks if .claude is properly installed in a repo
func (i *Installer) VerifyInstallation(repoPath, hubPath string) (bool, string) {
	claudePath := filepath.Join(repoPath, ".claude")

	if !util.FileExists(claudePath) && !util.IsSymlink(claudePath) {
		return false, "missing"
	}

	if util.IsSymlink(claudePath) {
		target, err := util.ReadSymlinkTarget(claudePath)
		if err != nil {
			return false, "broken symlink"
		}

		// Resolve the target relative to the repo
		absTarget := target
		if !filepath.IsAbs(target) {
			absTarget = filepath.Join(repoPath, target)
		}
		absTarget = filepath.Clean(absTarget)

		if !util.DirExists(absTarget) {
			return false, fmt.Sprintf("broken symlink (target missing: %s)", target)
		}

		// Check if it points to expected hub
		expectedHub := filepath.Clean(hubPath)
		if absTarget != expectedHub {
			return false, fmt.Sprintf("symlink points to unexpected location: %s", target)
		}

		return true, "ok (symlink)"
	}

	// It's a regular directory
	if util.DirExists(claudePath) {
		return true, "ok (local)"
	}

	return false, "unknown state"
}
