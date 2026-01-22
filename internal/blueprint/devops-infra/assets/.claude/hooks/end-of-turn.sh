#!/bin/bash
# ccflow-managed: true
# ccflow-template: devops-infra/end-of-turn@v1
#
# End-of-turn hook: Runs validation checks when Claude Code stops
# Validates Terraform and Kubernetes configurations

set -e

# Colors for output (if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Running Infrastructure validations..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

ERRORS=0
WARNINGS=0
SECURITY_ISSUES=0

# Check function - runs a command and reports result
check() {
    local name="$1"
    shift
    local cmd="$@"

    printf "%-20s" "$name"

    if output=$($cmd 2>&1); then
        echo -e "${GREEN}✓ PASS${NC}"
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        if [ -n "$output" ]; then
            echo "$output" | head -15
        fi
        return 1
    fi
}

# Security check function
security_check() {
    local name="$1"
    shift
    local cmd="$@"

    printf "%-20s" "$name"

    if output=$($cmd 2>&1); then
        echo -e "${GREEN}✓ PASS${NC}"
        return 0
    else
        local issues=$(echo "$output" | grep -c "CRITICAL\|HIGH" || true)
        if [ "$issues" -gt 0 ]; then
            echo -e "${RED}✗ FAIL ($issues issues)${NC}"
            echo "$output" | grep -E "CRITICAL|HIGH" | head -10
            SECURITY_ISSUES=$((SECURITY_ISSUES + issues))
            return 1
        else
            echo -e "${YELLOW}○ WARN${NC}"
            return 0
        fi
    fi
}

# ============================================================================
# TERRAFORM CHECKS
# ============================================================================

if ls *.tf 1> /dev/null 2>&1 || find . -name "*.tf" -not -path "./.terraform/*" | head -1 | grep -q .; then
    echo ""
    echo "Terraform Configuration Detected"
    echo "--------------------------------"

    # Terraform fmt check
    if command -v terraform &> /dev/null; then
        if check "Terraform Fmt" terraform fmt -check -recursive; then
            :
        else
            ((WARNINGS++)) || true
            echo "  Run 'terraform fmt -recursive' to fix formatting"
        fi

        # Terraform validate (requires init)
        if [ -d ".terraform" ]; then
            if check "Terraform Validate" terraform validate; then
                :
            else
                ((ERRORS++)) || true
            fi
        else
            echo -e "Terraform Validate: ${YELLOW}○ Run 'terraform init' first${NC}"
        fi
    fi

    # tfsec security scan
    if command -v tfsec &> /dev/null; then
        if security_check "tfsec" tfsec --soft-fail .; then
            :
        else
            ((WARNINGS++)) || true
        fi
    else
        echo -e "tfsec:              ${YELLOW}○ Not installed (recommended)${NC}"
    fi

    # checkov security scan
    if command -v checkov &> /dev/null; then
        if security_check "Checkov" checkov -d . --quiet --compact; then
            :
        else
            ((WARNINGS++)) || true
        fi
    else
        echo -e "Checkov:            ${YELLOW}○ Not installed (recommended)${NC}"
    fi

    # infracost estimate
    if command -v infracost &> /dev/null; then
        echo -e "Cost Estimate:      ${YELLOW}○ Run 'infracost breakdown --path .'${NC}"
    fi
fi

# ============================================================================
# KUBERNETES CHECKS
# ============================================================================

if ls *.yaml 1> /dev/null 2>&1 || find . -name "*.yaml" -not -path "./.terraform/*" | head -1 | grep -q .; then
    # Check if any YAML looks like Kubernetes
    if grep -r "apiVersion:" . --include="*.yaml" 2>/dev/null | head -1 | grep -q .; then
        echo ""
        echo "Kubernetes Configuration Detected"
        echo "----------------------------------"

        # yamllint
        if command -v yamllint &> /dev/null; then
            if check "YAML Lint" yamllint -d relaxed .; then
                :
            else
                ((WARNINGS++)) || true
            fi
        fi

        # kubeval or kubeconform
        if command -v kubeconform &> /dev/null; then
            if check "Kubeconform" kubeconform -summary .; then
                :
            else
                ((ERRORS++)) || true
            fi
        elif command -v kubeval &> /dev/null; then
            if check "Kubeval" kubeval .; then
                :
            else
                ((ERRORS++)) || true
            fi
        else
            echo -e "K8s Validation:     ${YELLOW}○ Install kubeconform (recommended)${NC}"
        fi

        # trivy for container/k8s security
        if command -v trivy &> /dev/null; then
            if security_check "Trivy" trivy config --severity HIGH,CRITICAL .; then
                :
            else
                ((WARNINGS++)) || true
            fi
        else
            echo -e "Trivy:              ${YELLOW}○ Not installed (recommended)${NC}"
        fi
    fi
fi

# ============================================================================
# HELM CHECKS
# ============================================================================

if [ -f "Chart.yaml" ]; then
    echo ""
    echo "Helm Chart Detected"
    echo "-------------------"

    if command -v helm &> /dev/null; then
        if check "Helm Lint" helm lint .; then
            :
        else
            ((ERRORS++)) || true
        fi

        if check "Helm Template" helm template . > /dev/null; then
            :
        else
            ((ERRORS++)) || true
        fi
    fi
fi

# ============================================================================
# SUMMARY
# ============================================================================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $SECURITY_ISSUES -gt 0 ]; then
    echo -e "${RED}Security issues found: $SECURITY_ISSUES${NC}"
    echo "Please address security issues before applying changes."
fi

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Validation completed with $ERRORS error(s)${NC}"
    echo "Please fix errors before proceeding."
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}Validation completed with $WARNINGS warning(s)${NC}"
else
    echo -e "${GREEN}All validations passed!${NC}"
fi

echo ""
echo "Remember: Infrastructure changes require human approval."
echo "Run 'terraform plan -out=plan.tfplan' to preview changes."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Don't fail the hook - just provide feedback
exit 0
