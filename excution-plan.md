Build the CLI tool “ccflow” end-to-end (code + tests + docs + CI + release pipeline), implementing ALL features in v1.0 and v1.1.

Non-negotiables:
- Target: macOS/Linux only for v1 (symlink install mode). Design abstractions so Windows can be added later (copy install mode), but do not ship Windows in v1.
- Default topology: multi-repo hub pattern. Must also support single-repo + later expansion.
- The tool is install-once CLI, Go-based, single binary. Homebrew distribution supported.
- “Deep search” requirement: you must research Claude Code’s latest official docs and best practices for:
  - custom slash commands (.claude/commands/*.md),
  - hooks wiring in settings.json and supported events,
  - permissions model and safe allow/deny patterns,
  - MCP integration best-practice guidance,
  and design robust templates accordingly.
  Cite sources in the repo docs as links (no need to paste large excerpts).

High-level deliverables:
1) ccflow CLI implemented with:
   - run/init wizard (supports blueprint selection)
   - list-blueprints
   - status
   - doctor
   - add-agent / add-command / add-hook (template OR stdin OR file OR print; safe overwrite semantics)
   - (v1.1) upgrade (apply new defaults without clobbering user edits)
   - (v1.1) optional workflow registry + ccflow list (convenience)
   - (v1.1) expand single-repo -> multi-repo (command or upgrade path)
2) Built-in blueprints:
   - web-dev (v1.0)
   - ios-dev (v1.1)
3) Embedded template assets for each blueprint:
   - .claude/agents/*.md
   - .claude/commands/*.md
   - .claude/hooks/*.sh
   - .claude/settings.json (or template)
   NOTE: You must research and craft “best templates” (deep search).
4) Deterministic generation:
   - multi-repo creates: workflow-hub/.claude + workflow-hub/workflow.yaml + docs/workflow/{state,designs}/ + symlinks .claude into repos
   - single-repo creates: repo/.claude + repo/.ccflow/workflow.yaml + repo/docs/workflow/{state,designs}/
5) Tests:
   - unit tests for discovery, blueprints, add-*, settings.json mutation
   - integration tests in temp dirs for run + symlink install + doctor
   - golden tests for generated trees (normalized paths)
6) CI/release:
   - PR CI: lint + test
   - Tag CI: goreleaser builds and publishes release artifacts
   - Homebrew tap publishing (preferred using goreleaser-supported approach)
7) Docs:
   - README: install, quickstart (ccflow run web-dev), managing agents/hooks/commands, how to use stdin/file
   - docs/ARCHITECTURE.md: workspace discovery rules, templates philosophy, safety considerations
   - docs/BLUEPRINTS.md: what web-dev/ios-dev do
   - docs/RELEASING.md: how to cut a release and update tap
   - docs/MCP.md: recommended MCP setup steps (do not try to automate MCP fully in v1)

Implementation plan (execute in order):

PHASE 0 — Repo bootstrap (day 0)
A) Create repo skeleton
- Initialize git repo “ccflow”
- Add Go module, choose Go version (stable)
- Choose CLI framework: Cobra
- Choose prompt library: survey (or minimal prompt if you prefer)
- Add common tooling:
  - golangci-lint config
  - editorconfig
  - Makefile (build/test/lint)
  - .gitignore

Acceptance:
- `go test ./...` passes on clean repo
- `ccflow --help` works with placeholder commands

PHASE 1 — Core architecture & types (v1.0 foundation)
B) Define internal packages (keep simple, no over-engineering):
- /cmd (cobra commands)
- /internal/workspace (discovery, markers, overrides)
- /internal/blueprint (embedded assets + blueprint manifest parsing)
- /internal/generator (write files from templates, ensure dirs)
- /internal/installer (symlink install mode)
- /internal/mutator (add-agent/command/hook + settings.json mutation)
- /internal/validator (status/doctor checks)
- /internal/util (io helpers: read stdin, safe write, chmod, relative symlink)

C) Define on-disk markers and config:
- Multi-repo marker: <root>/workflow-hub/workflow.yaml
- Single-repo marker: <root>/.ccflow/workflow.yaml
- workflow.yaml schema in YAML v3
- Config must record:
  - version, name, blueprint, topology
  - hub/docs paths
  - state dirs
  - repos list (name/path/kind)
  - hooks/gates enabled booleans
  - mcp preferences (for guidance printing only)

Acceptance:
- workspace resolver can locate root reliably:
  - precedence: --workspace > CCFLOW_WORKSPACE > walk-up markers (nearest)
- parsing/writing workflow.yaml works with tests

PHASE 2 — Blueprints system (v1.0: web-dev)
D) Implement embedded blueprint assets using Go embed:
- internal/blueprints/web-dev/blueprint.yaml
- internal/blueprints/web-dev/assets/.claude/...
- internal/blueprints/web-dev/templates/...
Blueprint manager requirements:
- `ccflow list-blueprints` prints id + display name + description
- `ccflow run web-dev` loads blueprint

Deep search requirement (do this BEFORE finalizing assets):
- Search official Claude Code docs for:
  - custom slash commands format/location
  - hook event names and structure in settings.json
  - permissions allow/deny patterns
  - recommended project vs user settings patterns
- Also search reputable community examples (avoid random gists unless widely referenced)
- Summarize findings in docs/ARCHITECTURE.md with links.

Acceptance:
- blueprint loader returns:
  - defaults for repos/agents/commands/hooks
  - mapping of hook name -> settings.json registration info (event, commands)
- Template token rendering works (OrgName, WorkflowName, dirs)

PHASE 3 — `ccflow run` wizard (v1.0)
E) Implement `ccflow run [blueprint]` interactive flow:
Prompt set (minimum):
1) Workflow name (default derived from folder)
2) Mode:
   - Use existing repos
   - Generate repos
3) Topology:
   - Multi-repo (default)
   - Single-repo
4) If existing repos: gather repo paths (allow multi-select from auto-detected git repos under chosen root)
5) If generate repos: confirm repo set from blueprint (allow user to add/remove)
6) Hooks enabled? (default yes)
7) Gates enabled? (default yes)
8) Tracker choice for MCP guidance: Linear/Jira/None
9) VCS choice for MCP guidance: GitHub/GitLab/None

Then execute:
- Create required folders
- Write hub `.claude` assets (agents/commands/hooks/settings.json)
- Write docs workflow dirs (state/designs)
- Write workflow.yaml marker in correct location
- Install `.claude` into repos:
  - multi-repo: create `.claude` symlink in each repo -> hub/.claude (relative symlink preferred)
  - single-repo: no symlink needed (local .claude exists)

Safety:
- Never overwrite existing `.claude` unless user confirms AND passes explicit `--force` (or a wizard confirm + explicit typed acknowledgment).
- If `.claude` exists and is not a symlink to expected hub, abort with remediation.

Acceptance:
- Fresh run in empty workspace creates correct multi-repo tree
- Run in existing repos attaches by symlink without clobbering
- Output prints next steps including:
  - how to open Claude Code and use /idea etc.
  - MCP recommended setup steps (commands/instructions only)

PHASE 4 — add-agent / add-command / add-hook (v1.0)
F) Implement `add-*` commands with content sourcing:
- `ccflow add-agent <name> [--file PATH | --stdin | --print] [--force]`
- same for add-command and add-hook
Rules:
- If --print: print template content to stdout (no writes)
- If --file: write file content as-is
- If --stdin: read stdin until EOF; write content
- Else: use built-in template if exists; error if not found with suggestion to use --stdin/--file
Overwrite:
- If destination exists, refuse unless --force

Hook special behavior:
- Writes script to hub/.claude/hooks/<name>.sh (chmod +x)
- Updates hub/.claude/settings.json to register hook events as defined in blueprint hook manifest
- Settings.json must remain valid JSON; preserve existing unrelated settings

Acceptance:
- Can add default devops-agent via template
- Can add custom agent via stdin
- Can add hook via file and see settings.json updated properly

PHASE 5 — status + doctor (v1.0)
G) Implement validators:
- `ccflow status`:
  - prints: workflow name, blueprint, topology, hub/docs paths
  - prints repos and .claude install status (OK/BROKEN/MISSING)
  - prints hooks enabled and whether scripts exist
  - exit codes: 0 healthy, 2 warnings, 3 errors
- `ccflow doctor`:
  - parses settings.json, checks hook executability, symlink correctness, marker presence
  - prints actionable remediation steps
  - non-zero exit if errors

Acceptance:
- doctor catches broken symlink
- doctor catches non-executable hook script
- status works from any subdirectory (workspace discovery)

PHASE 6 — Template quality (v1.0 completion gate)
H) Deep-search and craft “best templates” for web-dev:
- Agents: product, architect, implementation, review, devops + (frontend/test)
- Commands: idea/design/implement/review/release/status
- Hooks: post-edit/end-of-turn/safe-bash
Template requirements:
- Must be repo-agnostic but include sensible detection (node, java, terraform)
- Must reference the docs workflow state directories and gates policy
- Must encourage small PRs and systematic validation
- Must be safe by default (no dangerous rm -rf patterns; no writing secrets)
- Commands should define clear outputs/artifacts and how to update state files
- end-of-turn hook should run lightweight validation and fail loudly with instructions
- Document template assumptions in docs/ARCHITECTURE.md

Acceptance:
- web-dev blueprint feels “ready to use” without edits
- commands are coherent and match hook behavior
- docs explain what templates do and how to customize

PHASE 7 — CI + Release (v1.0)
I) Add GitHub Actions:
- PR workflow: go test + golangci-lint
- Release workflow: on tag v*, run goreleaser
J) Add GoReleaser config:
- Build for darwin/linux amd64/arm64
- Generate checksums
- Publish GitHub Release
- Publish Homebrew formula or cask via a tap repo (preferred modern approach)
K) Write docs:
- README: install from release, brew install instructions (using tap), quickstart
- RELEASING.md: tag process, tap requirements

Acceptance:
- `goreleaser --snapshot --clean` works locally
- Tagging publishes release artifacts (dry-run acceptable)
- Brew install instructions are correct and tested once manually

Stop here ONLY after v1.0 is complete and working.

========================================================
v1.1 FEATURES — implement in same build (continue)

PHASE 8 — ios-dev blueprint (v1.1)
L) Add ios-dev blueprint:
- blueprint.yaml + assets + templates
- repo defaults (e.g., ios-app, docs, maybe backend optional)
- hooks tailored for iOS tooling where possible (xcodebuild, swiftformat if installed)
- commands same set (idea/design/implement/review/release/status), but language-specific guidance in templates
Deep search requirement:
- research best practices for iOS project validation hooks and safe minimal checks
- keep hooks non-breaking if tools absent (detect and skip with warning)

Acceptance:
- `ccflow run ios-dev` produces correct tree and usable templates
- doctor/status work the same

PHASE 9 — Workflow registry + `ccflow list` (v1.1)
M) Implement optional registry:
- file: ~/.ccflow/registry.json
- on `run`, add/update entry: {name, path, blueprint, created_at, last_used_at}
- `ccflow list` prints registered workflows
Notes:
- registry is convenience only; all commands must still work via marker discovery even without registry

Acceptance:
- list shows workflows created
- removing registry does not break anything

PHASE 10 — upgrade command (v1.1)
N) Implement `ccflow upgrade`:
Goal: apply new default templates/settings to an existing workflow WITHOUT clobbering user edits.

Approach:
- For generated files, add a header marker comment in templates like:
  - “ccflow-managed: true”
  - “ccflow-template: web-dev/devops-agent@v1”
- Maintain a manifest in workflow.yaml or a managed.json file listing which files are managed and their template ids + hashes.
Upgrade rules:
- If file is ccflow-managed and unchanged by user (hash matches), overwrite with new version
- If file is ccflow-managed but user-modified, do NOT overwrite; write a .new file and print a diff suggestion
- If file is user-created (not managed), never touch it
Also upgrade settings.json:
- merge hook registrations if missing
- do not delete user-added hooks/permissions
- keep JSON valid and stable order if possible

Acceptance:
- upgrade updates unmodified templates
- upgrade preserves modified templates and writes *.new
- upgrade does not break settings.json

PHASE 11 — expand single-repo to multi-repo (v1.1)
O) Implement expansion:
- command: `ccflow expand --to multi-repo` (or integrate into upgrade with prompt)
Behavior:
- Create workflow-hub and move/copy existing .claude there (respect managed rules)
- Create docs repo/folder if not present
- Replace local .claude with symlink to hub/.claude
- Update workflow.yaml topology and paths
Safety:
- require confirmation; never delete anything without explicit user action
Acceptance:
- single-repo workflow can become multi-repo safely
- status/doctor pass after expansion

PHASE 12 — Final polish + documentation (v1.1)
P) Update docs:
- BLUEPRINTS.md includes ios-dev
- ARCHITECTURE.md includes upgrade/managed-files rules
- README includes v1.1 commands (list/upgrade/expand)
Q) Versioning:
- tag v1.0.0 after phase 7
- tag v1.1.0 after phase 12 (or cut both tags sequentially as appropriate)

========================================================
Implementation details & coding standards

- Use Go embed for blueprint assets.
- Use relative symlinks when possible; fall back to absolute if needed.
- Ensure hook scripts are chmod +x after writing.
- Provide helpful errors with remediation. No silent failures.
- Keep output concise but actionable.
- Add unit + integration tests. Use temp dirs and avoid network calls.

Definition of Done (v1.1):
- `ccflow run web-dev` works end-to-end in a fresh temp workspace.
- `ccflow run ios-dev` works end-to-end.
- add-agent/command/hook work with template, stdin, and file modes.
- status and doctor produce useful output and correct exit codes.
- registry list works.
- upgrade works with managed/unmanaged behavior.
- expand works (single -> multi).
- CI passes.
- goreleaser snapshot passes.
- Docs are complete and installation instructions are correct.
