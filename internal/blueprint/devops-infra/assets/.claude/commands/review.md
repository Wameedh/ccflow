# /review - Review Infrastructure Changes

Review IaC code, security scans, and terraform plan before approval.

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
3. **USE** the AskUserQuestion tool (this blocks until user responds)
4. **WAIT** for the response before ANY further action
5. **UPDATE** the state file with the answer
6. **THEN** continue with the workflow

---

## What This Command Does

1. **Reviews Code**: Checks IaC against design and standards
2. **Security Scan**: Verifies security scans pass
3. **Plan Review**: Analyzes terraform plan for expected changes
4. **Cost Review**: Validates cost estimate
5. **Updates State**: Records review results

## CRITICAL SECURITY GATE

This review MUST verify:
- No CRITICAL security findings
- No HIGH security findings (or documented exceptions)
- Terraform plan matches expected changes
- No unexpected resource deletions
- Cost within budget

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Load terraform plan

### Step 2: Run Security Scans

```bash
# tfsec scan
tfsec . --format json > tfsec-report.json

# checkov scan
checkov -d . --output json > checkov-report.json

# trivy for K8s configs
trivy config . --format json > trivy-report.json
```

### Step 3: Review Security Findings

| Severity | Action |
|----------|--------|
| CRITICAL | MUST fix before approval |
| HIGH | MUST fix or document exception |
| MEDIUM | SHOULD fix |
| LOW | May defer |

### Step 4: Review Terraform Plan

```bash
# Load saved plan
terraform show {{.DocsStateDir}}/<feature-id>.tfplan

# JSON format for detailed analysis
terraform show -json {{.DocsStateDir}}/<feature-id>.tfplan | jq '.resource_changes'
```

#### Plan Review Checklist

- [ ] Number of changes matches expectations
- [ ] No unexpected destroys
- [ ] No unexpected recreates
- [ ] Resource types are correct
- [ ] No sensitive data in outputs

### Step 5: Review Cost Estimate

```bash
# Run infracost
infracost breakdown --path . --format json > cost-report.json

# Compare with budget
infracost diff --path . --compare-to infracost-base.json
```

### Step 6: Code Review Checklist

#### Security
- [ ] tfsec scan: 0 CRITICAL, 0 HIGH
- [ ] checkov scan: 0 CRITICAL
- [ ] No secrets in code
- [ ] Encryption enabled
- [ ] IAM least privilege

#### Code Quality
- [ ] Follows module patterns
- [ ] Variables validated
- [ ] Resources tagged
- [ ] Outputs documented

#### Plan Verification
- [ ] Plan matches design
- [ ] No unexpected changes
- [ ] Cost acceptable

#### Blast Radius
- [ ] Assessment accurate
- [ ] Rollback plan documented
- [ ] Approvals obtained

### Step 7: Update State
```json
{
  "status": "plan_approved|changes_requested",
  "review_completed_at": "<ISO timestamp>",
  "security_scan": {
    "passed": true,
    "critical": 0,
    "high": 0,
    "medium": 2,
    "low": 5
  },
  "cost_estimate": {
    "monthly_cost": 150.00,
    "monthly_diff": 50.00,
    "currency": "USD"
  },
  "terraform_plan_path": "{{.DocsStateDir}}/<feature-id>.tfplan"
}
```

### Step 8: Output Summary
- Security scan summary
- Plan summary (create/update/delete counts)
- Cost summary
- Review status

---

## Phase Completion & Handoff

**CRITICAL: You must follow the configured transition behavior.**

After completing this phase successfully (review approved):

1. **Read the workflow configuration** from `.ccflow/workflow.yaml` or `workflow-hub/workflow.yaml`
2. **Check the `transitions.review_to_release.mode` value**
3. **Follow the corresponding behavior:**

### If mode is "auto":
- IMMEDIATELY invoke: `Skill(skill="release", args="<feature-id>")`

### If mode is "prompt":
- Ask the user: "Ready to proceed to /release <feature-id>?"
- If "Yes": invoke `Skill(skill="release", args="<feature-id>")`
- If "No": print "Run /release <feature-id> when ready."

### If mode is "manual":
- Print: "Next step: /release <feature-id>"

**Note:** If the review requested changes, do NOT proceed to release.
