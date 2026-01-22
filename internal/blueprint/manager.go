package blueprint

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed all:web-dev all:ios-dev all:go-cli-dev all:python-data-dev all:devops-infra
var blueprintsFS embed.FS

// Manager provides access to embedded blueprints
type Manager struct {
	blueprints map[string]*Blueprint
}

// NewManager creates a new blueprint manager
func NewManager() (*Manager, error) {
	m := &Manager{
		blueprints: make(map[string]*Blueprint),
	}

	// Load all embedded blueprints
	entries, err := blueprintsFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprints directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		bp, err := m.loadBlueprint(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to load blueprint %s: %w", entry.Name(), err)
		}
		m.blueprints[bp.ID] = bp
	}

	return m, nil
}

// loadBlueprint loads a blueprint from the embedded filesystem
func (m *Manager) loadBlueprint(name string) (*Blueprint, error) {
	manifestPath := filepath.Join(name, "blueprint.yaml")
	data, err := blueprintsFS.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint.yaml: %w", err)
	}

	var bp Blueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint.yaml: %w", err)
	}

	return &bp, nil
}

// List returns all available blueprints
func (m *Manager) List() []*Blueprint {
	var result []*Blueprint
	for _, bp := range m.blueprints {
		result = append(result, bp)
	}
	return result
}

// Get returns a blueprint by ID
func (m *Manager) Get(id string) (*Blueprint, error) {
	bp, ok := m.blueprints[id]
	if !ok {
		return nil, fmt.Errorf("blueprint not found: %s", id)
	}
	return bp, nil
}

// GetAsset retrieves an asset file from a blueprint
func (m *Manager) GetAsset(blueprintID, assetPath string) ([]byte, error) {
	fullPath := filepath.Join(blueprintID, "assets", assetPath)
	return blueprintsFS.ReadFile(fullPath)
}

// GetTemplate retrieves a template file from a blueprint
func (m *Manager) GetTemplate(blueprintID, templatePath string) ([]byte, error) {
	fullPath := filepath.Join(blueprintID, "templates", templatePath)
	return blueprintsFS.ReadFile(fullPath)
}

// RenderAsset renders an asset template with the given data
func (m *Manager) RenderAsset(blueprintID, assetPath string, data *TemplateData) ([]byte, error) {
	content, err := m.GetAsset(blueprintID, assetPath)
	if err != nil {
		return nil, err
	}

	// Check if file needs template rendering (ends in .tmpl)
	if strings.HasSuffix(assetPath, ".tmpl") {
		return m.renderTemplate(string(content), data)
	}

	// For non-template files, still do basic variable substitution
	return m.renderTemplate(string(content), data)
}

// renderTemplate renders a Go template string with the given data
func (m *Manager) renderTemplate(content string, data *TemplateData) ([]byte, error) {
	tmpl, err := template.New("asset").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return []byte(buf.String()), nil
}

// ListAssets lists all assets for a blueprint in a given subdirectory
func (m *Manager) ListAssets(blueprintID, subdir string) ([]string, error) {
	basePath := filepath.Join(blueprintID, "assets", subdir)
	var assets []string

	err := fs.WalkDir(blueprintsFS, basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Get relative path from assets directory
		relPath, _ := filepath.Rel(filepath.Join(blueprintID, "assets"), path)
		assets = append(assets, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return assets, nil
}

// HasAgent checks if a blueprint has a built-in agent template
func (m *Manager) HasAgent(blueprintID, agentName string) bool {
	path := filepath.Join(blueprintID, "assets", ".claude", "agents", agentName+".md")
	_, err := blueprintsFS.ReadFile(path)
	return err == nil
}

// HasCommand checks if a blueprint has a built-in command template
func (m *Manager) HasCommand(blueprintID, commandName string) bool {
	path := filepath.Join(blueprintID, "assets", ".claude", "commands", commandName+".md")
	_, err := blueprintsFS.ReadFile(path)
	return err == nil
}

// HasHook checks if a blueprint has a built-in hook template
func (m *Manager) HasHook(blueprintID, hookName string) bool {
	path := filepath.Join(blueprintID, "assets", ".claude", "hooks", hookName+".sh")
	_, err := blueprintsFS.ReadFile(path)
	return err == nil
}

// GetAgentContent returns the content of a built-in agent
func (m *Manager) GetAgentContent(blueprintID, agentName string, data *TemplateData) ([]byte, error) {
	return m.RenderAsset(blueprintID, filepath.Join(".claude", "agents", agentName+".md"), data)
}

// GetCommandContent returns the content of a built-in command
func (m *Manager) GetCommandContent(blueprintID, commandName string, data *TemplateData) ([]byte, error) {
	return m.RenderAsset(blueprintID, filepath.Join(".claude", "commands", commandName+".md"), data)
}

// GetHookContent returns the content of a built-in hook
func (m *Manager) GetHookContent(blueprintID, hookName string, data *TemplateData) ([]byte, error) {
	return m.RenderAsset(blueprintID, filepath.Join(".claude", "hooks", hookName+".sh"), data)
}
