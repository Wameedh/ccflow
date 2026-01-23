# /review - Review Implementation

Review Go code changes and prepare for release.

## Usage

```
/review [feature-id]
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

1. **Reviews Code**: Checks implementation against design and Go standards
2. **Runs Validations**: Executes go vet, build, test, and lint
3. **Security Check**: Scans for common security issues
4. **Prepares PR**: Drafts pull request description
5. **Updates State**: Records review results

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Identify all changed files

### Step 2: Run Automated Checks

```bash
# Format check
gofmt -l .

# Static analysis
go vet ./...

# Build verification
go build ./...

# Run tests
go test ./...

# Race detection
go test -race ./...

# Linting (if available)
golangci-lint run

# Cross-platform build check
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

### Step 3: Code Review Checklist

#### Code Quality
- [ ] Follows existing Cobra patterns
- [ ] No unnecessary complexity
- [ ] Appropriate error handling with wrapping
- [ ] No hardcoded values that should be config
- [ ] No debug code (fmt.Println, etc.)

#### Go-Specific
- [ ] `gofmt` formatting applied
- [ ] `goimports` imports organized
- [ ] `go vet` passes
- [ ] Proper context usage
- [ ] Error handling follows conventions

#### Testing
- [ ] Table-driven tests for new logic
- [ ] Tests cover happy path
- [ ] Tests cover error cases
- [ ] No flaky tests introduced

#### Security
- [ ] No secrets in code
- [ ] Input validation for CLI args
- [ ] Safe file operations
- [ ] Proper command escaping

#### CLI Quality
- [ ] Help text is clear and complete
- [ ] Examples are provided
- [ ] Flags are consistent with existing commands
- [ ] Exit codes are correct

#### Documentation
- [ ] README updated if needed
- [ ] Godoc comments for exported items
- [ ] Inline comments for complex logic

### Step 4: Generate PR Description

```markdown
## Summary
<Brief description of changes>

## Changes
- <Change 1>
- <Change 2>

## Command Interface
```
myapp <command> [flags]
```

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual CLI testing completed
- [ ] Cross-platform builds verified

## Related
- Design: {{.DocsDesignDir}}/<feature-id>-design.md
- State: {{.DocsStateDir}}/<feature-id>.json
```

### Step 5: Update State
```json
{
  "status": "approved|changes_requested",
  "review_completed_at": "<ISO timestamp>",
  "review_checklist": {
    "code_quality": true,
    "tests": true,
    "security": true,
    "documentation": true
  },
  "review_notes": []
}
```

### Step 6: Output Summary
- Review status (approved/changes requested)
- Issues found (if any)
- PR description draft

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully (review approved):

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.review_to_release.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- DO NOT just print "Next step: /release"
- IMMEDIATELY invoke the next command using the Skill tool:
  ```
  Skill(skill="release", args="<feature-id>")
  ```

### If mode is "prompt":
- Ask the user: "Ready to proceed to /release <feature-id>?"
- Use AskUserQuestion with options: ["Yes, proceed", "No, I'll do it later"]
- If "Yes": invoke `Skill(skill="release", args="<feature-id>")`
- If "No": print "Run /release <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /release <feature-id>"
- Do not invoke automatically

**Note:** If the review requested changes, do NOT proceed to release. Instead, inform the user about the requested changes.
