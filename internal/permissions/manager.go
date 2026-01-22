package permissions

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
)

// Manager handles agent permission operations
type Manager struct {
	workspace *workspace.Workspace
	bpManager *blueprint.Manager
}

// NewManager creates a new permission manager
func NewManager(ws *workspace.Workspace, bpManager *blueprint.Manager) *Manager {
	return &Manager{
		workspace: ws,
		bpManager: bpManager,
	}
}

// List returns all agent permissions in the workflow
func (m *Manager) List() map[string]config.AgentPermission {
	if m.workspace.Config.AgentPermissions == nil {
		return make(map[string]config.AgentPermission)
	}
	return m.workspace.Config.AgentPermissions
}

// Get returns permissions for a specific agent
func (m *Manager) Get(agentName string) (*config.AgentPermission, error) {
	if m.workspace.Config.AgentPermissions == nil {
		return nil, fmt.Errorf("no permissions configured for agent: %s", agentName)
	}

	perm, ok := m.workspace.Config.AgentPermissions[agentName]
	if !ok {
		return nil, fmt.Errorf("no permissions configured for agent: %s", agentName)
	}

	return &perm, nil
}

// Set sets permissions for an agent (replaces existing)
func (m *Manager) Set(agentName string, perm config.AgentPermission) error {
	if err := m.validateRepoNames(perm.Write); err != nil {
		return fmt.Errorf("invalid write repos: %w", err)
	}
	if err := m.validateRepoNames(perm.Read); err != nil {
		return fmt.Errorf("invalid read repos: %w", err)
	}

	if m.workspace.Config.AgentPermissions == nil {
		m.workspace.Config.AgentPermissions = make(map[string]config.AgentPermission)
	}

	m.workspace.Config.AgentPermissions[agentName] = perm
	return nil
}

// Grant adds additional access for an agent
func (m *Manager) Grant(agentName, accessType, repoName string) error {
	if err := m.validateRepoName(repoName); err != nil {
		return err
	}

	if m.workspace.Config.AgentPermissions == nil {
		m.workspace.Config.AgentPermissions = make(map[string]config.AgentPermission)
	}

	perm := m.workspace.Config.AgentPermissions[agentName]

	switch accessType {
	case "write":
		if !slices.Contains(perm.Write, repoName) {
			perm.Write = append(perm.Write, repoName)
		}
	case "read":
		if !slices.Contains(perm.Read, repoName) {
			perm.Read = append(perm.Read, repoName)
		}
	default:
		return fmt.Errorf("invalid access type: %s (must be 'write' or 'read')", accessType)
	}

	m.workspace.Config.AgentPermissions[agentName] = perm
	return nil
}

// Revoke removes access for an agent
func (m *Manager) Revoke(agentName, accessType, repoName string) error {
	if m.workspace.Config.AgentPermissions == nil {
		return fmt.Errorf("no permissions configured for agent: %s", agentName)
	}

	perm, ok := m.workspace.Config.AgentPermissions[agentName]
	if !ok {
		return fmt.Errorf("no permissions configured for agent: %s", agentName)
	}

	switch accessType {
	case "write":
		perm.Write = removeFromSlice(perm.Write, repoName)
	case "read":
		perm.Read = removeFromSlice(perm.Read, repoName)
	default:
		return fmt.Errorf("invalid access type: %s (must be 'write' or 'read')", accessType)
	}

	m.workspace.Config.AgentPermissions[agentName] = perm
	return nil
}

// Save writes the updated configuration to workflow.yaml
func (m *Manager) Save() error {
	return workspace.SaveConfig(m.workspace.ConfigPath, m.workspace.Config)
}

// RegenerateAgent regenerates a single agent's markdown file with updated permissions
func (m *Manager) RegenerateAgent(agentName string) error {
	bp, err := m.bpManager.Get(m.workspace.Config.Blueprint)
	if err != nil {
		return fmt.Errorf("failed to get blueprint: %w", err)
	}

	// Check if this is a built-in agent
	if !m.bpManager.HasAgent(bp.ID, agentName) {
		return fmt.Errorf("agent %s is not a built-in template", agentName)
	}

	// Build template data with permissions
	data := m.buildAgentTemplateData(agentName)

	// Render the agent template
	content, err := m.bpManager.GetAgentContent(bp.ID, agentName, data)
	if err != nil {
		return fmt.Errorf("failed to render agent template: %w", err)
	}

	// Write to the agent file
	agentPath := filepath.Join(m.workspace.GetHubPath(), "agents", agentName+".md")
	if err := util.SafeWriteFile(agentPath, content, true); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

// GetRepoNames returns all repository names in the workflow
func (m *Manager) GetRepoNames() []string {
	var names []string
	for _, repo := range m.workspace.Config.Repos {
		names = append(names, repo.Name)
	}
	return names
}

// GetAgentNames returns all agent names from the blueprint
func (m *Manager) GetAgentNames() ([]string, error) {
	bp, err := m.bpManager.Get(m.workspace.Config.Blueprint)
	if err != nil {
		return nil, err
	}
	return bp.Agents.Defaults, nil
}

// buildAgentTemplateData creates TemplateData with permission-aware repo lists
func (m *Manager) buildAgentTemplateData(agentName string) *blueprint.TemplateData {
	cfg := m.workspace.Config

	data := &blueprint.TemplateData{
		WorkflowName:    cfg.Name,
		DocsRoot:        cfg.State.Root,
		DocsStateDir:    cfg.State.StateDir,
		DocsDesignDir:   cfg.State.DesignsDir,
		TrackerProvider: string(cfg.MCP.Tracker),
		VCSProvider:     string(cfg.MCP.VCS),
		HooksEnabled:    cfg.Hooks.Enabled,
		GatesEnabled:    cfg.Gates.Enabled,
	}

	// Build permission-aware repo lists
	var allRepos []blueprint.RepoInfo
	var writeRepos []blueprint.RepoInfo
	var readRepos []blueprint.RepoInfo

	perm, hasPerm := cfg.AgentPermissions[agentName]

	for _, repo := range cfg.Repos {
		repoInfo := blueprint.RepoInfo{
			Name: repo.Name,
			Path: repo.Path,
			Kind: string(repo.Kind),
		}

		// If no permissions configured, default to full access
		if !hasPerm || len(perm.Write) == 0 && len(perm.Read) == 0 {
			repoInfo.CanWrite = true
			allRepos = append(allRepos, repoInfo)
			writeRepos = append(writeRepos, repoInfo)
		} else {
			// Check if agent has write access
			if slices.Contains(perm.Write, repo.Name) {
				repoInfo.CanWrite = true
				writeRepos = append(writeRepos, repoInfo)
				allRepos = append(allRepos, repoInfo)
			} else if slices.Contains(perm.Read, repo.Name) {
				repoInfo.CanWrite = false
				readRepos = append(readRepos, repoInfo)
				allRepos = append(allRepos, repoInfo)
			}
		}
	}

	data.AllRepos = allRepos
	data.WriteRepos = writeRepos
	data.ReadRepos = readRepos

	return data
}

// validateRepoNames validates that all repo names exist in the workflow
func (m *Manager) validateRepoNames(names []string) error {
	validNames := m.GetRepoNames()
	for _, name := range names {
		if !slices.Contains(validNames, name) {
			return fmt.Errorf("unknown repository: %s", name)
		}
	}
	return nil
}

// validateRepoName validates a single repo name
func (m *Manager) validateRepoName(name string) error {
	validNames := m.GetRepoNames()
	if !slices.Contains(validNames, name) {
		return fmt.Errorf("unknown repository: %s (available: %v)", name, validNames)
	}
	return nil
}

// removeFromSlice removes an element from a string slice
func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
