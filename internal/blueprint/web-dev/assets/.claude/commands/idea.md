# /idea - Capture a New Feature Idea

Create a new feature specification and initialize workflow state tracking.

## Usage

```
/idea [feature name or description]
```

## What This Command Does

1. **Gathers Requirements**: Asks clarifying questions about the feature
2. **Creates State File**: Initializes tracking in `{{.DocsStateDir}}/`
3. **Generates Spec**: Creates a structured specification document
4. **Suggests Next Steps**: Recommends moving to `/design`

## Process

### Step 1: Understand the Idea
Ask the user to describe:
- What problem does this solve?
- Who is the target user?
- What is the expected behavior?
- Are there any constraints?

### Step 2: Create State File
Create `{{.DocsStateDir}}/<feature-id>.json`:

```json
{
  "id": "<feature-id>",
  "title": "<Feature Title>",
  "description": "<User's description>",
  "status": "ideation",
  "acceptance_criteria": [],
  "dependencies": [],
  "created_at": "<ISO timestamp>",
  "updated_at": "<ISO timestamp>"
}
```

### Step 3: Define Acceptance Criteria
Work with the user to define clear, testable acceptance criteria:
- Each criterion should be independently verifiable
- Use "Given/When/Then" format when helpful
- Include edge cases and error scenarios

### Step 4: Output Summary
Print a summary including:
- Feature ID and title
- Brief description
- Acceptance criteria
- Suggested next step: `/design <feature-id>`

## Example

```
User: /idea Add dark mode support