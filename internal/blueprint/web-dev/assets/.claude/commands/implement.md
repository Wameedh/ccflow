# /implement - Implement a Feature

Implement code changes according to an approved design.

## Usage

```
/implement [feature-id]
```

## What This Command Does

1. **Loads Design**: Reads design doc from `{{.DocsDesignDir}}/`
2. **Implements Changes**: Makes code changes per the design
3. **Writes Tests**: Creates tests for new functionality
4. **Updates State**: Tracks implementation progress

## Prerequisites

Before using this command:
- Feature must have a design document
- Design should be approved (status = "design_approved" or explicit confirmation)
- Understand the codebase patterns

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Review acceptance criteria

### Step 2: Update State
```json
{
  "status": "implementation",
  "implementation_started_at": "<ISO timestamp>",
  "branch": "feature/<feature-id>"
}
```

### Step 3: Implement Changes
Follow the design document:
1. Create new files as specified
2. Modify existing files as planned
3. Follow existing code patterns
4. Add inline documentation where needed

### Step 4: Write Tests
For each change:
- Unit tests for new functions/methods
- Integration tests for new endpoints
- Component tests for new UI components

### Step 5: Verify Implementation
- Run existing tests to check for regressions
- Run new tests to verify functionality
- Check that build succeeds
- Verify against acceptance criteria

### Step 6: Update State
```json
{
  "status": "review",
  "implementation_completed_at": "<ISO timestamp>",
  "files_changed": ["list", "of", "files"],
  "tests_added": ["list", "of", "test", "files"]
}
```

### Step 7: Output Summary
- List of files changed
- Tests added
- Any deviations from design
- Suggested next step: `/review <feature-id>`

## Guidelines

- Make small, focused commits
- Don't deviate from design without documenting why
- Write tests alongside code, not after
- Keep backward compatibility unless design specifies otherwise
- Don't commit secrets or credentials
- Update documentation as needed
