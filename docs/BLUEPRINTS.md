# Blueprints

ccflow includes built-in blueprints for common development workflows.

## Available Blueprints

### web-dev

**Description**: Multi-repo workflow for TypeScript/JavaScript web applications

**Default Repositories**:
- `web` (node) - Frontend application
- `api` (node) - Backend API
- `docs` (docs) - Documentation

**Agents**:
- `product-agent` - Product specification and requirements
- `architect-agent` - Technical design and architecture
- `implementation-agent` - Code implementation
- `review-agent` - Code review and quality
- `devops-agent` - Build, deploy, and infrastructure
- `frontend-subagent` - Frontend/UI specialization
- `test-subagent` - Testing strategies

**Commands**:
- `/idea` - Capture new feature ideas
- `/design` - Create technical designs
- `/implement` - Implement features
- `/review` - Review implementations
- `/release` - Prepare releases
- `/status` - Check workflow status

**Hooks**:
- `post-edit.sh` - Auto-format on file changes
- `end-of-turn.sh` - Validation at end of turn

**Best For**:
- React/Vue/Angular applications
- Node.js backends
- Full-stack TypeScript projects
- Monorepos with web focus

### ios-dev

**Description**: Multi-repo workflow for iOS/Swift applications

**Default Repositories**:
- `ios-app` (swift) - iOS application
- `api` (node) - Backend API
- `docs` (docs) - Documentation

**Agents**:
- `product-agent` - iOS-focused product specs
- `architect-agent` - Swift/SwiftUI architecture
- `implementation-agent` - Swift implementation
- `review-agent` - iOS code review
- `devops-agent` - Xcode builds and App Store
- `ios-subagent` - iOS/Swift specialization
- `test-subagent` - iOS testing (XCTest, Swift Testing)

**Commands**:
Same as web-dev, with iOS-specific guidance in templates.

**Hooks**:
- `post-edit.sh` - SwiftFormat integration
- `end-of-turn.sh` - Swift build validation

**Best For**:
- iOS applications
- Swift packages
- iOS + backend projects
- App Store submissions

## Customizing Blueprints

### Modifying Templates

After running `ccflow run`, you can modify any file in `.claude/`:

```bash
# Edit an agent
vim .claude/agents/product-agent.md

# Add a new command
vim .claude/commands/my-command.md

# Modify hooks
vim .claude/hooks/post-edit.sh
```

### Adding Custom Agents

```bash
# From built-in template
ccflow add-agent devops-agent

# From file
ccflow add-agent my-agent --file ./my-agent.md

# From stdin
echo "# My Agent\n\nYou are..." | ccflow add-agent my-agent --stdin
```

### Preserving Customizations

When upgrading, ccflow preserves your changes:

1. **Unmodified files**: Updated to latest template
2. **Modified files**: Preserved; new version written as `.new`
3. **Custom files**: Never touched

Run `ccflow upgrade --dry-run` to preview changes.

## Creating Custom Blueprints

Currently, ccflow only supports built-in blueprints. To request a new blueprint:

1. Open an issue on GitHub
2. Describe the use case and target stack
3. Suggest default agents/commands/hooks

Future versions may support user-defined blueprints in `~/.ccflow/blueprints/`.
