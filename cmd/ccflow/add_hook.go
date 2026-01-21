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
	addHookFileFlag  string
	addHookStdinFlag bool
	addHookPrintFlag bool
)

var addHookCmd = &cobra.Command{
	Use:   "add-hook <name>",
	Short: "Add a hook to the workflow",
	Long: `Add a new hook to the workflow.

Hooks are bash scripts that run in response to Claude Code events.
Adding a hook also updates settings.json to register it.

Content sources (in order of precedence):
  --print   Print the template content without writing
  --stdin   Read content from stdin
  --file    Read content from a file
  (default) Use built-in template if available

Examples:
  ccflow add-hook post-edit              # Use built-in template
  ccflow add-hook my-hook --stdin        # Read from stdin
  ccflow add-hook my-hook --file ./hook.sh
  ccflow add-hook end-of-turn --print    # Print template to stdout`,
	Args: cobra.ExactArgs(1),
	Run:  addHook,
}

func init() {
	addHookCmd.Flags().StringVar(&addHookFileFlag, "file", "", "read content from file")
	addHookCmd.Flags().BoolVar(&addHookStdinFlag, "stdin", false, "read content from stdin")
	addHookCmd.Flags().BoolVar(&addHookPrintFlag, "print", false, "print template to stdout")
}

func addHook(cmd *cobra.Command, args []string) {
	hookName := args[0]

	// Initialize blueprint manager
	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprints: %v", err)
	}

	// Handle --print mode
	if addHookPrintFlag {
		printHookTemplate(bpManager, hookName)
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
		Name:        hookName,
		Force:       forceFlag,
		BlueprintID: ws.Config.Blueprint,
		HubPath:     ws.GetHubPath(),
		TemplateData: &blueprint.TemplateData{
			WorkflowName:  ws.Config.Name,
			DocsStateDir:  ws.Config.State.StateDir,
			DocsDesignDir: ws.Config.State.DesignsDir,
		},
	}

	if addHookStdinFlag {
		opts.Source = mutator.SourceStdin
		content, err := util.ReadStdin()
		if err != nil {
			exitWithError("failed to read from stdin: %v", err)
		}
		opts.Content = content
	} else if addHookFileFlag != "" {
		opts.Source = mutator.SourceFile
		opts.FilePath = addHookFileFlag
	} else {
		opts.Source = mutator.SourceTemplate
		if !mut.HasTemplate(ws.Config.Blueprint, "hook", hookName) {
			exitWithError("no built-in template for hook '%s'\nUse --stdin or --file to provide content, or --print to see available templates", hookName)
		}
	}

	// Add the hook
	if err := mut.AddHook(opts); err != nil {
		exitWithError("failed to add hook: %v", err)
	}

	printSuccess("Hook '%s' added to %s/hooks/%s.sh", hookName, ws.GetHubPath(), hookName)
	printInfo("Hook registered in settings.json")
}

func printHookTemplate(bpManager *blueprint.Manager, hookName string) {
	// Try to find template in any blueprint
	blueprints := bpManager.List()
	for _, bp := range blueprints {
		if bpManager.HasHook(bp.ID, hookName) {
			content, err := bpManager.GetHookContent(bp.ID, hookName, &blueprint.TemplateData{
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

	// List available hooks
	fmt.Fprintf(os.Stderr, "No template found for hook '%s'\n\nAvailable hook templates:\n", hookName)
	for _, bp := range blueprints {
		fmt.Fprintf(os.Stderr, "\n  %s:\n", bp.ID)
		for _, hook := range bp.Hooks.Defaults {
			fmt.Fprintf(os.Stderr, "    - %s\n", hook)
		}
	}
	os.Exit(1)
}
