# /design - Create Technical Design

Create a technical design document for a Go CLI feature.

## Usage

```
/design [feature-id]
```

## What This Command Does

1. **Loads Feature State**: Reads from `{{.DocsStateDir}}/<feature-id>.json`
2. **Creates Design Doc**: Generates `{{.DocsDesignDir}}/<feature-id>-design.md`
3. **Updates State**: Changes status to "design"
4. **Plans Implementation**: Identifies files to change and Cobra patterns to use

## Process

### Step 1: Load Feature Context
- Read the feature state file
- Understand requirements and acceptance criteria
- Identify related commands in the codebase

### Step 2: Analyze Codebase
- Find similar Cobra command patterns in existing code
- Identify packages that will need changes
- Note any dependencies or risks

### Step 3: Create Design Document

Write to `{{.DocsDesignDir}}/<feature-id>-design.md`:

```markdown
# Design: <Feature Title>

## Status
Design In Progress

## Problem Statement
<From the feature spec>

## Proposed Solution
<High-level description of the approach>

## Command Interface

```
myapp <command> [subcommand] [flags] [args]
```

### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --output | -o | string | table | Output format |

### Arguments
| Position | Name | Required | Description |
|----------|------|----------|-------------|

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |

## Technical Approach

### Package Layout
```
cmd/
  <command>.go      # Cobra command definition
internal/
  <package>/        # Business logic
```

### File Changes
| File | Change Type | Description |
|------|-------------|-------------|
| cmd/<command>.go | create | New command |

### New Packages
- `internal/<package>` - [purpose]

## Alternatives Considered
1. <Alternative 1>
   - Pros: ...
   - Cons: ...
   - Why rejected: ...

## Testing Strategy
- Unit tests: Table-driven tests for business logic
- Integration tests: Command execution tests
- Manual testing: CLI behavior verification

## Rollout Plan
1. Implement command
2. Add tests
3. Update documentation
4. Release
```

### Step 4: Update State
```json
{
  "status": "design",
  "design_started_at": "<ISO timestamp>",
  "design_doc": "{{.DocsDesignDir}}/<feature-id>-design.md"
}
```

### Step 5: Output Summary
- Link to design document
- Key decisions made
- Open questions to resolve
- Suggested next step: `/implement <feature-id>`

## Guidelines

- Keep designs focused and practical
- Follow existing Cobra patterns in the codebase
- Consider backward compatibility for existing flags
- Document all command-line interface decisions
- Get design approval before implementing
