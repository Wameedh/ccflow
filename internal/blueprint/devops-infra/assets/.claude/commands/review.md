# /review - Review Infrastructure Changes

Review IaC code, security scans, and terraform plan before approval.

## Usage

```
/review [feature-id]
```

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
- Next step: `/release <feature-id>` if approved
