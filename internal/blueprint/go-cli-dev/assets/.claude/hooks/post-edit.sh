#!/bin/bash
# ccflow-managed: true
# ccflow-template: go-cli-dev/post-edit@v1
#
# Post-edit hook: Runs after Write/Edit operations on Go files
# Auto-formats Go code with gofmt and goimports

set -e

# Get the edited file from environment
FILE_PATH="${CLAUDE_FILE_PATH:-}"

if [ -z "$FILE_PATH" ]; then
    exit 0
fi

# Only process Go files
if [[ ! "$FILE_PATH" =~ \.go$ ]]; then
    exit 0
fi

# Get the directory of the edited file
FILE_DIR=$(dirname "$FILE_PATH")

# Find project root by looking for go.mod
find_project_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/go.mod" ]; then
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

# Format the file with gofmt
if command -v gofmt &> /dev/null; then
    gofmt -w "$FILE_PATH" 2>/dev/null || true
fi

# Run goimports to organize imports
if command -v goimports &> /dev/null; then
    goimports -w "$FILE_PATH" 2>/dev/null || true
fi

exit 0
