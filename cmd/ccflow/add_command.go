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
	addCommandFileFlag  string
	addCommandStdinFlag bool
	addCommandPrintFlag bool
)

var addCommandCmd = &cobra.Command{
	Use:   "add-command <name>",
	Short: "Add a command to the workflow",
	Long: `Add a new slash command to the workflow.

Content sources (in order of precedence):
  --print   Print the template content without writing
  --stdin   Read content from stdin
  --file    Read content from a file
  (default) Use built-in template if available

Examples:
  ccflow add-command deploy              # Use built-in template
  ccflow add-command my-cmd --stdin      # Read from stdin
  ccflow add-command my-cmd --file ./cmd.md
  ccflow add-command idea --print        # Print template to stdout`,
	Args: cobra.ExactArgs(1),
	Run:  addCommand,
}

func init() {
	addCommandCmd.Flags().StringVar(&addCommandFileFlag, "file", "", "read content from file")
	addCommandCmd.Flags().BoolVar(&addCommandStdinFlag, "stdin", false, "read content from stdin")
	addCommandCmd.Flags().BoolVar(&addCommandPrintFlag, "print", false, "print template to stdout")
}

func addCommand(cmd *cobra.Command, args []string) {
	commandName := args[0]

	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	// Handle --print mode
	if addCommandPrintFlag {
		printCommandTemplate(bpManager, commandName)
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
		Name:        commandName,
		Force:       forceFlag,
		BlueprintID: ws.Config.Blueprint,
		HubPath:     ws.GetHubPath(),
		TemplateData: &blueprint.TemplateData{
			WorkflowName:  ws.Config.Name,
			DocsStateDir:  ws.Config.State.StateDir,
			DocsDesignDir: ws.Config.State.DesignsDir,
			GatesEnabled:  ws.Config.Gates.Enabled,
		},
	}

	if addCommandStdinFlag {
		opts.Source = mutator.SourceStdin
		content, err := util.ReadStdin()
		if err != nil {
			exitWithError("failed to read from stdin: %v", err)
		}
		opts.Content = content
	} else if addCommandFileFlag != "" {
		opts.Source = mutator.SourceFile
		opts.FilePath = addCommandFileFlag
	} else {
		opts.Source = mutator.SourceTemplate
		if !mut.HasTemplate(ws.Config.Blueprint, "command", commandName) {
			exitWithError("no built-in template for command '%s'\nUse --stdin or --file to provide content, or --print to see available templates", commandName)
		}
	}

	// Add the command
	if err := mut.AddCommand(opts); err != nil {
		exitWithError("failed to add command: %v", err)
	}

	printSuccess("Command '%s' added to %s/commands/%s.md", commandName, ws.GetHubPath(), commandName)
	printInfo("Use /%s in Claude Code to invoke this command", commandName)
}

func printCommandTemplate(bpManager *blueprint.Manager, commandName string) {
	// Try to find template in any blueprint
	blueprints := bpManager.List()
	for _, bp := range blueprints {
		if bpManager.HasCommand(bp.ID, commandName) {
			content, err := bpManager.GetCommandContent(bp.ID, commandName, &blueprint.TemplateData{
				WorkflowName:  "{{.WorkflowName}}",
				DocsStateDir:  "{{.DocsStateDir}}",
				DocsDesignDir: "{{.DocsDesignDir}}",
				GatesEnabled:  true,
			})
			if err != nil {
				exitWithError("failed to get template: %v", err)
			}
			fmt.Println(string(content))
			return
		}
	}

	// List available commands
	fmt.Fprintf(os.Stderr, "No template found for command '%s'\n\nAvailable command templates:\n", commandName)
	for _, bp := range blueprints {
		fmt.Fprintf(os.Stderr, "\n  %s:\n", bp.ID)
		for _, command := range bp.Commands.Defaults {
			fmt.Fprintf(os.Stderr, "    - %s\n", command)
		}
	}
	os.Exit(1)
}
