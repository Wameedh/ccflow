# /release - Release Infrastructure Changes

Coordinate the deployment of approved infrastructure changes.

## Usage

```
/release [feature-id]
```

## What This Command Does

1. **Verifies Approvals**: Ensures all gates passed
2. **Provides Instructions**: Generates deployment instructions
3. **Monitors Deployment**: Guides verification steps
4. **Updates State**: Records deployment status

## CRITICAL: HUMAN APPROVAL REQUIRED

**This command DOES NOT apply changes.**

It provides instructions for humans to execute:
- `terraform apply` must be run by a human
- Production changes require change window
- All deployments require verification

## Prerequisites

Before using this command:
- Feature must be in "plan_approved" status
- Security scans must pass
- Terraform plan must be saved
- Cost estimate must be acceptable
- Rollback plan must be documented

## Process

### Step 1: Verify Release Gates

{{if .GatesEnabled}}
**Gates are ENABLED for this workflow.**

Required gates:
- [ ] Plan approved
- [ ] Security scan passed (0 CRITICAL/HIGH)
- [ ] Cost approved
- [ ] Rollback plan documented
- [ ] Change window scheduled (for production)
{{else}}
**Gates are DISABLED for this workflow.**
{{end}}

### Step 2: Generate Deployment Instructions

#### For Development Environment
```bash
# 1. Navigate to terraform directory
cd infrastructure/environments/dev

# 2. Initialize terraform
terraform init

# 3. Verify the saved plan
terraform show {{.DocsStateDir}}/<feature-id>.tfplan

# 4. Apply the saved plan (HUMAN EXECUTES)
terraform apply {{.DocsStateDir}}/<feature-id>.tfplan

# 5. Verify outputs
terraform output
```

#### For Staging Environment
```bash
# 1. Wait for dev deployment verification
# 2. Get staging approval

# 3. Generate staging plan
terraform plan -out=staging.tfplan

# 4. Apply (HUMAN EXECUTES WITH APPROVAL)
terraform apply staging.tfplan
```

#### For Production Environment
```bash
# 1. Wait for staging deployment verification
# 2. Schedule change window
# 3. Notify on-call
# 4. Get production approval

# 5. Generate production plan
terraform plan -out=production.tfplan

# 6. Apply (HUMAN EXECUTES WITH APPROVAL)
terraform apply production.tfplan

# 7. Monitor dashboards
# 8. Verify application health
```

### Step 3: Verification Checklist

#### Post-Deployment Verification
- [ ] Terraform apply completed successfully
- [ ] Resources created as expected
- [ ] Security groups configured correctly
- [ ] Application health checks passing
- [ ] Monitoring dashboards showing healthy
- [ ] No alerts triggered

### Step 4: Update State
```json
{
  "status": "released",
  "release_completed_at": "<ISO timestamp>",
  "deployed_to": ["dev", "staging", "production"],
  "terraform_outputs": {}
}
```

### Step 5: Output Summary
- Deployment instructions
- Verification steps
- Rollback procedure
- Monitoring links

## Rollback Procedure

If issues are discovered after deployment:

### Immediate Rollback
```bash
# 1. Do not panic
# 2. Assess the issue

# 3. Option A: Apply previous configuration
git checkout HEAD~1 -- .
terraform plan -out=rollback.tfplan
# Review plan, then:
terraform apply rollback.tfplan

# 4. Option B: Targeted fix
# Only if you know exactly what to change
```

### Post-Rollback
1. Update state to "rollback"
2. Document what happened
3. Investigate root cause
4. Plan fix and re-deploy

## Guidelines

- Never apply without reviewing the plan
- Follow environment progression (dev -> staging -> prod)
- Monitor closely after deployment
- Have rollback plan ready
- Document everything
