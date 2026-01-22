# /release - Release a Feature

Prepare and execute a Go CLI release.

## Usage

```
/release [feature-id]
```

## What This Command Does

1. **Verifies Gates**: Ensures all quality gates pass
2. **Creates Release**: Prepares release with GoReleaser
3. **Deployment Steps**: Provides release checklist
4. **Updates State**: Marks feature as released

## Prerequisites

Before using this command:
- Feature must be in "approved" status
- All tests must pass
- Code review must be complete

## Process

### Step 1: Verify Release Gates

{{if .GatesEnabled}}
**Gates are ENABLED for this workflow.**

Required gates:
- [ ] All tests passing (`go test ./...`)
- [ ] Build succeeds for all platforms
- [ ] `go vet` passes
- [ ] Code review approved
- [ ] No blocking security issues
{{else}}
**Gates are DISABLED for this workflow.**
Proceeding without gate verification.
{{end}}

### Step 2: Pre-Release Checks

```bash
# Verify clean git state
git status

# Run full test suite
go test ./...

# Run with race detector
go test -race ./...

# Build verification
go build ./...

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o dist/myapp-linux-amd64 ./cmd/myapp
GOOS=darwin GOARCH=amd64 go build -o dist/myapp-darwin-amd64 ./cmd/myapp
GOOS=darwin GOARCH=arm64 go build -o dist/myapp-darwin-arm64 ./cmd/myapp
GOOS=windows GOARCH=amd64 go build -o dist/myapp-windows-amd64.exe ./cmd/myapp

# Lint check
golangci-lint run
```

### Step 3: Release Checklist

#### Before Release
- [ ] All automated checks pass
- [ ] Manual testing complete
- [ ] Rollback plan documented
- [ ] CHANGELOG updated
- [ ] Version number incremented

#### Release Steps

**Option A: GoReleaser (Recommended)**
```bash
# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Run GoReleaser
goreleaser release --clean
```

**Option B: Manual Release**
```bash
# Build binaries
./scripts/build.sh

# Create GitHub release
gh release create v1.0.0 dist/* --title "v1.0.0" --notes-file CHANGELOG.md
```

#### After Release
- [ ] GitHub release created
- [ ] Binaries uploaded
- [ ] Checksums verified
- [ ] Homebrew formula updated (if applicable)
- [ ] Documentation updated

### Step 4: Update State
```json
{
  "status": "released",
  "release_completed_at": "<ISO timestamp>",
  "version": "<version>",
  "release_url": "https://github.com/org/repo/releases/tag/vx.y.z",
  "artifacts": [
    "myapp-linux-amd64",
    "myapp-darwin-amd64",
    "myapp-darwin-arm64",
    "myapp-windows-amd64.exe"
  ]
}
```

### Step 5: Output Summary
- Release status
- Version number
- Release URL
- List of artifacts
- Post-release notes

## Rollback Procedure

If issues are discovered:
1. Identify the issue and severity
2. If critical:
   ```bash
   # Delete the release
   gh release delete v1.0.0

   # Delete the tag
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```
3. Update state to "rollback"
4. Investigate and fix issues
5. Re-release when ready

## Guidelines

- Never skip quality gates for "urgent" releases
- Always have a rollback plan
- Use semantic versioning (major.minor.patch)
- Test release artifacts before publishing
- Document any breaking changes in CHANGELOG
