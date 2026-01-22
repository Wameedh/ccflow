#!/bin/bash
# ccflow-managed: true
# ccflow-template: python-data-dev/post-edit@v1
#
# Post-edit hook: Runs after Write/Edit operations on Python files
# Auto-formats Python code with black, isort, and ruff

set -e

# Get the edited file from environment
FILE_PATH="${CLAUDE_FILE_PATH:-}"

if [ -z "$FILE_PATH" ]; then
    exit 0
fi

# Only process Python files
if [[ ! "$FILE_PATH" =~ \.py$ ]]; then
    exit 0
fi

# Get the directory of the edited file
FILE_DIR=$(dirname "$FILE_PATH")

# Find project root by looking for common markers
find_project_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/pyproject.toml" ] || [ -f "$dir/setup.py" ] || [ -f "$dir/requirements.txt" ]; then
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

# Run black (code formatter)
if command -v black &> /dev/null; then
    black --quiet "$FILE_PATH" 2>/dev/null || true
fi

# Run isort (import sorter)
if command -v isort &> /dev/null; then
    isort --quiet "$FILE_PATH" 2>/dev/null || true
fi

# Run ruff format (if using ruff as formatter)
if command -v ruff &> /dev/null; then
    ruff format --quiet "$FILE_PATH" 2>/dev/null || true
fi

exit 0
