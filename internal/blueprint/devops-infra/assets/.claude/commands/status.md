# /status - Check Infrastructure Status

Display the current state of infrastructure changes and environments.

## Usage

```
/status [feature-id]
```

If no feature-id is provided, shows overview of all infrastructure changes.

## What This Command Does

1. **Lists Changes**: Shows all infrastructure changes in progress
2. **Shows Environment Status**: Displays state of each environment
3. **Security Status**: Shows security scan results
4. **Next Steps**: Suggests appropriate actions

## Process

### Overview Mode (no feature-id)

Scan `{{.DocsStateDir}}/` for all state files and display:

```
Workflow: {{.WorkflowName}}
Blueprint: devops-infra
State Directory: {{.DocsStateDir}}

Infrastructure Changes:
┌─────────────────┬──────────────────┬───────────────┬──────────────┬─────────────────┐
│ ID              │ Title            │ Status        │ Blast Radius │ Last Updated    │
├─────────────────┼──────────────────┼───────────────┼──────────────┼─────────────────┤
│ add-rds         │ Analytics RDS    │ plan_approved │ Medium       │ 2 hours ago     │
│ update-vpc      │ VPC Peering      │ review        │ High         │ 1 day ago       │
│ new-eks         │ EKS Cluster      │ design        │ Critical     │ 3 days ago      │
└─────────────────┴──────────────────┴───────────────┴──────────────┴─────────────────┘

Environment Status:
┌─────────────────┬────────────────┬────────────────┬────────────────┐
│ Environment     │ Last Deploy    │ Status         │ Drift          │
├─────────────────┼────────────────┼────────────────┼────────────────┤
│ dev             │ 2024-01-15     │ ✓ Healthy      │ No drift       │
│ staging         │ 2024-01-14     │ ✓ Healthy      │ No drift       │
│ production      │ 2024-01-10     │ ✓ Healthy      │ 2 resources    │
└─────────────────┴────────────────┴────────────────┴────────────────┘

Summary:
- Ideation: 0
- Design: 1
- Implementation: 0
- Review: 1
- Plan Approved: 1
- Released: 0
```

### Feature Detail Mode (with feature-id)

Display detailed status for a specific infrastructure change:

```
Feature: add-rds
Title: Analytics RDS Instance
Status: plan_approved
Blast Radius: Medium
Created: 2024-01-15T10:30:00Z
Last Updated: 2024-01-16T14:22:00Z

Target Environments: dev, staging, production

Security Scan:
┌──────────────┬────────┐
│ Severity     │ Count  │
├──────────────┼────────┤
│ Critical     │ 0      │
│ High         │ 0      │
│ Medium       │ 2      │
│ Low          │ 5      │
└──────────────┴────────┘
Status: ✓ PASSED

Terraform Plan Summary:
┌──────────────┬────────┐
│ Action       │ Count  │
├──────────────┼────────┤
│ Create       │ 5      │
│ Update       │ 0      │
│ Delete       │ 0      │
│ Replace      │ 0      │
└──────────────┴────────┘

Cost Estimate:
- Monthly Cost: $150.00
- Monthly Change: +$150.00

Deployment Status:
┌──────────────┬────────────┬────────────────┐
│ Environment  │ Status     │ Deployed At    │
├──────────────┼────────────┼────────────────┤
│ dev          │ ○ Pending  │ -              │
│ staging      │ ○ Pending  │ -              │
│ production   │ ○ Pending  │ -              │
└──────────────┴────────────┴────────────────┘

Timeline:
- Ideation: 2024-01-15T10:30:00Z
- Design Started: 2024-01-15T14:00:00Z
- Design Completed: 2024-01-15T17:30:00Z
- Implementation Started: 2024-01-16T09:00:00Z
- Review Completed: 2024-01-16T14:00:00Z
- Plan Approved: 2024-01-16T14:22:00Z

Files Changed:
- terraform/environments/dev/rds.tf
- terraform/environments/staging/rds.tf
- terraform/environments/production/rds.tf
- terraform/modules/rds/main.tf

Terraform Plan: {{.DocsStateDir}}/add-rds.tfplan
Design Doc: {{.DocsDesignDir}}/add-rds-design.md

Next Step: Run /release add-rds to begin deployment
```

## Guidelines

- Run `/status` before starting any infrastructure work
- Check drift regularly with `/drift`
- Monitor cost with `/cost`
- Address stuck changes promptly
