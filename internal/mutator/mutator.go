package mutator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/util"
)

// Mutator handles adding agents, commands, and hooks to a workflow
type Mutator struct {
	bpManager *blueprint.Manager
}

// New creates a new Mutator
func New(bpManager *blueprint.Manager) *Mutator {
	return &Mutator{bpManager: bpManager}
}

// ContentSource represents where content comes from
type ContentSource int

const (
	// SourceTemplate uses built-in template
	SourceTemplate ContentSource = iota
	// SourceFile reads from a file
	SourceFile
	// SourceStdin reads from stdin
	SourceStdin
)

// AddOptions contains options for adding an artifact
type AddOptions struct {
	Name         string
	Source       ContentSource
	FilePath     string // For SourceFile
	Content      []byte // For SourceStdin (pre-read)
	Force        bool
	BlueprintID  string // For template lookups
	HubPath      string // Path to .claude directory
	TemplateData *blueprint.TemplateData
}

// AddAgent adds an agent to the workflow
func (m *Mutator) AddAgent(opts AddOptions) error {
	content, err := m.resolveContent(opts, "agent")
	if err != nil {
		return err
	}

	agentPath := filepath.Join(opts.HubPath, "agents", opts.Name+".md")
	return util.SafeWriteFile(agentPath, content, opts.Force)
}

// AddCommand adds a command to the workflow
func (m *Mutator) AddCommand(opts AddOptions) error {
	content, err := m.resolveContent(opts, "command")
	if err != nil {
		return err
	}

	cmdPath := filepath.Join(opts.HubPath, "commands", opts.Name+".md")
	return util.SafeWriteFile(cmdPath, content, opts.Force)
}

// AddHook adds a hook to the workflow and updates settings.json
func (m *Mutator) AddHook(opts AddOptions) error {
	content, err := m.resolveContent(opts, "hook")
	if err != nil {
		return err
	}

	// Write the hook script
	hookPath := filepath.Join(opts.HubPath, "hooks", opts.Name+".sh")
	if err := util.SafeWriteExecutable(hookPath, content, opts.Force); err != nil {
		return err
	}

	// Update settings.json to register the hook
	if err := m.registerHook(opts); err != nil {
		return fmt.Errorf("hook script written but failed to update settings.json: %w", err)
	}

	return nil
}

// GetTemplateContent returns the template content for an artifact without writing
func (m *Mutator) GetTemplateContent(blueprintID, artifactType, name string, data *blueprint.TemplateData) ([]byte, error) {
	switch artifactType {
	case "agent":
		return m.bpManager.GetAgentContent(blueprintID, name, data)
	case "command":
		return m.bpManager.GetCommandContent(blueprintID, name, data)
	case "hook":
		return m.bpManager.GetHookContent(blueprintID, name, data)
	default:
		return nil, fmt.Errorf("unknown artifact type: %s", artifactType)
	}
}

// resolveContent determines where to get content from based on options
func (m *Mutator) resolveContent(opts AddOptions, artifactType string) ([]byte, error) {
	switch opts.Source {
	case SourceStdin:
		if len(opts.Content) == 0 {
			return nil, fmt.Errorf("no content provided from stdin")
		}
		return opts.Content, nil

	case SourceFile:
		if opts.FilePath == "" {
			return nil, fmt.Errorf("no file path provided")
		}
		content, err := os.ReadFile(opts.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return content, nil

	case SourceTemplate:
		return m.GetTemplateContent(opts.BlueprintID, artifactType, opts.Name, opts.TemplateData)

	default:
		return nil, fmt.Errorf("unknown content source")
	}
}

// registerHook updates settings.json to include the hook
func (m *Mutator) registerHook(opts AddOptions) error {
	settingsPath := filepath.Join(opts.HubPath, "settings.json")

	// Read existing settings
	var settings map[string]interface{}
	if util.FileExists(settingsPath) {
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			return fmt.Errorf("failed to read settings.json: %w", err)
		}
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse settings.json: %w", err)
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Get hook registration info from blueprint
	bp, err := m.bpManager.Get(opts.BlueprintID)
	if err != nil {
		// If blueprint not found, use default registration
		return m.addDefaultHookRegistration(settings, settingsPath, opts.Name)
	}

	hookReg, ok := bp.HooksManifest[opts.Name]
	if !ok {
		// Hook not in manifest, use default registration
		return m.addDefaultHookRegistration(settings, settingsPath, opts.Name)
	}

	// Add hook events to settings
	return m.addHookRegistration(settings, settingsPath, hookReg)
}

// addHookRegistration adds a hook registration to settings
func (m *Mutator) addHookRegistration(settings map[string]interface{}, settingsPath string, hookReg blueprint.HookRegistration) error {
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
	}

	// Add each event from the hook registration
	for _, event := range hookReg.Events {
		// Build the hook entry in new format
		hookEntry := map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": "./" + hookReg.Script,
				},
			},
		}

		// Add matcher if there are tool filters
		if len(event.Commands) > 0 {
			hookEntry["matcher"] = map[string]interface{}{
				"tools": event.Commands,
			}
		}

		// Get or create the event array
		eventHooks, ok := hooks[event.Event].([]interface{})
		if !ok {
			eventHooks = []interface{}{}
		}

		// Check if this hook already exists
		if !m.hookExists(eventHooks, hookReg.Script) {
			eventHooks = append(eventHooks, hookEntry)
		}
		hooks[event.Event] = eventHooks
	}

	settings["hooks"] = hooks

	// Write updated settings
	return m.writeSettings(settingsPath, settings)
}

// addDefaultHookRegistration adds a default hook registration
func (m *Mutator) addDefaultHookRegistration(settings map[string]interface{}, settingsPath, hookName string) error {
	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
	}

	// Default: register as Stop event
	hookScript := "./hooks/" + hookName + ".sh"
	hookEntry := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": hookScript,
			},
		},
	}

	// Get or create the Stop event array
	stopHooks, ok := hooks["Stop"].([]interface{})
	if !ok {
		stopHooks = []interface{}{}
	}

	if !m.hookExists(stopHooks, hookScript) {
		stopHooks = append(stopHooks, hookEntry)
	}

	hooks["Stop"] = stopHooks
	settings["hooks"] = hooks
	return m.writeSettings(settingsPath, settings)
}

// hookExists checks if a hook with the same script already exists in an event's hook list
func (m *Mutator) hookExists(eventHooks []interface{}, script string) bool {
	for _, h := range eventHooks {
		hookEntry, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		// Check in the hooks array within this entry
		hooksArr, ok := hookEntry["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, hk := range hooksArr {
			hook, ok := hk.(map[string]interface{})
			if !ok {
				continue
			}
			if hook["command"] == script || hook["command"] == "./"+script {
				return true
			}
		}
	}
	return false
}

// writeSettings writes the settings map to file
func (m *Mutator) writeSettings(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// HasTemplate checks if a template exists for the given artifact
func (m *Mutator) HasTemplate(blueprintID, artifactType, name string) bool {
	switch artifactType {
	case "agent":
		return m.bpManager.HasAgent(blueprintID, name)
	case "command":
		return m.bpManager.HasCommand(blueprintID, name)
	case "hook":
		return m.bpManager.HasHook(blueprintID, name)
	default:
		return false
	}
}
