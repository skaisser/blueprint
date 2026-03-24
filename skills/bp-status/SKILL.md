---
name: bp-status
description: >
  Show quick repo status — branch, base, plan, PR info in one shot.
  Triggers on "/bp-status", "/status", "show status", "repo status",
  "what branch am I on", "current status", "where am I", "project status",
  "what's the current state", "which branch", "is there a PR open",
  "what plan am I on". Uses blueprint CLI with git fallback.
---

# Status: Quick Repo Context

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Show quick repo status (branch/base/plan/PR/config/changes) in one shot.

## Execution Chain

Try each source in order — use the first that succeeds.

### 1. Primary: `blueprint meta` (Go CLI)

```bash
~/.blueprint/bin/blueprint meta 2>/dev/null
```

Parses JSON output with: `branch`, `base_branch`, `plan_file`, `project`, `git_remote`.
Reads from `blueprint/live/` for active plans and `blueprint/.config.yml` for config.

### 2. Fallback: raw git + gh commands

```bash
git branch --show-current
git log --oneline -3
git status --short
gh pr view --json number,title,url,state 2>/dev/null || echo "No open PR"
ls blueprint/live/*.md 2>/dev/null || echo "No active plan"
# Read config with the Read tool instead of cat:
# Read("blueprint/.config.yml")
```

## Rules

- Do NOT modify anything — this is a read-only operation.
- Do NOT scan or analyze code — status is metadata only.
- Data must be fresh — always run commands, never rely on cached or prior results.
- Try CLI first, fall back gracefully — never error out if a source is missing.
- Show `blueprint/.config.yml` info (staging_branch, language).
- Show active plan from `blueprint/live/` if any.
- Show uncommitted changes summary.

## Output

Present results in a clean, scannable format:

```
Branch:    feat/auth-flow
Base:      staging
Plan:      blueprint/live/0002-auth-flow.md (in-progress, 12/18 tasks)
PR:        #42 — Auth flow refactor (open)
Config:    staging_branch=staging, language=auto
Changes:   3 modified, 1 untracked
```

All info in one response — no follow-up needed.
