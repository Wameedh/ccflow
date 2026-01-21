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

// WorkflowConfig represents the workflow.yaml configuration
type WorkflowConfig struct {
	Version   int         `yaml:"version" json:"version"`
	Name      string      `yaml:"name" json:"name"`
	Topology  Topology    `yaml:"topology" json:"topology"`
	Blueprint string      `yaml:"blueprint" json:"blueprint"`
	Paths     PathsConfig `yaml:"paths" json:"paths"`
	State     StateConfig `yaml:"state" json:"state"`
	Repos     []RepoConfig `yaml:"repos" json:"repos"`
	Hooks     struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
	} `yaml:"hooks" json:"hooks"`
	Gates struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
	} `yaml:"gates" json:"gates"`
	MCP MCPConfig `yaml:"mcp" json:"mcp"`
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
