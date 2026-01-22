# /implement - Implement a Feature

Implement code changes according to an approved design for Go CLI tools.

## Usage

```
/implement [feature-id]
```

## What This Command Does

1. **Loads Design**: Reads design doc from `{{.DocsDesignDir}}/`
2. **Implements Changes**: Makes code changes per the design
3. **Writes Tests**: Creates table-driven tests for new functionality
4. **Updates State**: Tracks implementation progress

## Prerequisites

Before using this command:
- Feature must have a design document
- Design should be approved (status = "design_approved" or explicit confirmation)
- Understand the existing Cobra patterns in the codebase

## Process

### Step 1: Load Context
- Read feature state from `{{.DocsStateDir}}/<feature-id>.json`
- Read design doc from `{{.DocsDesignDir}}/<feature-id>-design.md`
- Review acceptance criteria

### Step 2: Update State
```json
{
  "status": "implementation",
  "implementation_started_at": "<ISO timestamp>",
  "branch": "feature/<feature-id>"
}
```

### Step 3: Implement Cobra Command
Follow the design document:

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var featureCmd = &cobra.Command{
    Use:   "feature [flags]",
    Short: "Brief description",
    Long:  `Detailed description with usage examples.`,
    Example: `  myapp feature --flag value`,
    RunE: runFeature,
}

func init() {
    rootCmd.AddCommand(featureCmd)
    featureCmd.Flags().StringP("output", "o", "table", "Output format")
}

func runFeature(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

### Step 4: Write Tests
For each new function/command:

```go
func TestFeatureCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
        wantOut string
    }{
        {
            name:    "success case",
            args:    []string{"feature", "--flag", "value"},
            wantErr: false,
            wantOut: "expected output",
        },
        {
            name:    "error case",
            args:    []string{"feature", "--invalid"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Step 5: Verify Implementation
- Run `go vet ./...` to check for issues
- Run `go build ./...` to verify compilation
- Run `go test ./...` to verify tests pass
- Test manually with the CLI
- Verify against acceptance criteria

### Step 6: Update State
```json
{
  "status": "review",
  "implementation_completed_at": "<ISO timestamp>",
  "files_changed": ["cmd/feature.go", "internal/feature/feature.go"],
  "tests_added": ["cmd/feature_test.go", "internal/feature/feature_test.go"]
}
```

### Step 7: Output Summary
- List of files changed
- Tests added
- Any deviations from design
- Suggested next step: `/review <feature-id>`

## Guidelines

- Make small, focused commits
- Don't deviate from design without documenting why
- Write tests alongside code, not after
- Keep backward compatibility for existing CLI flags
- Don't commit secrets or credentials
- Use `go fmt` and `goimports` before committing
- Handle errors explicitly with proper wrapping
