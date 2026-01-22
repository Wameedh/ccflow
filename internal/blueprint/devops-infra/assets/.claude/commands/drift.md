# /drift - Detect Infrastructure Drift

Detect configuration drift between Terraform state and actual infrastructure.

## Usage

```
/drift [environment]
```

If no environment specified, checks all environments.

## What This Command Does

1. **Detects Drift**: Compares Terraform state to actual infrastructure
2. **Reports Changes**: Lists resources that have drifted
3. **Suggests Actions**: Recommends remediation steps
4. **Updates State**: Records drift detection results

## What is Drift?

Drift occurs when the actual infrastructure differs from what Terraform expects:
- Manual changes made outside Terraform
- Changes made by other automation
- Resource modifications by cloud provider

## Process

### Step 1: Run Drift Detection

```bash
# For each environment
cd infrastructure/environments/<env>

# Refresh state and plan
terraform plan -refresh-only -out=drift.tfplan

# Show drift details
terraform show drift.tfplan
```

### Step 2: Analyze Drift Report

```
Drift Detection Report
Environment: production
Timestamp: 2024-01-16T14:30:00Z

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Drift Summary:
┌──────────────────────┬────────────────┐
│ Status               │ Count          │
├──────────────────────┼────────────────┤
│ No Drift             │ 45             │
│ Drifted              │ 2              │
│ Unknown              │ 0              │
└──────────────────────┴────────────────┘

Drifted Resources:
┌─────────────────────────────┬──────────────────┬─────────────────────────────┐
│ Resource                    │ Type             │ Drift Description           │
├─────────────────────────────┼──────────────────┼─────────────────────────────┤
│ aws_instance.web[0]         │ aws_instance     │ instance_type changed       │
│ aws_security_group.api      │ aws_security_group│ ingress rule added         │
└─────────────────────────────┴──────────────────┴─────────────────────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Step 3: Drift Details

For each drifted resource:

```
Resource: aws_instance.web[0]
Type: aws_instance
Address: aws_instance.web[0]

Changes Detected:
  ~ instance_type: "t3.medium" -> "t3.large"

Possible Causes:
  - Manual change via AWS Console
  - Change by another automation tool
  - AWS maintenance/migration

Recommended Action:
  Option 1: Update Terraform config to match current state
            (if the change is intentional)
  Option 2: Apply Terraform to revert to desired state
            (if the change is unintentional)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Resource: aws_security_group.api
Type: aws_security_group
Address: aws_security_group.api

Changes Detected:
  + ingress rule: 0.0.0.0/0:22 (SSH)

!! SECURITY CONCERN !!
  SSH access opened to 0.0.0.0/0 is a security risk.

Recommended Action:
  1. Investigate who made this change
  2. Apply Terraform to remove the rule
  3. Review security practices
```

### Step 4: Remediation Options

#### Option 1: Accept Drift (Update Config)
If the change was intentional:
```bash
# Import the current state
# Update Terraform config to match
# Run terraform plan to verify no changes
```

#### Option 2: Revert Drift (Apply Config)
If the change was unintentional:
```bash
# Run terraform plan
# Review the changes
# Apply to revert (HUMAN EXECUTES)
terraform apply
```

### Step 5: Record Drift Detection
```json
{
  "drift_detection": {
    "timestamp": "ISO timestamp",
    "environment": "production",
    "total_resources": 47,
    "drifted_resources": 2,
    "resources": [
      {
        "address": "aws_instance.web[0]",
        "type": "aws_instance",
        "drift_type": "update"
      }
    ]
  }
}
```

## Scheduling Drift Detection

Recommended schedule:
- **Development**: Weekly
- **Staging**: Daily
- **Production**: Daily or continuous

## Guidelines

- Run drift detection regularly
- Investigate all drift immediately
- Document intentional drift
- Security-related drift requires immediate action
- Consider GitOps for drift prevention
