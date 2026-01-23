# /design - Create Technical Design

Create a technical design document for a Go CLI feature.

## Usage

```
/design [feature-id]
```

## CRITICAL: Question Protocol

**YOU MUST FOLLOW THIS PROTOCOL - VIOLATIONS BREAK THE WORKFLOW**

### Before Starting Any Work

1. Read the state file for this feature
2. Check if `pending_questions` array exists with any `answered: false` items
3. If yes: Use AskUserQuestion tool for EACH unanswered question, then STOP
4. If no: Proceed with the command

### When You Need User Input

1. **STOP** all other work immediately
2. **DO NOT** write code, create files, or make decisions without user input
3. **USE** the AskUserQuestion tool (this blocks until user responds):

```
AskUserQuestion(questions=[{
  "question": "Your question here?",
  "header": "Short Label",
  "options": [
    {"label": "Option 1", "description": "What this means"},
    {"label": "Option 2", "description": "What this means"}
  ],
  "multiSelect": false
}])
```

4. **WAIT** for the response before ANY further action
5. **UPDATE** the state file with the answer
6. **THEN** continue with the workflow

---

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

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully:

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.design_to_implement.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- DO NOT just print "Next step: /implement"
- IMMEDIATELY invoke the next command using the Skill tool:
  ```
  Skill(skill="implement", args="<feature-id>")
  ```

### If mode is "prompt":
- Ask the user: "Ready to proceed to /implement <feature-id>?"
- Use AskUserQuestion with options: ["Yes, proceed", "No, I'll do it later"]
- If "Yes": invoke `Skill(skill="implement", args="<feature-id>")`
- If "No": print "Run /implement <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /implement <feature-id>"
- Do not invoke automatically
