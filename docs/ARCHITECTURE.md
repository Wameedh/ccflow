# Architecture

This document describes the internal architecture of ccflow.

## Overview

ccflow is a Go CLI tool built with:
- **Cobra** for command parsing
- **Survey** for interactive prompts
- **Go embed** for bundling blueprint assets

## Package Structure

```
ccflow/
├── cmd/ccflow/          # Cobra commands
├── internal/
│   ├── blueprint/       # Blueprint loading and management
│   ├── config/          # Configuration types
│   ├── generator/       # Workflow generation
│   ├── installer/       # .claude installation (symlinks)
│   ├── mutator/         # Adding agents/commands/hooks
│   ├── util/            # File utilities
│   ├── validator/       # Status and doctor checks
│   └── workspace/       # Workspace discovery
└── main.go
```

## Workspace Discovery

### Resolution Order

1. `--workspace` flag
2. `CCFLOW_WORKSPACE` environment variable
3. Walk up from CWD looking for markers:
   - `workflow-hub/workflow.yaml` (multi-repo)
   - `.ccflow/workflow.yaml` (single-repo)

If multiple markers exist in the ancestor chain, the nearest one wins.

### Marker Files

**Multi-repo**: `<workspace>/workflow-hub/workflow.yaml`
- The hub contains the canonical `.claude` directory
- Other repos symlink to `../workflow-hub/.claude`

**Single-repo**: `<repo>/.ccflow/workflow.yaml`
- The `.claude` directory lives directly in the repo
- No symlinks needed

## Blueprint System

Blueprints are embedded in the binary using Go's `embed` package.

### Blueprint Structure

```
internal/blueprint/
├── web-dev/
│   ├── blueprint.yaml
│   ├── assets/
│   │   └── .claude/
│   │       ├── agents/*.md
│   │       ├── commands/*.md
│   │       ├── hooks/*.sh
│   │       └── settings.json
│   └── templates/
└── ios-dev/
    └── ...
```

### Template Rendering

Templates support Go `text/template` syntax:
- `{{.WorkflowName}}` - Workflow name
- `{{.DocsStateDir}}` - State directory path
- `{{.DocsDesignDir}}` - Designs directory path
- `{{.GatesEnabled}}` - Whether gates are enabled
- `{{.HooksEnabled}}` - Whether hooks are enabled

## Claude Code Integration

### settings.json

ccflow generates a `settings.json` that configures:

```json
{
  "hooks": [
    {
      "event": "PostToolUse",
      "commands": ["Write", "Edit"],
      "script": "./hooks/post-edit.sh"
    },
    {
      "event": "Stop",
      "script": "./hooks/end-of-turn.sh"
    }
  ],
  "permissions": {
    "allow": [...],
    "deny": [...]
  }
}
```

### Hook Events

- `PostToolUse` - After Write/Edit operations (formatting)
- `Stop` - End of conversation turn (validation)
- `PreToolUse` - Before tool execution (not used by default)
- `Notification` - For notifications

### Permissions

The default permissions:
- **Allow**: Common development tools (npm, git, go, etc.)
- **Deny**: Dangerous operations and sensitive files

## Upgrade Mechanism

The upgrade command tracks which files are managed by ccflow:

1. Files have a header comment: `# ccflow-managed: true`
2. A manifest tracks file hashes: `.ccflow-managed.json`
3. On upgrade:
   - Unchanged managed files are updated
   - Modified files are preserved; new versions written as `.new`
   - User-created files are never touched

## Safety Considerations

### File Safety
- Never overwrite without `--force`
- Symlinks verified before operations
- Manifest tracking prevents data loss

### Permissions
- Conservative default allow list
- Explicit deny for dangerous operations
- No secrets in generated files

### Hook Safety
- Hooks run in user context
- Exit codes respected but don't block
- No network calls in default hooks

## References

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Custom Slash Commands](https://docs.anthropic.com/en/docs/claude-code/slash-commands)
- [Hooks Configuration](https://docs.anthropic.com/en/docs/claude-code/hooks)
