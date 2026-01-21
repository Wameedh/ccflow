package blueprint

// Blueprint represents a workflow blueprint configuration
type Blueprint struct {
	ID              string          `yaml:"id"`
	DisplayName     string          `yaml:"display_name"`
	Description     string          `yaml:"description"`
	DefaultTopology string          `yaml:"default_topology"`
	DefaultRepos    []DefaultRepo   `yaml:"default_repos"`
	Agents          AgentDefaults   `yaml:"agents"`
	Commands        CommandDefaults `yaml:"commands"`
	Hooks           HookDefaults    `yaml:"hooks"`
	HooksManifest   HooksManifest   `yaml:"hooks_manifest"`
	MCPSuggestions  MCPSuggestions  `yaml:"mcp_suggestions"`
}

// DefaultRepo represents a default repository in a blueprint
type DefaultRepo struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"`
}

// AgentDefaults defines default agents for a blueprint
type AgentDefaults struct {
	Defaults []string `yaml:"defaults"`
}

// CommandDefaults defines default commands for a blueprint
type CommandDefaults struct {
	Defaults []string `yaml:"defaults"`
}

// HookDefaults defines default hooks for a blueprint
type HookDefaults struct {
	Defaults []string `yaml:"defaults"`
}

// HooksManifest maps hook names to their settings.json registration info
type HooksManifest map[string]HookRegistration

// HookRegistration defines how a hook should be registered in settings.json
type HookRegistration struct {
	Script string      `yaml:"script"`
	Events []HookEvent `yaml:"events"`
}

// HookEvent defines a single hook event configuration
type HookEvent struct {
	Event    string   `yaml:"event"`
	Commands []string `yaml:"commands,omitempty"`
}

// MCPSuggestions defines MCP integration suggestions
type MCPSuggestions struct {
	VCS     []string `yaml:"vcs"`
	Tracker []string `yaml:"tracker"`
	Deploy  []string `yaml:"deploy"`
}

// TemplateData contains data for template rendering
type TemplateData struct {
	OrgName         string
	WorkflowName    string
	DocsRoot        string
	DocsStateDir    string
	DocsDesignDir   string
	TrackerProvider string
	VCSProvider     string
	Repos           []DefaultRepo
	HooksEnabled    bool
	GatesEnabled    bool
}
