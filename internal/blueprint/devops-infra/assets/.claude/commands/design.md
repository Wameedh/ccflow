# /design - Create Infrastructure Design

Create a technical design document for infrastructure changes.

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
3. **USE** the AskUserQuestion tool (this blocks until user responds)
4. **WAIT** for the response before ANY further action
5. **UPDATE** the state file with the answer
6. **THEN** continue with the workflow

---

## What This Command Does

1. **Loads Request**: Reads from `{{.DocsStateDir}}/<feature-id>.json`
2. **Creates Design Doc**: Generates `{{.DocsDesignDir}}/<feature-id>-design.md`
3. **Plans Infrastructure**: Identifies resources, security, and cost
4. **Updates State**: Changes status to "design"

## Process

### Step 1: Load Request Context
- Read the feature state file
- Understand requirements and blast radius
- Identify existing infrastructure patterns

### Step 2: Generate Terraform Plan Preview
```bash
# Preview what will change
terraform plan -out=preview.tfplan
terraform show -json preview.tfplan
```

### Step 3: Create Design Document

Write to `{{.DocsDesignDir}}/<feature-id>-design.md`:

```markdown
# Design: [Infrastructure Request Title]

## Status
Design In Progress

## Request Summary
<From the request spec>

## Blast Radius Assessment
- **Level**: [low|medium|high|critical]
- **Affected Systems**: [list]
- **Affected Environments**: [list]
- **Estimated Users Impacted**: [count]

## Proposed Infrastructure

### Resources
| Resource | Type | Provider | Environment | Cost/Month |
|----------|------|----------|-------------|------------|

### Architecture Diagram
```
[ASCII diagram or link]
```

### Terraform Modules
| Module | Source | Purpose |
|--------|--------|---------|

## Security Considerations
- [ ] Encryption at rest configured
- [ ] Encryption in transit configured
- [ ] IAM follows least privilege
- [ ] Security groups properly scoped
- [ ] Audit logging enabled
- [ ] Compliance requirements met

## Cost Estimate
| Resource | Monthly Cost |
|----------|--------------|
| **Total** | **$X** |

## Alternatives Considered
1. [Alternative 1] - [why rejected]

## Rollback Plan
1. [Step 1]
2. [Step 2]

## Deployment Strategy
1. Apply to dev
2. Verify and test
3. Apply to staging (requires approval)
4. Verify and test
5. Apply to production (requires approval + change window)

## File Changes
| File | Change Type | Description |
|------|-------------|-------------|
```

### Step 4: Update State
```json
{
  "status": "design",
  "design_started_at": "<ISO timestamp>",
  "design_doc": "{{.DocsDesignDir}}/<feature-id>-design.md",
  "cost_estimate": {
    "monthly_cost": 150.00,
    "currency": "USD"
  }
}
```

### Step 5: Output Summary
- Link to design document
- Cost estimate
- Security considerations
- Suggested next step: `/implement <feature-id>`

## Guidelines

- Run `infracost` for cost estimates
- Run security scans (tfsec, checkov) on proposed configs
- Document rollback procedures
- Consider multi-environment deployment
- Get appropriate approvals for blast radius

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully:

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.design_to_implement.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- IMMEDIATELY invoke: `Skill(skill="implement", args="<feature-id>")`

### If mode is "prompt":
- Ask the user: "Ready to proceed to /implement <feature-id>?"
- If "Yes": invoke `Skill(skill="implement", args="<feature-id>")`
- If "No": print "Run /implement <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /implement <feature-id>"
