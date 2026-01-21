# Design Document — ccflow (Implementation Plan)

## 1. Tech stack

**Language:** Go
**CLI framework:** Cobra (commands, flags, help)
**Templating:** Go `text/template` for templates + `embed` package to ship blueprint assets
**Interactive prompts:** one of:

* `github.com/AlecAivazis/survey/v2` (good wizard UX), or
* minimal in-house prompt helper (stdin reads + defaults)

**Filesystem:** standard library (`os`, `path/filepath`)
**YAML:** `gopkg.in/yaml.v3`
**JSON:** `encoding/json`

## 2. Architecture overview

### Components

1. **Command Layer**

   * Cobra commands map to handlers
2. **Workspace Resolver**

   * Finds workflow root (marker-based; optional registry)
3. **Blueprint Manager**

   * Lists built-ins, loads blueprint manifest + asset files (embedded)
4. **Generator**

   * Creates workflow structure and writes templates
5. **Installer**

   * Installs `.claude` into repos (symlink mode in v1)
6. **Mutators**

   * add-agent/command/hook create files + update config as needed
7. **Validators**

   * status/doctor: symlink checks, file existence, executable bits, JSON parse

### Data flow

* `ccflow run`: prompt → select blueprint → generator writes hub/docs → installer links repos → write workflow.yaml → register workflow (optional)
* `ccflow add-*`: resolve workspace → locate hub `.claude` path → choose template/file/stdin → write artifact → update workflow.yaml (optional) → print result

## 3. Files and on-disk schema

### Markers

* Multi-repo marker: `<workspace>/workflow-hub/workflow.yaml`
* Single-repo marker: `<repo>/.ccflow/workflow.yaml`

### ccflow config (workflow.yaml)

Used for re-runs, upgrades, and discovery.

```yaml
version: 1
name: my-workflow
topology: multi-repo   # multi-repo | single-repo
paths:
  hub: workflow-hub
  docs: docs
state:
  root: docs/workflow
  state_dir: docs/workflow/state
  designs_dir: docs/workflow/designs
repos:
  - name: backend
    path: backend
    kind: java
  - name: web
    path: web
    kind: node
hooks:
  enabled: true
gates:
  enabled: true
blueprint: web-dev
mcp:
  vcs: github|gitlab|none
  tracker: linear|jira|none
  deploy: argocd|none
```

## 4. Blueprint system

### Blueprint directory layout (embedded)

```
internal/blueprints/
  web-dev/
    blueprint.yaml
    assets/
      .claude/agents/*.md
      .claude/commands/*.md
      .claude/hooks/*.sh
      .claude/settings.json.tmpl
      templates/state.schema.json
      templates/ticket.state.json.tmpl
  ios-dev/
    ...
```

### blueprint.yaml schema

```yaml
id: web-dev
display_name: Web Development
description: Multi-repo workflow for TS web apps (+ optional backend)
default_topology: multi-repo
default_repos:
  - name: web
    kind: node
  - name: admin
    kind: node
  - name: docs
    kind: docs
agents:
  defaults: [product-agent, architect-agent, implementation-agent, review-agent, devops-agent, frontend-subagent, test-subagent]
commands:
  defaults: [idea, design, implement, review, release, status]
hooks:
  defaults: [post-edit, end-of-turn]
mcp_suggestions:
  vcs: [github]
  tracker: [linear, jira]
  deploy: [none]
```

### Template tokens

Templates should support token replacement:

* `{{.OrgName}}`, `{{.WorkflowName}}`
* `{{.DocsStateDir}}`, `{{.DocsDesignDir}}`
* `{{.TrackerProvider}}`, `{{.VCSProvider}}`
* `{{.Repos}}`

## 5. Workspace discovery & multiple workflows

### Resolution order

1. `--workspace` if provided
2. `CCFLOW_WORKSPACE` if set
3. Marker walk-up from CWD:

   * look for `workflow-hub/workflow.yaml`
   * else look for `.ccflow/workflow.yaml`

If multiple found in ancestor chain, choose nearest.
If none found, error: “No workflow found. Run `ccflow run` to create one.”

### Optional registry

* Path: `~/.ccflow/registry.json`
* Updated on `run` with `{name, path, blueprint, last_used_at}`
* Used for `ccflow list` only (not required for operations)

## 6. Installer behavior

### v1: symlink mode

For each selected repo:

* Ensure repo path exists
* If `.claude` exists:

  * if it’s a symlink pointing to hub: ok
  * else: refuse unless `--force-install` (future flag)
* Create symlink:

  * `<repo>/.claude -> <relative path to hub/.claude>` (prefer relative symlink)

Also consider:

* Ensure hook scripts in hub are executable (`chmod +x`)

### Future: copy mode

* Detect Windows, or allow `--install-mode copy`
* Copy `.claude` folder into repos
* Add a `.ccflow/managed.json` marker so ccflow can update later

## 7. add-agent / add-command / add-hook

### Shared logic: “content source selection”

Inputs:

* `<name>` plus flags: `--file`, `--stdin`, `--print`, `--force`

Rules:

* If `--print`: output template content (if built-in), else error if no source
* If `--file`: read file content
* If `--stdin`: read all from stdin
* Else: use built-in template by `<name>` (and error if not found)
* If destination exists: refuse unless `--force`

Destinations:

* agent → `<hub>/.claude/agents/<name>.md`
* command → `<hub>/.claude/commands/<name>.md`
* hook → `<hub>/.claude/hooks/<name>.sh` AND update `<hub>/.claude/settings.json`

### Hook registration update

Settings JSON structure:

```json
{
  "hooks": [
    {"event": "PostToolUse", "commands": ["Write", "Edit"], "script": "./hooks/post-edit.sh"},
    {"event": "Stop", "script": "./hooks/end-of-turn.sh"}
  ],
  "permissions": { ... }
}
```

When adding a hook template, ccflow should know which event(s) it belongs to:

* `post-edit` → PostToolUse (Write/Edit)
* `end-of-turn` → Stop

The blueprint’s hook manifest should define:

```yaml
hooks_manifest:
  post-edit:
    script: hooks/post-edit.sh
    events:
      - event: PostToolUse
        commands: [Write, Edit]
  end-of-turn:
    script: hooks/end-of-turn.sh
    events:
      - event: Stop
```

## 8. Default assets: commands, agents, hooks (v1)

You’ll embed a “good-enough” baseline; users can extend later.

### Default commands (Markdown)

* `idea.md`: create spec + state JSON + (optional) ticket instructions
* `design.md`: create design doc in `docs/workflow/designs/`
* `implement.md`: implement changes + update state
* `review.md`: run validations + prepare PR steps + update state
* `release.md`: enforce gates + health check + promote steps (prints checklist)
* `status.md`: read state and show progress/gates

These should be written as Claude Code custom command markdown files.

### Default hooks (bash)

* `post-edit.sh`:

  * detect repo type (node/java/go/terraform) by file presence
  * run format commands if available
* `end-of-turn.sh`:

  * run validations (typecheck/test/build) if available
  * write a small “validation summary” to stdout
  * optionally update workflow state if ticket env var present (v1 can keep this simple)

### settings.json

* hooks wired
* conservative permissions allowlist, deny sensitive patterns
* do not try to perfectly secure everything; `doctor` and docs will recommend keeping secrets out of tree

## 9. MCP integration behavior (v1)

Do **not** attempt to fully configure MCP programmatically in v1.
Instead:

* Record choices in `workflow.yaml`
* Print “Next steps” instructions:

  * recommended MCP servers and typical setup steps
  * example environment variables / config locations if needed
    This avoids brittle cross-machine assumptions.

## 10. `status` and `doctor`

### `ccflow status`

Outputs:

* workflow name + blueprint
* topology (multi/single)
* hub path + docs path
* repos list + symlink health (OK/BROKEN)
* enabled hooks and whether scripts exist

Exit codes:

* 0: healthy
* 2: warnings (non-fatal)
* 3: errors (broken workflow)

### `ccflow doctor`

Does deeper checks:

* parse JSON for `.claude/settings.json`
* check hook executability
* check symlink points to correct target
* check marker files and required dirs exist

## 11. Error handling principles

* Always show the path that caused the issue
* Provide one concrete fix command when possible
* Never silently overwrite

## 12. Testing strategy

### Unit tests

* workspace discovery:

  * nearest marker selection
  * `--workspace` overrides
* template loading:

  * blueprint listing
  * unknown template errors
* add-* behavior:

  * no overwrite without `--force`
  * stdin/file/template sources

### Integration tests (temp dirs)

* `run web-dev` creates expected tree and symlinks
* `add-agent --stdin` writes file correctly
* `add-hook` updates settings.json correctly and keeps valid JSON

### Golden tests

* Snapshot expected generated files for each blueprint (with normalized tokens)

## 13. Build & release

* GitHub Actions:

  * on PR: lint + test
  * on tag: GoReleaser build + create release artifacts
* Homebrew distribution:

  * publish via a tap repo
  * provide install instructions in README

## 14. Milestones

### v1.0

* `run`, `list-blueprints`, `status`, `doctor`
* `add-agent`, `add-command`, `add-hook`
* `web-dev` blueprint (primary)
* symlink installer (mac/linux)
* release pipeline (GitHub Releases)

### v1.1

* `ios-dev` blueprint
* optional workflow registry and `ccflow list`
* `upgrade` (apply new defaults without overwriting user edits)
