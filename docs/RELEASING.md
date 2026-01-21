# Releasing

This document describes how to release new versions of ccflow.

## Prerequisites

- Go 1.21+
- GoReleaser installed (`brew install goreleaser`)
- GitHub CLI installed (`brew install gh`)
- Write access to the repository
- `GITHUB_TOKEN` with `repo` scope
- `HOMEBREW_TAP_TOKEN` for tap updates

## Version Numbering

ccflow follows semantic versioning:

- **Major** (1.x.x): Breaking changes
- **Minor** (x.1.x): New features, backward compatible
- **Patch** (x.x.1): Bug fixes, backward compatible

## Release Process

### 1. Prepare the Release

```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Run tests
make test

# Run linter
make lint

# Build and verify
make build
./ccflow --version
```

### 2. Update Changelog

Create or update `CHANGELOG.md`:

```markdown
## [1.1.0] - 2024-01-15

### Added
- ios-dev blueprint
- Workflow registry and `ccflow list`
- Upgrade command

### Fixed
- Symlink handling on network drives
```

### 3. Create and Push Tag

```bash
# Create tag
git tag -a v1.1.0 -m "Release v1.1.0"

# Push tag
git push origin v1.1.0
```

### 4. Automated Release

The GitHub Actions workflow automatically:

1. Builds binaries for all platforms
2. Creates GitHub Release with assets
3. Updates Homebrew tap formula

### 5. Verify Release

```bash
# Check GitHub Release
gh release view v1.1.0

# Test Homebrew install (after tap updates)
brew update
brew install wameedh/tap/ccflow
ccflow --version
```

## Manual Release (if needed)

If automation fails:

```bash
# Build snapshot locally
goreleaser build --snapshot --clean

# Create release manually
goreleaser release --clean

# Update tap manually
# Edit homebrew-tap/Formula/ccflow.rb with new version and SHA
```

## Homebrew Tap

The tap repository is at `github.com/wameedh/homebrew-tap`.

### Formula Location

```
homebrew-tap/
└── Formula/
    └── ccflow.rb
```

### Manual Formula Update

```ruby
class Ccflow < Formula
  desc "CLI tool for creating and managing Claude Code workflows"
  homepage "https://github.com/wameedh/ccflow"
  url "https://github.com/wameedh/ccflow/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "<sha256-of-tarball>"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X github.com/wameedh/ccflow/internal/config.Version=#{version}
    ]
    system "go", "build", *std_go_args(ldflags: ldflags)
  end

  test do
    assert_match "ccflow version", shell_output("#{bin}/ccflow version")
  end
end
```

### Get SHA256

```bash
curl -sL https://github.com/wameedh/ccflow/archive/refs/tags/v1.1.0.tar.gz | sha256sum
```

## Rollback

If a release has issues:

```bash
# Delete the release
gh release delete v1.1.0 --yes

# Delete the tag
git push --delete origin v1.1.0
git tag -d v1.1.0

# Revert Homebrew tap if needed
cd homebrew-tap
git revert HEAD
git push
```

## Checklist

- [ ] All tests passing
- [ ] Linter clean
- [ ] CHANGELOG updated
- [ ] Version tag created
- [ ] Release workflow completed
- [ ] GitHub Release has all assets
- [ ] Homebrew formula updated
- [ ] Installation tested
