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
Blueprint: web-dev
State Directory: {{.DocsStateDir}}

Active Features:
┌─────────────┬────────────────┬───────────────┬─────────────────┐
│ ID          │ Title          │ Status        │ Last Updated    │
├─────────────┼────────────────┼───────────────┼─────────────────┤
│ dark-mode   │ Dark Mode      │ implementation│ 2 hours ago     │
│ user-prefs  │ User Prefs     │ design        │ 1 day ago       │
│ api-cache   │ API Caching    │ review        │ 30 minutes ago  │
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
Feature: dark-mode
Title: Dark Mode Support
Status: implementation
Created: 2024-01-15T10:30:00Z
Last Updated: 2024-01-16T14:22:00Z

Acceptance Criteria:
✓ 1. User can toggle dark mode from settings
○ 2. Dark mode preference persists across sessions
○ 3. Dark mode respects system preference by default
○ 4. All components support dark mode styling

Timeline:
- Ideation: 2024-01-15T10:30:00Z
- Design Started: 2024-01-15T14:00:00Z
- Design Completed: 2024-01-15T17:30:00Z
- Implementation Started: 2024-01-16T09:00:00Z

Files Changed:
- src/components/ThemeProvider.tsx
- src/hooks/useTheme.ts
- src/styles/themes/dark.css

Design Doc: {{.DocsDesignDir}}/dark-mode-design.md

Next Step: Continue implementation or run /review dark-mode
```

### Gate Status

{{if .GatesEnabled}}
**Quality Gates: ENABLED**

| Gate | Status |
|------|--------|
| Tests Pass | ✓ |
| Build Succeeds | ✓ |
| Lint Clean | ✓ |
| Type Check | ✓ |
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
