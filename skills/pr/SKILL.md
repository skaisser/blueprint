---
name: pr
description: >
  Create a Pull Request with plan context, GitHub Issues integration, and proper base branch detection.
  Use this skill whenever the user says "/pr", "create a PR", "open a pull request",
  "make a PR", "create pull request", or any request to create a PR for the current branch.
  Also triggers on "open PR", "submit PR", "PR for this branch", "push and create PR",
  "I'm ready for a PR", or "let's open a pull request for this".
  ALWAYS runs /ship first and detects the correct base branch (feature→{staging_branch}→main flow).
---

# PR: Create Pull Request

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Create a Pull Request to base branch with descriptive summary.

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Critical Rules

- You MUST run `/ship` at Step 1. NEVER create a PR with uncommitted changes.
- You MUST detect the correct base branch. PRs NEVER target `main` directly (unless from the staging branch).
- Follow steps in order. DO NOT skip or reorder steps.
- **PR title MUST be under 70 characters.** Use the body for details, not the title.
- **NEVER add AI signatures** to PR title or body. No "Generated with Claude Code", no "Co-Authored-By", no AI attribution of any kind. The audit hook will BLOCK this.
- Do NOT run tests — CI handles that.
- Do NOT modify application code — PR is a git/GitHub operation only.
- Do NOT re-read the entire codebase — summarize from commits and plan context.

## Step 0: Plan Check Gate

```bash
~/.blueprint/bin/blueprint meta
```

Use `plan_file` from JSON output to check if a plan exists. If a plan file exists, check if `/plan-check` was run:
```bash
grep -q "Plan vs Implementation\|Plan check[:\—–-]" "$PLAN_FILE" 2>/dev/null
```

- **Plan exists + NOT checked** → STOP. Tell the user: "Run `/plan-check` first."
- **Plan exists + checked** → Continue.
- **No plan file** → This was a `/quick` task. Continue.

## Step 1: Ship Changes — MANDATORY

Run: `echo "🔷 BP: pr [1/2] shipping changes before PR"`

Run `/ship` first to commit and push all current changes. DO NOT skip this.

## Step 2: Determine Base Branch

**CRITICAL: PRs NEVER target `main` directly. Always go through the staging branch first.**

```bash
CURRENT_BRANCH=$(git branch --show-current)
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml 2>/dev/null | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}

if [ "$CURRENT_BRANCH" = "$STAGING_BRANCH" ]; then
    BASE_BRANCH="main"
elif git show-ref --verify --quiet refs/heads/$STAGING_BRANCH || git show-ref --verify --quiet refs/remotes/origin/$STAGING_BRANCH; then
    BASE_BRANCH="$STAGING_BRANCH"
else
    BASE_BRANCH="main"
fi
```

**Flow: `feature/* → {staging_branch} → main`**

## Step 3: Gather Context

```bash
~/.blueprint/bin/blueprint context "$BASE_BRANCH"
```

Read plan file from `blueprint/live/` if exists. Extract GitHub issue numbers from plan header.

## Step 4: Create PR

Run: `echo "🔷 BP: pr [2/2] creating pull request"`

Title format: `<emoji> <type>: <description>` — **MUST be under 70 characters total.**

### GitHub Issue Detection

Before composing the PR body, check the plan frontmatter for an `issue:` field:

```bash
# Read issue number(s) from plan frontmatter
ISSUE_REF=""
if [ -n "$PLAN_FILE" ]; then
    ISSUE_RAW=$(grep '^issue:' "$PLAN_FILE" | sed 's/^issue: *//')
    if [ -n "$ISSUE_RAW" ] && [ "$ISSUE_RAW" != "null" ]; then
        # Handle array format: [42, 43] or single number: 42
        if echo "$ISSUE_RAW" | grep -q '\['; then
            # Array — extract numbers and build "Closes #N" for each
            ISSUE_REF=$(echo "$ISSUE_RAW" | tr -d '[]' | tr ',' '\n' | sed 's/ //g' | while read -r n; do echo "Closes #$n"; done | paste -sd ', ' -)
        else
            ISSUE_REF="Closes #$ISSUE_RAW"
        fi
    fi
fi
```

If `ISSUE_REF` is non-empty, include it in the **References** section of the PR body.

### Create the PR

```bash
PR_URL=$(gh pr create --base "$BASE_BRANCH" --title "<emoji> <type>: <title>" --body "$(cat <<'EOF'
## Summary
[What and why — 1-3 bullet points]

## Changes
- [Change 1]
- [Change 2]

## Technical Notes
[Important details, patterns, decisions]

## Test Plan
- [ ] [How to verify change 1]
- [ ] [How to verify change 2]

## References
[ISSUE_REF if present, e.g. Closes #42]
EOF
)")
echo "$PR_URL"
```

**Important:** When composing the actual PR body, replace `[ISSUE_REF if present]` with the real `$ISSUE_REF` value. If there is no issue, omit the line entirely.

Display the PR URL to the user after creation.

## Step 5: After Creation

**STOP. You MUST use `AskUserQuestion` tool here.**

- **Question:** "PR created. What's next?"
- **Option 1:** "Run /review" — Trigger @claude code review on the PR
- **Option 2:** "I'll handle review manually"

Use $ARGUMENTS for any additional context.
