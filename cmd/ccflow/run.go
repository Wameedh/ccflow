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

type wizardAnswers struct {
	workflowName  string
	workspacePath string
	topology      config.Topology
	repos         []config.RepoConfig
	hooksEnabled  bool
	gatesEnabled  bool
	vcs           config.VCSProvider
	tracker       config.TrackerProvider
	transitions   config.TransitionsConfig
	parallel      config.ParallelConfig
}

func runWorkflow(cmd *cobra.Command, args []string) {
	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	// Show welcome message
	printWelcome()

	// Determine setup mode (quick vs custom)
	setupMode := promptSetupMode()

	// Determine blueprint
	var blueprintID string
	if len(args) > 0 {
		blueprintID = args[0]
		if _, getErr := bpManager.Get(blueprintID); getErr != nil {
			exitWithError("unknown blueprint: %s\nRun 'ccflow list-blueprints' to see available blueprints", blueprintID)
		}
	} else {
		blueprintID = promptBlueprintSection(bpManager)
	}

	bp, _ := bpManager.Get(blueprintID)

	// Run wizard prompts based on setup mode
	var answers wizardAnswers
	if setupMode == "quick" {
		answers = runQuickSetup(bp)
	} else {
		answers = runCustomSetup(bp)
	}

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
		Transitions:   answers.transitions,
		Parallel:      answers.parallel,
		Force:         forceFlag,
	}

	// Pre-flight check: warn about existing files
	if !forceFlag {
		existingFiles, checkErr := gen.CheckExistingFiles(opts)
		if checkErr != nil {
			exitWithError("failed to check existing files: %v", checkErr)
		}

		if len(existingFiles) > 0 {
			fmt.Printf("\n  %d file(s) already exist:\n", len(existingFiles))
			for _, f := range existingFiles {
				relPath, relErr := filepath.Rel(answers.workspacePath, f)
				if relErr != nil {
					relPath = f
				}
				fmt.Printf("    - %s\n", relPath)
			}

			var overwrite bool
			promptErr := survey.AskOne(&survey.Confirm{
				Message: "Overwrite existing files?",
				Default: false,
			}, &overwrite)

			if promptErr != nil {
				exitWithError("files already exist. Use --force to overwrite, or run interactively to confirm")
			}

			if !overwrite {
				fmt.Println("\n  Aborted. No files were modified.")
				return
			}
			fmt.Println("\n  Overwriting existing files...")
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

		fmt.Println("\n  Linking workflow to repositories:")
		for _, r := range results {
			if r.Success {
				if r.Skipped {
					printInfo("  %s: %s", r.RepoName, r.Message)
				} else {
					printSuccess("  %s: %s", r.RepoName, r.Message)
				}
			} else {
				printWarning("  %s: %v", r.RepoName, r.Error)
			}
		}
	}

	// Register workflow in global registry
	registerWorkflow(answers.workspacePath, cfg)

	// Print success summary with workflow diagram
	printWorkflowSummary(answers, cfg, bp)
}

// printWelcome displays the welcome message
func printWelcome() {
	fmt.Println()
	fmt.Println("  Welcome to ccflow!")
	fmt.Println()
	fmt.Println("  This wizard sets up an AI-assisted development workflow")
	fmt.Println("  for your project using Claude Code.")
	fmt.Println()
}

// promptSetupMode asks user to choose between quick and custom setup
func promptSetupMode() string {
	var mode string
	survey.AskOne(&survey.Select{
		Message: "How would you like to set up your workflow?",
		Options: []string{
			"Quick setup (recommended) - Use smart defaults, get started fast",
			"Custom setup - Configure each option yourself",
		},
		Default: "Quick setup (recommended) - Use smart defaults, get started fast",
	}, &mode)

	if strings.Contains(mode, "Quick") {
		return "quick"
	}
	return "custom"
}

// promptBlueprintSection displays the project type selection with full details
func promptBlueprintSection(bpManager *blueprint.Manager) string {
	printSectionHeader("PROJECT TYPE", "Choose a template that matches your project")

	blueprints := bpManager.List()

	// Build rich options with full agent list
	options := make([]string, len(blueprints))
	bpIDs := make([]string, len(blueprints))

	for i, bp := range blueprints {
		bpIDs[i] = bp.ID

		// Format commands list
		commands := formatList(bp.Commands.Defaults, "/")

		// Format agents list (full list as requested)
		agents := formatAgentList(bp.Agents.Defaults)

		options[i] = fmt.Sprintf("%s\n    %s\n    Commands: %s\n    Agents:   %s",
			bp.DisplayName,
			bp.Description,
			commands,
			agents,
		)
	}

	var selection string
	survey.AskOne(&survey.Select{
		Message: "What type of project are you building?",
		Options: options,
	}, &selection)

	// Find matching blueprint ID
	for i, opt := range options {
		if opt == selection {
			return bpIDs[i]
		}
	}

	// Fallback: extract from display name
	for i, bp := range blueprints {
		if strings.HasPrefix(selection, bp.DisplayName) {
			return bpIDs[i]
		}
	}

	return blueprints[0].ID
}

// runQuickSetup returns sensible defaults without prompting
func runQuickSetup(bp *blueprint.Blueprint) wizardAnswers {
	cwd, _ := os.Getwd()

	fmt.Println()
	fmt.Println("  Using quick setup with these defaults:")
	fmt.Printf("    - Workflow name: %s\n", filepath.Base(cwd))
	fmt.Println("    - Single repository mode")
	fmt.Println("    - Auto-formatting enabled")
	fmt.Println("    - Review checkpoints enabled")
	fmt.Println("    - Phase transitions: Ask before proceeding")
	fmt.Println("    - Parallel execution: Disabled")
	fmt.Println("    - GitHub integration")
	fmt.Println()

	return wizardAnswers{
		workflowName:  filepath.Base(cwd),
		workspacePath: cwd,
		topology:      config.TopologySingleRepo,
		repos:         nil,
		hooksEnabled:  true,
		gatesEnabled:  true,
		vcs:           config.VCSGitHub,
		tracker:       config.TrackerNone,
		transitions: config.TransitionsConfig{
			IdeaToDesign:      config.TransitionConfig{Mode: config.TransitionPrompt},
			DesignToImplement: config.TransitionConfig{Mode: config.TransitionPrompt},
			ImplementToReview: config.TransitionConfig{Mode: config.TransitionPrompt},
			ReviewToRelease:   config.TransitionConfig{Mode: config.TransitionPrompt},
		},
		parallel: config.ParallelConfig{
			Enabled:  false,
			SyncGate: "all",
		},
	}
}

// runCustomSetup runs the full interactive wizard
func runCustomSetup(bp *blueprint.Blueprint) wizardAnswers {
	answers := wizardAnswers{
		hooksEnabled: true,
		gatesEnabled: true,
	}

	// Get current directory
	cwd, _ := os.Getwd()

	// === PROJECT SETUP SECTION ===
	printSectionHeader("PROJECT SETUP", "Basic information about your project")

	// Workflow name
	survey.AskOne(&survey.Input{
		Message: "What should we call this workflow?",
		Default: filepath.Base(cwd),
		Help:    "This name identifies your workflow in ccflow commands (e.g., ccflow status)",
	}, &answers.workflowName)

	// Workspace path
	survey.AskOne(&survey.Input{
		Message: "Where is your project located?",
		Default: cwd,
		Help:    "The root directory containing your code",
	}, &answers.workspacePath)
	answers.workspacePath, _ = filepath.Abs(answers.workspacePath)

	// === PROJECT STRUCTURE SECTION ===
	printSectionHeader("PROJECT STRUCTURE", "How your code is organized")

	// Repository mode (replaces confusing "topology" and "setup mode")
	var repoMode string
	survey.AskOne(&survey.Select{
		Message: "How is your code organized?",
		Options: []string{
			"Single repository - All code in one place (simpler setup)",
			"Multiple repositories - Separate repos that work together",
		},
		Default: "Single repository - All code in one place (simpler setup)",
		Help: `Single repository: Best for most projects. All your code lives in one git repo.

Multiple repositories: For teams with separate repos (e.g., frontend, backend,
shared libraries). The AI workflow can work across all repos with shared context.`,
	}, &repoMode)

	if strings.Contains(repoMode, "Multiple") {
		answers.topology = config.TopologyMultiRepo

		// Ask about existing vs new repos
		var existingMode string
		survey.AskOne(&survey.Select{
			Message: "Do you have existing repositories to connect?",
			Options: []string{
				"Yes, connect existing repositories",
				"No, I'll set up repositories later",
			},
			Default: "Yes, connect existing repositories",
		}, &existingMode)

		if strings.Contains(existingMode, "Yes") {
			answers.repos = selectExistingRepos(answers.workspacePath)
		} else {
			answers.repos = configureNewRepos(bp)
		}
	} else {
		answers.topology = config.TopologySingleRepo
	}

	// === AUTOMATION SECTION ===
	printSectionHeader("AUTOMATION", "Configure automatic code quality features")

	// Hooks (renamed to "Auto-formatting")
	fmt.Println("  Auto-formatting automatically formats and validates code after the AI")
	fmt.Println("  writes it, ensuring consistent style and catching errors early.")
	fmt.Println()

	var enableFormatting bool
	survey.AskOne(&survey.Confirm{
		Message: "Enable automatic code formatting?",
		Default: true,
	}, &enableFormatting)
	answers.hooksEnabled = enableFormatting

	// Gates (renamed to "Review checkpoints")
	fmt.Println()
	fmt.Println("  Review checkpoints pause the AI at important moments (like before")
	fmt.Println("  deploying) so you can review changes and approve the next step.")
	fmt.Println()

	var enableCheckpoints bool
	survey.AskOne(&survey.Confirm{
		Message: "Enable review checkpoints?",
		Default: true,
	}, &enableCheckpoints)
	answers.gatesEnabled = enableCheckpoints

	// === INTEGRATIONS SECTION ===
	printSectionHeader("INTEGRATIONS", "Connect to your development tools (optional)")

	fmt.Println("  Connecting your code host lets the AI create pull requests,")
	fmt.Println("  understand your branching strategy, and assist with code reviews.")
	fmt.Println()

	// VCS
	var vcsStr string
	survey.AskOne(&survey.Select{
		Message: "Where do you host your code?",
		Options: []string{
			"GitHub",
			"GitLab",
			"Other / Skip this",
		},
		Default: "GitHub",
	}, &vcsStr)
	switch {
	case strings.Contains(vcsStr, "GitHub"):
		answers.vcs = config.VCSGitHub
	case strings.Contains(vcsStr, "GitLab"):
		answers.vcs = config.VCSGitLab
	default:
		answers.vcs = config.VCSNone
	}

	fmt.Println()
	fmt.Println("  Connecting your issue tracker helps the AI understand your")
	fmt.Println("  project requirements and link work to tickets.")
	fmt.Println()

	// Tracker
	var trackerStr string
	survey.AskOne(&survey.Select{
		Message: "What issue tracker do you use?",
		Options: []string{
			"None / Skip this",
			"Linear",
			"Jira",
		},
		Default: "None / Skip this",
	}, &trackerStr)
	switch {
	case strings.Contains(trackerStr, "Linear"):
		answers.tracker = config.TrackerLinear
	case strings.Contains(trackerStr, "Jira"):
		answers.tracker = config.TrackerJira
	default:
		answers.tracker = config.TrackerNone
	}

	// === WORKFLOW TRANSITIONS SECTION ===
	printSectionHeader("WORKFLOW TRANSITIONS", "Configure how phases hand off to each other")

	fmt.Println("  When one phase completes, the AI can automatically start the next")
	fmt.Println("  phase, ask you first, or just suggest what to do next.")
	fmt.Println()

	answers.transitions = promptTransitions()

	// === PARALLEL EXECUTION SECTION ===
	printSectionHeader("PARALLEL EXECUTION", "Run multiple agents simultaneously")

	fmt.Println("  For larger projects, multiple agents can work in parallel")
	fmt.Println("  (e.g., backend and frontend agents implementing together).")
	fmt.Println()

	answers.parallel = promptParallelConfig()

	return answers
}

// printSectionHeader prints a formatted section header
func printSectionHeader(title, subtitle string) {
	fmt.Println()
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Printf("  %s\n", title)
	fmt.Printf("  %s\n", subtitle)
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Println()
}

// formatList formats a list with a prefix
func formatList(items []string, prefix string) string {
	if len(items) == 0 {
		return "(none)"
	}
	formatted := make([]string, len(items))
	for i, item := range items {
		formatted[i] = prefix + item
	}
	return strings.Join(formatted, ", ")
}

// formatAgentList formats agents in a readable way, wrapping if needed
func formatAgentList(agents []string) string {
	if len(agents) == 0 {
		return "(none)"
	}

	// Join all agents, but format nicely
	result := strings.Join(agents, ", ")

	// If too long, truncate with count
	if len(result) > 50 {
		// Show first few and count
		shown := agents[:3]
		remaining := len(agents) - 3
		result = strings.Join(shown, ", ")
		if remaining > 0 {
			result += fmt.Sprintf(" (+%d more)", remaining)
		}
	}

	return result
}

func selectExistingRepos(workspacePath string) []config.RepoConfig {
	var repos []config.RepoConfig

	// Find git repos in workspace
	gitRepos, err := util.FindGitRepos(workspacePath)
	if err != nil || len(gitRepos) == 0 {
		fmt.Println()
		fmt.Println("  No git repositories found in this directory.")
		fmt.Println("  You can add repositories to workflow.yaml later.")
		fmt.Println()
		return repos
	}

	fmt.Println()
	fmt.Println("  Found these repositories in your workspace:")
	fmt.Println()

	// Build options
	options := make([]string, len(gitRepos))
	for i, repo := range gitRepos {
		relPath, _ := filepath.Rel(workspacePath, repo)
		options[i] = relPath
	}

	var selected []string
	survey.AskOne(&survey.MultiSelect{
		Message: "Which repositories should be part of this workflow?",
		Options: options,
		Help:    "Use space to select, enter to confirm. Selected repos will share the AI workflow.",
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

	if len(repos) > 0 {
		fmt.Println()
		fmt.Println("  This template suggests these repositories:")
		for _, r := range repos {
			fmt.Printf("    - %s (%s)\n", r.Name, r.Kind)
		}
		fmt.Println()

		var modify bool
		survey.AskOne(&survey.Confirm{
			Message: "Would you like to modify this list?",
			Default: false,
		}, &modify)

		if modify {
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
					Message: "What type of code will this repository contain?",
					Options: []string{"node", "java", "go", "python", "swift", "docs"},
				}, &kind)

				repos = append(repos, config.RepoConfig{
					Name: name,
					Path: name,
					Kind: config.RepoKind(kind),
				})
			}
		}
	}

	return repos
}

// describeTransitionMode returns a human-readable description of the transition config
func describeTransitionMode(t config.TransitionsConfig) string {
	// Check if all transitions are the same mode
	modes := []config.TransitionMode{
		t.IdeaToDesign.Mode,
		t.DesignToImplement.Mode,
		t.ImplementToReview.Mode,
		t.ReviewToRelease.Mode,
	}

	allSame := true
	for _, m := range modes[1:] {
		if m != modes[0] {
			allSame = false
			break
		}
	}

	if allSame {
		switch modes[0] {
		case config.TransitionAuto:
			return "Automatic (all phases)"
		case config.TransitionPrompt:
			return "Ask before proceeding (all phases)"
		case config.TransitionManual:
			return "Manual (all phases)"
		}
	}

	return "Mixed (custom per phase)"
}

// promptTransitions prompts user to configure workflow transitions
func promptTransitions() config.TransitionsConfig {
	transitions := config.TransitionsConfig{}

	phaseTransitions := []struct {
		field *config.TransitionConfig
		from  string
		to    string
	}{
		{&transitions.IdeaToDesign, "/idea", "/design"},
		{&transitions.DesignToImplement, "/design", "/implement"},
		{&transitions.ImplementToReview, "/implement", "/review"},
		{&transitions.ReviewToRelease, "/review", "/release"},
	}

	for _, pt := range phaseTransitions {
		var mode string
		survey.AskOne(&survey.Select{
			Message: fmt.Sprintf("%s → %s:", pt.from, pt.to),
			Options: []string{
				"Ask me first (recommended)",
				"Automatic (proceed immediately)",
				"Manual (I'll run the command myself)",
			},
			Default: "Ask me first (recommended)",
		}, &mode)

		switch {
		case strings.Contains(mode, "Automatic"):
			pt.field.Mode = config.TransitionAuto
		case strings.Contains(mode, "Manual"):
			pt.field.Mode = config.TransitionManual
		default:
			pt.field.Mode = config.TransitionPrompt
		}
	}

	return transitions
}

// promptParallelConfig prompts user to configure parallel execution
func promptParallelConfig() config.ParallelConfig {
	parallel := config.ParallelConfig{
		Enabled:  false,
		SyncGate: "all",
	}

	var enableParallel bool
	survey.AskOne(&survey.Confirm{
		Message: "Enable parallel agent execution?",
		Default: false,
		Help:    "Allow multiple agents to work simultaneously on different parts of the feature",
	}, &enableParallel)
	parallel.Enabled = enableParallel

	if enableParallel {
		var syncMode string
		survey.AskOne(&survey.Select{
			Message: "Default sync behavior:",
			Options: []string{
				"Wait for all agents to complete before next phase (recommended)",
				"Let each agent proceed independently",
			},
			Default: "Wait for all agents to complete before next phase (recommended)",
		}, &syncMode)

		if strings.Contains(syncMode, "independently") {
			parallel.SyncGate = "any"
		} else {
			parallel.SyncGate = "all"
		}

		fmt.Println()
		fmt.Println("  You can define parallel groups in workflow.yaml later.")
		fmt.Println("  Example: backend-agent and frontend-agent working together.")
	}

	return parallel
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

// printWorkflowSummary displays the success message with ASCII workflow diagram
func printWorkflowSummary(answers wizardAnswers, cfg *config.WorkflowConfig, bp *blueprint.Blueprint) {
	fmt.Println()
	fmt.Println("  " + strings.Repeat("═", 70))
	fmt.Println("  SETUP COMPLETE!")
	fmt.Println("  " + strings.Repeat("═", 70))
	fmt.Println()

	// Summary
	fmt.Printf("  Workflow:  %s\n", cfg.Name)
	fmt.Printf("  Template:  %s\n", bp.DisplayName)
	if cfg.Topology == config.TopologyMultiRepo {
		fmt.Printf("  Structure: Multiple repositories\n")
	} else {
		fmt.Printf("  Structure: Single repository\n")
	}
	fmt.Println()

	// ASCII Workflow Diagram
	fmt.Println("  Your Development Workflow")
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Println()
	fmt.Println("    ┌────────┐    ┌────────┐    ┌───────────┐    ┌────────┐    ┌─────────┐")
	fmt.Println("    │ /idea  │───>│/design │───>│/implement │───>│/review │───>│/release │")
	fmt.Println("    └────────┘    └────────┘    └───────────┘    └────────┘    └─────────┘")
	fmt.Println("        │             │              │               │              │")
	fmt.Println("        v             v              v               v              v")
	fmt.Println("     Capture      Technical      AI writes        Human        Deploy &")
	fmt.Println("     feature      planning       the code        approval       ship it")
	fmt.Println()

	// Show enabled features
	fmt.Println("  Enabled Features")
	fmt.Println("  " + strings.Repeat("─", 70))
	if answers.hooksEnabled {
		fmt.Println("    [x] Auto-formatting    - Code is automatically formatted and validated")
	} else {
		fmt.Println("    [ ] Auto-formatting    - Disabled (you'll format code manually)")
	}
	if answers.gatesEnabled {
		fmt.Println("    [x] Review checkpoints - AI pauses for your approval at key moments")
	} else {
		fmt.Println("    [ ] Review checkpoints - Disabled (AI works without pausing)")
	}
	fmt.Printf("    [x] Phase transitions  - %s\n", describeTransitionMode(answers.transitions))
	if answers.parallel.Enabled {
		fmt.Printf("    [x] Parallel agents    - Enabled (sync: %s)\n", answers.parallel.SyncGate)
	} else {
		fmt.Println("    [ ] Parallel agents    - Disabled (agents run sequentially)")
	}
	fmt.Println()

	// What was created
	fmt.Println("  What Was Created")
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Printf("    Commands:  %s\n", formatList(bp.Commands.Defaults, "/"))
	fmt.Printf("    Agents:    %s\n", strings.Join(bp.Agents.Defaults, ", "))
	fmt.Println()

	// Files created
	fmt.Println("  Files Created")
	fmt.Println("  " + strings.Repeat("─", 70))
	if cfg.Topology == config.TopologyMultiRepo {
		fmt.Printf("    %s/\n", filepath.Base(answers.workspacePath))
		fmt.Printf("    ├── %s/.claude/     (AI workflow configuration)\n", cfg.Paths.Hub)
		fmt.Printf("    ├── %s/workflow/    (feature tracking)\n", cfg.Paths.Docs)
		for i, repo := range cfg.Repos {
			prefix := "├──"
			if i == len(cfg.Repos)-1 {
				prefix = "└──"
			}
			fmt.Printf("    %s %s/            (linked to workflow)\n", prefix, repo.Path)
		}
	} else {
		fmt.Printf("    %s/\n", filepath.Base(answers.workspacePath))
		fmt.Println("    ├── .claude/              (AI workflow configuration)")
		fmt.Println("    ├── .ccflow/workflow.yaml (workflow settings)")
		fmt.Println("    └── docs/workflow/        (feature tracking)")
	}
	fmt.Println()

	// Getting started
	fmt.Println("  Getting Started")
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Println("    1. Open your project in Claude Code")
	fmt.Println("       (or VS Code with the Claude extension)")
	fmt.Println()
	fmt.Println("    2. Type /idea and describe your first feature")
	fmt.Println("       Example: \"/idea Add user authentication with OAuth\"")
	fmt.Println()
	fmt.Println("    3. Follow the workflow: /design -> /implement -> /review -> /release")
	fmt.Println("       The AI will guide you through each step")
	fmt.Println()

	// Command reference
	fmt.Println("  Command Reference")
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Println("    /idea       Describe a new feature you want to build")
	fmt.Println("    /design     Create a technical plan for the feature")
	fmt.Println("    /implement  Start coding with AI assistance")
	fmt.Println("    /review     Get your code reviewed before merging")
	fmt.Println("    /release    Prepare and deploy your changes")
	fmt.Println("    /status     Check the status of all features")
	fmt.Println()

	// Integration hints
	switch answers.vcs {
	case config.VCSGitHub:
		fmt.Println("  Optional: Enhanced Integration")
		fmt.Println("  " + strings.Repeat("─", 70))
		fmt.Println("    For full GitHub integration (auto-create PRs, etc.), install the")
		fmt.Println("    GitHub MCP server: https://github.com/modelcontextprotocol/servers")
		fmt.Println()
	case config.VCSGitLab:
		fmt.Println("  Optional: Enhanced Integration")
		fmt.Println("  " + strings.Repeat("─", 70))
		fmt.Println("    For full GitLab integration, install the GitLab MCP server")
		fmt.Println()
	}

	// Helpful commands
	fmt.Println("  Helpful Commands")
	fmt.Println("  " + strings.Repeat("─", 70))
	fmt.Println("    ccflow status    Check workflow health")
	fmt.Println("    ccflow doctor    Diagnose issues")
	fmt.Println("    ccflow --help    See all commands")
	fmt.Println()
}
