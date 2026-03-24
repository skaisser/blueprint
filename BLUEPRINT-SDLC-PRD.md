---
title: BLUEPRINT SDLC — Open-Source Edition
type: prd
version: 1.0
status: draft
author: Shirleyson Kaisser
repo: github.com/skaisser/blueprint
date: 2026-03-24
tags: [prd, blueprint, sdlc, open-source]
---

# BLUEPRINT SDLC
## Open-Source Edition — Product Requirements Document

> v1.0 · March 2026 · Shirleyson Kaisser · github.com/skaisser/blueprint

---

## 1. Executive Summary

BLUEPRINT SDLC is a complete, portable software development lifecycle built on top of Claude Code. It turns Claude from a code assistant into a disciplined engineering partner — with planning, execution, review, and merge all governed by a structured pipeline of slash commands, an audit hook enforcing 14 rules on every tool call, and a Go CLI binary shipping pre-compiled for macOS and Linux.

The project is extracted from a private 59-skill monorepo (Kaisser SDLC 4.0) and published as a focused, stack-agnostic open-source release covering the core pipeline — everything needed to take an idea from backlog to merged PR.

The name **BLUEPRINT** is both the project identity and a dual mnemonic: all 9 pipeline steps map to letters, and the first 4 (B-L-U-E) name the 4 folders inside the `blueprint/` workspace directory.

### The GTD Gap

GTD (Getting Things Done) for Claude Code was the right idea but the wrong execution — too slow, too manual, too much overhead for what should be a frictionless flow. BLUEPRINT fills that gap directly:

| GTD for Claude | BLUEPRINT |
|----------------|-----------|
| Manual task capture | `/backlog` — one command, file created instantly |
| You manage the system | Audit hook enforces the system for you |
| Context switches kill flow | 1M context + `/flow` keeps everything in one session |
| Folders you maintain by hand | BLUE folders move automatically on each phase transition |
| Trust yourself to follow the process | Trust the pipeline — 14 rules catch you when you don't |
| Someday/Maybe pile | `blueprint/expired/` — archived, not lost |

> **"GTD taught you to capture everything. BLUEPRINT ships it."**

GTD users already living in Obsidian have a natural migration path — their inbox becomes `/backlog`, their projects become `blueprint/live/`, their Someday/Maybe becomes `blueprint/expired/`. Zero new habits, just a faster system.

### The BLUEPRINT Acronym

| # | Letter | Command | Phase |
|---|--------|---------|-------|
| 1 | **B** | `/backlog` | Backlog — capture and prioritise ideas |
| 2 | **L** | `/plan` | Layout — create branch + blueprint file |
| 3 | **U** | `/plan-review` | Unpack — validate, assign complexity [H]/[S]/[O] |
| 4 | **E** | `/plan-approved` | Endorse — execute, spawn parallel subagents |
| 5 | **P** | `/plan-check` | Preflight — audit code vs blueprint |
| 6 | **R** | `/pr` | Raise — open pull request with full context |
| 7 | **I** | `/review` | Inspect — trigger @claude code review |
| 8 | **N** | `/address-pr` | Negotiate — fetch feedback, fix, push |
| 9 | **T** | `/finish` | Tag — merge, rename blueprint to `upstream/` |

---

## 2. Goals & Non-Goals

### 2.1 Goals

- Publish a self-contained, installable SDLC that works on any Claude Code project regardless of tech stack
- Position BLUEPRINT as the successor to GTD for Claude Code — faster, enforced, automatic
- Rename all internals from "Kaisser SDLC" / `claude-cli` to "BLUEPRINT" / `blueprint` CLI
- Replace `.planning/` with `blueprint/` — no dot, visible in Finder, Obsidian vault, and GitHub tree
- Ship pre-compiled Go binaries: macOS Apple Silicon (arm64), macOS Intel (amd64), Linux amd64
- Provide a cross-platform smart installer (macOS + Linux) with interactive gum multi-select menu
- Provide an Obsidian-ready `blueprint/` workspace using the BLUE folder structure with Dataview-compatible frontmatter
- Document the autoresearch eval framework so contributors can benchmark and improve skills
- Offer a clear extension path — users can add custom skills without modifying the core

### 2.2 Non-Goals for v1

- Personal integrations will **not** be open-sourced: UniFi, Home Assistant, TikTok Shop, NotebookLM, Voice Cloning, n8n
- Laravel/TALL-specific audit rules will be factored into an optional `presets/laravel/` — not in the default config
- No GUI, web dashboard, or cloud sync in v1 — purely local + git
- No dependency on paid services — the core pipeline runs on the free Claude Code tier

---

## 3. Naming & Branding

### 3.1 The BLUEPRINT Mnemonic

Every letter maps to one of the 9 pipeline phases. A developer can recite B-L-U-E-P-R-I-N-T and immediately know where they are in the lifecycle — no diagram required.

### 3.2 The BLUE Folder Structure

The `blueprint/` workspace uses the first 4 letters as subfolder names — a second mnemonic layer embedded in the file system:

| Letter | Folder | Trigger | Status meaning |
|--------|--------|---------|----------------|
| **B** | `blueprint/backlog/` | `/backlog` | Ideas not yet planned |
| **L** | `blueprint/live/` | `/plan` | Currently in development |
| **U** | `blueprint/upstream/` | `/finish` | Shipped and merged to main |
| **E** | `blueprint/expired/` | `/backlog --archive` | Cancelled or deferred |

> Files move between folders as status transitions. `/plan` moves `backlog/` → `live/`. `/finish` moves `live/` → `upstream/`. The file path alone communicates status at any point in git history. Obsidian automatically tracks file moves and updates all internal links.

### 3.3 CLI Rename: `claude-cli` → `blueprint`

| Old (`claude-cli`) | New (`blueprint`) | Purpose |
|--------------------|-------------------|---------|
| `claude-cli sdlc meta` | `blueprint meta` | Repo metadata summary |
| `claude-cli sdlc context` | `blueprint context` | Generate CLAUDE.md |
| `claude-cli sdlc sync` | `blueprint sync` | Sync to `~/.blueprint` |
| `claude-cli sdlc commit` | `blueprint commit` | Staged commit helper |
| `claude-cli sdlc full` | `blueprint status` | Full SDLC status |
| `claude-cli audit` | `blueprint audit` | Audit hook (14 rules) |
| `claude-cli pr-review` | `blueprint pr-review` | PR review formatter |

---

## 4. Open-Source Scope

### 4.1 Included — Core SDLC (24 skills)

| Category | Skills | Notes |
|----------|--------|-------|
| Pipeline | `/backlog` `/plan` `/plan-review` `/plan-approved` `/plan-check` `/pr` `/review` `/address-pr` `/finish` | Core BLUEPRINT loop — stack agnostic |
| Automation | `/flow` `/flow-auto` `/flow-auto-wt` `/batch-flow` | Full pipeline chaining with checkpoints |
| Fast Tracks | `/quick` `/hotfix-push` `/resume` | Escape hatches for small or urgent work |
| Git & PR | `/commit` `/ship` `/push` `/branch` `/complete` | Git workflow helpers |
| Testing | `/test` `/run-tests` `/coverage` `/tdd-review` | Stack-agnostic test generation |
| Project Setup | `/start` `/workflow-sync` `/context` `/sync` `/status` | Init and maintenance |
| Skill Factory | `/skill-creator` | Create and benchmark new skills |

### 4.2 Included — Standalone Skills (optional)

- `/brand-generator` — DaisyUI + Tailwind 4 design systems from inspiration URLs
- `/cf-pages-deploy` — Deploy static sites to Cloudflare Pages via Wrangler CLI
- `/remotion-video` — Full video production pipeline (VDLC): plan → script → storyboard → render
- `/yt-search` — YouTube search via yt-dlp (no API key required)
- `/excalidraw-diagram` — Architecture and workflow diagrams, Playwright-validated PNG output
- `/firecrawl` — Web scraping router (requires Firecrawl CLI — user provides API key)

### 4.3 Excluded — Private Integrations

| Integration | Skill | Reason |
|-------------|-------|--------|
| Home Assistant | `/home-assistant-manager` | Personal smart home — hardware-specific (MacStudio) |
| UniFi Network | `/unifi` | Hardware-specific (UDM) + 1Password auth baked in |
| TikTok Shop | `/tiktok-shop` | Business-specific + proprietary HMAC API signing |
| NotebookLM | `/notebooklm` | Google OAuth — complex per-user setup |
| Voice Cloning | `/voice-cloning-skill` | GPU-heavy, highly personal, model-specific |
| n8n Workflows | `/n8n` | Server-specific + 1Password auth baked in |

---

## 5. Architecture

### 5.1 Repository Structure

```
blueprint-sdlc/
├── README.md
├── install.sh                    # macOS + Linux installer
├── CLAUDE.md                     # Global Claude Code instructions
├── assets/
│   └── sdlc-infographic.webp
├── cli/
│   ├── main.go  cmd/  internal/
│   ├── blueprint-darwin-arm64    # macOS Apple Silicon
│   ├── blueprint-darwin-amd64    # macOS Intel
│   └── blueprint-linux-amd64     # Linux x86_64
├── hooks/
│   └── audit.py                  # PreToolUse fallback (Python)
├── skills/                       # 24 SDLC + optional standalones
│   ├── backlog/ plan/ plan-review/ plan-approved/ plan-check/
│   ├── pr/ review/ address-pr/ finish/
│   ├── flow/ flow-auto/ flow-auto-wt/ batch-flow/
│   ├── quick/ hotfix-push/ resume/
│   ├── commit/ ship/ push/ branch/ complete/
│   ├── test/ run-tests/ coverage/ tdd-review/
│   ├── start/ workflow-sync/ context/ sync/ status/
│   └── skill-creator/
├── templates/
│   ├── .githooks/                # commit-msg, pre-push
│   ├── .github/workflows/claude.yml
│   ├── CLAUDE.md
│   └── blueprint/                # Scaffold: backlog/ live/ upstream/ expired/
├── presets/
│   ├── laravel/                  # Laravel TALL audit rules
│   └── node/                     # Node/TS overrides (stub v1)
└── evals/
    └── sdlc/                     # Autoresearch eval framework
```

### 5.2 Per-Project `blueprint/` Workspace

```
my-project/
├── src/
├── blueprint/
│   ├── backlog/
│   │   └── 0001-dark-mode.md
│   ├── live/
│   │   └── 0002-auth-flow.md
│   ├── upstream/
│   │   └── 0003-onboarding-complete.md
│   └── expired/
│       └── 0004-old-feature.md
└── CLAUDE.md
```

### 5.3 Blueprint File Frontmatter

Every blueprint file uses YAML frontmatter compatible with Obsidian Dataview:

```yaml
---
id: 0002
title: Auth flow refactor
status: live              # backlog | live | upstream | expired
complexity: H             # H (fast) | S (balanced) | O (deep)
linear_issue: KPG-42
branch: feat/auth-flow-refactor
strategy: parallel        # parallel | team | single | leader
started_at: 2026-03-20
completed_at:
tags: [auth, security, refactor]
---
```

### 5.4 Go CLI Binaries

| Binary | Platform | Arch | Est. size |
|--------|----------|------|-----------|
| `blueprint-darwin-arm64` | macOS | Apple Silicon (M1–M4) | ~6 MB |
| `blueprint-darwin-amd64` | macOS | Intel | ~6 MB |
| `blueprint-linux-amd64` | Linux | x86_64 | ~5.5 MB |

No Go toolchain required on target machines — the installer copies the correct binary to `~/.blueprint/bin/` and adds it to PATH.

### 5.5 Audit Hook — 14 Rules

`blueprint audit` fires on every Claude Code tool call via PreToolUse. The Go binary is the only hook — fast, compiled, zero dependencies.

| # | Rule | Enforcement |
|---|------|-------------|
| 1 | Skill read gate | Block writes to `.jsx/.tsx/.html/.css` without reading `frontend-design` SKILL.md |
| 2 | Reference tracking | Track reads of SKILL.md, plan-template.md, team-execution.md |
| 3 | Team compliance | Warn if TeamCreate used without reading team-execution.md |
| 4 | Standalone task count | Warn if 3+ Task calls with no TeamCreate |
| 5 | Handoff tracking | Track AskUserQuestion at required `/flow` checkpoints |
| 6 | Checkpoint audit trail | Record every `🔷 BP` checkpoint; enforce `/plan-check` before `/pr` |
| 7 | Workflow creation gate | Block creating `claude.yml` if project has no homolog branch |
| 8 | Test suite enforcement | Block test runner without `--parallel` or `--filter` flag |
| 9 | Plan task deletion | Warn when unchecked `[ ]` tasks removed instead of implemented |
| 10 | Dangerous command block | Block `migrate:fresh`, AI signatures, direct `git push main` |
| 11 | @claude review enforcement | Block short `@claude review` comments — require full prompt |
| 12 | Plan-check skip detection | Warn if `/pr` invoked after `/plan-approved` without `/plan-check` |
| 13 | Acceptance criteria gate | Warn if PR created with unchecked acceptance criteria |
| 14 | Flow-auto step enforcement | Block PR creation if mandatory steps were skipped |

---

## 6. Execution Strategies

`/plan-review` assigns one of four execution modes. Decision rule: can you write a complete, self-contained prompt for each worker upfront? **YES → Parallel. NO → Team.**

| Mode | When | How |
|------|------|-----|
| Parallel Subagents *(default)* | 2+ independent phases | Multiple Agent calls in ONE message — true parallelism |
| Coordinated Team | Workers need mid-task handoffs | Team messaging or blueprint-file handoffs between agents |
| Single Subagent | 1 phase or strictly sequential | One Agent call — no spawn overhead |
| Leader Direct | ≤3 [H] tasks total | Lead model handles directly — fastest for trivial work |

**Complexity markers assigned during `/plan-review`:**

- `[H]` — Fast: small scope, high skill. Execute immediately.
- `[S]` — Balanced: medium scope. Standard parallel execution.
- `[O]` — Deep reasoning: architectural decisions, sequential-only, full context required.

---

## 7. Git Hooks & Commit Format

### 7.1 commit-msg hook

- Validates emoji + type format on every commit
- Blocks AI signatures (`Co-Authored-By`, `Generated by Claude`)
- Required format: `<emoji> <type>: <description>` (present tense, lowercase)

| Emoji | Type | Use case |
|-------|------|----------|
| ✨ | `feat` | New feature |
| 🐛 | `fix` | Bug fix |
| 📚 | `docs` | Documentation only |
| ♻️ | `refactor` | Restructuring, no behavior change |
| 🧪 | `test` | Tests only |
| 📋 | `plan` | Blueprint file updates |
| 🔀 | `merge` | Branch merge |
| 🩹 | `hotfix` | Urgent production fix |
| 🚀 | `deploy` | Deployment / CI changes |

### 7.2 pre-push hook

- Auto-detects if project has a homolog branch (local or remote)
- If homolog exists: blocks direct push to `main` — must use `homolog → main` PR flow
- If no homolog: allows direct push to `main` (config repos, docs, single-branch projects)
- Always allows push to `homolog` and feature branches

---

## 8. Obsidian Integration

`blueprint/` is designed as a first-class Obsidian vault subfolder. With the Dataview plugin, developers get a living project dashboard with zero additional tooling.

### 8.1 Example Dataview Queries

Active blueprints:

```dataview
TABLE complexity, linear_issue, started_at
FROM "blueprint/live"
SORT started_at DESC
```

Backlog by complexity:

```dataview
TABLE title, complexity, tags
FROM "blueprint/backlog"
SORT complexity ASC
```

### 8.2 File Lifecycle

- `backlog/` → `live/`: `/plan` creates branch, moves file, fills frontmatter
- `live/` → `upstream/`: `/finish` after PR merge, file renamed with `-complete` suffix
- `live/` → `expired/`: `/backlog --archive` or manual move

---

## 9. Pipeline Automation

### 9.1 `/flow` — Guided Pipeline

Chains the full pipeline with two human review pauses:

```
/plan → /plan-review → ⏸ review → /plan-approved → /plan-check → /pr → ⏸ review → /finish
```

- Context-aware: with 1M context most plans complete in one session without breaking
- Resumable: `/flow --from plan-approved` picks up mid-pipeline after a context clear

### 9.2 `/flow-auto` — Zero-Touch Pipeline

| Behaviour | `/flow` | `/flow-auto` |
|-----------|---------|--------------|
| User decisions | 2 checkpoints | Zero — model decides everything |
| AskUserQuestion | Yes, at checkpoints | Never (breaks autonomy) |
| Review loop | Manual | Automatic, up to 3 cycles |
| Merges PR | Yes, via `/finish` | Never — leaves for human review |
| Best for | Staying in control | Fire and forget |

### 9.3 `/batch-flow` — Multi-Plan Sequential Execution

```bash
blueprint batch-flow 2-6               # Execute blueprints 0002 through 0006
blueprint batch-flow 2-6 --auto-merge  # Full merge chain after each
```

- Auto context compaction between plans
- Crash recovery: skips completed, resumes from last incomplete
- Stops batch on merge conflicts

---

## 10. Installation & Quick Start

### 10.1 Install

```bash
git clone git@github.com:skaisser/blueprint.git ~/Sites/blueprint
cd ~/Sites/blueprint && ./install.sh
```

The installer auto-detects platform + arch and:

- Copies `blueprint` binary to `~/.blueprint/bin/` and adds to PATH
- Installs 24 SDLC skills to `~/.claude/skills/`
- Registers the audit hook via `claude settings`
- macOS: removes Linux-only MCP servers; Linux: removes macOS-only entries (Herd, BrowserMCP)

### 10.2 New Project

```bash
/start
```

Sets up: git hooks, CLAUDE.md template, `blueprint/` workspace (BLUE folders), GitHub Action, homolog branch.

### 10.3 Full Pipeline Walk-through

```bash
/backlog              # Capture idea → blueprint/backlog/0001-feature.md
/plan                 # Promote → blueprint/live/0001-feature.md + git branch
/plan-review          # Validate, assign [H]/[S]/[O], pick execution strategy
/plan-approved        # Execute — spawn parallel subagents per strategy
/plan-check           # Audit code vs blueprint, two-commit snapshot
/pr                   # Open pull request with full context
/review               # Trigger @claude review on the PR
/address-pr           # Fetch all feedback, fix, push
/finish               # Merge PR → blueprint/upstream/0001-feature-complete.md
```

---

## 11. Autoresearch — Skill Self-Improvement

Inspired by [karpathy/autoresearch](https://github.com/karpathy/autoresearch). BLUEPRINT optimizes skill prompts and trigger descriptions rather than neural network weights — using the same boolean scoring loop.

| Dimension | What it measures (boolean — yes/no only) |
|-----------|------------------------------------------|
| Trigger | Does the skill fire for the right inputs and stay silent for wrong ones? |
| Fast | Did execution stay within time and tool-call budget? |
| Necessary | Were all steps needed? No wasted work? |
| Better | Is output quality equal or better than the previous version? |

```bash
cd evals/sdlc && uv run streamlit run dashboard.py        # Live results dashboard
cd evals/sdlc && uv run python optimize.py plan -n 3      # 3 iterations on plan skill
cd evals/sdlc && uv run python optimize.py --all          # All SDLC skills
```

---

## 12. Roadmap

| Phase | Version | Scope |
|-------|---------|-------|
| Launch | v1.0 | 24 SDLC skills + `blueprint` CLI + 14-rule audit hook + Obsidian BLUE workspace |
| Presets | v1.1 | Laravel TALL preset, Node/TS preset — stack-specific audit rule extensions |
| Standalone Pack | v1.2 | Brand generator, cf-pages-deploy, yt-search, remotion-video as optional add-ons |
| Eval Suite | v1.3 | Autoresearch dashboard + optimizer publicly documented and community-extensible |
| Multi-Agent v2 | v2.0 | Improved team execution, inter-agent blueprint handoffs, `/flow-auto` stability |

---

## 13. Open Questions

- **License**: Apache 2.0 ✓ — explicit patent grant protects contributors, trademark clause protects the BLUEPRINT name from impersonation forks.
- **Laravel audit rules**: `presets/laravel/` with env flag, or comments in default config?
- **Obsidian**: should `install.sh` auto-create `.obsidian/` with Dataview queries, or leave to user?
- **`/laravel-db-diagram`**: migration-reader is Laravel-specific — ship under `presets/laravel/` or skip v1?
- **Firecrawl**: bundle as optional standalone or document as external dependency only?

---

## 14. Smart Installer Design

The installer uses [gum](https://github.com/charmbracelet/gum) for an interactive TUI — same pattern as devterm but cross-platform (macOS + Linux) and with meaningful component selection rather than a single confirm.

### Platform detection

```bash
if [[ "$(uname)" == "Darwin" ]]; then
    # bootstrap: brew install gum
    # offer: macOS-only options (Herd, BrowserMCP)
elif [[ "$(uname)" == "Linux" ]]; then
    # bootstrap: curl gum binary from github.com/charmbracelet/gum/releases
    # skip macOS-only options silently
else
    echo "✗ BLUEPRINT requires macOS or Linux." && exit 1
fi
```

### Interactive menu (gum multi-select)

```
  BLUEPRINT SDLC — installer
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Select components to install:

  [x] Core SDLC skills (24 skills)     ← required, always on
  [x] blueprint CLI binary              ← required, always on
  [x] Audit hook (14 rules)
  [x] Git hooks (commit-msg, pre-push)
  [x] GitHub Action (@claude review)
  [ ] Obsidian setup (Dataview queries)
  [ ] Laravel preset (audit rules)
  [ ] Node/TS preset
  [ ] Standalone skills (brand, cf-pages, yt-search, remotion)

  Platform: macOS Apple Silicon  ✓  auto-detected
```

### Installer architecture (mirrors devterm's lib/ pattern)

```
install.sh                        # entry point — self-clone via curl
lib/
├── utils.sh                      # logging, colors, step()
├── banner.sh                     # figlet BLUEPRINT header
├── checks.sh                     # Claude Code installed? git? platform?
├── menu.sh                       # gum multi-select + summary
├── platform.sh                   # macOS vs Linux detection + gum bootstrap
└── install/
    ├── core.sh                   # skills + binary + hooks
    ├── audit.sh                  # audit hook registration
    ├── git-hooks.sh              # commit-msg, pre-push
    ├── github-action.sh          # claude.yml
    ├── obsidian.sh               # .obsidian/ + dataview queries
    ├── preset-laravel.sh         # Laravel audit rules
    ├── preset-node.sh            # Node/TS overrides
    └── standalones.sh            # brand, cf-pages, yt-search, remotion
```

### Key design decisions vs devterm

| devterm | BLUEPRINT |
|---------|-----------|
| macOS only | macOS + Linux |
| One confirm, install everything | Multi-select — user picks components |
| Global machine setup | Global `~/.blueprint/` + per-project via `/start` |
| iTerm2 / terminal focus | Claude Code focus — any terminal |
| MIT | Apache 2.0 |

### Install paths

There are two supported install paths depending on the user's intent:

**Path A — one-liner (most users)**

```bash
curl -fsSL https://raw.githubusercontent.com/skaisser/blueprint/main/install.sh | bash
```

- No git clone required
- Self-downloads the install script, bootstraps gum, shows the interactive menu
- Pulls pre-compiled `blueprint` binary for the detected platform — no Go toolchain needed
- Skills install to `~/.claude/skills/`, binary to `~/.blueprint/bin/`
- On a fresh machine with no Claude Code installed, detects the gap and links to the Claude Code install page before proceeding
- Updates work the same way — re-running the one-liner pulls the latest binary and skills

**Path B — clone (power users / CLI contributors)**

```bash
git clone git@github.com:skaisser/blueprint.git ~/Sites/blueprint
cd ~/Sites/blueprint && ./install.sh
```

- For users who want to fork the repo, modify the Go CLI, and build their own binary
- Required when: customising audit rules at source level, contributing back to the project, or building a team-specific fork of BLUEPRINT
- The `cli/` directory contains full Go source + a `Makefile` with `make build` targets for all three platforms:

```makefile
make build-mac-arm   # blueprint-darwin-arm64
make build-mac-intel # blueprint-darwin-amd64
make build-linux     # blueprint-linux-amd64
make build-all       # all three
```

### Decision rule for the README

> Just want BLUEPRINT? One line.
> Want to hack the CLI? Clone it.

---

## 14. Success Metrics

| Metric | 3 months | 6 months |
|--------|----------|----------|
| GitHub stars | 500 | 2,000 |
| `install.sh` runs | 200 | 1,000 |
| Contributors (PRs merged) | 5 | 20 |
| Community skills added | 3 | 15 |
| Avg issue close time | — | < 2 weeks |
