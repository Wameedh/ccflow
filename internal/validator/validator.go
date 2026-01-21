package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/installer"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
)

// Validator performs health checks on a workflow
type Validator struct {
	installer *installer.Installer
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		installer: installer.New(),
	}
}

// StatusResult contains the result of a status check
type StatusResult struct {
	WorkflowName string
	Blueprint    string
	Topology     config.Topology
	HubPath      string
	DocsPath     string
	Repos        []RepoStatus
	Hooks        []HookStatus
	HooksEnabled bool
	GatesEnabled bool
	Errors       []string
	Warnings     []string
}

// RepoStatus represents the status of a single repo
type RepoStatus struct {
	Name    string
	Path    string
	Kind    config.RepoKind
	Status  string // "ok", "broken", "missing"
	Message string
}

// HookStatus represents the status of a hook
type HookStatus struct {
	Name       string
	ScriptPath string
	Exists     bool
	Executable bool
	Events     []string
}

// DoctorResult contains the result of a doctor check
type DoctorResult struct {
	Checks   []Check
	Passed   int
	Failed   int
	Warnings int
}

// Check represents a single doctor check
type Check struct {
	Name        string
	Status      string // "pass", "fail", "warn"
	Message     string
	Remediation string
}

// Status performs a status check on the workflow
func (v *Validator) Status(ws *workspace.Workspace) *StatusResult {
	result := &StatusResult{
		WorkflowName: ws.Config.Name,
		Blueprint:    ws.Config.Blueprint,
		Topology:     ws.Topology,
		HubPath:      ws.GetHubPath(),
		DocsPath:     ws.GetDocsPath(),
		HooksEnabled: ws.Config.Hooks.Enabled,
		GatesEnabled: ws.Config.Gates.Enabled,
	}

	// Check repos
	for _, repo := range ws.Config.Repos {
		repoStatus := v.checkRepo(ws, repo)
		result.Repos = append(result.Repos, repoStatus)
		if repoStatus.Status == "broken" || repoStatus.Status == "missing" {
			result.Errors = append(result.Errors, fmt.Sprintf("repo %s: %s", repo.Name, repoStatus.Message))
		}
	}

	// Check hooks
	result.Hooks = v.checkHooks(ws)
	for _, hook := range result.Hooks {
		if !hook.Exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("hook script missing: %s", hook.ScriptPath))
		} else if !hook.Executable {
			result.Warnings = append(result.Warnings, fmt.Sprintf("hook script not executable: %s", hook.ScriptPath))
		}
	}

	return result
}

// checkRepo checks the status of a single repository
func (v *Validator) checkRepo(ws *workspace.Workspace, repo config.RepoConfig) RepoStatus {
	status := RepoStatus{
		Name: repo.Name,
		Path: filepath.Join(ws.Root, repo.Path),
		Kind: repo.Kind,
	}

	// Check if repo directory exists
	if !util.DirExists(status.Path) {
		status.Status = "missing"
		status.Message = "directory does not exist"
		return status
	}

	// For multi-repo, check symlink
	if ws.Topology == config.TopologyMultiRepo {
		ok, msg := v.installer.VerifyInstallation(status.Path, ws.GetHubPath())
		if ok {
			status.Status = "ok"
			status.Message = msg
		} else {
			status.Status = "broken"
			status.Message = msg
		}
	} else {
		// Single repo - just check .claude exists
		claudePath := filepath.Join(status.Path, ".claude")
		if util.DirExists(claudePath) {
			status.Status = "ok"
			status.Message = "local .claude directory"
		} else {
			status.Status = "missing"
			status.Message = ".claude directory missing"
		}
	}

	return status
}

// checkHooks checks the status of all configured hooks
func (v *Validator) checkHooks(ws *workspace.Workspace) []HookStatus {
	var hooks []HookStatus

	settingsPath := filepath.Join(ws.GetHubPath(), "settings.json")
	if !util.FileExists(settingsPath) {
		return hooks
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return hooks
	}

	var settings struct {
		Hooks []struct {
			Event    string   `json:"event"`
			Script   string   `json:"script"`
			Commands []string `json:"commands,omitempty"`
		} `json:"hooks"`
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return hooks
	}

	// Group by script
	hookMap := make(map[string]*HookStatus)
	for _, h := range settings.Hooks {
		scriptPath := h.Script
		if !filepath.IsAbs(scriptPath) {
			scriptPath = filepath.Join(ws.GetHubPath(), h.Script)
		}

		existing, ok := hookMap[scriptPath]
		if !ok {
			hookStatus := &HookStatus{
				Name:       filepath.Base(scriptPath),
				ScriptPath: scriptPath,
				Exists:     util.FileExists(scriptPath),
				Executable: util.IsExecutable(scriptPath),
				Events:     []string{h.Event},
			}
			hookMap[scriptPath] = hookStatus
		} else {
			existing.Events = append(existing.Events, h.Event)
		}
	}

	for _, h := range hookMap {
		hooks = append(hooks, *h)
	}

	return hooks
}

// Doctor performs comprehensive health checks
func (v *Validator) Doctor(ws *workspace.Workspace) *DoctorResult {
	result := &DoctorResult{}

	// Check 1: Marker file exists
	result.addCheck(v.checkMarker(ws))

	// Check 2: settings.json is valid
	result.addCheck(v.checkSettingsJSON(ws))

	// Check 3: Hook scripts exist and are executable
	for _, hookCheck := range v.checkHookScripts(ws) {
		result.addCheck(hookCheck)
	}

	// Check 4: Symlinks are correct (multi-repo only)
	if ws.Topology == config.TopologyMultiRepo {
		for _, symlinkCheck := range v.checkSymlinks(ws) {
			result.addCheck(symlinkCheck)
		}
	}

	// Check 5: Required directories exist
	result.addCheck(v.checkDirectories(ws))

	return result
}

// addCheck adds a check result and updates counts
func (r *DoctorResult) addCheck(check Check) {
	r.Checks = append(r.Checks, check)
	switch check.Status {
	case "pass":
		r.Passed++
	case "fail":
		r.Failed++
	case "warn":
		r.Warnings++
	}
}

// checkMarker verifies the workflow marker file
func (v *Validator) checkMarker(ws *workspace.Workspace) Check {
	check := Check{Name: "Workflow marker"}

	if util.FileExists(ws.ConfigPath) {
		check.Status = "pass"
		check.Message = fmt.Sprintf("workflow.yaml exists at %s", ws.ConfigPath)
	} else {
		check.Status = "fail"
		check.Message = "workflow.yaml not found"
		check.Remediation = "Run 'ccflow run' to create a workflow"
	}

	return check
}

// checkSettingsJSON verifies settings.json is valid
func (v *Validator) checkSettingsJSON(ws *workspace.Workspace) Check {
	check := Check{Name: "Settings JSON"}

	settingsPath := filepath.Join(ws.GetHubPath(), "settings.json")

	if !util.FileExists(settingsPath) {
		check.Status = "warn"
		check.Message = "settings.json not found"
		check.Remediation = "Run 'ccflow run' to generate settings.json"
		return check
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("cannot read settings.json: %v", err)
		return check
	}

	var settings interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		check.Status = "fail"
		check.Message = fmt.Sprintf("invalid JSON in settings.json: %v", err)
		check.Remediation = "Fix JSON syntax errors in settings.json"
		return check
	}

	check.Status = "pass"
	check.Message = "settings.json is valid JSON"
	return check
}

// checkHookScripts verifies hook scripts
func (v *Validator) checkHookScripts(ws *workspace.Workspace) []Check {
	var checks []Check

	hooks := v.checkHooks(ws)
	for _, hook := range hooks {
		check := Check{Name: fmt.Sprintf("Hook: %s", hook.Name)}

		if !hook.Exists {
			check.Status = "fail"
			check.Message = fmt.Sprintf("script not found: %s", hook.ScriptPath)
			check.Remediation = fmt.Sprintf("Create the hook script or run 'ccflow add-hook %s'", hook.Name)
		} else if !hook.Executable {
			check.Status = "warn"
			check.Message = fmt.Sprintf("script not executable: %s", hook.ScriptPath)
			check.Remediation = fmt.Sprintf("Run 'chmod +x %s'", hook.ScriptPath)
		} else {
			check.Status = "pass"
			check.Message = fmt.Sprintf("script exists and is executable (events: %v)", hook.Events)
		}

		checks = append(checks, check)
	}

	return checks
}

// checkSymlinks verifies symlinks in multi-repo setup
func (v *Validator) checkSymlinks(ws *workspace.Workspace) []Check {
	var checks []Check

	for _, repo := range ws.Config.Repos {
		repoPath := filepath.Join(ws.Root, repo.Path)
		check := Check{Name: fmt.Sprintf("Symlink: %s", repo.Name)}

		if !util.DirExists(repoPath) {
			check.Status = "warn"
			check.Message = fmt.Sprintf("repository directory not found: %s", repoPath)
			checks = append(checks, check)
			continue
		}

		ok, msg := v.installer.VerifyInstallation(repoPath, ws.GetHubPath())
		if ok {
			check.Status = "pass"
			check.Message = msg
		} else {
			check.Status = "fail"
			check.Message = msg
			check.Remediation = fmt.Sprintf("Run 'ccflow run' to fix symlinks, or manually create: ln -s <hub>/.claude %s/.claude", repoPath)
		}

		checks = append(checks, check)
	}

	return checks
}

// checkDirectories verifies required directories exist
func (v *Validator) checkDirectories(ws *workspace.Workspace) Check {
	check := Check{Name: "Required directories"}

	missing := []string{}

	// Check hub .claude directory
	hubPath := ws.GetHubPath()
	if !util.DirExists(hubPath) {
		missing = append(missing, hubPath)
	}

	// Check subdirectories
	subdirs := []string{"agents", "commands", "hooks"}
	for _, subdir := range subdirs {
		path := filepath.Join(hubPath, subdir)
		if !util.DirExists(path) {
			missing = append(missing, path)
		}
	}

	// Check state directories
	if !util.DirExists(ws.GetStatePath()) {
		missing = append(missing, ws.GetStatePath())
	}
	if !util.DirExists(ws.GetDesignsPath()) {
		missing = append(missing, ws.GetDesignsPath())
	}

	if len(missing) == 0 {
		check.Status = "pass"
		check.Message = "all required directories exist"
	} else {
		check.Status = "fail"
		check.Message = fmt.Sprintf("missing directories: %v", missing)
		check.Remediation = "Run 'ccflow run' to create missing directories"
	}

	return check
}
