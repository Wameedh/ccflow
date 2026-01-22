# /implement - Implement Infrastructure Changes

Implement Terraform and Kubernetes configurations according to approved design.

## Usage

```
/implement [feature-id]
```

## What This Command Does

1. **Loads Design**: Reads design doc from `{{.DocsDesignDir}}/`
2. **Implements IaC**: Writes Terraform/Kubernetes configurations
3. **Runs Validations**: Executes security scans
4. **Updates State**: Tracks implementation progress

## CRITICAL SAFETY NOTICE

**This command ONLY writes configuration files.**

It does NOT:
- Apply Terraform changes
- Create/modify Kubernetes resources
- Execute any mutating commands

All infrastructure mutations require human approval.

## Prerequisites

Before using this command:
- Design must be approved
- Appropriate approvals for blast radius obtained
- Rollback plan documented

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Verify blast radius approvals

### Step 2: Update State
```json
{
  "status": "implementation",
  "implementation_started_at": "<ISO timestamp>",
  "branch": "feature/<feature-id>"
}
```

### Step 3: Implement Terraform Resources
Follow the design document:

```hcl
# main.tf
resource "aws_instance" "example" {
  ami           = var.ami_id
  instance_type = var.instance_type

  tags = {
    Name        = "${var.project}-${var.environment}"
    Environment = var.environment
    ManagedBy   = "terraform"
    Feature     = "<feature-id>"
  }
}
```

### Step 4: Run Validation Commands

```bash
# Format check
terraform fmt -check -recursive

# Validate configuration
terraform validate

# Security scan
tfsec .
checkov -d .

# Preview changes (READ-ONLY)
terraform plan -out=plan.tfplan
```

### Step 5: Document Plan Output
Save and review the terraform plan:

```bash
# Save plan
terraform plan -out={{.DocsStateDir}}/<feature-id>.tfplan

# Show plan for review
terraform show {{.DocsStateDir}}/<feature-id>.tfplan
```

### Step 6: Update State
```json
{
  "status": "review",
  "implementation_completed_at": "<ISO timestamp>",
  "files_changed": ["list", "of", "files"],
  "terraform_plan_path": "{{.DocsStateDir}}/<feature-id>.tfplan",
  "security_scan": {
    "passed": true,
    "critical": 0,
    "high": 0
  }
}
```

### Step 7: Output Summary
- List of files changed
- Security scan results
- Terraform plan summary
- Cost estimate
- Suggested next step: `/review <feature-id>`

## Guidelines

- Never run `terraform apply` directly
- Always save plans to files
- Run security scans before every commit
- Document all resources with descriptions
- Tag all resources appropriately
- Follow existing patterns in the codebase
