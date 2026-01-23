package blueprint

import (
	"testing"
)

// TestAllTemplatesRender validates that all templates in all blueprints render without error.
// This test would catch bugs like undefined template variables (e.g., {{.Version}} instead of correct field).
func TestAllTemplatesRender(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create template data with all fields populated
	data := &TemplateData{
		OrgName:         "test-org",
		WorkflowName:    "test-workflow",
		DocsRoot:        "docs/workflow",
		DocsStateDir:    "docs/workflow/state",
		DocsDesignDir:   "docs/workflow/designs",
		TrackerProvider: "linear",
		VCSProvider:     "github",
		HooksEnabled:    true,
		GatesEnabled:    true,
		Repos: []DefaultRepo{
			{Name: "test-repo", Kind: "go"},
		},
		AllRepos: []RepoInfo{
			{Name: "test-repo", Path: "test-repo", Kind: "go", CanWrite: true},
		},
		WriteRepos: []RepoInfo{
			{Name: "test-repo", Path: "test-repo", Kind: "go", CanWrite: true},
		},
		ReadRepos: []RepoInfo{},
	}

	blueprints := mgr.List()
	if len(blueprints) == 0 {
		t.Fatal("No blueprints found")
	}

	for _, bp := range blueprints {
		t.Run(bp.ID, func(t *testing.T) {
			// Test all default agents render without error
			for _, agentName := range bp.Agents.Defaults {
				content, err := mgr.GetAgentContent(bp.ID, agentName, data)
				if err != nil {
					t.Errorf("Agent %s failed to render: %v", agentName, err)
				}
				if len(content) == 0 {
					t.Errorf("Agent %s rendered empty content", agentName)
				}
			}

			// Test all default commands render without error
			for _, cmdName := range bp.Commands.Defaults {
				content, err := mgr.GetCommandContent(bp.ID, cmdName, data)
				if err != nil {
					t.Errorf("Command %s failed to render: %v", cmdName, err)
				}
				if len(content) == 0 {
					t.Errorf("Command %s rendered empty content", cmdName)
				}
			}

			// Test all default hooks render without error
			for _, hookName := range bp.Hooks.Defaults {
				content, err := mgr.GetHookContent(bp.ID, hookName, data)
				if err != nil {
					t.Errorf("Hook %s failed to render: %v", hookName, err)
				}
				if len(content) == 0 {
					t.Errorf("Hook %s rendered empty content", hookName)
				}
			}

			// Test settings.json renders without error
			settingsContent, err := mgr.GetAsset(bp.ID, ".claude/settings.json")
			if err != nil {
				t.Errorf("settings.json failed to load: %v", err)
			}
			if len(settingsContent) == 0 {
				t.Errorf("settings.json is empty")
			}
		})
	}
}

// TestTemplateDataFields ensures TemplateData has all expected fields set.
// This serves as documentation of what fields templates can use.
func TestTemplateDataFields(t *testing.T) {
	data := &TemplateData{
		OrgName:         "test-org",
		WorkflowName:    "test-workflow",
		DocsRoot:        "docs/workflow",
		DocsStateDir:    "docs/workflow/state",
		DocsDesignDir:   "docs/workflow/designs",
		TrackerProvider: "linear",
		VCSProvider:     "github",
		HooksEnabled:    true,
		GatesEnabled:    true,
	}

	// Verify fields are set correctly
	if data.OrgName != "test-org" {
		t.Errorf("OrgName not set correctly")
	}
	if data.WorkflowName != "test-workflow" {
		t.Errorf("WorkflowName not set correctly")
	}
	if data.DocsRoot != "docs/workflow" {
		t.Errorf("DocsRoot not set correctly")
	}
	if data.DocsStateDir != "docs/workflow/state" {
		t.Errorf("DocsStateDir not set correctly")
	}
	if data.DocsDesignDir != "docs/workflow/designs" {
		t.Errorf("DocsDesignDir not set correctly")
	}
	if data.TrackerProvider != "linear" {
		t.Errorf("TrackerProvider not set correctly")
	}
	if data.VCSProvider != "github" {
		t.Errorf("VCSProvider not set correctly")
	}
	if !data.HooksEnabled {
		t.Errorf("HooksEnabled not set correctly")
	}
	if !data.GatesEnabled {
		t.Errorf("GatesEnabled not set correctly")
	}
}

// TestBlueprintHasExpectedAssets verifies each blueprint has required assets.
func TestBlueprintHasExpectedAssets(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	for _, bp := range mgr.List() {
		t.Run(bp.ID, func(t *testing.T) {
			// Every blueprint should have at least one agent
			if len(bp.Agents.Defaults) == 0 {
				t.Errorf("Blueprint %s has no default agents", bp.ID)
			}

			// Every blueprint should have at least one command
			if len(bp.Commands.Defaults) == 0 {
				t.Errorf("Blueprint %s has no default commands", bp.ID)
			}

			// Verify each declared agent exists
			for _, agentName := range bp.Agents.Defaults {
				if !mgr.HasAgent(bp.ID, agentName) {
					t.Errorf("Blueprint %s declares agent %s but file is missing", bp.ID, agentName)
				}
			}

			// Verify each declared command exists
			for _, cmdName := range bp.Commands.Defaults {
				if !mgr.HasCommand(bp.ID, cmdName) {
					t.Errorf("Blueprint %s declares command %s but file is missing", bp.ID, cmdName)
				}
			}

			// Verify each declared hook exists
			for _, hookName := range bp.Hooks.Defaults {
				if !mgr.HasHook(bp.ID, hookName) {
					t.Errorf("Blueprint %s declares hook %s but file is missing", bp.ID, hookName)
				}
			}
		})
	}
}

// TestRenderAssetWithEmptyData ensures templates handle minimal data gracefully.
func TestRenderAssetWithEmptyData(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Minimal data - templates should not fail with empty optional fields
	data := &TemplateData{
		WorkflowName:    "minimal",
		DocsRoot:        "docs",
		DocsStateDir:    "docs/state",
		DocsDesignDir:   "docs/designs",
		TrackerProvider: "none",
		VCSProvider:     "none",
		HooksEnabled:    false,
		GatesEnabled:    false,
	}

	// Test with the first available blueprint
	blueprints := mgr.List()
	if len(blueprints) == 0 {
		t.Skip("No blueprints available")
	}

	bp := blueprints[0]
	for _, agentName := range bp.Agents.Defaults {
		_, err := mgr.GetAgentContent(bp.ID, agentName, data)
		if err != nil {
			t.Errorf("Agent %s failed with minimal data: %v", agentName, err)
		}
	}
}
