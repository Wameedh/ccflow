#!/bin/bash
# ccflow-managed: true
# ccflow-template: devops-infra/post-edit@v1
#
# Post-edit hook: Runs after Write/Edit operations on IaC files
# Auto-formats Terraform and YAML files

set -e

# Get the edited file from environment
FILE_PATH="${CLAUDE_FILE_PATH:-}"

if [ -z "$FILE_PATH" ]; then
    exit 0
fi

# Get file extension
EXT="${FILE_PATH##*.}"

# Get the directory of the edited file
FILE_DIR=$(dirname "$FILE_PATH")

# Find project root by looking for common markers
find_project_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -f "$dir/main.tf" ] || [ -f "$dir/Chart.yaml" ] || [ -f "$dir/kustomization.yaml" ]; then
            echo "$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    return 1
}

PROJECT_ROOT=$(find_project_root "$FILE_DIR") || PROJECT_ROOT=""

# Format Terraform files
if [ "$EXT" = "tf" ] || [ "$EXT" = "tfvars" ]; then
    if command -v terraform &> /dev/null; then
        terraform fmt "$FILE_PATH" 2>/dev/null || true
    fi
fi

# Format YAML files
if [ "$EXT" = "yaml" ] || [ "$EXT" = "yml" ]; then
    if command -v yamlfmt &> /dev/null; then
        yamlfmt "$FILE_PATH" 2>/dev/null || true
    fi
fi

# Format JSON files
if [ "$EXT" = "json" ]; then
    if command -v jq &> /dev/null; then
        # Format in place with jq
        TMP_FILE=$(mktemp)
        if jq '.' "$FILE_PATH" > "$TMP_FILE" 2>/dev/null; then
            mv "$TMP_FILE" "$FILE_PATH"
        else
            rm -f "$TMP_FILE"
        fi
    fi
fi

exit 0
