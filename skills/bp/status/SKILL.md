---
name: bp:status
description: >
  Show quick repo status — branch, base, plan, PR info in one shot.
  Triggers on "/bp:status", "/status", "show status", "repo status",
  "what branch am I on", or any request to see the current project state.
  Also triggers on "current status", "where am I", "project status", "git status",
  "what's the current state", "which branch", "is there a PR open", or "what plan am I on".
---

# Status: Quick Repo Context

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Show quick repo status (branch/base/plan/PR) in one shot.

## Data Sources
This skill uses **git** and **gh** commands to gather real-time repo state. Data must reflect the current git state — never cache or reuse stale data from prior calls.

## Run

Try `blueprint status` first. If it fails, fall back to direct commands.

### Primary
```bash
~/.blueprint/bin/blueprint status
```

### Fallback (if blueprint CLI is unavailable or errors)
Run these git/gh commands directly:
```bash
git branch --show-current
git log --oneline -3
git status --short
gh pr view --json number,title,url,state 2>/dev/null || echo "No open PR"
```
And check `blueprint/` for any active plan file.

## Rules
- Do NOT modify anything — this is a read-only operation.
- Do NOT scan or analyze code — status is metadata only.
- Data must be fresh — always run commands, never rely on cached or prior results.

## Output
Present results in a scannable format — not a wall of text. Show:
- Current branch + base branch
- Active plan (if any)
- Open PR (if any)
- Uncommitted changes summary

All info in one response — no follow-up needed.
