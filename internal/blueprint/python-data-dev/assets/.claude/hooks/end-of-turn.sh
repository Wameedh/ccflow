#!/bin/bash
# ccflow-managed: true
# ccflow-template: python-data-dev/end-of-turn@v1
#
# End-of-turn hook: Runs validation checks when Claude Code stops
# Validates Python code with mypy, ruff, and pytest

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
echo "Running Python validations..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

ERRORS=0
WARNINGS=0

# Find project root
find_project_root() {
    local dir="${1:-$(pwd)}"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/pyproject.toml" ] || [ -f "$dir/setup.py" ] || [ -f "$dir/requirements.txt" ]; then
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

# Only run checks if Python project markers exist
if [ ! -f "pyproject.toml" ] && [ ! -f "setup.py" ] && [ ! -f "requirements.txt" ]; then
    echo -e "${YELLOW}No Python project markers found - skipping validations${NC}"
    exit 0
fi

echo ""
echo "Python Project Detected"
echo "-----------------------"

# Syntax check (basic)
if command -v python3 &> /dev/null; then
    if check "Syntax Check" bash -c 'find . -name "*.py" -not -path "./.venv/*" -not -path "./venv/*" -exec python3 -m py_compile {} \; 2>&1 | head -5'; then
        :
    else
        ((ERRORS++)) || true
    fi
fi

# Ruff (linting)
if command -v ruff &> /dev/null; then
    if check "Ruff" ruff check .; then
        :
    else
        ((WARNINGS++)) || true
    fi
else
    echo -e "Ruff:               ${YELLOW}○ Not installed (recommended)${NC}"
fi

# Mypy (type checking)
if command -v mypy &> /dev/null; then
    if check "Mypy" mypy --ignore-missing-imports .; then
        :
    else
        ((WARNINGS++)) || true
    fi
else
    echo -e "Mypy:               ${YELLOW}○ Not installed (optional)${NC}"
fi

# Black (formatting check)
if command -v black &> /dev/null; then
    if check "Black" black --check --quiet .; then
        :
    else
        ((WARNINGS++)) || true
        echo "  Run 'black .' to fix formatting"
    fi
else
    echo -e "Black:              ${YELLOW}○ Not installed (recommended)${NC}"
fi

# isort (import sorting check)
if command -v isort &> /dev/null; then
    if check "isort" isort --check-only --quiet .; then
        :
    else
        ((WARNINGS++)) || true
        echo "  Run 'isort .' to fix import order"
    fi
fi

# Tests suggestion
echo -e "Tests:              ${YELLOW}○ Run 'pytest' to verify${NC}"

# Notebook check
if ls *.ipynb 1> /dev/null 2>&1 || find . -name "*.ipynb" -not -path "./.venv/*" | head -1 | grep -q .; then
    echo -e "Notebooks:          ${YELLOW}○ Run 'nbconvert --execute' to verify${NC}"
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
