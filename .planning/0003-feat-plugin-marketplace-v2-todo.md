---
id: "0003"
title: "feat: Transform Blueprint into Claude Code native plugin + Homebrew tap"
type: feat
status: todo
project: blueprint
branch: feat/plugin-marketplace-v2
base: main
tags: [plugin, marketplace, homebrew, distribution, v2]
linear: null
backlog: null
created: "27/03/2026 18:30"
completed: null
pr: null
session: null
---

# feat: Transform Blueprint into Claude Code native plugin + Homebrew tap

## Goal
Convert Blueprint from a curl|bash installer into a Claude Code native plugin installable via `/plugin marketplace add skaisser/blueprint` + `/plugin install blueprint`. Add Homebrew tap for the Go CLI binary. Get listed in the official Anthropic marketplace for maximum discoverability.

## Non-Goals
- Rewriting the Go CLI (stays as-is)
- Changing skill content/behavior (only paths change)
- Dropping install.sh entirely (keep as legacy/CI option)
- Windows support (not yet)
- Creating a separate repo (transform the existing one)

## Context
- Plugin system supports: `skills/*/SKILL.md`, `hooks/hooks.json` (PreToolUse, Stop, etc.), `.mcp.json`, `scripts/setup.sh`
- Current `skills/` directory structure already matches plugin format exactly
- Go CLI binary can't be bundled — needs external distribution (Homebrew tap + setup script download)
- `${CLAUDE_PLUGIN_ROOT}` env var available in hooks and scripts, points to plugin install directory
- Official marketplace is at `anthropics/claude-plugins-public` — entry in marketplace.json
- Ralph-loop plugin proves hooks work: `hooks/hooks.json` → `${CLAUDE_PLUGIN_ROOT}/hooks/stop-hook.sh`
- Context7 plugin proves MCP registration works: `.mcp.json` at plugin root
- GoReleaser is the standard Go release automation tool (binaries + Homebrew formula)
- Current `release.yml` GitHub Action builds binaries on tag push — will be replaced by GoReleaser

## Phases

### Phase 1: Plugin Scaffold
**Touches:** `.claude-plugin/`, `hooks/`, `.mcp.json`

- [ ] Create `.claude-plugin/plugin.json` with name `blueprint`, description, author (Shirleyson Kaisser)
- [ ] Create `hooks/hooks.json` registering `PreToolUse` hook with `".*"` matcher calling `${CLAUDE_PLUGIN_ROOT}/hooks/audit-wrapper.sh`
- [ ] Create `hooks/audit-wrapper.sh` — checks if `~/.blueprint/bin/blueprint` exists, if not runs setup script first, then pipes stdin to `blueprint audit`
- [ ] Create `.mcp.json` registering Context7 (`@upstash/context7-mcp`) and Sequential Thinking (`@anthropic-ai/mcp-sequential-thinking`)
- [ ] Add `.claude-plugin/` to the repo (not gitignored)

**Verify:** `cat .claude-plugin/plugin.json && cat hooks/hooks.json && cat .mcp.json` — all valid JSON

### Phase 2: Setup Script (First-Run Bootstrap)
**Touches:** `scripts/setup.sh`

- [ ] Create `scripts/setup.sh` that detects platform (darwin-arm64, darwin-amd64, linux-amd64)
- [ ] Download latest binary from GitHub Releases (`skaisser/blueprint`) to `~/.blueprint/bin/blueprint`
- [ ] Add `~/.blueprint/bin` to PATH in shell RC file if not already present
- [ ] Configure statusline in `~/.claude/settings.json` (statusLine command pointing to `${CLAUDE_PLUGIN_ROOT}/config/statusline.sh`)
- [ ] Make setup script idempotent (skip steps already done, don't overwrite newer versions)
- [ ] Add version check — skip download if installed version >= release version

**Verify:** `bash scripts/setup.sh && which blueprint && blueprint --version`

### Phase 3: GoReleaser + Homebrew Tap
**Touches:** `cli/.goreleaser.yml`, `.github/workflows/release.yml`, new repo `skaisser/homebrew-tap`

- [ ] Create `cli/.goreleaser.yml` with builds for darwin-arm64, darwin-amd64, linux-amd64 and Homebrew tap config
- [ ] Replace `.github/workflows/release.yml` with GoReleaser-based workflow (triggered on tag `v*`)
- [ ] Create `skaisser/homebrew-tap` GitHub repo with initial README
- [ ] Add GoReleaser `brews` section pointing to `skaisser/homebrew-tap`
- [ ] Test with a dry-run: `cd cli && goreleaser check`

**Verify:** `goreleaser check` passes, homebrew-tap repo exists

### Phase 4: Skill & Template Path Updates
**Touches:** `skills/start/SKILL.md`, `skills/plan/SKILL.md`, `skills/plan-approved/SKILL.md`, `skills/plan-review/SKILL.md`

- [ ] Update `/start` skill to copy templates from `${CLAUDE_PLUGIN_ROOT}/templates/` with fallback to `~/.blueprint/templates/` (backward compat)
- [ ] Update skills that reference `~/.claude/skills/{name}/references/` to use relative paths or `${CLAUDE_PLUGIN_ROOT}/skills/{name}/references/`
- [ ] Update `/plan` skill reference path for `plan-template.md` to work from plugin root
- [ ] Verify `config/statusline.sh` works when referenced via `${CLAUDE_PLUGIN_ROOT}/config/statusline.sh`
- [ ] Add `${CLAUDE_PLUGIN_ROOT}` documentation comment to skills that use external file references

**Verify:** `grep -r "CLAUDE_PLUGIN_ROOT" skills/` shows updated references

### Phase 5: README + Marketplace Submission
**Touches:** `README.md`, external PR to `anthropics/claude-plugins-public`

- [ ] Rewrite README.md install section: primary = `/plugin install blueprint`, secondary = `brew install`, tertiary = `curl|bash` (legacy)
- [ ] Add "Plugin vs Homebrew vs Manual" comparison table
- [ ] Add migration guide section for existing install.sh users
- [ ] Update roadmap section (v2.0 = plugin marketplace)
- [ ] Prepare marketplace submission: draft PR adding blueprint entry to `anthropics/claude-plugins-public` marketplace.json with source pointing to `skaisser/blueprint`

**Verify:** README renders correctly on GitHub, marketplace entry JSON is valid

### Phase 6: Integration Testing
**Touches:** test scripts, verification

- [ ] Test fresh plugin install: `/plugin marketplace add skaisser/blueprint` → `/plugin install blueprint` → verify 27 skills load
- [ ] Test audit hook fires on tool calls after plugin install
- [ ] Test MCP servers (Context7, Sequential Thinking) auto-register after plugin install
- [ ] Test setup script downloads binary correctly on first audit hook trigger
- [ ] Test `brew install skaisser/tap/blueprint` installs CLI binary
- [ ] Test backward compat: existing install.sh still works independently
- [ ] Test statusline displays correctly from plugin path

**Verify:** All 7 integration tests pass manually

## Acceptance
- [ ] `/plugin marketplace add skaisser/blueprint` + `/plugin install blueprint` installs all 27 skills
- [ ] PreToolUse audit hook works via plugin hooks.json (no manual settings.json edit needed)
- [ ] Context7 and Sequential Thinking MCP servers auto-register via .mcp.json
- [ ] `brew install skaisser/tap/blueprint` installs the Go CLI binary
- [ ] Setup script auto-downloads binary on first use if not installed via brew
- [ ] Statusline works from plugin path
- [ ] Existing install.sh users are not broken
- [ ] Blueprint is listed in the official Anthropic plugin marketplace
