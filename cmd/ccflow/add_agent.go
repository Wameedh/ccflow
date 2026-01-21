package ccflow

import (
	"fmt"
	"os"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/mutator"
	"github.com/Wameedh/ccflow/internal/util"
	"github.com/Wameedh/ccflow/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	addAgentFileFlag  string
	addAgentStdinFlag bool
	addAgentPrintFlag bool
)

var addAgentCmd = &cobra.Command{
	Use:   "add-agent <name>",
	Short: "Add an agent to the workflow",
	Long: `Add a new agent to the workflow.

Content sources (in order of precedence):
  --print   Print the template content without writing
  --stdin   Read content from stdin
  --file    Read content from a file
  (default) Use built-in template if available

Examples:
  ccflow add-agent devops-agent           # Use built-in template
  ccflow add-agent my-agent --stdin       # Read from stdin
  ccflow add-agent my-agent --file ./agent.md
  ccflow add-agent devops-agent --print   # Print template to stdout`,
	Args: cobra.ExactArgs(1),
	Run:  addAgent,
}

func init() {
	addAgentCmd.Flags().StringVar(&addAgentFileFlag, "file", "", "read content from file")
	addAgentCmd.Flags().BoolVar(&addAgentStdinFlag, "stdin", false, "read content from stdin")
	addAgentCmd.Flags().BoolVar(&addAgentPrintFlag, "print", false, "print template to stdout")
}

func addAgent(cmd *cobra.Command, args []string) {
	agentName := args[0]

	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	// Handle --print mode
	if addAgentPrintFlag {
		printAgentTemplate(bpManager, agentName)
		return
	}

	// Discover workspace
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	// Determine content source
	mut := mutator.New(bpManager)
	opts := mutator.AddOptions{
		Name:        agentName,
		Force:       forceFlag,
		BlueprintID: ws.Config.Blueprint,
		HubPath:     ws.GetHubPath(),
		TemplateData: &blueprint.TemplateData{
			WorkflowName:  ws.Config.Name,
			DocsStateDir:  ws.Config.State.StateDir,
			DocsDesignDir: ws.Config.State.DesignsDir,
		},
	}

	if addAgentStdinFlag {
		opts.Source = mutator.SourceStdin
		content, err := util.ReadStdin()
		if err != nil {
			exitWithError("failed to read from stdin: %v", err)
		}
		opts.Content = content
	} else if addAgentFileFlag != "" {
		opts.Source = mutator.SourceFile
		opts.FilePath = addAgentFileFlag
	} else {
		opts.Source = mutator.SourceTemplate
		if !mut.HasTemplate(ws.Config.Blueprint, "agent", agentName) {
			exitWithError("no built-in template for agent '%s'\nUse --stdin or --file to provide content, or --print to see available templates", agentName)
		}
	}

	// Add the agent
	if err := mut.AddAgent(opts); err != nil {
		exitWithError("failed to add agent: %v", err)
	}

	printSuccess("Agent '%s' added to %s/agents/%s.md", agentName, ws.GetHubPath(), agentName)
}

func printAgentTemplate(bpManager *blueprint.Manager, agentName string) {
	// Try to find template in any blueprint
	blueprints := bpManager.List()
	for _, bp := range blueprints {
		if bpManager.HasAgent(bp.ID, agentName) {
			content, err := bpManager.GetAgentContent(bp.ID, agentName, &blueprint.TemplateData{
				WorkflowName:  "{{.WorkflowName}}",
				DocsStateDir:  "{{.DocsStateDir}}",
				DocsDesignDir: "{{.DocsDesignDir}}",
			})
			if err != nil {
				exitWithError("failed to get template: %v", err)
			}
			fmt.Println(string(content))
			return
		}
	}

	// List available agents
	fmt.Fprintf(os.Stderr, "No template found for agent '%s'\n\nAvailable agent templates:\n", agentName)
	for _, bp := range blueprints {
		fmt.Fprintf(os.Stderr, "\n  %s:\n", bp.ID)
		for _, agent := range bp.Agents.Defaults {
			fmt.Fprintf(os.Stderr, "    - %s\n", agent)
		}
	}
	os.Exit(1)
}
