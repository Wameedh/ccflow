#!/bin/bash
# ccflow-managed: true
# ccflow-template: go-cli-dev/end-of-turn@v1
#
# End-of-turn hook: Runs validation checks when Claude Code stops
# Validates Go code with go vet, go build, and optional golangci-lint

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
echo "Running Go validations..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

ERRORS=0
WARNINGS=0

# Find project root
find_project_root() {
    local dir="${1:-$(pwd)}"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/go.mod" ]; then
            echo "$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    echo "$(pwd)"
}

PROJECT_ROOT=$(find_project_root)
cd "$PROJECT_ROOT"

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

# Only run checks if go.mod exists
if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}No go.mod found - skipping Go validations${NC}"
    exit 0
fi

echo ""
echo "Go Project Detected"
echo "-------------------"

# Go fmt check (verify files are formatted)
if check "Go Fmt" bash -c 'test -z "$(gofmt -l . 2>/dev/null | head -5)"'; then
    :
else
    ((WARNINGS++)) || true
    echo "  Run 'gofmt -w .' to fix formatting"
fi

# Go vet
if check "Go Vet" go vet ./...; then
    :
else
    ((ERRORS++)) || true
fi

# Go build
if check "Go Build" go build ./...; then
    :
else
    ((ERRORS++)) || true
fi

# golangci-lint (if available)
if command -v golangci-lint &> /dev/null; then
    if check "Golangci-lint" golangci-lint run --timeout 2m; then
        :
    else
        ((WARNINGS++)) || true
    fi
else
    echo -e "Golangci-lint:      ${YELLOW}○ Not installed (optional)${NC}"
fi

# Tests suggestion
echo -e "Tests:              ${YELLOW}○ Run 'go test ./...' to verify${NC}"

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Validation completed with $ERRORS error(s)${NC}"
    echo "Please fix errors before committing."
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}Validation completed with $WARNINGS warning(s)${NC}"
else
    echo -e "${GREEN}All validations passed!${NC}"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Don't fail the hook - just provide feedback
exit 0
