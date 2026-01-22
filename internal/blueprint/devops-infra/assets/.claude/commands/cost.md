# /cost - Estimate Infrastructure Costs

Estimate and track infrastructure costs for proposed and existing resources.

## Usage

```
/cost [feature-id|environment]
```

- With feature-id: Shows cost estimate for a specific change
- With environment: Shows current costs for an environment
- No argument: Shows overview of all costs

## What This Command Does

1. **Estimates Costs**: Uses infracost to estimate resource costs
2. **Compares Costs**: Shows cost difference for proposed changes
3. **Tracks Budget**: Compares against budget thresholds
4. **Generates Reports**: Creates cost breakdown reports

## Prerequisites

Requires `infracost` CLI:
```bash
# Install infracost
brew install infracost

# Authenticate (free tier available)
infracost auth login
```

## Process

### Cost Estimate for Change

```bash
# Navigate to terraform directory
cd infrastructure/environments/<env>

# Generate cost estimate
infracost breakdown --path .
```

### Output: Cost Breakdown

```
Cost Estimate Report
Feature: add-rds
Environment: production
Generated: 2024-01-16T14:30:00Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Current Monthly Cost: $1,250.00

Proposed Changes:
┌─────────────────────────────────────┬──────────────┬──────────────┬──────────────┐
│ Resource                            │ Current      │ Proposed     │ Change       │
├─────────────────────────────────────┼──────────────┼──────────────┼──────────────┤
│ aws_db_instance.analytics           │ $0           │ $150.00      │ +$150.00     │
│ aws_db_subnet_group.analytics       │ $0           │ $0           │ $0           │
│ aws_security_group.rds              │ $0           │ $0           │ $0           │
├─────────────────────────────────────┼──────────────┼──────────────┼──────────────┤
│ Total Change                        │              │              │ +$150.00     │
└─────────────────────────────────────┴──────────────┴──────────────┴──────────────┘

New Monthly Cost: $1,400.00 (+12%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Resource Details:

aws_db_instance.analytics
  Type: db.t3.medium
  Engine: PostgreSQL
  Storage: 100 GB gp3
  Monthly: $150.00
    - Instance: $50.00
    - Storage: $10.00
    - Backup: $5.00
    - Data transfer: ~$85.00 (estimated)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Environment Cost Overview

```
/cost production
```

```
Environment Cost Report
Environment: production
Period: January 2024

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Monthly Cost Summary:
┌─────────────────────────────────────┬──────────────┐
│ Category                            │ Monthly Cost │
├─────────────────────────────────────┼──────────────┤
│ Compute (EC2, ECS, Lambda)          │ $500.00      │
│ Database (RDS, DynamoDB)            │ $300.00      │
│ Storage (S3, EBS)                   │ $150.00      │
│ Networking (VPC, ALB, NAT)          │ $200.00      │
│ Other                               │ $100.00      │
├─────────────────────────────────────┼──────────────┤
│ Total                               │ $1,250.00    │
└─────────────────────────────────────┴──────────────┘

Budget Status:
  Budget: $1,500.00/month
  Current: $1,250.00/month
  Remaining: $250.00 (16.7%)
  Status: ✓ Under budget

Cost Trend:
  Last Month: $1,180.00
  This Month: $1,250.00
  Change: +$70.00 (+5.9%)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Cost Comparison (Diff)

```bash
# Generate baseline
infracost breakdown --path . --format json > infracost-base.json

# After making changes, compare
infracost diff --path . --compare-to infracost-base.json
```

### All Environments Overview

```
/cost
```

```
Infrastructure Cost Overview
Generated: 2024-01-16T14:30:00Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Environment Costs:
┌─────────────────┬──────────────┬──────────────┬────────────────┐
│ Environment     │ Monthly Cost │ Budget       │ Status         │
├─────────────────┼──────────────┼──────────────┼────────────────┤
│ dev             │ $200.00      │ $300.00      │ ✓ Under (33%)  │
│ staging         │ $400.00      │ $500.00      │ ✓ Under (20%)  │
│ production      │ $1,250.00    │ $1,500.00    │ ✓ Under (17%)  │
├─────────────────┼──────────────┼──────────────┼────────────────┤
│ Total           │ $1,850.00    │ $2,300.00    │ ✓ Under (20%)  │
└─────────────────┴──────────────┴──────────────┴────────────────┘

Pending Changes:
┌─────────────────┬─────────────────────┬──────────────┐
│ Feature         │ Description         │ Cost Change  │
├─────────────────┼─────────────────────┼──────────────┤
│ add-rds         │ Analytics RDS       │ +$150.00     │
│ eks-upgrade     │ EKS to 1.28         │ $0           │
└─────────────────┴─────────────────────┴──────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Cost Governance

### Budget Thresholds

| Threshold | Action |
|-----------|--------|
| > 80% of budget | Warning alert |
| > 90% of budget | Review required |
| > 100% of budget | Approval required |

### Cost Review Checklist

- [ ] Change cost is reasonable
- [ ] Budget can accommodate change
- [ ] No unexpected cost drivers
- [ ] Cost optimization considered

## Guidelines

- Run cost estimate before every change
- Review costs during design phase
- Set up budget alerts
- Consider reserved instances for stable workloads
- Review unused resources regularly
