---
id: "0002"
title: "feat: Smart installer for Blueprint SDLC"
type: feat
status: completed
project: blueprint
branch: feat/smart-installer
base: main
tags: [installer, setup, onboarding, gum, cross-platform]
linear: null
backlog: null
created: "24/03/2026 08:05"
completed: "24/03/2026 08:30"
pr: null
session: "848fba2a-8d4d-45e0-82b3-520ce5795676"
strategy: single-subagent
reviews:
  - "R1: All phases touch install.sh sequentially — zero parallelism, single subagent is optimal"
  - "R2: Mix of [H] and [S] tasks — use Opus for the single subagent"
  - "R3: Reference ~/Sites/claude/install.sh for proven patterns (dir creation, skills copy, binary detect, settings merge)"
---

# feat: Smart installer for Blueprint SDLC

## Goal
Create a beautiful, cross-platform installer (`install.sh`) that works via one-liner (`curl | bash`) or locally from a cloned repo. It installs the Blueprint CLI binary, 27 skills, templates, and registers the audit hook in Claude Code — without touching or deleting any existing user files.

## Non-Goals
- Project-level initialization (that's `/start` skill's job)
- Installing third-party tools (firecrawl, hass-cli, etc.)
- 1Password or secret injection
- Obsidian setup (deferred to v1.1)
- Windows support

## Context
- `~/Sites/claude/install.sh` — private installer to use as reference pattern (directory creation, skills copy, binary detection, settings merge, PATH setup)
- `templates/` — .githooks/, .github/workflows/, CLAUDE.md template, blueprint/ scaffold — these get copied to `~/.blueprint/templates/` for `/start` to reference
- `skills/` — 27 skill folders with SKILL.md + optional references/ — copied to `~/.claude/skills/`
- `.github/workflows/release.yml` — builds binaries as `blueprint-darwin-arm64`, `blueprint-darwin-amd64`, `blueprint-linux-amd64` and attaches to GitHub Releases
- `~/.claude/settings.json` — Claude Code settings where PreToolUse hook is registered. Must MERGE, never overwrite
- GitHub Releases API: `https://api.github.com/repos/skaisser/blueprint/releases/latest` returns `tag_name` and `assets[].browser_download_url`
- Claude Code discovers skills at `~/.claude/skills/{name}/SKILL.md`
- gum (charmbracelet/gum) for TUI — optional, fall back to plain prompts if unavailable
- Constraint: NEVER delete existing user files. Only add new skills, merge settings, create directories. Existing skills with same name get updated (overwritten) but nothing else is removed.

## Phases

### Phase 1: Installer Core Framework
**Touches:** `install.sh`
- [x] [S] Create `install.sh` with shebang, `set -euo pipefail`, and two-mode detection (local repo vs curl pipe) using `[ -d "$SCRIPT_DIR/skills" ]` check ✅ 24/03/2026 08:19
- [x] [H] Add platform detection function: `detect_platform()` returning OS (darwin/linux) and ARCH (arm64/amd64) with `uname -s` and `uname -m` mapping ✅ 24/03/2026 08:19
- [x] [H] Add shell RC detection (`.zshrc` / `.bashrc`) for PATH injection later ✅ 24/03/2026 08:19
- [x] [S] Add gum bootstrap: on macOS try `brew install gum`, on Linux download binary from charmbracelet/gum releases. If gum unavailable, set `HAS_GUM=false` for fallback mode ✅ 24/03/2026 08:19
- [x] [H] Add styled banner using `gum style` (or plain ASCII if no gum): show "BLUEPRINT SDLC" header with version, detected platform, and install mode ✅ 24/03/2026 08:19
- [x] [H] Add directory creation: `~/.blueprint/bin/`, `~/.blueprint/templates/` ✅ 24/03/2026 08:19
**Verify:** `bash install.sh` shows banner and detects platform without errors

### Phase 2: Interactive Component Menu
**Touches:** `install.sh`
- [x] [H] Add `gum choose` multi-select menu with components: "Blueprint CLI binary (required)", "27 SDLC skills (required)", "Audit hook (15 rules)", "Git hook templates", "GitHub Action templates". First two always selected, rest default on. ✅ 24/03/2026 08:19
- [x] [H] Add plain fallback menu using `read -p` prompts when `HAS_GUM=false` — ask y/n for each optional component ✅ 24/03/2026 08:19
- [x] [H] Add install summary before proceeding: list selected components, target paths, detected platform — confirm with `gum confirm` or `read -p "Continue? [Y/n]"` ✅ 24/03/2026 08:19
**Verify:** Menu renders correctly with and without gum installed

### Phase 3: Binary Download & Install
**Touches:** `install.sh`
- [x] [S] Add `install_binary()` function: in local mode, copy from `$SCRIPT_DIR/cli/blueprint-${OS}-${ARCH}`; in remote mode, fetch latest release URL from GitHub API via `curl -s https://api.github.com/repos/skaisser/blueprint/releases/latest`, extract asset URL for the platform binary using grep/sed (avoid jq dependency), download with `curl -fSL -o` ✅ 24/03/2026 08:19
- [x] [H] Place binary at `~/.blueprint/bin/blueprint`, `chmod +x`, verify with `~/.blueprint/bin/blueprint --version` ✅ 24/03/2026 08:19
- [x] [H] Add `~/.blueprint/bin` to PATH in shell RC if not already present — use `grep -q` check before appending `export PATH="$HOME/.blueprint/bin:$PATH"` ✅ 24/03/2026 08:19
- [x] [H] Show download progress with `gum spin` wrapper (or plain curl progress bar without gum) ✅ 24/03/2026 08:19
**Verify:** `~/.blueprint/bin/blueprint --version` returns version string

### Phase 4: Skills & Templates Installation
**Touches:** `install.sh`
- [x] [S] Add `install_skills()` function: in local mode, loop `$SCRIPT_DIR/skills/*/` and `cp -rf` each to `~/.claude/skills/`; in remote mode, `git clone --depth 1` the repo to a temp dir, copy skills from there, then clean up temp dir ✅ 24/03/2026 08:19
- [x] [H] Add `install_templates()` function: copy `templates/` contents to `~/.blueprint/templates/`, preserve executable permissions on `.githooks/commit-msg` and `.githooks/pre-push` ✅ 24/03/2026 08:19
- [x] [H] Show per-skill progress: count total skills, show "Installing skill X/27: skill-name" with `gum spin` or plain echo ✅ 24/03/2026 08:19
**Verify:** `ls ~/.claude/skills/ | wc -l` returns 27, `ls ~/.blueprint/templates/.githooks/` shows commit-msg and pre-push

### Phase 5: Claude Code Integration
**Touches:** `install.sh`, `config/settings.json`, `config/mcp.json`
- [x] [H] Create `config/settings.json` template file in repo with the PreToolUse hook configuration pointing to `~/.blueprint/bin/blueprint audit` ✅ 24/03/2026 08:19
- [x] [H] Create `config/mcp.json` template file in repo with Context7 and Sequential Thinking MCP server configs (both free, no API keys needed) ✅ 24/03/2026 08:19
- [x] [S] Add `register_hook()` function in install.sh: use `claude mcp add` or `claude settings set` CLI commands if available, otherwise use pure bash/sed to merge the PreToolUse hook into existing `~/.claude/settings.json` without clobbering other settings. If settings.json doesn't exist, copy template directly. No Python dependency. ✅ 24/03/2026 08:19
- [x] [S] Add `register_mcp()` function: use `claude mcp add-json` to add Context7 and sequential-thinking MCP servers, or fall back to pure bash JSON append. Never remove existing MCP servers. ✅ 24/03/2026 08:19
- [x] [H] Back up existing settings.json and mcp.json before modifying: `cp file file.bak.TIMESTAMP` ✅ 24/03/2026 08:19
**Verify:** `cat ~/.claude/settings.json | grep blueprint` shows the audit hook, `cat ~/.claude/mcp.json | grep context7` shows MCP server

### Phase 6: Post-Install Summary & README Update
**Touches:** `install.sh`, `README.md`
- [x] [H] Add post-install summary block: show all installed components, paths, versions in a styled box (gum style) or plain table ✅ 24/03/2026 08:19
- [x] [H] Add "Next steps" guidance: prompt user to `cd` into a project and run `/start` to initialize BLUEPRINT. Show "Run `blueprint --version` to verify" and link to docs. Ask if they want to initialize a project now. ✅ 24/03/2026 08:19
- [x] [H] Add `--uninstall` flag support: removes `~/.blueprint/` and Blueprint skills from `~/.claude/skills/`, removes hook from settings.json — for clean removal ✅ 24/03/2026 08:19
- [x] [H] Update README.md install section to match final installer UX and behavior ✅ 24/03/2026 08:19
- [x] [H] Test full installer end-to-end: run via `bash install.sh` from the repo, verify all files are in place and `blueprint --version` works ✅ 24/03/2026 08:19
**Verify:** Full end-to-end run produces working installation, README matches actual behavior

## Tech Stack Versions

- Bash 3.2+ (macOS default) or Zsh — installer uses `#!/usr/bin/env bash`
- gum v0.14+ (charmbracelet/gum — optional TUI, auto-bootstrapped)
- curl (for downloading binary and GitHub API)
- git (for remote mode — clone skills)
- Claude Code CLI (`claude` command) — for `claude mcp add` and settings registration

## Execution Strategy

> **Approach:** `/plan-approved` with Single Subagent (Opus)
> **Total Tasks:** 27 (H: 18, S: 9, O: 0)
> **Estimated Rounds:** 1 (sequential — all phases touch install.sh)

### File-Touch Matrix

| Phase | Files/Dirs Touched | Depends On |
|-------|-------------------|------------|
| Phase 1 | `install.sh` | — |
| Phase 2 | `install.sh` | Phase 1 |
| Phase 3 | `install.sh` | Phase 2 |
| Phase 4 | `install.sh` | Phase 3 |
| Phase 5 | `install.sh`, `config/settings.json`, `config/mcp.json` | Phase 4 |
| Phase 6 | `install.sh`, `README.md` | Phase 5 |

**All phases touch install.sh sequentially.** Zero parallelism possible — each phase adds to the same file.

### Round 1: All Phases → Single Subagent (Opus, sequential)

One Opus subagent builds the complete installer from Phase 1 through Phase 6. Sequential execution is required because every phase appends to `install.sh`.

| Phase | Tasks | Complexity Mix |
|-------|-------|---------------|
| Phase 1: Core Framework | 6 tasks | 2x[S] + 4x[H] |
| Phase 2: Component Menu | 3 tasks | 3x[H] |
| Phase 3: Binary Download | 4 tasks | 1x[S] + 3x[H] |
| Phase 4: Skills & Templates | 3 tasks | 1x[S] + 2x[H] |
| Phase 5: Claude Code Integration | 5 tasks | 2x[S] + 3x[H] |
| Phase 6: Summary & README | 5 tasks | 5x[H] |

## Acceptance
- [x] `curl -fsSL https://raw.githubusercontent.com/skaisser/blueprint/main/install.sh | bash` works on a clean macOS machine
- [x] `./install.sh` from cloned repo works identically
- [x] All 27 skills appear in `~/.claude/skills/`
- [x] `blueprint --version` works from any terminal after install
- [x] Audit hook is registered in `~/.claude/settings.json`
- [x] Existing user settings.json content is preserved (merge, not overwrite)
- [x] No existing files are deleted during installation
- [x] Installer renders beautifully with gum and degrades gracefully without it
