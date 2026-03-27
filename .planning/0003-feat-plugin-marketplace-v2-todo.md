---
id: "0003"
title: "feat: Blueprint v2 — native Claude Code plugin in separate repo + Homebrew tap"
type: feat
status: approved
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
strategy: parallel-subagents
reviews:
  - "Phase 1: added LICENSE file task"
  - "Phase 2: session-start hook type TBD — may not exist in plugin system, added fallback note"
  - "Phase 4/5: require GitHub PATs — user must create before execution"
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

## Tech Stack Versions
- Go 1.24 (CLI binary)
- Bash 5.x (installer, hooks, scripts)
- GitHub Actions (CI/CD)
- GoReleaser v2 (release automation)

## Phases

### Phase 1: Create Plugin Repo Structure
**Touches:** new repo `skaisser/blueprint-plugin`

- [ ] [H] Create `skaisser/blueprint-plugin` GitHub repo via `gh repo create` with description "Blueprint SDLC — Claude Code Plugin"
- [ ] [H] Create `.claude-plugin/plugin.json` with name `blueprint`, description (focused on SDLC value prop), author (Shirleyson Kaisser)
- [ ] [H] Create `.mcp.json` registering Context7 (`@upstash/context7-mcp`) and Sequential Thinking (`@anthropic-ai/mcp-sequential-thinking`)
- [ ] [H] Copy all 27 `skills/*/` directories (SKILL.md + references/) from main repo to plugin repo
- [ ] [H] Copy `config/statusline.sh` and `templates/` directory from main repo to plugin repo
- [ ] [H] Add Apache 2.0 LICENSE file to plugin repo
- [ ] [H] Create initial README.md with quick-start install instructions

**Verify:** `cat .claude-plugin/plugin.json && cat .mcp.json` — valid JSON, skills/ has 27 directories

### Phase 2: Plugin Hooks (Audit + Auto-Update)
**Touches:** `hooks/` in plugin repo

- [ ] [H] Create `hooks/hooks.json` registering `PreToolUse` hook calling `${CLAUDE_PLUGIN_ROOT}/hooks/audit-wrapper.sh`
- [ ] [S] Create `hooks/audit-wrapper.sh` — checks if `~/.blueprint/bin/blueprint` exists; if missing, runs setup script (lazy install); then pipes stdin to `blueprint audit`. Must handle: binary not found gracefully (exit 0), setup script failure (exit 0), and successful audit (forward exit code)
- [ ] [H] Add auto-update check to hooks.json — register a `Stop` hook calling `${CLAUDE_PLUGIN_ROOT}/hooks/auto-update.sh` (NOTE: if Stop hook doesn't work for this purpose, fall back to embedding version check in audit-wrapper.sh with 24h cache)
- [ ] [S] Create `hooks/auto-update.sh` — adapted from ~/Sites/claude's `skills-check.sh`: check GitHub API for latest release, compare with installed version, print stderr notification if update available. Cache result for 24h in `~/.blueprint/.update-check`
- [ ] [H] Ensure all hook scripts have `chmod +x` set via git (use `git update-index --chmod=+x`)

**Verify:** `json_pp < hooks/hooks.json` valid, `bash -n hooks/audit-wrapper.sh` passes, `bash -n hooks/auto-update.sh` passes

### Phase 3: Setup Script (Binary Bootstrap)
**Touches:** `scripts/setup.sh` in plugin repo

- [ ] [S] Create `scripts/setup.sh` that detects platform (darwin-arm64, darwin-amd64, linux-amd64) and downloads latest `blueprint` binary from GitHub Releases (`skaisser/blueprint`) to `~/.blueprint/bin/blueprint`
- [ ] [H] Add `~/.blueprint/bin` to PATH in shell RC file (~/.zshrc or ~/.bashrc) if not already present
- [ ] [S] Configure statusline in `~/.claude/settings.json` — smart-merge statusLine command pointing to `${CLAUDE_PLUGIN_ROOT}/config/statusline.sh` (read existing, merge, write back — never overwrite)
- [ ] [H] Make script idempotent — skip steps already done, compare installed version vs release version before downloading
- [ ] [H] Add `--force` flag to re-download even if version matches

**Verify:** `bash scripts/setup.sh && ~/.blueprint/bin/blueprint --version`

### Phase 4: GoReleaser + Homebrew Tap
**Touches:** `.goreleaser.yml` in main repo, `.github/workflows/release.yml` in main repo, new repo `skaisser/homebrew-tap`

- [ ] [S] Create `.goreleaser.yml` in main repo root with builds for darwin-arm64, darwin-amd64, linux-amd64 — source dir `cli/`, binary name `blueprint`, ldflags for version injection
- [ ] [S] Add GoReleaser `brews` section pointing to `skaisser/homebrew-tap` with formula name `blueprint`, install instructions, and test block
- [ ] [S] Replace `.github/workflows/release.yml` with GoReleaser-based workflow (triggered on tag `v*`, uses `goreleaser/goreleaser-action@v6`)
- [ ] [H] Create `skaisser/homebrew-tap` GitHub repo via `gh repo create` with initial README explaining `brew tap skaisser/tap && brew install blueprint`
- [ ] [H] User action: create GitHub PAT `HOMEBREW_TAP_TOKEN` with repo scope, add as secret to skaisser/blueprint repo

**Verify:** `goreleaser check` passes in main repo

### Phase 5: Sync Action (Main → Plugin Repo)
**Touches:** `.github/workflows/sync-plugin.yml` in main repo

- [ ] [S] Create `.github/workflows/sync-plugin.yml` — triggered on tag push `v*` (runs after release). Checks out both repos, copies `skills/`, `config/`, `templates/`, `hooks/` (from plugin source) to plugin repo
- [ ] [S] Action updates plugin repo's `plugin.json` version field to match release tag using `jq`
- [ ] [H] Action commits and pushes to `skaisser/blueprint-plugin` main branch with message `🔄 ci: sync from blueprint v{tag}`
- [ ] [H] User action: create GitHub PAT `PLUGIN_REPO_TOKEN` with repo scope, add as secret to skaisser/blueprint repo

**Verify:** Manually trigger action via `workflow_dispatch`, verify plugin repo gets updated files

### Phase 6: Skill Path Updates
**Touches:** `skills/start/SKILL.md`, `skills/plan/SKILL.md`, `skills/plan-approved/SKILL.md`, `skills/plan-review/SKILL.md` in main repo

- [ ] [H] Update `/start` skill to reference templates from `${CLAUDE_PLUGIN_ROOT}/templates/` with fallback to `~/.blueprint/templates/` and `~/.claude/templates/`
- [ ] [H] Update `/plan` skill to reference `plan-template.md` via relative path (bundled with skill) — verify "references/plan-template.md" works from both plugin root and ~/.claude/skills/
- [ ] [H] Update skills that hardcode `~/.claude/skills/{name}/references/` to use `${CLAUDE_PLUGIN_ROOT}/skills/{name}/references/` with fallback
- [ ] [H] Add path-resolution helper comment at top of skills that reference external files

**Verify:** `grep -rn "CLAUDE_PLUGIN_ROOT\|~/.blueprint\|~/.claude/skills" skills/ | head -20` shows dual-path references

### Phase 7: README + Marketplace Submission
**Touches:** `README.md` in both repos, external PR to `anthropics/claude-plugins-public`

- [ ] [S] Write plugin repo README.md: hero section with badge, 2-second install command, feature list ("What's included": 27 skills, audit hook, MCP servers, statusline), "How it works" section, link to main repo for contributors
- [ ] [S] Update main repo README.md: add "Install via Plugin (Recommended)" as primary method, keep curl|bash as secondary, add Homebrew as tertiary, add migration guide for existing users
- [ ] [S] Draft marketplace submission PR to `anthropics/claude-plugins-public` — add entry to marketplace.json with source `skaisser/blueprint-plugin`, category `development`, description focused on SDLC pipeline value prop
- [ ] [H] Add badges to plugin repo: GitHub stars, license, marketplace listing link

**Verify:** Both READMEs render correctly on GitHub, marketplace JSON entry is valid

### Phase 8: Integration Testing
**Touches:** verification across all install methods

- [ ] [H] Test fresh plugin install: `/plugin marketplace add skaisser/blueprint-plugin` → `/plugin install blueprint` → verify 27 skills register as slash commands
- [ ] [H] Test audit hook fires on PreToolUse after plugin install (write a file, check hook ran)
- [ ] [H] Test MCP servers auto-register (Context7 and Sequential Thinking appear in tool list)
- [ ] [H] Test setup script bootstraps binary on first audit hook trigger when binary is missing
- [ ] [H] Test `brew tap skaisser/tap && brew install blueprint` installs CLI, `blueprint --version` works
- [ ] [H] Test auto-update hook detects when plugin is behind and notifies user
- [ ] [H] Test backward compat: main repo's install.sh still works independently for non-plugin users

**Verify:** All 7 tests pass, plugin installs cleanly on a fresh Claude Code session

## Execution Strategy

> **Approach:** `/plan-approved` with parallel subagents (5 rounds)
> **Total Tasks:** 42 (H: 30, S: 12, O: 0)
> **Estimated Rounds:** 5 (2 parallel, 3 sequential)

### File-Touch Matrix

| Phase | Repo | Files/Dirs Touched | Depends On |
|-------|------|--------------------|------------|
| Phase 1 | NEW (blueprint-plugin) | .claude-plugin/, .mcp.json, skills/, config/, templates/, README.md, LICENSE | — |
| Phase 4 | blueprint (main) | .goreleaser.yml, .github/workflows/release.yml + NEW repo (homebrew-tap) | — |
| Phase 6 | blueprint (main) | skills/start/, skills/plan/, skills/plan-approved/, skills/plan-review/ | — |
| Phase 2 | blueprint-plugin | hooks/ | Phase 1 |
| Phase 3 | blueprint-plugin | scripts/ | Phase 1 |
| Phase 5 | blueprint (main) | .github/workflows/sync-plugin.yml | Phase 1, Phase 4 |
| Phase 7 | Both repos | README.md (both), external PR | Phase 1-6 |
| Phase 8 | Both repos | verification only | Phase 1-7 |

### Round 1: Phase 1 + Phase 4 + Phase 6 → Parallel Subagents (3 workers, dispatched together)
Independent workstreams — Phase 1 creates new plugin repo, Phase 4 works on main repo CI/release, Phase 6 updates main repo skills. Zero file overlap.

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 1: Plugin repo scaffold | Sonnet | 1.1-1.7 (7x[H]) | File creation/copying only |
| Phase 4: GoReleaser + Homebrew | Opus | 4.1-4.5 (3x[S] + 2x[H]) | GoReleaser config needs precision |
| Phase 6: Skill path updates | Sonnet | 6.1-6.4 (4x[H]) | Simple text replacements |

### Round 2: Phase 2 + Phase 3 → Parallel Subagents (2 workers, dispatched together)
Both in plugin repo but different directories (hooks/ vs scripts/). Depends on Round 1 (Phase 1).

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 2: Plugin hooks | Opus | 2.1-2.5 (2x[S] + 3x[H]) | Audit wrapper needs careful error handling |
| Phase 3: Setup script | Opus | 3.1-3.5 (2x[S] + 3x[H]) | Platform detection + settings merge |

### Round 3: Phase 5 → Single Subagent (depends on Round 1 + Round 2)

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 5: Sync action | Opus | 5.1-5.4 (2x[S] + 2x[H]) | Cross-repo GitHub Action |

### Round 4: Phase 7 → Single Subagent (depends on all previous)

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 7: README + marketplace | Opus | 7.1-7.4 (3x[S] + 1x[H]) | Marketing copy + marketplace PR |

### Round 5: Phase 8 → Leader Direct (user-driven manual testing)
All 7 integration tests are manual verification steps — user runs them interactively.

Tasks: 8.1-8.7 (7x[H]) — manual verification, no code changes.

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
