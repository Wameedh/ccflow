#!/bin/bash
# ccflow-managed: true
# ccflow-template: ios-dev/post-edit@v1
#
# Post-edit hook: Runs after Write/Edit operations
# Detects project type and runs appropriate formatters

set -e

# Get the edited file from environment
FILE_PATH="${CLAUDE_FILE_PATH:-}"

if [ -z "$FILE_PATH" ]; then
    exit 0
fi

# Get the directory of the edited file
FILE_DIR=$(dirname "$FILE_PATH")
EXT="${FILE_PATH##*.}"

# Find project root by looking for common markers
find_project_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/Package.swift" ] || [ -d "$dir/*.xcodeproj" ] || [ -d "$dir/*.xcworkspace" ] || [ -f "$dir/package.json" ]; then
            echo "$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    return 1
}

PROJECT_ROOT=$(find_project_root "$FILE_DIR") || PROJECT_ROOT=""

if [ -z "$PROJECT_ROOT" ]; then
    exit 0
fi

cd "$PROJECT_ROOT"

# Format Swift files
if [ "$EXT" = "swift" ]; then
    # Try SwiftFormat if available
    if command -v swiftformat &> /dev/null; then
        swiftformat "$FILE_PATH" 2>/dev/null || true
    fi

    # Try swift-format (Apple's formatter) if available
    if command -v swift-format &> /dev/null; then
        swift-format format -i "$FILE_PATH" 2>/dev/null || true
    fi
fi

# Format other file types (for backend/docs)
case "$EXT" in
    ts|tsx|js|jsx|json|css|scss|md)
        if [ -f "package.json" ] && command -v npx &> /dev/null; then
            if [ -f "node_modules/.bin/prettier" ]; then
                npx prettier --write "$FILE_PATH" 2>/dev/null || true
            fi
        fi
        ;;
    go)
        if command -v gofmt &> /dev/null; then
            gofmt -w "$FILE_PATH" 2>/dev/null || true
        fi
        ;;
esac

exit 0
