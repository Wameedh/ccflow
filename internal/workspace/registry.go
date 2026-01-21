package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/wameedh/ccflow/internal/util"
)

const (
	// RegistryDir is the directory for ccflow global config
	RegistryDir = ".ccflow"
	// RegistryFile is the filename for the workflow registry
	RegistryFile = "registry.json"
)

// RegistryEntry represents a workflow entry in the registry
type RegistryEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Blueprint  string    `json:"blueprint"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// Registry represents the global workflow registry
type Registry struct {
	Version   int             `json:"version"`
	Workflows []RegistryEntry `json:"workflows"`
}

// GetRegistryPath returns the path to the global registry file
func GetRegistryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, RegistryDir, RegistryFile), nil
}

// LoadRegistry loads the global registry
func LoadRegistry() (*Registry, error) {
	path, err := GetRegistryPath()
	if err != nil {
		return nil, err
	}

	if !util.FileExists(path) {
		// Return empty registry if file doesn't exist
		return &Registry{Version: 1, Workflows: []RegistryEntry{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}

	return &reg, nil
}

// SaveRegistry saves the global registry
func SaveRegistry(reg *Registry) error {
	path, err := GetRegistryPath()
	if err != nil {
		return err
	}

	if err := util.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddOrUpdateWorkflow adds or updates a workflow in the registry
func (r *Registry) AddOrUpdateWorkflow(entry RegistryEntry) {
	for i, existing := range r.Workflows {
		if existing.Path == entry.Path {
			// Update existing entry
			r.Workflows[i] = entry
			return
		}
	}
	// Add new entry
	r.Workflows = append(r.Workflows, entry)
}

// RemoveWorkflow removes a workflow from the registry by path
func (r *Registry) RemoveWorkflow(path string) bool {
	for i, existing := range r.Workflows {
		if existing.Path == path {
			r.Workflows = append(r.Workflows[:i], r.Workflows[i+1:]...)
			return true
		}
	}
	return false
}

// FindByName finds a workflow by name
func (r *Registry) FindByName(name string) *RegistryEntry {
	for _, entry := range r.Workflows {
		if entry.Name == name {
			return &entry
		}
	}
	return nil
}
