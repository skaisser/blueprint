---
name: finish
description: >
  Mark feature complete, merge PR, move plan to upstream, and handle staging→main flow.
  Use this skill whenever the user says "/finish", "finish this", "merge and finish",
  "wrap up", "close this out", or any request to complete a feature and merge the PR.
  Also triggers on "merge PR", "finish the feature", "we're done", "mark as complete",
  "the PR is approved let's merge", or "feature is complete".
  Run AFTER PR is approved. Handles plan move to upstream and main merge.
---

# Finish: Complete Feature and Merge

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Mark the current feature as complete and merge to base branch (and optionally to main).

Run AFTER PR is approved.

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Critical Rules

- You MUST use `AskUserQuestion` at Step 6. ALWAYS ask about merging to main. NEVER skip it.
- You MUST read the plan file at Step 2. NEVER fabricate plan details.
- Follow steps in order. DO NOT skip or reorder steps.
- Do NOT modify application code — finish is a git/project management operation only.
- Use `blueprint` CLI first, fall back to `gh`/`git` commands if CLI unavailable.

## Step 0: Verify PR is Approved

```bash
PR_STATE=$(gh pr view --json reviewDecision --jq '.reviewDecision' 2>/dev/null)
PR_MERGEABLE=$(gh pr view --json mergeable --jq '.mergeable' 2>/dev/null)
```

If `reviewDecision` is not `APPROVED` and the PR has required reviews, warn the user:
"PR is not yet approved. Are you sure you want to merge?"

Only proceed if the user confirms or the PR has no required review policy.

## Step 1: Determine Base Branch

Run: `echo "🔷 BP: finish [1/6] determining base branch"`

```bash
CURRENT_BRANCH=$(git branch --show-current)

# Read staging branch from config — CLI first, fallback to grep
STAGING_BRANCH=$(~/.blueprint/bin/blueprint meta 2>/dev/null | grep -o '"base_branch":"[^"]*"' | cut -d'"' -f4)
if [ -z "$STAGING_BRANCH" ]; then
    STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml 2>/dev/null | awk '{print $2}')
fi
STAGING_BRANCH=${STAGING_BRANCH:-staging}

if [ "$CURRENT_BRANCH" = "$STAGING_BRANCH" ]; then
    BASE_BRANCH="main"
    HAS_STAGING=false
elif git show-ref --verify --quiet refs/heads/$STAGING_BRANCH || git show-ref --verify --quiet refs/remotes/origin/$STAGING_BRANCH; then
    BASE_BRANCH="$STAGING_BRANCH"
    HAS_STAGING=true
else
    BASE_BRANCH="main"
    HAS_STAGING=false
fi
```

## Step 2: Read Plan and Mark Complete

Run: `echo "🔷 BP: finish [2/6] reading plan"`

```bash
# CLI first — returns plan_file path
PLAN_FILE=$(~/.blueprint/bin/blueprint meta 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('plan_file',''))" 2>/dev/null)

# Fallback — find active plan in blueprint/live/
if [ -z "$PLAN_FILE" ]; then
    PLAN_FILE=$(ls blueprint/live/[0-9]*-*.md 2>/dev/null | head -1)
fi
```

Read the plan file. Update frontmatter:
- `status: completed`
- `completed_at: DD/MM/YYYY HH:MM`
- `session: "${CLAUDE_SESSION_ID}"` (for `claude -r` resume — preserves the session that finished the feature)

Find the PR number:

```bash
PR_NUMBER=$(gh pr view --json number --jq '.number' 2>/dev/null)
if [ -z "$PR_NUMBER" ]; then
    PR_NUMBER=$(grep '^pr:' "$PLAN_FILE" | awk '{print $2}')
fi
```

## Step 3: Archive Backlog Item (if applicable)

Run: `echo "🔷 BP: finish [3/6] archiving backlog item"`

Read the plan frontmatter for a `backlog:` field (e.g., `backlog: "0014"`).

If a backlog ID is found:
1. Locate the file in `blueprint/backlog/` matching the ID
2. Update its frontmatter: `status: archived`
3. Move it: `git mv blueprint/backlog/NNNN-*.md blueprint/expired/`

If no `backlog:` field exists, skip silently.

## Step 4: Move Plan to Upstream

Move the plan from `blueprint/live/` to `blueprint/upstream/` with `-complete` suffix:

```bash
PLAN_FILE_BASENAME=$(basename "$PLAN_FILE")
DONE_FILE="blueprint/upstream/${PLAN_FILE_BASENAME%.md}-complete.md"
mkdir -p blueprint/upstream
if [ -f "$PLAN_FILE" ]; then
    git mv "$PLAN_FILE" "$DONE_FILE"
elif [ -f "$DONE_FILE" ]; then
    echo "Plan already moved to upstream"
fi
```

## Step 5: Commit, Push, and Pre-merge Checks

```bash
echo "🔷 BP: finish [4/6] committing and pushing"

# Commit plan file changes
git add blueprint/ && git commit -m "🧹 chore: finish NNNN-<description>"

# Push before merging
git push
```

### 5a: Trigger CI Tests (if available)

Check if the project has an on-demand test workflow on the default branch:

```bash
gh api "repos/{owner}/{repo}/contents/.github/workflows/tests.yml" --jq '.name' 2>/dev/null
```

If it exists, trigger and wait:

```bash
gh pr comment "$PR_NUMBER" --body "@tests"
sleep 15
RUN_ID=$(gh run list --workflow=tests.yml --limit=1 --json databaseId -q '.[0].databaseId')
if [ -n "$RUN_ID" ] && [ "$RUN_ID" != "null" ]; then
    gh run watch "$RUN_ID" --exit-status
    if [ $? -ne 0 ]; then
        # STOP. AskUserQuestion: "CI tests failed. Fix and retry / Merge anyway / Abort"
    fi
fi
```

If the workflow doesn't exist, skip silently.

### 5b: Pre-merge Conflict Check

```bash
BEHIND=$(git rev-list --count HEAD..origin/$BASE_BRANCH 2>/dev/null || echo "0")
if [ "$BEHIND" -gt 0 ]; then
    echo "Branch is $BEHIND commits behind $BASE_BRANCH — rebasing"
    git fetch origin "$BASE_BRANCH"
    git merge "origin/$BASE_BRANCH" --no-edit || {
        CONFLICTED=$(git diff --name-only --diff-filter=U)
        if echo "$CONFLICTED" | grep -qv "^blueprint/"; then
            echo "Code conflicts detected — manual resolution needed"
            # STOP. AskUserQuestion to inform user.
        else
            echo "Resolving blueprint/ conflicts (taking base branch version)"
            git checkout --theirs blueprint/ && git add blueprint/
            git commit -m "🔀 merge: resolve plan file conflicts with $BASE_BRANCH"
        fi
    }
    git push
fi
```

## Step 6: Merge PR and Clean Up

Run: `echo "🔷 BP: finish [5/6] merging PR"`

```bash
gh pr merge "$PR_NUMBER" --merge --delete-branch
```

Verify merge succeeded:
```bash
MERGE_STATE=$(gh pr view "$PR_NUMBER" --json state --jq '.state')
if [ "$MERGE_STATE" != "MERGED" ]; then
    # STOP. AskUserQuestion: "PR merge failed — check for conflicts or failing checks"
fi
```

Only after confirmed merge:
```bash
FEATURE_BRANCH=$(git branch --show-current)
git checkout "$BASE_BRANCH"
git pull
git branch -d "$FEATURE_BRANCH"
```

### 6b: Verify GitHub Issue Closure

If the plan has an `issue:` field in frontmatter, verify the issue was auto-closed by the PR merge:

```bash
# Read issue number(s) from plan frontmatter
if [ -n "$PLAN_FILE" ]; then
    ISSUE_RAW=$(grep '^issue:' "$PLAN_FILE" | sed 's/^issue: *//')
    if [ -n "$ISSUE_RAW" ] && [ "$ISSUE_RAW" != "null" ]; then
        # Handle array format: [42, 43] or single number: 42
        ISSUE_NUMBERS=$(echo "$ISSUE_RAW" | tr -d '[]' | tr ',' '\n' | sed 's/ //g')
        for ISSUE_NUMBER in $ISSUE_NUMBERS; do
            ISSUE_STATE=$(gh issue view "$ISSUE_NUMBER" --json state --jq '.state')
            if [ "$ISSUE_STATE" != "CLOSED" ]; then
                echo "Warning: Issue #$ISSUE_NUMBER not auto-closed — closing manually"
                gh issue close "$ISSUE_NUMBER"
            else
                echo "Issue #$ISSUE_NUMBER confirmed closed"
            fi
        done
    fi
fi
```

## Step 7: Handle Main Branch — MANDATORY

Run: `echo "🔷 BP: finish [6/6] asking user about main merge"`

**STOP. You MUST use `AskUserQuestion` here. ALWAYS.**

If `HAS_STAGING` is true (merged to staging branch):
- **Question:** "PR merged to {STAGING_BRANCH}. What about main?"
- **Option 1:** "Merge to main now" — Create PR and merge
- **Option 2:** "Create PR to main, I'll merge manually"
- **Option 3:** "I'll do it later"

**Merge to main commands:**
```bash
gh pr create --base main --head "$STAGING_BRANCH" \
  --title "<emoji> <type>: <description>" \
  --body "Merges $STAGING_BRANCH → main. Contains PR #$PR_NUMBER: <brief>"

MAIN_PR=$(gh pr list --base main --head "$STAGING_BRANCH" --json number --jq '.[0].number')
gh pr merge "$MAIN_PR" --merge
```

If `HAS_STAGING` is false (merged directly to main):
- **Question:** "PR merged to main. Anything else?"
- **Option 1:** "All done"
- **Option 2:** "Deploy / run post-merge tasks"

## Step 8: Done

Report what happened:
- PR merged to which branch
- Plan moved to `blueprint/upstream/`
- Backlog item archived (if applicable)
- Main merge status

If worktree: remind to run `/complete` from the main repo terminal for cleanup.

Use $ARGUMENTS for any additional context.
