<p align="center">
  <a href="https://github.com/skaisser/blueprint/releases/latest"><img src="https://img.shields.io/github/v/release/skaisser/blueprint?style=flat-square&color=0C447C" alt="Release" /></a>
  <img src="https://img.shields.io/badge/skills-27-185FA5?style=flat-square" alt="Skills" />
  <img src="https://img.shields.io/badge/audit_rules-15-534AB7?style=flat-square" alt="Audit Rules" />
  <img src="https://img.shields.io/badge/Claude_Code-required-blueviolet?style=flat-square" alt="Claude Code" />
  <img src="https://img.shields.io/badge/macOS-arm64_·_amd64-black?style=flat-square&logo=apple" alt="macOS" />
  <img src="https://img.shields.io/badge/Linux-amd64-black?style=flat-square&logo=linux" alt="Linux" />
  <img src="https://img.shields.io/badge/license-Apache_2.0-0F6E56?style=flat-square" alt="License" />
</p>

<p align="center">
  <img src="assets/blueprint.png" alt="Blueprint — SDLC Pipeline for Claude Code" width="720" />
</p>

<h1 align="center">BLUEPRINT SDLC</h1>

<p align="center">
  A complete, portable software development lifecycle for Claude Code.<br/>
  <strong>Stack-agnostic. Zero paid dependencies. Works on any Claude Code project.</strong>
</p>

---

BLUEPRINT turns Claude Code from a code assistant into a disciplined engineering partner — with planning, execution, review, and merge all governed by a structured pipeline of slash commands, an audit hook enforcing 15 rules on every tool call, and a Go CLI binary shipping pre-compiled for macOS and Linux.

> If you've tried GTD-style workflows with Claude Code, you know the pain: too slow, too manual, too much overhead. BLUEPRINT fixes that.
>
> **GTD taught you to capture everything. BLUEPRINT ships it.**

---

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/skaisser/blueprint/main/install.sh | bash
```

No Go toolchain required. The installer auto-detects your platform (macOS + Linux), shows an interactive menu via [gum](https://github.com/charmbracelet/gum), and installs only what you choose:

- `blueprint` CLI binary → `~/.blueprint/bin/`
- 27 SDLC skills → `~/.claude/skills/`
- Audit hook (15 rules) via Claude Code settings
- Git hooks (commit-msg, pre-push)
- GitHub Action (`@claude` review)
- Obsidian setup with Dataview queries *(optional)*

> **Just want BLUEPRINT? One line. Want to hack the CLI? Clone it.**

### Clone (contributors / CLI hackers)

```bash
git clone git@github.com:skaisser/blueprint.git ~/Sites/blueprint
cd ~/Sites/blueprint && ./install.sh
```

```bash
cd cli && make build-all   # builds arm64 · amd64 · linux
```

---

## The Pipeline

Every letter in **BLUEPRINT** maps to a pipeline phase:

| # | | Command | Phase |
|---|---|---------|-------|
| 1 | **B** | `/backlog` | **B**acklog — capture and prioritise ideas |
| 2 | **L** | `/plan` | **L**ayout — create branch + blueprint file |
| 3 | **U** | `/plan-review` | **U**npack — validate, assign complexity |
| 4 | **E** | `/plan-approved` | **E**ndorse — execute, spawn parallel subagents |
| 5 | **P** | `/plan-check` | **P**reflight — audit code vs blueprint |
| 6 | **R** | `/pr` | **R**aise — open pull request with full context |
| 7 | **I** | `/review` | **I**nspect — trigger @claude code review |
| 8 | **N** | `/address-pr` | **N**egotiate — fetch feedback, fix, push |
| 9 | **T** | `/finish` | **T**ag — merge, rename blueprint to `upstream/` |

---

## The BLUE Workspace

The `blueprint/` directory uses the first four letters as folder names — your file path _is_ your status:

| | Folder | Trigger | Meaning |
|---|--------|---------|---------|
| **B** | `blueprint/backlog/` | `/backlog` | Ideas not yet planned |
| **L** | `blueprint/live/` | `/plan` | Currently in development |
| **U** | `blueprint/upstream/` | `/finish` | Shipped and merged |
| **E** | `blueprint/expired/` | `/backlog --archive` | Cancelled or deferred |

Files move between folders automatically on each phase transition. Compatible with [Obsidian](https://obsidian.md) + Dataview out of the box — open `blueprint/` as a vault and your kanban is ready.

---

## Quick Start

### 1. Initialize a project

```bash
/start
```

Sets up git hooks, CLAUDE.md, `blueprint/` workspace (BLUE folders), GitHub Action, and a configurable staging branch.

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
/flow-auto            # Zero-touch — model decides everything, PR ready for merge
/batch-flow 2-6       # Execute blueprints 0002–0006 sequentially
```

---

## Why BLUEPRINT?

| Before | BLUEPRINT |
|--------|-----------|
| Manual task capture | `/backlog` — one command, file created instantly |
| You manage the system | Audit hook enforces the system for you |
| Context switches kill flow | 1M context + `/flow` keeps everything in one session |
| Folders you maintain by hand | BLUE folders move automatically on phase transitions |
| Trust yourself to follow the process | 15 rules catch you when you don't |
| Someday/Maybe pile | `blueprint/expired/` — archived, not lost |

---

## What's Included

### Core SDLC — 27 skills

| Category | Skills |
|----------|--------|
| Pipeline | `/backlog` `/plan` `/plan-review` `/plan-approved` `/plan-check` `/pr` `/review` `/address-pr` `/finish` |
| Automation | `/flow` `/flow-auto` `/flow-auto-wt` `/batch-flow` |
| Fast Tracks | `/quick` `/hotfix` `/resume` |
| Git & PR | `/bp-commit` `/bp-ship` `/bp-push` `/bp-branch` |
| Testing | `/bp-test` `/bp-tdd-review` |
| Project Setup | `/start` `/bp-context` `/bp-status` `/complete` |
| Skill Factory | `/skill-creator` |

---

## Execution Strategies

`/plan-review` assigns complexity and picks the fastest execution mode automatically:

| Complexity | Meaning |
|------------|---------|
| `[H]` | Fast — small scope, execute immediately |
| `[S]` | Balanced — standard parallel execution |
| `[O]` | Deep reasoning — sequential, full context required |

| Strategy | When | How |
|----------|------|-----|
| Parallel Subagents *(default)* | 2+ independent phases | Multiple Agent calls in one message — true parallelism |
| Coordinated Team | Workers need mid-task handoffs | Team messaging between agents |
| Single Subagent | 1 phase or strictly sequential | One Agent call, no spawn overhead |
| Leader Direct | ≤3 `[H]` tasks total | Lead model handles directly |

---

## Audit Hook — 15 Rules

`blueprint audit` fires on every Claude Code tool call via PreToolUse. The Go binary is the only hook — fast, compiled, zero dependencies.

| # | Rule | What it enforces |
|---|------|-----------------|
| 1 | Skill read gate | Block writes without reading relevant SKILL.md |
| 2 | Reference tracking | Track reads of key reference files |
| 3 | Team compliance | Warn if teams used without reading team-execution.md |
| 4 | Standalone task count | Warn if 3+ tasks with no team |
| 5 | Handoff tracking | Track checkpoints at `/flow` pauses |
| 6 | Checkpoint audit trail | Enforce `/plan-check` before `/pr` |
| 7 | Workflow creation gate | Block `claude-pr-reviewer.yml` without a staging branch |
| 8 | Test suite enforcement | Block test runner without `--parallel` or `--filter` |
| 9 | Plan task deletion | Warn when unchecked tasks removed instead of implemented |
| 10 | Dangerous command block | Block `migrate:fresh`, AI signatures, direct push to main |
| 11 | Review enforcement | Block short `@claude review` — require full prompt |
| 12 | Plan-check skip detection | Warn if `/pr` invoked without `/plan-check` |
| 13 | Acceptance criteria gate | Warn if PR has unchecked acceptance criteria |
| 14 | Flow-auto enforcement | Block PR if mandatory steps were skipped |
| 15 | Backlog CLI enforcement | Block manual backlog file parsing — use `blueprint backlog` CLI |

---

## Commit Format

BLUEPRINT enforces emoji + type on every commit via the `commit-msg` hook. AI signatures (`Co-Authored-By`, `Generated by Claude`) are blocked.

```
<emoji> <type>: <description>   (present tense, lowercase)
```

| Emoji | Type | Use case |
|-------|------|----------|
| ✨ | `feat` | New feature |
| 🐛 | `fix` | Bug fix |
| 📚 | `docs` | Documentation |
| ♻️ | `refactor` | Restructuring, no behavior change |
| 🧪 | `test` | Tests only |
| 📋 | `plan` | Blueprint file updates |
| 🔀 | `merge` | Branch merge |
| 🩹 | `hotfix` | Urgent production fix |
| 🚀 | `deploy` | Deployment / CI |

---

## Blueprint Frontmatter

Every blueprint file is Obsidian + Dataview compatible:

```yaml
---
id: 0002
title: Auth flow refactor
type: feature
status: live              # backlog | live | upstream | expired
issue: null
branch: feat/auth-flow-refactor
base: main
strategy: parallel
session: null
pr: null
created: "20/03/2026 14:00"
completed: null
tags: [auth, security]
---
```

```dataview
TABLE type, issue, created
FROM "blueprint/live"
SORT created DESC
```

---

## Platforms

| Binary | Platform | Arch |
|--------|----------|------|
| `blueprint-darwin-arm64` | macOS | Apple Silicon (M1–M4) |
| `blueprint-darwin-amd64` | macOS | Intel |
| `blueprint-linux-amd64` | Linux | x86_64 |

---

## Roadmap

| Version | Scope |
|---------|-------|
| **v1.0** | 27 SDLC skills · `blueprint` CLI · 15-rule audit hook · Obsidian BLUE workspace |
| v1.1 | Laravel TALL preset · Node/TS preset |
| v1.2 | Standalone skills as optional add-ons |
| v1.3 | Autoresearch eval dashboard + optimizer |
| v2.0 | Improved multi-agent execution · inter-agent blueprint handoffs |

---

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

---

<p align="center">
  Built by <a href="https://github.com/skaisser">Shirleyson Kaisser</a> ·
  <a href="https://blueprint.skaisser.dev">blueprint.skaisser.dev</a>
</p>
