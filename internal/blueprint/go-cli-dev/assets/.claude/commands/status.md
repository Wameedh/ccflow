# /status - Check Workflow Status

Display the current state of the workflow and active features.

## Usage

```
/status [feature-id]
```

If no feature-id is provided, shows overview of all features.

## What This Command Does

1. **Lists Features**: Shows all features in the workflow
2. **Shows Progress**: Displays status of each feature
3. **Gate Status**: Shows quality gate compliance
4. **Next Steps**: Suggests appropriate actions

## Process

### Overview Mode (no feature-id)

Scan `{{.DocsStateDir}}/` for all state files and display:

```
Workflow: {{.WorkflowName}}
Blueprint: go-cli-dev
State Directory: {{.DocsStateDir}}

Active Features:
┌─────────────┬────────────────┬───────────────┬─────────────────┐
│ ID          │ Title          │ Status        │ Last Updated    │
├─────────────┼────────────────┼───────────────┼─────────────────┤
│ config-cmd  │ Config Command │ implementation│ 2 hours ago     │
│ list-cmd    │ List Command   │ design        │ 1 day ago       │
│ init-cmd    │ Init Command   │ review        │ 30 minutes ago  │
└─────────────┴────────────────┴───────────────┴─────────────────┘

Summary:
- Ideation: 0
- Design: 1
- Implementation: 1
- Review: 1
- Released: 0
```

### Feature Detail Mode (with feature-id)

Display detailed status for a specific feature:

```
Feature: config-cmd
Title: Config Validation Command
Status: implementation
Created: 2024-01-15T10:30:00Z
Last Updated: 2024-01-16T14:22:00Z

Acceptance Criteria:
✓ 1. Command validates config file syntax
○ 2. Command reports validation errors clearly
○ 3. Command supports --format flag for output
○ 4. Command returns proper exit codes

Timeline:
- Ideation: 2024-01-15T10:30:00Z
- Design Started: 2024-01-15T14:00:00Z
- Design Completed: 2024-01-15T17:30:00Z
- Implementation Started: 2024-01-16T09:00:00Z

Files Changed:
- cmd/config.go
- cmd/config_test.go
- internal/config/validator.go

Design Doc: {{.DocsDesignDir}}/config-cmd-design.md

Next Step: Continue implementation or run /review config-cmd
```

### Build Status

```
Build Status:
┌─────────────────┬────────┐
│ Check           │ Status │
├─────────────────┼────────┤
│ go vet          │ ✓ PASS │
│ go build        │ ✓ PASS │
│ golangci-lint   │ ✓ PASS │
│ go test         │ ○ SKIP │
└─────────────────┴────────┘
```

### Gate Status

{{if .GatesEnabled}}
**Quality Gates: ENABLED**

| Gate | Status |
|------|--------|
| Go Vet | ✓ |
| Go Build | ✓ |
| Tests Pass | ✓ |
| Lint Clean | ✓ |
{{else}}
**Quality Gates: DISABLED**
{{end}}

## Output Format

The command outputs:
- Formatted table for overview
- Detailed view for single feature
- Suggestions for next actions based on current status

## Guidelines

- Run `/status` regularly to track progress
- Address stuck features promptly
- Keep state files updated
- Use feature IDs consistently
