package ccflow

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Wameedh/ccflow/internal/blueprint"
	"github.com/Wameedh/ccflow/internal/config"
	"github.com/Wameedh/ccflow/internal/permissions"
	"github.com/Wameedh/ccflow/internal/workspace"
)

var (
	writeReposFlag string
	readReposFlag  string
)

var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Manage agent repository permissions",
	Long: `Manage per-agent repository access permissions in your workflow.

Examples:
  ccflow permissions list                      List all agent permissions
  ccflow permissions show backend-agent        Show permissions for an agent
  ccflow permissions set backend-agent --write backend --read frontend
  ccflow permissions grant backend-agent --write frontend
  ccflow permissions revoke backend-agent --write frontend`,
}

var permissionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agent permissions",
	Long: `Display the repository permissions for all agents in the current workflow.

If no explicit permissions are configured, agents have full read/write access
to all repositories (default behavior for backward compatibility).`,
	Run: runPermissionsList,
}

var permissionsShowCmd = &cobra.Command{
	Use:   "show <agent-name>",
	Short: "Show permissions for a specific agent",
	Args:  cobra.ExactArgs(1),
	Run:   runPermissionsShow,
}

var permissionsSetCmd = &cobra.Command{
	Use:   "set <agent-name>",
	Short: "Set permissions for an agent",
	Long: `Set repository permissions for an agent. This replaces any existing permissions.

Examples:
  ccflow permissions set backend-agent --write backend --read frontend,shared
  ccflow permissions set architect-agent --read backend,frontend,shared`,
	Args: cobra.ExactArgs(1),
	Run:  runPermissionsSet,
}

var permissionsGrantCmd = &cobra.Command{
	Use:   "grant <agent-name>",
	Short: "Grant additional access to an agent",
	Long: `Add write or read access to additional repositories for an agent.

Examples:
  ccflow permissions grant backend-agent --write frontend
  ccflow permissions grant architect-agent --read docs`,
	Args: cobra.ExactArgs(1),
	Run:  runPermissionsGrant,
}

var permissionsRevokeCmd = &cobra.Command{
	Use:   "revoke <agent-name>",
	Short: "Revoke access from an agent",
	Long: `Remove write or read access to repositories from an agent.

Examples:
  ccflow permissions revoke backend-agent --write frontend
  ccflow permissions revoke architect-agent --read docs`,
	Args: cobra.ExactArgs(1),
	Run:  runPermissionsRevoke,
}

func init() {
	// Add flags to set command
	permissionsSetCmd.Flags().StringVar(&writeReposFlag, "write", "", "comma-separated list of repos with write access")
	permissionsSetCmd.Flags().StringVar(&readReposFlag, "read", "", "comma-separated list of repos with read-only access")

	// Add flags to grant command
	permissionsGrantCmd.Flags().StringVar(&writeReposFlag, "write", "", "repo to grant write access")
	permissionsGrantCmd.Flags().StringVar(&readReposFlag, "read", "", "repo to grant read-only access")

	// Add flags to revoke command
	permissionsRevokeCmd.Flags().StringVar(&writeReposFlag, "write", "", "repo to revoke write access")
	permissionsRevokeCmd.Flags().StringVar(&readReposFlag, "read", "", "repo to revoke read-only access")

	// Add subcommands
	permissionsCmd.AddCommand(permissionsListCmd)
	permissionsCmd.AddCommand(permissionsShowCmd)
	permissionsCmd.AddCommand(permissionsSetCmd)
	permissionsCmd.AddCommand(permissionsGrantCmd)
	permissionsCmd.AddCommand(permissionsRevokeCmd)
}

func runPermissionsList(cmd *cobra.Command, args []string) {
	ws, mgr := initPermissionsManager()

	perms := mgr.List()
	repoNames := mgr.GetRepoNames()
	agentNames, _ := mgr.GetAgentNames()

	fmt.Printf("\nAgent Permissions for workflow: %s\n", ws.Config.Name)
	fmt.Println(strings.Repeat("─", 50))

	if len(perms) == 0 {
		fmt.Println("\nNo explicit permissions configured.")
		fmt.Println("All agents have full read/write access to all repositories.")
		fmt.Printf("\nRepositories: %s\n", strings.Join(repoNames, ", "))
		fmt.Printf("Agents: %s\n", strings.Join(agentNames, ", "))
		return
	}

	// Show configured permissions
	for agentName, perm := range perms {
		fmt.Printf("\n%s:\n", agentName)
		if len(perm.Write) > 0 {
			fmt.Printf("  Write: %s\n", strings.Join(perm.Write, ", "))
		} else {
			fmt.Println("  Write: (none)")
		}
		if len(perm.Read) > 0 {
			fmt.Printf("  Read:  %s\n", strings.Join(perm.Read, ", "))
		} else {
			fmt.Println("  Read:  (none)")
		}
	}

	// Show agents without explicit permissions
	unconfigured := []string{}
	for _, agent := range agentNames {
		if _, ok := perms[agent]; !ok {
			unconfigured = append(unconfigured, agent)
		}
	}
	if len(unconfigured) > 0 {
		fmt.Printf("\nAgents with full access (no restrictions):\n")
		for _, agent := range unconfigured {
			fmt.Printf("  %s\n", agent)
		}
	}
}

func runPermissionsShow(cmd *cobra.Command, args []string) {
	agentName := args[0]
	ws, mgr := initPermissionsManager()

	perm, err := mgr.Get(agentName)
	if err != nil {
		// Check if agent exists in blueprint
		agentNames, _ := mgr.GetAgentNames()
		found := false
		for _, name := range agentNames {
			if name == agentName {
				found = true
				break
			}
		}
		if !found {
			exitWithError("unknown agent: %s", agentName)
		}

		// No explicit permissions configured
		fmt.Printf("\nAgent: %s (no explicit permissions configured)\n", agentName)
		fmt.Printf("Status: Full read/write access to all repositories\n")
		fmt.Printf("Repositories: %s\n", strings.Join(mgr.GetRepoNames(), ", "))
		return
	}

	fmt.Printf("\nPermissions for %s in workflow: %s\n", agentName, ws.Config.Name)
	fmt.Println(strings.Repeat("─", 50))
	if len(perm.Write) > 0 {
		fmt.Printf("Write access: %s\n", strings.Join(perm.Write, ", "))
	} else {
		fmt.Println("Write access: (none)")
	}
	if len(perm.Read) > 0 {
		fmt.Printf("Read access:  %s\n", strings.Join(perm.Read, ", "))
	} else {
		fmt.Println("Read access:  (none)")
	}
}

func runPermissionsSet(cmd *cobra.Command, args []string) {
	agentName := args[0]
	_, mgr := initPermissionsManager()

	// Validate agent exists
	agentNames, _ := mgr.GetAgentNames()
	found := false
	for _, name := range agentNames {
		if name == agentName {
			found = true
			break
		}
	}
	if !found {
		exitWithError("unknown agent: %s (available: %s)", agentName, strings.Join(agentNames, ", "))
	}

	// Parse repo lists
	perm := config.AgentPermission{}
	if writeReposFlag != "" {
		perm.Write = parseRepoList(writeReposFlag)
	}
	if readReposFlag != "" {
		perm.Read = parseRepoList(readReposFlag)
	}

	// Set permissions
	if err := mgr.Set(agentName, perm); err != nil {
		exitWithError("failed to set permissions: %v", err)
	}

	// Save config
	if err := mgr.Save(); err != nil {
		exitWithError("failed to save workflow config: %v", err)
	}

	printSuccess("Updated permissions for %s", agentName)

	// Regenerate agent if it's a built-in template
	if err := mgr.RegenerateAgent(agentName); err != nil {
		printWarning("Could not regenerate agent template: %v", err)
	} else {
		printSuccess("Regenerated %s.md", agentName)
	}
}

func runPermissionsGrant(cmd *cobra.Command, args []string) {
	agentName := args[0]
	_, mgr := initPermissionsManager()

	// Validate agent exists
	agentNames, _ := mgr.GetAgentNames()
	found := false
	for _, name := range agentNames {
		if name == agentName {
			found = true
			break
		}
	}
	if !found {
		exitWithError("unknown agent: %s", agentName)
	}

	if writeReposFlag == "" && readReposFlag == "" {
		exitWithError("specify --write or --read flag with repository name")
	}

	// Grant write access
	if writeReposFlag != "" {
		for _, repo := range parseRepoList(writeReposFlag) {
			if err := mgr.Grant(agentName, "write", repo); err != nil {
				exitWithError("failed to grant write access: %v", err)
			}
			printSuccess("Granted write access to %s for %s", repo, agentName)
		}
	}

	// Grant read access
	if readReposFlag != "" {
		for _, repo := range parseRepoList(readReposFlag) {
			if err := mgr.Grant(agentName, "read", repo); err != nil {
				exitWithError("failed to grant read access: %v", err)
			}
			printSuccess("Granted read access to %s for %s", repo, agentName)
		}
	}

	// Save config
	if err := mgr.Save(); err != nil {
		exitWithError("failed to save workflow config: %v", err)
	}

	// Regenerate agent
	if err := mgr.RegenerateAgent(agentName); err != nil {
		printWarning("Could not regenerate agent template: %v", err)
	} else {
		printSuccess("Regenerated %s.md", agentName)
	}
}

func runPermissionsRevoke(cmd *cobra.Command, args []string) {
	agentName := args[0]
	_, mgr := initPermissionsManager()

	if writeReposFlag == "" && readReposFlag == "" {
		exitWithError("specify --write or --read flag with repository name")
	}

	// Revoke write access
	if writeReposFlag != "" {
		for _, repo := range parseRepoList(writeReposFlag) {
			if err := mgr.Revoke(agentName, "write", repo); err != nil {
				exitWithError("failed to revoke write access: %v", err)
			}
			printSuccess("Revoked write access to %s from %s", repo, agentName)
		}
	}

	// Revoke read access
	if readReposFlag != "" {
		for _, repo := range parseRepoList(readReposFlag) {
			if err := mgr.Revoke(agentName, "read", repo); err != nil {
				exitWithError("failed to revoke read access: %v", err)
			}
			printSuccess("Revoked read access to %s from %s", repo, agentName)
		}
	}

	// Save config
	if err := mgr.Save(); err != nil {
		exitWithError("failed to save workflow config: %v", err)
	}

	// Regenerate agent
	if err := mgr.RegenerateAgent(agentName); err != nil {
		printWarning("Could not regenerate agent template: %v", err)
	} else {
		printSuccess("Regenerated %s.md", agentName)
	}
}

// initPermissionsManager discovers workspace and creates permission manager
func initPermissionsManager() (*workspace.Workspace, *permissions.Manager) {
	ws, err := workspace.Discover(workspaceFlag)
	if err != nil {
		exitWithError("%v", err)
	}

	bpManager, err := blueprint.NewManager()
	if err != nil {
		exitWithError("failed to initialize blueprint manager: %v", err)
	}

	return ws, permissions.NewManager(ws, bpManager)
}

// parseRepoList parses a comma-separated list of repo names
func parseRepoList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
