# PRD — ccflow (Claude Code Flow Wizard)

## 1. Purpose

`ccflow` is an install-once CLI tool that helps developers **create and manage Claude Code workflows** (agents, slash commands, hooks, settings, multi-repo orchestration). It provides:

* An interactive wizard to scaffold a workflow
* Post-install “management commands” to add agents/commands/hooks later
* Built-in **workflow blueprints** (e.g., `web-dev`, `ios-dev`) that standardize best practices

## 2. Target users

* Solo devs using Claude Code who want a repeatable workflow
* Small teams that want a shared `.claude` setup across multiple repos
* Platform/tooling engineers creating standardized Claude Code workflows

## 3. Platforms

* **v1**: macOS + Linux only (bash + symlinks)
* **future**: Windows supported (copy-based install mode; PowerShell scripts)

## 4. Core concepts

### Workflow

A “workflow” is a workspace containing:

* A canonical `.claude/` config (agents/commands/hooks/settings)
* A shared state location (workflow state JSON + design docs)
* One or more git repos that “use” the workflow via `.claude` install (symlink in v1)

### Workflow Hub pattern

Default multi-repo layout:

* `workflow-hub/` contains `.claude/` (single source of truth)
* `docs/` contains workflow state under `docs/workflow/…`
* Each repo has `.claude -> ../workflow-hub/.claude` symlink

### Blueprint

A blueprint is a preset that defines:

* Default repos to create (optional)
* Default agents, commands, hooks
* Default settings.json permissions
* Workflow state schema + gate policy
* Suggested MCP integrations (guidance, not fully automated)

## 5. Goals / Non-goals

### Goals (v1)

* Installable CLI (`ccflow`) with interactive wizard
* Create workflows from blueprints (multi-repo by default)
* Support existing repos or generate repos
* Manage workflow after creation:

  * add-agent, add-command, add-hook
  * allow custom content via file or stdin
* Workspace discovery with multiple workflows on same machine
* Deterministic output and safe overwrite behavior
* Release pipeline producing binaries and Homebrew installability

### Non-goals (v1)

* Fully automated MCP provisioning across all environments
* Deep understanding of each repo’s tech stack; we’ll use simple “detectors”
* GitHub Actions / ArgoCD generation beyond minimal skeleton (blueprint-specific optional)
* Windows support (planned)

## 6. User stories

### Setup

1. As a user with no repos, I run `ccflow run web-dev` and it creates repos + hub + docs + symlinks.
2. As a user with existing repos, I run `ccflow run` and select repos to attach to a shared hub.

### Manage

3. As a user, I run `ccflow add-agent devops-agent` to add the default devops agent.
4. As a user, I run `ccflow add-agent custom-agent --stdin` to paste markdown and create it.
5. As a user, I run `ccflow add-hook end-of-turn --file ./hook.sh` and it registers the hook in settings.json.
6. As a user, I run `ccflow status` to verify symlinks and show current workflow config.

### Multiple workflows

7. As a user with multiple workflows on my machine, `ccflow` uses the nearest workflow marker by default, and I can override with `--workspace`.

## 7. CLI UX requirements

### Commands

* `ccflow run [blueprint]` (alias `init`)

  * Starts interactive setup wizard
  * Default blueprint if omitted
* `ccflow list-blueprints`
* `ccflow status`
* `ccflow doctor`
* `ccflow add-agent <name> [--file PATH | --stdin | --print] [--force]`
* `ccflow add-command <name> [--file PATH | --stdin | --print] [--force]`
* `ccflow add-hook <name> [--file PATH | --stdin | --print] [--force]`
* (optional niceties) `ccflow list` (registry-based list of workflows)

### Input modes for add-*

* Template mode: `ccflow add-agent devops-agent` uses built-in template if available
* File mode: `--file` writes file content as-is
* Stdin mode: `--stdin` reads content until EOF
* Print mode: `--print` outputs template to stdout without writing
* Overwrite: refuse by default if file exists; `--force` overwrites

### Wizard prompts (minimum)

* Workspace mode: existing repos vs generate new repos
* Topology: multi-repo (default) vs single repo
* Workspace root folder name/location
* Repo selection or repo set creation
* Blueprint selection/confirmation (if not passed)
* Hooks enabled? (default yes)
* Gate policy enabled? (default yes)
* Tracker preference (Linear/Jira/None) for **MCP guidance only**
* VCS preference (GitHub/GitLab/None) for **MCP guidance only**

## 8. Workspace discovery requirements

* By default, `ccflow` searches upward from CWD for a marker file:

  * multi-repo: `<root>/workflow-hub/workflow.yaml`
  * single-repo: `<root>/.ccflow/workflow.yaml`
* If multiple markers exist in ancestors, choose the **nearest**.
* Override selection with:

  * `--workspace /absolute/or/relative/path`
  * `CCFLOW_WORKSPACE` env var
* Optional: maintain global registry `~/.ccflow/registry.json` to list known workflows (convenience only).

## 9. Output artifacts requirements

### Multi-repo default structure

```
<workspace>/
  workflow-hub/
    .claude/
      agents/
      commands/
      hooks/
      settings.json
    workflow.yaml
  docs/
    workflow/
      state/
      designs/
  <repo1>/
  <repo2>/
```

### Single-repo structure

```
<repo>/
  .claude/...
  .ccflow/workflow.yaml
  docs/workflow/state/
  docs/workflow/designs/
```

### Default content requirements

* A usable default set of commands: `/idea /design /implement /review /release /status`
* Default agents (minimum):

  * product-agent, architect-agent, implementation-agent, review-agent, devops-agent
  * plus 2–4 subagents depending on blueprint (web-dev: frontend/test; ios-dev: ios/test; etc.)
* Hooks:

  * `post-edit.sh` (formatting)
  * `end-of-turn.sh` (validate + update state)
  * `safe-bash.sh` (simple allowlist helper; may be referenced in docs)
* `settings.json` wires hooks and defines a conservative permissions baseline.

## 10. Quality, safety, and UX

* Must never overwrite user content without `--force`
* Must show clear next steps after `run`, including:

  * How to symlink or confirm `.claude` usage
  * How to add MCP servers (printed instructions)
* `doctor` must validate:

  * workflow marker exists
  * `.claude/settings.json` valid JSON
  * hooks scripts exist and executable
  * symlink targets correct (multi-repo)
* Provide readable errors with remediation steps.

## 11. Success criteria (v1)

* A user can install ccflow, run `ccflow run web-dev`, and immediately have:

  * a valid Claude Code `.claude` layout
  * working hooks scripts
  * commands available in Claude Code
  * state folders created
* A user can run `ccflow add-agent <name> --stdin` and see the agent file created.
* `ccflow status` and `ccflow doctor` give actionable output and catch broken symlinks.
