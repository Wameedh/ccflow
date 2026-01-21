#!/bin/bash
# ccflow-managed: true
# ccflow-template: web-dev/post-edit@v1
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

# Find project root by looking for common markers
find_project_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/package.json" ] || [ -f "$dir/go.mod" ] || [ -f "$dir/pom.xml" ] || [ -f "$dir/Cargo.toml" ]; then
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

# Detect project type and run formatter
format_file() {
    local file="$1"
    local ext="${file##*.}"

    # Node.js / TypeScript / JavaScript
    if [ -f "package.json" ]; then
        case "$ext" in
            ts|tsx|js|jsx|json|css|scss|md)
                if command -v npx &> /dev/null; then
                    # Try prettier if available
                    if [ -f "node_modules/.bin/prettier" ] || npx prettier --version &> /dev/null 2>&1; then
                        npx prettier --write "$file" 2>/dev/null || true
                    fi
                    # Try eslint fix if available
                    if [ -f "node_modules/.bin/eslint" ]; then
                        npx eslint --fix "$file" 2>/dev/null || true
                    fi
                fi
                ;;
        esac
    fi

    # Go
    if [ -f "go.mod" ] && [ "$ext" = "go" ]; then
        if command -v gofmt &> /dev/null; then
            gofmt -w "$file" 2>/dev/null || true
        fi
        if command -v goimports &> /dev/null; then
            goimports -w "$file" 2>/dev/null || true
        fi
    fi

    # Python
    if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
        if [ "$ext" = "py" ]; then
            if command -v black &> /dev/null; then
                black "$file" 2>/dev/null || true
            fi
            if command -v isort &> /dev/null; then
                isort "$file" 2>/dev/null || true
            fi
        fi
    fi

    # Terraform
    if [ "$ext" = "tf" ]; then
        if command -v terraform &> /dev/null; then
            terraform fmt "$file" 2>/dev/null || true
        fi
    fi
}

# Run formatter on the edited file
format_file "$FILE_PATH"

exit 0
