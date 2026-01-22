#!/bin/bash
# ccflow-managed: true
# ccflow-template: devops-infra/pre-apply@v1
#
# CRITICAL SAFETY HOOK: Blocks dangerous infrastructure operations
# This hook prevents accidental terraform apply/destroy and kubectl mutations
#
# This hook runs BEFORE Bash commands and can BLOCK execution

set -e

# Get the command that's about to be executed
COMMAND="${CLAUDE_BASH_COMMAND:-}"

if [ -z "$COMMAND" ]; then
    exit 0
fi

# Colors for output
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to block a command
block_command() {
    local reason="$1"
    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}BLOCKED: $reason${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Command attempted: $COMMAND"
    echo ""
    echo "This command has been blocked by the pre-apply safety hook."
    echo "Infrastructure mutations require human approval and should be"
    echo "performed through the proper change management process."
    echo ""
    exit 1
}

# Function to warn about a command
warn_command() {
    local warning="$1"
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}WARNING: $warning${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# ============================================================================
# TERRAFORM SAFETY CHECKS
# ============================================================================

# Block terraform apply without a saved plan file
if echo "$COMMAND" | grep -qE '^terraform\s+apply'; then
    # Check if it's using a saved plan file
    if ! echo "$COMMAND" | grep -qE '\.tfplan|\.plan|plan\.out'; then
        block_command "terraform apply without a saved plan file is not allowed"
    fi
    block_command "terraform apply requires human approval"
fi

# Block terraform destroy
if echo "$COMMAND" | grep -qE '^terraform\s+destroy'; then
    block_command "terraform destroy is blocked - this is a destructive operation"
fi

# Block terraform import (can modify state)
if echo "$COMMAND" | grep -qE '^terraform\s+import'; then
    block_command "terraform import is blocked - state modifications require approval"
fi

# Block terraform state modifications
if echo "$COMMAND" | grep -qE '^terraform\s+state\s+(rm|mv|push|pull)'; then
    block_command "terraform state modifications are blocked"
fi

# Block terraform taint/untaint
if echo "$COMMAND" | grep -qE '^terraform\s+(taint|untaint)'; then
    block_command "terraform taint/untaint is blocked"
fi

# Block terraform force-unlock
if echo "$COMMAND" | grep -qE '^terraform\s+force-unlock'; then
    block_command "terraform force-unlock is blocked - lock management requires approval"
fi

# ============================================================================
# KUBERNETES SAFETY CHECKS
# ============================================================================

# Block kubectl apply
if echo "$COMMAND" | grep -qE '^kubectl\s+apply'; then
    block_command "kubectl apply is blocked - use GitOps or manual approval"
fi

# Block kubectl create
if echo "$COMMAND" | grep -qE '^kubectl\s+create'; then
    block_command "kubectl create is blocked - use declarative configs"
fi

# Block kubectl delete
if echo "$COMMAND" | grep -qE '^kubectl\s+delete'; then
    block_command "kubectl delete is blocked - this is a destructive operation"
fi

# Block kubectl patch/edit
if echo "$COMMAND" | grep -qE '^kubectl\s+(patch|edit)'; then
    block_command "kubectl mutations are blocked"
fi

# Block kubectl exec
if echo "$COMMAND" | grep -qE '^kubectl\s+exec'; then
    block_command "kubectl exec is blocked - use proper debugging procedures"
fi

# Block kubectl run
if echo "$COMMAND" | grep -qE '^kubectl\s+run'; then
    block_command "kubectl run is blocked"
fi

# Block kubectl scale
if echo "$COMMAND" | grep -qE '^kubectl\s+scale'; then
    block_command "kubectl scale is blocked - scaling should be automated"
fi

# Block kubectl rollout (except status)
if echo "$COMMAND" | grep -qE '^kubectl\s+rollout' && ! echo "$COMMAND" | grep -qE 'status|history'; then
    block_command "kubectl rollout mutations are blocked"
fi

# Block kubectl cordon/drain/taint
if echo "$COMMAND" | grep -qE '^kubectl\s+(cordon|drain|taint|uncordon)'; then
    block_command "kubectl node management is blocked"
fi

# ============================================================================
# HELM SAFETY CHECKS
# ============================================================================

# Block helm install
if echo "$COMMAND" | grep -qE '^helm\s+install'; then
    block_command "helm install is blocked - use GitOps"
fi

# Block helm upgrade
if echo "$COMMAND" | grep -qE '^helm\s+upgrade'; then
    block_command "helm upgrade is blocked - use GitOps"
fi

# Block helm uninstall/delete
if echo "$COMMAND" | grep -qE '^helm\s+(uninstall|delete)'; then
    block_command "helm uninstall is blocked - this is a destructive operation"
fi

# Block helm rollback
if echo "$COMMAND" | grep -qE '^helm\s+rollback'; then
    block_command "helm rollback is blocked"
fi

# ============================================================================
# GENERAL SAFETY CHECKS
# ============================================================================

# Block dangerous rm commands
if echo "$COMMAND" | grep -qE '^rm\s+-rf\s+/'; then
    block_command "Dangerous rm command blocked"
fi

# Block sudo/su
if echo "$COMMAND" | grep -qE '^(sudo|su)\s+'; then
    block_command "Privilege escalation is blocked"
fi

# Warn about state file access
if echo "$COMMAND" | grep -qE '\.tfstate|terraform\.tfstate'; then
    warn_command "Accessing state files - ensure you're not exposing secrets"
fi

# Command is allowed - proceed
exit 0
