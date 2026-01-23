package config

// Topology represents the workflow topology type
type Topology string

const (
	TopologyMultiRepo  Topology = "multi-repo"
	TopologySingleRepo Topology = "single-repo"
)

// RepoKind represents the type of repository
type RepoKind string

const (
	RepoKindNode      RepoKind = "node"
	RepoKindJava      RepoKind = "java"
	RepoKindGo        RepoKind = "go"
	RepoKindPython    RepoKind = "python"
	RepoKindSwift     RepoKind = "swift"
	RepoKindTerraform RepoKind = "terraform"
	RepoKindDocs      RepoKind = "docs"
	RepoKindUnknown   RepoKind = "unknown"
)

// VCSProvider represents version control system provider
type VCSProvider string

const (
	VCSGitHub VCSProvider = "github"
	VCSGitLab VCSProvider = "gitlab"
	VCSNone   VCSProvider = "none"
)

// TrackerProvider represents issue tracker provider
type TrackerProvider string

const (
	TrackerLinear TrackerProvider = "linear"
	TrackerJira   TrackerProvider = "jira"
	TrackerNone   TrackerProvider = "none"
)

// DeployProvider represents deployment provider
type DeployProvider string

const (
	DeployArgoCD DeployProvider = "argocd"
	DeployNone   DeployProvider = "none"
)

// TransitionMode defines how phase transitions behave
type TransitionMode string

const (
	TransitionAuto   TransitionMode = "auto"   // Immediately invoke next command
	TransitionPrompt TransitionMode = "prompt" // Ask user before proceeding
	TransitionManual TransitionMode = "manual" // Just print suggestion
)

// TransitionConfig defines a single phase transition
type TransitionConfig struct {
	Mode TransitionMode `yaml:"mode" json:"mode"`
}

// TransitionsConfig defines all workflow transitions
type TransitionsConfig struct {
	IdeaToDesign      TransitionConfig `yaml:"idea_to_design" json:"idea_to_design"`
	DesignToImplement TransitionConfig `yaml:"design_to_implement" json:"design_to_implement"`
	ImplementToReview TransitionConfig `yaml:"implement_to_review" json:"implement_to_review"`
	ReviewToRelease   TransitionConfig `yaml:"review_to_release" json:"review_to_release"`
}

// ParallelGroup defines a group of agents that run in parallel
type ParallelGroup struct {
	Name   string   `yaml:"name" json:"name"`
	Agents []string `yaml:"agents" json:"agents"`
	Sync   string   `yaml:"sync" json:"sync"` // "all" or "any"
}

// ParallelConfig defines parallel execution settings
type ParallelConfig struct {
	Enabled  bool            `yaml:"enabled" json:"enabled"`
	SyncGate string          `yaml:"sync_gate" json:"sync_gate"`
	Groups   []ParallelGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
}

// RepoConfig represents a repository in the workflow
type RepoConfig struct {
	Name string   `yaml:"name" json:"name"`
	Path string   `yaml:"path" json:"path"`
	Kind RepoKind `yaml:"kind" json:"kind"`
}

// PathsConfig represents the paths configuration
type PathsConfig struct {
	Hub  string `yaml:"hub" json:"hub"`
	Docs string `yaml:"docs" json:"docs"`
}

// StateConfig represents the state directories configuration
type StateConfig struct {
	Root       string `yaml:"root" json:"root"`
	StateDir   string `yaml:"state_dir" json:"state_dir"`
	DesignsDir string `yaml:"designs_dir" json:"designs_dir"`
}

// MCPConfig represents MCP integration preferences (guidance only)
type MCPConfig struct {
	VCS     VCSProvider     `yaml:"vcs" json:"vcs"`
	Tracker TrackerProvider `yaml:"tracker" json:"tracker"`
	Deploy  DeployProvider  `yaml:"deploy" json:"deploy"`
}

// AgentPermission defines repository access permissions for an agent
type AgentPermission struct {
	Write []string `yaml:"write,omitempty" json:"write,omitempty"`
	Read  []string `yaml:"read,omitempty" json:"read,omitempty"`
}

// WorkflowConfig represents the workflow.yaml configuration
type WorkflowConfig struct {
	Version   int          `yaml:"version" json:"version"`
	Name      string       `yaml:"name" json:"name"`
	Topology  Topology     `yaml:"topology" json:"topology"`
	Blueprint string       `yaml:"blueprint" json:"blueprint"`
	Paths     PathsConfig  `yaml:"paths" json:"paths"`
	State     StateConfig  `yaml:"state" json:"state"`
	Repos     []RepoConfig `yaml:"repos" json:"repos"`
	Hooks     struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
	} `yaml:"hooks" json:"hooks"`
	Gates struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
	} `yaml:"gates" json:"gates"`
	MCP              MCPConfig                  `yaml:"mcp" json:"mcp"`
	AgentPermissions map[string]AgentPermission `yaml:"agent_permissions,omitempty" json:"agent_permissions,omitempty"`
	Transitions      TransitionsConfig          `yaml:"transitions" json:"transitions"`
	Parallel         ParallelConfig             `yaml:"parallel" json:"parallel"`
}

// NewDefaultWorkflowConfig creates a new workflow config with sensible defaults
func NewDefaultWorkflowConfig(name string) *WorkflowConfig {
	return &WorkflowConfig{
		Version:   1,
		Name:      name,
		Topology:  TopologyMultiRepo,
		Blueprint: "web-dev",
		Paths: PathsConfig{
			Hub:  "workflow-hub",
			Docs: "docs",
		},
		State: StateConfig{
			Root:       "docs/workflow",
			StateDir:   "docs/workflow/state",
			DesignsDir: "docs/workflow/designs",
		},
		Repos: []RepoConfig{},
		Hooks: struct {
			Enabled bool `yaml:"enabled" json:"enabled"`
		}{Enabled: true},
		Gates: struct {
			Enabled bool `yaml:"enabled" json:"enabled"`
		}{Enabled: true},
		MCP: MCPConfig{
			VCS:     VCSNone,
			Tracker: TrackerNone,
			Deploy:  DeployNone,
		},
		AgentPermissions: make(map[string]AgentPermission),
		Transitions: TransitionsConfig{
			IdeaToDesign:      TransitionConfig{Mode: TransitionPrompt},
			DesignToImplement: TransitionConfig{Mode: TransitionPrompt},
			ImplementToReview: TransitionConfig{Mode: TransitionPrompt},
			ReviewToRelease:   TransitionConfig{Mode: TransitionPrompt},
		},
		Parallel: ParallelConfig{
			Enabled:  false,
			SyncGate: "all",
			Groups:   nil,
		},
	}
}

// ManagedFileInfo tracks managed files for upgrade functionality
type ManagedFileInfo struct {
	TemplateID string `json:"template_id"`
	Hash       string `json:"hash"`
	Version    string `json:"version"`
}

// ManagedFilesManifest tracks all ccflow-managed files
type ManagedFilesManifest struct {
	Version int                        `json:"version"`
	Files   map[string]ManagedFileInfo `json:"files"`
}
