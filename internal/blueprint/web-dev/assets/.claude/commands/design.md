# /design - Create Technical Design

Create a technical design document for a feature.

## Usage

```
/design [feature-id]
```

## What This Command Does

1. **Loads Feature State**: Reads from `{{.DocsStateDir}}/<feature-id>.json`
2. **Creates Design Doc**: Generates `{{.DocsDesignDir}}/<feature-id>-design.md`
3. **Updates State**: Changes status to "design"
4. **Plans Implementation**: Identifies files to change and approach

## Process

### Step 1: Load Feature Context
- Read the feature state file
- Understand requirements and acceptance criteria
- Identify related code in the codebase

### Step 2: Analyze Codebase
- Find similar patterns in existing code
- Identify files that will need changes
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

## Technical Approach

### Architecture Changes
<Any architectural changes needed>

### File Changes
| File | Change Type | Description |
|------|-------------|-------------|
| path/to/file | modify | What changes |
| path/to/new | create | New file purpose |

### Data Model Changes
<Any database or state changes>

### API Changes
<Any API endpoint changes>

## Alternatives Considered
1. <Alternative 1>
   - Pros: ...
   - Cons: ...
   - Why rejected: ...

## Testing Strategy
- Unit tests: ...
- Integration tests: ...
- Manual testing: ...

## Rollout Plan
1. <Phase 1>
2. <Phase 2>

## Open Questions
- [ ] Question 1
- [ ] Question 2
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
- Prefer simple solutions over complex ones
- Consider backward compatibility
- Document assumptions clearly
- Get design approval before implementing
