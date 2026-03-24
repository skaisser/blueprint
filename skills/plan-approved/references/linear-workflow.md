# Linear Integration Workflow

## Overview

Linear issues are tracked throughout the SDLC pipeline. Use `~/.claude/scripts/linear.py` for all Linear operations — it calls the GraphQL API directly, which is much faster than MCP.

## Project Config Detection

**Every command that touches Linear MUST run this first:**

```bash
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
LINEAR_FILE="$REPO_ROOT/.linear"
if [ -f "$LINEAR_FILE" ]; then
    source "$LINEAR_FILE"   # exports: TEAM, PROJECT_ID, PROJECT_URL (optional)
    HAS_LINEAR=true
else
    HAS_LINEAR=false
fi
```

- `HAS_LINEAR=false` → skip ALL Linear steps silently, no warnings, no friction
- `HAS_LINEAR=true` → `$TEAM` overrides `DEFAULT_TEAM` from `~/.linear-config`
- If explicit issue IDs were given in `$ARGUMENTS` (e.g. `/plan KPG-26 ...`), always process them regardless of `HAS_LINEAR` (user knows what they're doing)

## Quick Reference

### Compound Commands

```bash
~/.claude/scripts/linear.py start KPG-26 "Implementation started — branch: feat/foo"
~/.claude/scripts/linear.py review KPG-26 "PR created: https://github.com/..."
~/.claude/scripts/linear.py finish KPG-26 "Merged to staging branch — plan 0012 complete"
```

### Atomic Commands

```bash
~/.claude/scripts/linear.py get KPG-26
~/.claude/scripts/linear.py status KPG-26 "In Progress"
~/.claude/scripts/linear.py comment KPG-26 "message"
~/.claude/scripts/linear.py create "Title" "Description"
~/.claude/scripts/linear.py search "login bug"
~/.claude/scripts/linear.py states KPG
~/.claude/scripts/linear.py teams
```

## Status Mapping

| Workflow Step | Linear Status | Command |
|---------------|---------------|---------|
| `/plan` creates plan | **In Progress** | `linear.py start KPG-26 "message"` |
| `/plan-approved` starts | Verify **In Progress** | `linear.py get KPG-26` |
| `/pr` creates PR | **In Review** | `linear.py review KPG-26 "PR URL"` |
| PR merged (GitHub) | **Done** | GitHub auto-closes via `Closes KPG-XX` |
| `/finish` after merge | Verify **Done** | `linear.py finish KPG-26 "message"` |

## Error Handling

- If `linear.py` fails: warn user, don't block workflow
- Linear integration is **additive** — workflow works without it
