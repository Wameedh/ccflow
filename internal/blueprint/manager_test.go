package blueprint

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	blueprints := mgr.List()
	if len(blueprints) == 0 {
		t.Error("Expected at least one blueprint")
	}

	// Check web-dev blueprint
	webDev, err := mgr.Get("web-dev")
	if err != nil {
		t.Fatalf("Failed to get web-dev blueprint: %v", err)
	}

	if webDev.DisplayName != "Web Development" {
		t.Errorf("Expected 'Web Development', got %s", webDev.DisplayName)
	}
}

func TestHasAgent(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Test existing agent
	if !mgr.HasAgent("web-dev", "devops-agent") {
		t.Error("Expected HasAgent to return true for devops-agent in web-dev")
	}

	// Test non-existing agent
	if mgr.HasAgent("web-dev", "nonexistent-agent") {
		t.Error("Expected HasAgent to return false for nonexistent-agent")
	}
}

func TestGetAgentContent(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	data := &TemplateData{
		WorkflowName:  "test-workflow",
		DocsStateDir:  "docs/workflow/state",
		DocsDesignDir: "docs/workflow/designs",
	}

	content, err := mgr.GetAgentContent("web-dev", "devops-agent", data)
	if err != nil {
		t.Fatalf("GetAgentContent failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty content")
	}

	// Check that template was rendered
	if string(content) == "" {
		t.Error("Template rendering produced empty content")
	}
}

func TestListBlueprints(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	blueprints := mgr.List()

	// Should have both web-dev and ios-dev
	foundWebDev := false
	foundIosDev := false

	for _, bp := range blueprints {
		if bp.ID == "web-dev" {
			foundWebDev = true
		}
		if bp.ID == "ios-dev" {
			foundIosDev = true
		}
	}

	if !foundWebDev {
		t.Error("Expected to find web-dev blueprint")
	}
	if !foundIosDev {
		t.Error("Expected to find ios-dev blueprint")
	}
}
