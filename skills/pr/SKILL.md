---
name: pr
description: >
  Create a Pull Request with plan context, GitHub issue integration, and proper base branch detection.
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
blueprint meta
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
BASE_BRANCH=$(blueprint base-branch)
CURRENT_BRANCH=$(git branch --show-current)
```

**Flow: `feature/* → {staging_branch} → main`**

## Step 3: Gather Context

```bash
blueprint context "$BASE_BRANCH"
```

Read plan file from `blueprint/live/` if exists. Extract GitHub issue numbers from plan header.

## Step 4: Create PR

Run: `echo "🔷 BP: pr [2/2] creating pull request"`

Title format: `<emoji> <type>: <description>` — **MUST be under 70 characters total.**

### GitHub Issue Detection

Before composing the PR body, check the plan frontmatter for an `issue:` field:

```bash
ISSUE_REF=$(blueprint meta issue 2>/dev/null || echo "")
```

`blueprint meta issue` returns a single number (e.g. `42`) or a JSON array (e.g. `[42, 43]`). If non-empty and not `null`, build `Closes #N` lines for each issue number and include them in the **References** section of the PR body.

### Create the PR

```bash
PR_BODY=$(blueprint pr-body --base "$BASE_BRANCH" 2>/dev/null)
# Falls back to manual body template if blueprint pr-body unavailable

PR_URL=$(gh pr create --base "$BASE_BRANCH" --title "<emoji> <type>: <title>" --body "$PR_BODY")
echo "$PR_URL"
```

If `blueprint pr-body` is unavailable or returns empty, compose the body manually:

```
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
[Closes #N for each GitHub issue if ISSUE_REF is non-empty]
```

Display the PR URL to the user after creation.

## Step 5: After Creation

**STOP. You MUST use `AskUserQuestion` tool here.**

- **Question:** "PR created. What's next?"
- **Option 1:** "Run /review" — Trigger @claude code review on the PR
- **Option 2:** "I'll handle review manually"

Use $ARGUMENTS for any additional context.
