package ccflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/generator"
	"github.com/Wameedh/ccflow/internal/installer"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var runCmd = &cobra.Command{
	Use:     "run [blueprint]",
	Aliases: []string{"init"},
	Short:   "Create a new workflow",
	Long: `Create a new Claude Code workflow using an interactive wizard.

If a blueprint is specified, it will be used as the starting point.
Otherwise, you will be prompted to select one.

Examples:
  ccflow run                  # Interactive blueprint selection
  ccflow run web-dev          # Use web-dev blueprint
  ccflow run ios-dev          # Use ios-dev blueprint`,
	Args: cobra.MaximumNArgs(1),
	Run:  runWorkflow,
}

func runWorkflow(cmd *cobra.Command, args []string) {
	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	// Determine blueprint
	var blueprintID string
	if len(args) > 0 {
		blueprintID = args[0]
		if _, getErr := bpManager.Get(blueprintID); getErr != nil {
			exitWithError("unknown blueprint: %s\nRun 'ccflow list-blueprints' to see available blueprints", blueprintID)
		}
	} else {
		blueprintID = promptBlueprintSelection(bpManager)
	}

	bp, _ := bpManager.Get(blueprintID)

	// Run wizard prompts
	answers := runWizard(bp)

	// Generate workflow
	gen := generator.New(bpManager)
	opts := generator.GenerateOptions{
		WorkspacePath: answers.workspacePath,
		WorkflowName:  answers.workflowName,
		Blueprint:     blueprintID,
		Topology:      answers.topology,
		Repos:         answers.repos,
		HooksEnabled:  answers.hooksEnabled,
		GatesEnabled:  answers.gatesEnabled,
		VCS:           answers.vcs,
		Tracker:       answers.tracker,
		Force:         forceFlag,
	}

	// Pre-flight check: warn about existing files
	if !forceFlag {
		existingFiles, checkErr := gen.CheckExistingFiles(opts)
		if checkErr != nil {
			exitWithError("failed to check existing files: %v", checkErr)
		}

		if len(existingFiles) > 0 {
			fmt.Printf("\n%d file(s) already exist:\n", len(existingFiles))
			for _, f := range existingFiles {
				// Show relative path if possible
				relPath, relErr := filepath.Rel(answers.workspacePath, f)
				if relErr != nil {
					relPath = f
				}
				fmt.Printf("  - %s\n", relPath)
			}

			var overwrite bool
			promptErr := survey.AskOne(&survey.Confirm{
				Message: "Overwrite existing files?",
				Default: false,
			}, &overwrite)

			if promptErr != nil {
				// If prompt fails (non-interactive), require --force flag
				exitWithError("files already exist. Use --force to overwrite, or run interactively to confirm")
			}

			if !overwrite {
				fmt.Println("\nAborted. No files were modified.")
				return
			}
			fmt.Println("\nOverwriting existing files...")
			opts.Force = true
		}
	}

	cfg, err := gen.Generate(opts)
	if err != nil {
		exitWithError("failed to generate workflow: %v", err)
	}

	// Install .claude to repos (multi-repo only)
	if answers.topology == config.TopologyMultiRepo && len(answers.repos) > 0 {
		inst := installer.New()
		hubPath := filepath.Join(answers.workspacePath, cfg.Paths.Hub, ".claude")

		results := inst.Install(installer.InstallOptions{
			HubPath:       hubPath,
			Repos:         answers.repos,
			WorkspacePath: answers.workspacePath,
			Mode:          installer.InstallModeSymlink,
			Force:         opts.Force,
		})

		fmt.Println("\nInstalling .claude to repositories:")
		for _, r := range results {
			if r.Success {
				if r.Skipped {
					printInfo("%s: %s", r.RepoName, r.Message)
				} else {
					printSuccess("%s: %s", r.RepoName, r.Message)
				}
			} else {
				printWarning("%s: %v", r.RepoName, r.Error)
			}
		}
	}

	// Register workflow in global registry
	registerWorkflow(answers.workspacePath, cfg)

	// Print success and next steps
	printNextSteps(answers, cfg, bp)
}

type wizardAnswers struct {
	workflowName  string
	workspacePath string
	topology      config.Topology
	repos         []config.RepoConfig
	hooksEnabled  bool
	gatesEnabled  bool
	vcs           config.VCSProvider
	tracker       config.TrackerProvider
}

func runWizard(bp *blueprint.Blueprint) wizardAnswers {
	answers := wizardAnswers{
		hooksEnabled: true,
		gatesEnabled: true,
	}

	// Get current directory
	cwd, _ := os.Getwd()

	// Workflow name
	workflowNamePrompt := &survey.Input{
		Message: "Workflow name:",
		Default: filepath.Base(cwd),
	}
	survey.AskOne(workflowNamePrompt, &answers.workflowName)

	// Workspace path
	workspacePathPrompt := &survey.Input{
		Message: "Workspace path:",
		Default: cwd,
	}
	survey.AskOne(workspacePathPrompt, &answers.workspacePath)
	answers.workspacePath, _ = filepath.Abs(answers.workspacePath)

	// Mode
	var mode string
	modePrompt := &survey.Select{
		Message: "Setup mode:",
		Options: []string{"Use existing repositories", "Generate new repositories"},
		Default: "Use existing repositories",
	}
	survey.AskOne(modePrompt, &mode)
	useExisting := mode == "Use existing repositories"

	// Topology
	var topologyStr string
	topologyPrompt := &survey.Select{
		Message: "Topology:",
		Options: []string{"Multi-repo (recommended)", "Single-repo"},
		Default: "Multi-repo (recommended)",
	}
	survey.AskOne(topologyPrompt, &topologyStr)
	if strings.Contains(topologyStr, "Multi") {
		answers.topology = config.TopologyMultiRepo
	} else {
		answers.topology = config.TopologySingleRepo
	}

	// Repos
	if answers.topology == config.TopologyMultiRepo {
		if useExisting {
			answers.repos = selectExistingRepos(answers.workspacePath)
		} else {
			answers.repos = configureNewRepos(bp)
		}
	}

	// Hooks
	var enableHooks bool
	survey.AskOne(&survey.Confirm{
		Message: "Enable hooks (formatting, validation)?",
		Default: true,
	}, &enableHooks)
	answers.hooksEnabled = enableHooks

	// Gates
	var enableGates bool
	survey.AskOne(&survey.Confirm{
		Message: "Enable quality gates?",
		Default: true,
	}, &enableGates)
	answers.gatesEnabled = enableGates

	// VCS (for MCP guidance)
	var vcsStr string
	survey.AskOne(&survey.Select{
		Message: "Version control system (for MCP guidance):",
		Options: []string{"GitHub", "GitLab", "None"},
		Default: "GitHub",
	}, &vcsStr)
	switch vcsStr {
	case "GitHub":
		answers.vcs = config.VCSGitHub
	case "GitLab":
		answers.vcs = config.VCSGitLab
	default:
		answers.vcs = config.VCSNone
	}

	// Tracker (for MCP guidance)
	var trackerStr string
	survey.AskOne(&survey.Select{
		Message: "Issue tracker (for MCP guidance):",
		Options: []string{"Linear", "Jira", "None"},
		Default: "None",
	}, &trackerStr)
	switch trackerStr {
	case "Linear":
		answers.tracker = config.TrackerLinear
	case "Jira":
		answers.tracker = config.TrackerJira
	default:
		answers.tracker = config.TrackerNone
	}

	return answers
}

func promptBlueprintSelection(bpManager *blueprint.Manager) string {
	blueprints := bpManager.List()
	options := make([]string, len(blueprints))
	for i, bp := range blueprints {
		options[i] = fmt.Sprintf("%s - %s", bp.ID, bp.Description)
	}

	var selection string
	survey.AskOne(&survey.Select{
		Message: "Select a blueprint:",
		Options: options,
	}, &selection)

	// Extract ID from selection
	return strings.Split(selection, " - ")[0]
}

func selectExistingRepos(workspacePath string) []config.RepoConfig {
	var repos []config.RepoConfig

	// Find git repos in workspace
	gitRepos, err := util.FindGitRepos(workspacePath)
	if err != nil || len(gitRepos) == 0 {
		fmt.Println("No git repositories found. You can add repos to workflow.yaml later.")
		return repos
	}

	// Build options
	options := make([]string, len(gitRepos))
	for i, repo := range gitRepos {
		relPath, _ := filepath.Rel(workspacePath, repo)
		options[i] = relPath
	}

	var selected []string
	survey.AskOne(&survey.MultiSelect{
		Message: "Select repositories to include:",
		Options: options,
	}, &selected)

	for _, sel := range selected {
		fullPath := filepath.Join(workspacePath, sel)
		kind := util.DetectRepoKind(fullPath)
		repos = append(repos, config.RepoConfig{
			Name: filepath.Base(sel),
			Path: sel,
			Kind: kind,
		})
	}

	return repos
}

func configureNewRepos(bp *blueprint.Blueprint) []config.RepoConfig {
	// Start with blueprint defaults
	repos := make([]config.RepoConfig, len(bp.DefaultRepos))
	for i, dr := range bp.DefaultRepos {
		repos[i] = config.RepoConfig{
			Name: dr.Name,
			Path: dr.Name,
			Kind: config.RepoKind(dr.Kind),
		}
	}

	// Show defaults and allow modification
	fmt.Println("\nDefault repositories from blueprint:")
	for _, r := range repos {
		fmt.Printf("  - %s (%s)\n", r.Name, r.Kind)
	}

	var modify bool
	survey.AskOne(&survey.Confirm{
		Message: "Modify repository list?",
		Default: false,
	}, &modify)

	if modify {
		// Allow adding/removing repos
		// For simplicity, just allow adding
		for {
			var addMore bool
			survey.AskOne(&survey.Confirm{
				Message: "Add another repository?",
				Default: false,
			}, &addMore)

			if !addMore {
				break
			}

			var name, kind string
			survey.AskOne(&survey.Input{Message: "Repository name:"}, &name)
			survey.AskOne(&survey.Select{
				Message: "Repository type:",
				Options: []string{"node", "java", "go", "python", "swift", "docs"},
			}, &kind)

			repos = append(repos, config.RepoConfig{
				Name: name,
				Path: name,
				Kind: config.RepoKind(kind),
			})
		}
	}

	return repos
}

func registerWorkflow(workspacePath string, cfg *config.WorkflowConfig) {
	reg, err := workspace.LoadRegistry()
	if err != nil {
		return // Silently fail - registry is optional
	}

	entry := workspace.RegistryEntry{
		Name:       cfg.Name,
		Path:       workspacePath,
		Blueprint:  cfg.Blueprint,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	reg.AddOrUpdateWorkflow(entry)
	_ = workspace.SaveRegistry(reg) // Best effort, non-critical
}

func printNextSteps(answers wizardAnswers, cfg *config.WorkflowConfig, bp *blueprint.Blueprint) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Workflow created successfully!")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\nWorkflow: %s\n", cfg.Name)
	fmt.Printf("Blueprint: %s\n", cfg.Blueprint)
	fmt.Printf("Topology: %s\n", cfg.Topology)

	fmt.Println("\n📁 Structure created:")
	if cfg.Topology == config.TopologyMultiRepo {
		fmt.Printf("  %s/\n", answers.workspacePath)
		fmt.Printf("  ├── %s/.claude/\n", cfg.Paths.Hub)
		fmt.Printf("  ├── %s/workflow/\n", cfg.Paths.Docs)
		for _, repo := range cfg.Repos {
			fmt.Printf("  └── %s/ (.claude -> hub)\n", repo.Path)
		}
	} else {
		fmt.Printf("  %s/\n", answers.workspacePath)
		fmt.Println("  ├── .claude/")
		fmt.Println("  ├── .ccflow/workflow.yaml")
		fmt.Println("  └── docs/workflow/")
	}

	fmt.Println("\n🚀 Next steps:")
	fmt.Println("  1. Open your repository in Claude Code")
	fmt.Println("  2. Try these commands:")
	fmt.Println("     /idea    - Start a new feature")
	fmt.Println("     /design  - Create a technical design")
	fmt.Println("     /status  - Check workflow status")

	if answers.vcs != config.VCSNone || answers.tracker != config.TrackerNone {
		fmt.Println("\n🔌 MCP Integration (optional):")
		if answers.vcs == config.VCSGitHub {
			fmt.Println("  GitHub: Install the GitHub MCP server for PR integration")
			fmt.Println("  https://github.com/modelcontextprotocol/servers")
		}
		if answers.tracker == config.TrackerLinear {
			fmt.Println("  Linear: Install the Linear MCP server for issue tracking")
		}
		if answers.tracker == config.TrackerJira {
			fmt.Println("  Jira: Install the Jira MCP server for issue tracking")
		}
	}

	fmt.Println("\n📚 Documentation:")
	fmt.Println("  Run 'ccflow status' to check workflow health")
	fmt.Println("  Run 'ccflow doctor' for detailed diagnostics")
	fmt.Println("  Run 'ccflow add-agent --help' to add custom agents")
}
