#!/bin/bash
# ccflow-managed: true
# ccflow-template: web-dev/end-of-turn@v1
#
# End-of-turn hook: Runs validation checks when Claude Code stops
# Provides feedback on code health without blocking

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
echo "Running end-of-turn validations..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

ERRORS=0
WARNINGS=0

# Find project root
find_project_root() {
    local dir="${1:-$(pwd)}"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/package.json" ] || [ -f "$dir/go.mod" ] || [ -f "$dir/pom.xml" ]; then
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
            echo "$output" | head -10
        fi
        return 1
    fi
}

# Node.js project checks
if [ -f "package.json" ]; then
    echo ""
    echo "Node.js Project Detected"
    echo "------------------------"

    # Type checking
    if [ -f "tsconfig.json" ]; then
        if check "TypeScript" npx tsc --noEmit; then
            :
        else
            ((ERRORS++)) || true
        fi
    fi

    # Linting
    if [ -f "node_modules/.bin/eslint" ]; then
        if check "ESLint" npx eslint . --max-warnings=0; then
            :
        else
            ((WARNINGS++)) || true
        fi
    fi

    # Tests (quick check - just verify they exist and can run)
    if grep -q '"test"' package.json 2>/dev/null; then
        echo -e "Tests:              ${YELLOW}○ Run 'npm test' to verify${NC}"
    fi

    # Build check
    if grep -q '"build"' package.json 2>/dev/null; then
        if check "Build" npm run build; then
            :
        else
            ((ERRORS++)) || true
        fi
    fi
fi

# Go project checks
if [ -f "go.mod" ]; then
    echo ""
    echo "Go Project Detected"
    echo "-------------------"

    if check "Go Vet" go vet ./...; then
        :
    else
        ((ERRORS++)) || true
    fi

    if check "Go Build" go build ./...; then
        :
    else
        ((ERRORS++)) || true
    fi

    echo -e "Tests:              ${YELLOW}○ Run 'go test ./...' to verify${NC}"
fi

# Python project checks
if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then
    echo ""
    echo "Python Project Detected"
    echo "-----------------------"

    if command -v python3 &> /dev/null; then
        if check "Syntax" python3 -m py_compile *.py 2>/dev/null; then
            :
        else
            ((WARNINGS++)) || true
        fi
    fi

    echo -e "Tests:              ${YELLOW}○ Run 'pytest' to verify${NC}"
fi

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
