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

### `.linear` file format (project root)

```ini
TEAM=KPG
PROJECT_ID=9d9ba44425f7
PROJECT_URL=https://linear.app/kpgsa/project/paylog-9d9ba44425f7/overview
```

- `TEAM` — required. The issue prefix (KPG, ARK, etc.)
- `PROJECT_ID` — optional. Linear project UUID for linking new issues
- `PROJECT_URL` — optional. For reference only

### Setup per project

```bash
# In the project root:
cat > .linear << 'EOF'
TEAM=KPG
PROJECT_ID=9d9ba44425f7
EOF
```

Commit it — it's not sensitive. Projects without `.linear` simply don't use Linear.

## Global Setup

1. Get API key: https://linear.app/settings/api
2. Save it: `echo "lin_api_YOURKEY" > ~/.linear-api-key && chmod 600 ~/.linear-api-key`
3. Set fallback defaults (used when no `.linear` file exists):
   ```bash
   echo "DEFAULT_TEAM=KPG" > ~/.linear-config
   echo "DEFAULT_PROJECT=PayLog" >> ~/.linear-config
   ```

## Quick Reference

### Compound Commands (use these — one call does status + comment)

```bash
# /plan — after branch created: In Progress + comment
~/.claude/scripts/linear.py start KPG-26 "Implementation started — branch: feat/foo"

# /pr — after PR created: In Review + comment
~/.claude/scripts/linear.py review KPG-26 "PR created: https://github.com/..."

# /finish — after merge: verify Done (handles GitHub auto-close), add comment
~/.claude/scripts/linear.py finish KPG-26 "Merged to staging branch — plan 0012 complete"
```

### Atomic Commands (use when you need just one action)

```bash
# Fetch issue details
~/.claude/scripts/linear.py get KPG-26

# Update status only
~/.claude/scripts/linear.py status KPG-26 "In Progress"

# Add comment only
~/.claude/scripts/linear.py comment KPG-26 "message"

# Create issue (uses DEFAULT_TEAM from ~/.linear-config)
~/.claude/scripts/linear.py create "Feat: TikTok Shop Integration" "Plan overview + acceptance criteria"

# Create issue for specific team
~/.claude/scripts/linear.py create "Title" "Description" KPG

# Search issues
~/.claude/scripts/linear.py search "login bug"

# List workflow states and teams
~/.claude/scripts/linear.py states KPG
~/.claude/scripts/linear.py teams
```

## Status Mapping

| Workflow Step | Linear Status | Command |
|---------------|---------------|---------|
| `/plan` creates plan | **In Progress** | `linear.py start KPG-26 "message"` |
| `/plan-approved` starts | Verify **In Progress** | `linear.py get KPG-26` (check status) |
| `/pr` creates PR | **In Review** | `linear.py review KPG-26 "PR URL"` |
| PR merged (GitHub) | **Done** | GitHub auto-closes via `Closes KPG-XX` |
| `/finish` after merge | Verify **Done** | `linear.py finish KPG-26 "message"` (checks + fallback) |

## Plan Header Format

```markdown
> **Linear:** KPG-26, KPG-11, KPG-24
```

Multiple issues can be linked to one plan. Each issue ID should match the Linear project prefix + number.

## Branch Naming

Linear's GitHub integration matches issue IDs in:
1. **Branch name**: `feat/kpg-26-tiktok-shop` (preferred — auto-links)
2. **PR title/body**: `Closes KPG-26` or `Fixes KPG-26`
3. **Commit messages**: `KPG-26` anywhere in the message

For plans with multiple issues, use a descriptive branch name and put all issue IDs in the PR body:
```
Branch: feat/status-code-improvements
PR body: Closes KPG-26, KPG-11, KPG-24, KPG-17, KPG-25, KPG-16
```

## PR Body Format

Include Linear references in the PR body for GitHub integration:

```markdown
## Summary
[What and why]

## Linear
Closes KPG-26, Closes KPG-11, Closes KPG-24

## Changes
- [Change 1]
- [Change 2]
```

Each `Closes KPG-XX` triggers GitHub integration to move the issue to Done on merge.

## Commands That Touch Linear

All commands check `HAS_LINEAR` (from `.linear` detection above) before doing anything.
Exception: explicit issue IDs in `$ARGUMENTS` always run regardless.

| Command | Linear Action | Gated by `.linear`? |
|---------|--------------|---------------------|
| `/plan` | Fetch details, create issue if none, move to **In Progress** | Yes — Case 2 (no IDs) skipped if `HAS_LINEAR=false` |
| `/plan-review` | Verify issues exist (optional) | Yes |
| `/plan-approved` | Verify **In Progress**, add execution comment | Yes |
| `/pr` | Move to **In Review**, comment with PR URL | Yes (already gated by `> **Linear:**` in plan) |
| `/finish` | Verify **Done**, fallback update | Yes (already gated by `> **Linear:**` in plan) |
| `/quick` | Accept issue ID, move through In Progress → Done | Explicit ID only — no auto-detection needed |

## Error Handling

- If `linear.py` fails (no API key, network error): warn user, don't block workflow
- If issue ID not found: warn and continue (might be typo)
- If status update fails: log warning, don't block workflow
- Linear integration is **additive** — workflow works without it, just without status sync

## Caching

The script caches team IDs and workflow states for 24 hours in `~/.claude/cache/linear/`. To refresh:
```bash
~/.claude/scripts/linear.py cache-clear
```
