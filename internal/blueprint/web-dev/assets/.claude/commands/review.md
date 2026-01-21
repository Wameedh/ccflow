# /review - Review Implementation

Review code changes and prepare for release.

## Usage

```
/review [feature-id]
```

## What This Command Does

1. **Reviews Code**: Checks implementation against design and standards
2. **Runs Validations**: Executes tests, linting, type checking
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
# TypeScript/JavaScript projects
npm run lint
npm run typecheck
npm test
npm run build

# Check for security issues
npm audit
```

### Step 3: Code Review Checklist

#### Code Quality
- [ ] Follows existing code patterns
- [ ] No unnecessary complexity
- [ ] Appropriate error handling
- [ ] No hardcoded values that should be config
- [ ] No console.logs or debug code

#### Testing
- [ ] Unit tests for new logic
- [ ] Integration tests for new endpoints
- [ ] Tests cover edge cases
- [ ] No flaky tests introduced

#### Security
- [ ] No secrets in code
- [ ] Input validation present
- [ ] No injection vulnerabilities
- [ ] Auth/authz properly implemented

#### Documentation
- [ ] README updated if needed
- [ ] API docs updated
- [ ] Inline comments for complex logic

#### Design Compliance
- [ ] Implementation matches design
- [ ] Deviations documented and justified

### Step 4: Generate PR Description

```markdown
## Summary
<Brief description of changes>

## Changes
- <Change 1>
- <Change 2>

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Screenshots (if UI changes)
<Add screenshots>

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
- Suggested next step: `/release <feature-id>` or address feedback
