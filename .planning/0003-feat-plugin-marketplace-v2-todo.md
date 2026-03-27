---
id: "0003"
title: "feat: Blueprint v2 — native Claude Code plugin in separate repo + Homebrew tap"
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

# feat: Blueprint v2 — native Claude Code plugin in separate repo + Homebrew tap

## Goal
Create a new `skaisser/blueprint-plugin` repo that packages Blueprint as a Claude Code native plugin — installable via `/plugin marketplace add skaisser/blueprint-plugin` + `/plugin install blueprint`. Add Homebrew tap for the Go CLI binary. Get listed in the official Anthropic marketplace. Include auto-update hook from ~/Sites/claude pattern.

## Non-Goals
- Rewriting the Go CLI (stays as-is in skaisser/blueprint)
- Changing skill prompt content (only paths/references change)
- Windows support (not yet)
- Merging claude-cli into blueprint CLI (they stay separate)
- Dropping install.sh from main repo (keep as dev/legacy option)

## Context
- **Separate repo**: `skaisser/blueprint-plugin` — clean plugin-only, no Go source/binaries (~500KB vs ~40MB)
- **Main repo**: `skaisser/blueprint` stays as development source — Go CLI, install.sh, build tooling
- **Sync**: GitHub Action in main repo auto-publishes to plugin repo on release
- Plugin system supports: `skills/*/SKILL.md`, `hooks/hooks.json` (PreToolUse, Stop, etc.), `.mcp.json`, `scripts/`
- `${CLAUDE_PLUGIN_ROOT}` env var available in hooks and scripts, points to plugin install directory
- Official marketplace: `anthropics/claude-plugins-public` — entry in marketplace.json
- Ralph-loop proves hooks work: `hooks/hooks.json` → `${CLAUDE_PLUGIN_ROOT}/hooks/stop-hook.sh`
- Context7 proves MCP works: `.mcp.json` at plugin root
- **~/Sites/claude pattern**: `hooks/skills-check.sh` auto-updates on session start by git-fetching and running install if behind — we adapt this for the plugin's auto-update hook
- **claude-cli `skills outdated`**: compares source vs deployed by mtime — useful pattern for plugin health check
- GoReleaser automates: binaries + Homebrew formula + GitHub Releases

## Phases

### Phase 1: Create Plugin Repo Structure
**Touches:** new repo `skaisser/blueprint-plugin`

- [ ] Create `skaisser/blueprint-plugin` GitHub repo with description "Blueprint SDLC — Claude Code Plugin"
- [ ] Create `.claude-plugin/plugin.json` with name `blueprint`, description (focused on SDLC value prop), author (Shirleyson Kaisser)
- [ ] Create `.mcp.json` registering Context7 (`@upstash/context7-mcp`) and Sequential Thinking (`@anthropic-ai/mcp-sequential-thinking`)
- [ ] Copy all 27 `skills/*/` directories (SKILL.md + references/) from main repo
- [ ] Copy `config/statusline.sh` and `templates/` directory
- [ ] Create initial README.md with quick-start install instructions

**Verify:** `cat .claude-plugin/plugin.json && cat .mcp.json` — valid JSON, skills/ has 27 directories

### Phase 2: Plugin Hooks (Audit + Auto-Update)
**Touches:** `hooks/` in plugin repo

- [ ] Create `hooks/hooks.json` registering `PreToolUse` hook calling `${CLAUDE_PLUGIN_ROOT}/hooks/audit-wrapper.sh`
- [ ] Create `hooks/audit-wrapper.sh` — checks if `~/.blueprint/bin/blueprint` exists; if missing, runs setup script (lazy install); then pipes stdin to `blueprint audit`
- [ ] Create `hooks/session-start.json` registering a `Stop` or startup-equivalent hook for auto-update check
- [ ] Create `hooks/auto-update.sh` — adapted from ~/Sites/claude's `skills-check.sh` pattern: on session start, check if plugin repo has newer version, notify user if update available
- [ ] Ensure all hook scripts have `chmod +x` and proper shebang

**Verify:** `json_pp < hooks/hooks.json` valid, `bash -n hooks/audit-wrapper.sh` passes, `bash -n hooks/auto-update.sh` passes

### Phase 3: Setup Script (Binary Bootstrap)
**Touches:** `scripts/setup.sh` in plugin repo

- [ ] Create `scripts/setup.sh` that detects platform (darwin-arm64, darwin-amd64, linux-amd64) and downloads latest `blueprint` binary from GitHub Releases to `~/.blueprint/bin/blueprint`
- [ ] Add `~/.blueprint/bin` to PATH in shell RC file (~/.zshrc or ~/.bashrc) if not already present
- [ ] Configure statusline in `~/.claude/settings.json` pointing to `${CLAUDE_PLUGIN_ROOT}/config/statusline.sh`
- [ ] Make script idempotent — skip steps already done, compare installed version vs release version before downloading
- [ ] Add `--force` flag to re-download even if version matches

**Verify:** `bash scripts/setup.sh && ~/.blueprint/bin/blueprint --version`

### Phase 4: GoReleaser + Homebrew Tap
**Touches:** `cli/.goreleaser.yml` in main repo, `.github/workflows/release.yml` in main repo, new repo `skaisser/homebrew-tap`

- [ ] Create `.goreleaser.yml` in main repo root with builds for darwin-arm64, darwin-amd64, linux-amd64 targeting `cli/` as source
- [ ] Add GoReleaser `brews` section pointing to `skaisser/homebrew-tap` with formula name `blueprint`
- [ ] Replace `.github/workflows/release.yml` with GoReleaser-based workflow (triggered on tag `v*`)
- [ ] Create `skaisser/homebrew-tap` GitHub repo with initial README explaining `brew tap skaisser/tap && brew install blueprint`
- [ ] Add GitHub Action secret `HOMEBREW_TAP_TOKEN` (PAT with repo scope for homebrew-tap repo)

**Verify:** `goreleaser check` passes in main repo

### Phase 5: Sync Action (Main → Plugin Repo)
**Touches:** `.github/workflows/sync-plugin.yml` in main repo

- [ ] Create `.github/workflows/sync-plugin.yml` — triggered on tag push `v*` (same as release)
- [ ] Action copies `skills/`, `config/`, `templates/` from main repo to plugin repo
- [ ] Action updates plugin repo's `plugin.json` version field to match release tag
- [ ] Action commits and pushes to `skaisser/blueprint-plugin` main branch
- [ ] Add GitHub Action secret `PLUGIN_REPO_TOKEN` (PAT with repo scope for plugin repo)

**Verify:** Manually run action, verify plugin repo gets updated files

### Phase 6: Skill Path Updates
**Touches:** `skills/start/SKILL.md`, `skills/plan/SKILL.md`, `skills/plan-approved/SKILL.md`, `skills/plan-review/SKILL.md` in main repo

- [ ] Update `/start` skill to reference templates from `${CLAUDE_PLUGIN_ROOT}/templates/` with fallback to `~/.blueprint/templates/`
- [ ] Update `/plan` skill to reference `plan-template.md` via relative path (bundled with skill) — already works as "references/plan-template.md" relative to skill
- [ ] Update skills that hardcode `~/.claude/skills/{name}/references/` to use `${CLAUDE_PLUGIN_ROOT}/skills/{name}/references/` with fallback
- [ ] Add path-resolution helper comment at top of skills that reference external files

**Verify:** `grep -rn "CLAUDE_PLUGIN_ROOT\|~/.blueprint\|~/.claude/skills" skills/ | head -20` shows dual-path references

### Phase 7: README + Marketplace Submission
**Touches:** `README.md` in both repos, external PR to `anthropics/claude-plugins-public`

- [ ] Write plugin repo README.md: hero section, 2-second install, feature list, "What's included" (27 skills, audit hook, MCP servers, statusline), link to main repo for contributors
- [ ] Update main repo README.md: add "Install via Plugin (Recommended)" as primary method, keep curl|bash as secondary, add Homebrew as tertiary
- [ ] Draft marketplace submission PR to `anthropics/claude-plugins-public` — add entry with source `skaisser/blueprint-plugin`, category `development`, description focused on SDLC pipeline
- [ ] Add badges to plugin repo: stars, license, marketplace listing

**Verify:** Both READMEs render correctly, marketplace JSON entry is valid

### Phase 8: Integration Testing
**Touches:** verification across all install methods

- [ ] Test fresh plugin install: `/plugin marketplace add skaisser/blueprint-plugin` → `/plugin install blueprint` → verify 27 skills register as slash commands
- [ ] Test audit hook fires on PreToolUse after plugin install (write a file, check hook ran)
- [ ] Test MCP servers auto-register (Context7 and Sequential Thinking appear in tool list)
- [ ] Test setup script bootstraps binary on first audit hook trigger when binary is missing
- [ ] Test `brew tap skaisser/tap && brew install blueprint` installs CLI, `blueprint --version` works
- [ ] Test auto-update hook detects when plugin is behind and notifies user
- [ ] Test backward compat: main repo's install.sh still works independently for non-plugin users

**Verify:** All 7 tests pass, plugin installs cleanly on a fresh Claude Code session

## Acceptance
- [ ] `/plugin marketplace add skaisser/blueprint-plugin` + `/plugin install blueprint` installs all 27 skills as working slash commands
- [ ] PreToolUse audit hook works via plugin hooks.json — no manual settings.json needed
- [ ] Context7 and Sequential Thinking MCP servers auto-register via .mcp.json
- [ ] `brew tap skaisser/tap && brew install blueprint` installs the Go CLI binary
- [ ] Setup script auto-downloads binary on first use if not installed via brew
- [ ] Auto-update hook notifies user when plugin has updates available
- [ ] Statusline works from plugin path
- [ ] Existing install.sh users in main repo are not broken
- [ ] GitHub Action syncs main repo changes to plugin repo on release
- [ ] Blueprint is submitted to official Anthropic plugin marketplace
