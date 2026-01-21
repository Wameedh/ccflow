# ccflow - Claude Code Flow Wizard

A CLI tool for creating and managing Claude Code workflows.

## Features

- **Interactive wizard** to scaffold workflows with agents, commands, and hooks
- **Blueprint system** with pre-built templates (web-dev, ios-dev)
- **Multi-repo support** with centralized workflow hub
- **Quality gates** and validation hooks
- **Upgrade support** to update templates without losing customizations

## Installation

### Homebrew (macOS/Linux)

```bash
brew install wameedh/tap/ccflow
```

### From Source

```bash
go install github.com/wameedh/ccflow@latest
```

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/wameedh/ccflow/releases).

## Quick Start

```bash
# Create a new web development workflow
ccflow run web-dev

# Or use the interactive wizard
ccflow run
```

This creates:
- A `.claude` directory with agents, commands, and hooks
- Workflow state directories for tracking features
- A `workflow.yaml` configuration file

## Usage

### Creating Workflows

```bash
# Interactive mode
ccflow run

# With specific blueprint
ccflow run web-dev
ccflow run ios-dev

# List available blueprints
ccflow list-blueprints
```

### Managing Workflows

```bash
# Check workflow status
ccflow status

# Run diagnostics
ccflow doctor

# List registered workflows
ccflow list
```

### Adding Components

```bash
# Add an agent (using built-in template)
ccflow add-agent devops-agent

# Add an agent from a file
ccflow add-agent my-agent --file ./agent.md

# Add an agent from stdin
cat agent.md | ccflow add-agent my-agent --stdin

# Print template without writing
ccflow add-agent devops-agent --print

# Same options for commands and hooks
ccflow add-command deploy
ccflow add-hook pre-commit --file ./hook.sh
```

### Upgrading Workflows

```bash
# Preview changes
ccflow upgrade --dry-run

# Apply updates
ccflow upgrade
```

### Expanding Topology

```bash
# Convert single-repo to multi-repo
ccflow expand
```

## Workflow Structure

### Multi-repo (default)

```
workspace/
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
  repo1/
    .claude -> ../workflow-hub/.claude
  repo2/
    .claude -> ../workflow-hub/.claude
```

### Single-repo

```
repo/
  .claude/
    agents/
    commands/
    hooks/
    settings.json
  .ccflow/
    workflow.yaml
  docs/
    workflow/
      state/
      designs/
```

## Built-in Commands

After setup, use these commands in Claude Code:

- `/idea` - Capture a new feature idea
- `/design` - Create a technical design
- `/implement` - Implement a feature
- `/review` - Review implementation
- `/release` - Prepare for release
- `/status` - Check workflow status

## Configuration

### workflow.yaml

```yaml
version: 1
name: my-workflow
topology: multi-repo
blueprint: web-dev
paths:
  hub: workflow-hub
  docs: docs
state:
  root: docs/workflow
  state_dir: docs/workflow/state
  designs_dir: docs/workflow/designs
repos:
  - name: web
    path: web
    kind: node
hooks:
  enabled: true
gates:
  enabled: true
mcp:
  vcs: github
  tracker: linear
  deploy: none
```

### Workspace Discovery

ccflow finds your workflow by looking for marker files:
1. `--workspace` flag (highest priority)
2. `CCFLOW_WORKSPACE` environment variable
3. Walk up from current directory for:
   - `workflow-hub/workflow.yaml` (multi-repo)
   - `.ccflow/workflow.yaml` (single-repo)

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Design and internals
- [Blueprints](docs/BLUEPRINTS.md) - Blueprint details
- [MCP Integration](docs/MCP.md) - MCP setup guide
- [Releasing](docs/RELEASING.md) - Release process

## License

MIT
