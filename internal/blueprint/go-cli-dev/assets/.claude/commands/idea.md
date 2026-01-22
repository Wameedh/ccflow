# /idea - Capture a New CLI Feature Idea

Create a new feature specification and initialize workflow state tracking for Go CLI tools.

## Usage

```
/idea [feature name or description]
```

## What This Command Does

1. **Gathers Requirements**: Asks clarifying questions about the CLI feature
2. **Creates State File**: Initializes tracking in `{{.DocsStateDir}}/`
3. **Generates Spec**: Creates a structured specification document
4. **Suggests Next Steps**: Recommends moving to `/design`

## Process

### Step 1: Understand the Idea
Ask the user to describe:
- What problem does this command/feature solve?
- Who is the target user (developers, ops, etc.)?
- What is the expected CLI interface (command, flags, args)?
- What output format(s) should be supported?
- Are there any constraints (performance, compatibility)?

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
- Consider CLI-specific requirements:
  - Command syntax and flags
  - Exit codes for success/failure
  - Output format and content
  - Error messages
  - Cross-platform behavior

### Step 4: Output Summary
Print a summary including:
- Feature ID and title
- Brief description
- Proposed command syntax
- Acceptance criteria
- Suggested next step: `/design <feature-id>`

## Example

```
User: /idea Add a config validation command