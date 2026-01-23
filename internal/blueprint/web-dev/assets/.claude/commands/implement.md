# /implement - Implement a Feature

Implement code changes according to an approved design.

## Usage

```
/implement [feature-id]
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

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully:

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.implement_to_review.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- DO NOT just print "Next step: /review"
- IMMEDIATELY invoke the next command using the Skill tool:
  ```
  Skill(skill="review", args="<feature-id>")
  ```

### If mode is "prompt":
- Ask the user: "Ready to proceed to /review <feature-id>?"
- Use AskUserQuestion with options: ["Yes, proceed", "No, I'll do it later"]
- If "Yes": invoke `Skill(skill="review", args="<feature-id>")`
- If "No": print "Run /review <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /review <feature-id>"
- Do not invoke automatically
