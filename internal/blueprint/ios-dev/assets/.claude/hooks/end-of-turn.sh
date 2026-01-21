#!/bin/bash
# ccflow-managed: true
# ccflow-template: ios-dev/end-of-turn@v1
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
        if [ -f "$dir/Package.swift" ] || ls "$dir"/*.xcodeproj 1>/dev/null 2>&1 || ls "$dir"/*.xcworkspace 1>/dev/null 2>&1 || [ -f "$dir/package.json" ]; then
            echo "$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    echo "$(pwd)"
}

PROJECT_ROOT=$(find_project_root)
cd "$PROJECT_ROOT"

# Check function
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
            echo "$output" | head -5
        fi
        return 1
    fi
}

# Swift Package project checks
if [ -f "Package.swift" ]; then
    echo ""
    echo "Swift Package Detected"
    echo "----------------------"

    if check "Swift Build" swift build; then
        :
    else
        ((ERRORS++)) || true
    fi

    echo -e "Tests:              ${YELLOW}○ Run 'swift test' to verify${NC}"
fi

# Xcode project checks
if ls *.xcodeproj 1>/dev/null 2>&1 || ls *.xcworkspace 1>/dev/null 2>&1; then
    echo ""
    echo "Xcode Project Detected"
    echo "----------------------"

    # Find scheme name
    SCHEME=""
    if ls *.xcworkspace 1>/dev/null 2>&1; then
        WORKSPACE=$(ls *.xcworkspace | head -1)
        SCHEME=$(xcodebuild -workspace "$WORKSPACE" -list 2>/dev/null | grep -A 100 "Schemes:" | tail -n +2 | head -1 | xargs)
    elif ls *.xcodeproj 1>/dev/null 2>&1; then
        PROJECT=$(ls *.xcodeproj | head -1)
        SCHEME=$(xcodebuild -project "$PROJECT" -list 2>/dev/null | grep -A 100 "Schemes:" | tail -n +2 | head -1 | xargs)
    fi

    if [ -n "$SCHEME" ]; then
        echo "  Scheme: $SCHEME"
        echo -e "Build:              ${YELLOW}○ Run 'xcodebuild -scheme $SCHEME build' to verify${NC}"
        echo -e "Tests:              ${YELLOW}○ Run 'xcodebuild -scheme $SCHEME test' to verify${NC}"
    else
        echo -e "  ${YELLOW}Could not determine scheme${NC}"
    fi
fi

# Node.js project checks (for backend)
if [ -f "package.json" ]; then
    echo ""
    echo "Node.js Project Detected"
    echo "------------------------"

    if [ -f "tsconfig.json" ]; then
        if check "TypeScript" npx tsc --noEmit; then
            :
        else
            ((ERRORS++)) || true
        fi
    fi

    if grep -q '"build"' package.json 2>/dev/null; then
        if check "Build" npm run build; then
            :
        else
            ((ERRORS++)) || true
        fi
    fi
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
