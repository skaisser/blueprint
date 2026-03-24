# BLUEPRINT SDLC

> A complete, portable software development lifecycle for Claude Code.

BLUEPRINT turns Claude Code from a code assistant into a disciplined engineering partner — with planning, execution, review, and merge all governed by a structured pipeline of slash commands, an audit hook enforcing 14 rules on every tool call, and a Go CLI binary shipping pre-compiled for macOS and Linux.

**Stack-agnostic. Zero dependencies on paid services. Works on any Claude Code project.**

---

## The Pipeline

Every letter in **BLUEPRINT** maps to a pipeline phase:

| # | Letter | Command | Phase |
|---|--------|---------|-------|
| 1 | **B** | `/backlog` | Backlog — capture and prioritise ideas |
| 2 | **L** | `/plan` | Layout — create branch + blueprint file |
| 3 | **U** | `/plan-review` | Unpack — validate, assign complexity |
| 4 | **E** | `/plan-approved` | Endorse — execute, spawn parallel subagents |
| 5 | **P** | `/plan-check` | Preflight — audit code vs blueprint |
| 6 | **R** | `/pr` | Raise — open pull request with full context |
| 7 | **I** | `/review` | Inspect — trigger @claude code review |
| 8 | **N** | `/address-pr` | Negotiate — fetch feedback, fix, push |
| 9 | **T** | `/finish` | Tag — merge, rename blueprint to upstream |

---

## The BLUE Workspace

The `blueprint/` directory uses the first four letters as folder names — your file path _is_ your status:

| Folder | Trigger | Meaning |
|--------|---------|---------|
| `blueprint/backlog/` | `/backlog` | Ideas not yet planned |
| `blueprint/live/` | `/plan` | Currently in development |
| `blueprint/upstream/` | `/finish` | Shipped and merged |
| `blueprint/expired/` | `/backlog --archive` | Cancelled or deferred |

Files move between folders as work progresses. Compatible with [Obsidian](https://obsidian.md) + Dataview out of the box.

---

## Why BLUEPRINT?

If you've tried GTD-style workflows with Claude Code, you know the pain: too slow, too manual, too much overhead.

| Before | BLUEPRINT |
|--------|-----------|
| Manual task capture | `/backlog` — one command, file created |
| You manage the system | Audit hook enforces the system for you |
| Context switches kill flow | 1M context + `/flow` keeps everything in one session |
| Folders you maintain by hand | BLUE folders move automatically on phase transitions |
| Trust yourself to follow the process | 14 rules catch you when you don't |

---

## Install

### Quick install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/skaisser/blueprint/main/install.sh | bash
```

No Go toolchain required. The installer auto-detects your platform, shows an interactive menu via [gum](https://github.com/charmbracelet/gum), and installs:

- `blueprint` CLI binary to `~/.blueprint/bin/`
- 24 SDLC skills to `~/.claude/skills/`
- Audit hook (14 rules) via Claude Code settings
- Git hooks (commit-msg, pre-push)

### Clone (contributors / CLI hackers)

```bash
git clone git@github.com:skaisser/blueprint.git ~/Sites/blueprint
cd ~/Sites/blueprint && ./install.sh
```

Build your own binary:

```bash
cd cli && make build-all
```

> **Just want BLUEPRINT? One line. Want to hack the CLI? Clone it.**

---

## Quick Start

### 1. Initialize a project

```bash
/start
```

Sets up git hooks, CLAUDE.md template, `blueprint/` workspace, GitHub Action, and homolog branch.

### 2. Run the full pipeline

```bash
/backlog              # Capture idea → blueprint/backlog/0001-feature.md
/plan                 # Promote → blueprint/live/0001-feature.md + git branch
/plan-review          # Validate, assign complexity, pick execution strategy
/plan-approved        # Execute — spawn parallel subagents
/plan-check           # Audit code vs blueprint
/pr                   # Open pull request
/review               # Trigger @claude code review
/address-pr           # Fetch feedback, fix, push
/finish               # Merge → blueprint/upstream/0001-feature-complete.md
```

### 3. Or let it run itself

```bash
/flow                 # Guided pipeline with 2 review pauses
/flow-auto            # Zero-touch — model decides everything
/batch-flow 2-6       # Execute blueprints 0002 through 0006 sequentially
```

---

## What's Included

### Core SDLC (24 skills)

| Category | Skills |
|----------|--------|
| Pipeline | `/backlog` `/plan` `/plan-review` `/plan-approved` `/plan-check` `/pr` `/review` `/address-pr` `/finish` |
| Automation | `/flow` `/flow-auto` `/flow-auto-wt` `/batch-flow` |
| Fast Tracks | `/quick` `/hotfix-push` `/resume` |
| Git & PR | `/commit` `/ship` `/push` `/branch` `/complete` |
| Testing | `/test` `/run-tests` `/coverage` `/tdd-review` |
| Project Setup | `/start` `/workflow-sync` `/context` `/sync` `/status` |
| Skill Factory | `/skill-creator` |

### Standalone Skills (optional)

- `/brand-generator` — DaisyUI + Tailwind 4 design systems from inspiration URLs
- `/cf-pages-deploy` — Deploy static sites to Cloudflare Pages
- `/remotion-video` — Full video production pipeline: plan → script → storyboard → render
- `/yt-search` — YouTube search via yt-dlp (no API key)
- `/excalidraw-diagram` — Architecture diagrams, Playwright-validated PNG output
- `/firecrawl` — Web scraping router (requires Firecrawl CLI)

---

## Execution Strategies

`/plan-review` assigns complexity and picks the fastest execution mode:

| Complexity | Meaning |
|------------|---------|
| `[H]` | Fast — small scope, execute immediately |
| `[S]` | Balanced — standard parallel execution |
| `[O]` | Deep reasoning — sequential, full context required |

| Strategy | When | How |
|----------|------|-----|
| Parallel Subagents | 2+ independent phases | Multiple Agent calls in one message |
| Coordinated Team | Workers need mid-task handoffs | Team messaging between agents |
| Single Subagent | 1 phase or strictly sequential | One Agent call |
| Leader Direct | ≤3 `[H]` tasks total | Lead model handles directly |

---

## Audit Hook — 14 Rules

`blueprint audit` fires on every Claude Code tool call via PreToolUse:

| # | Rule | What it does |
|---|------|--------------|
| 1 | Skill read gate | Block writes without reading relevant SKILL.md |
| 2 | Reference tracking | Track reads of key reference files |
| 3 | Team compliance | Warn if teams used without reading team-execution.md |
| 4 | Standalone task count | Warn if 3+ tasks with no team |
| 5 | Handoff tracking | Track checkpoints at `/flow` pauses |
| 6 | Checkpoint audit trail | Enforce `/plan-check` before `/pr` |
| 7 | Workflow creation gate | Block `claude.yml` without homolog branch |
| 8 | Test suite enforcement | Block test runner without `--parallel` or `--filter` |
| 9 | Plan task deletion | Warn when unchecked tasks removed |
| 10 | Dangerous command block | Block `migrate:fresh`, AI signatures, direct push to main |
| 11 | Review enforcement | Block short `@claude review` comments |
| 12 | Plan-check skip detection | Warn if `/pr` invoked without `/plan-check` |
| 13 | Acceptance criteria gate | Warn if PR has unchecked acceptance criteria |
| 14 | Flow-auto enforcement | Block PR if mandatory steps were skipped |

---

## Commit Format

BLUEPRINT enforces emoji + type on every commit:

```
<emoji> <type>: <description>
```

| Emoji | Type | Use case |
|-------|------|----------|
| ✨ | `feat` | New feature |
| 🐛 | `fix` | Bug fix |
| 📚 | `docs` | Documentation |
| ♻️ | `refactor` | Restructuring |
| 🧪 | `test` | Tests |
| 📋 | `plan` | Blueprint file updates |
| 🔀 | `merge` | Branch merge |
| 🩹 | `hotfix` | Urgent fix |
| 🚀 | `deploy` | Deployment / CI |

---

## Platforms

| Binary | Platform | Arch |
|--------|----------|------|
| `blueprint-darwin-arm64` | macOS | Apple Silicon (M1–M4) |
| `blueprint-darwin-amd64` | macOS | Intel |
| `blueprint-linux-amd64` | Linux | x86_64 |

---

## Obsidian Integration

`blueprint/` is designed as a first-class Obsidian vault subfolder. With Dataview:

```dataview
TABLE complexity, linear_issue, started_at
FROM "blueprint/live"
SORT started_at DESC
```

---

## Roadmap

| Version | Scope |
|---------|-------|
| **v1.0** | 24 SDLC skills + CLI + 14-rule audit hook + Obsidian BLUE workspace |
| v1.1 | Laravel TALL preset, Node/TS preset |
| v1.2 | Standalone skills as optional add-ons |
| v1.3 | Autoresearch eval dashboard |
| v2.0 | Improved multi-agent execution, inter-agent handoffs |

---

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

---

Built by [Shirleyson Kaisser](https://github.com/skaisser)
