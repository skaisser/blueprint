---
name: finish
description: >
  Mark feature complete, merge PR, move plan to upstream, update Linear, and handle staging→main flow.
  Use this skill whenever the user says "/finish", "finish this", "merge and finish",
  "wrap up", "close this out", or any request to complete a feature and merge the PR.
  Also triggers on "merge PR", "finish the feature", "we're done", "mark as complete",
  "the PR is approved let's merge", or "feature is complete".
  Run AFTER PR is approved. Handles plan move to upstream, Linear verification, and main merge.
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

- You MUST use `AskUserQuestion` at Step 7. ALWAYS ask about merging to main. NEVER skip it.
- You MUST read the plan file at Step 2. NEVER fabricate plan details.
- Follow steps in order. DO NOT skip or reorder steps.
- Do NOT modify application code — finish is a git/project management operation only.

## Step 0: Verify PR is Approved

```bash
PR_STATE=$(gh pr view --json reviewDecision --jq '.reviewDecision' 2>/dev/null)
PR_MERGEABLE=$(gh pr view --json mergeable --jq '.mergeable' 2>/dev/null)
```

If `reviewDecision` is not `APPROVED` and the PR has required reviews, warn the user:
"PR is not yet approved. Are you sure you want to merge?"

Only proceed if the user confirms or the PR has no required review policy.

## Step 1: Determine Base Branch

Run: `echo "🏁 [finish:1] determining base branch"`

```bash
CURRENT_BRANCH=$(git branch --show-current)
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml 2>/dev/null | awk '{print $2}')
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

## Step 2: Mark Plan as Complete

Run: `echo "🏁 [finish:2] reading and marking plan complete"`

```bash
~/.blueprint/bin/blueprint meta
```

Add final session entry with `${CLAUDE_SESSION_ID}`:
- Work Sessions: `` `${CLAUDE_SESSION_ID}` DD/MM/YYYY HH:MM - Finished — `claude -r ${CLAUDE_SESSION_ID}` ``
- YAML `sessions:` array: `- id: "${CLAUDE_SESSION_ID}" date: "DD/MM/YYYY HH:MM" note: "Finished"`

Find the PR number and update the plan:

```bash
# Try current branch PR, then fall back to frontmatter pr: field
PR_NUMBER=$(gh pr view --json number --jq '.number' 2>/dev/null)
if [ -z "$PR_NUMBER" ]; then
    # PR may already be merged — check plan frontmatter
    PR_NUMBER=$(grep '^pr:' "$PLAN_FILE" | awk '{print $2}')
fi
```

## Step 3: Update Linear Ticket Status

Run: `echo "🏁 [finish:3] updating Linear ticket status"`

Read the plan frontmatter for a `linear:` field (e.g., `linear: ENG-123`).

If a Linear issue ID is found, update its status to "Done" using the Linear MCP tool:

```
Use mcp__linear__save_issue to update the issue:
  - id: <the Linear issue UUID> (look up via mcp__linear__get_issue if you only have the identifier)
  - stateId: find the "Done" state ID via mcp__linear__list_issue_statuses for the issue's team
```

If no `linear:` field exists in the plan frontmatter, skip this step silently — not all features have Linear tickets.

## Step 3b: Archive Backlog Item

Run: `echo "🏁 [finish:3b] archiving backlog item"`

Read the plan frontmatter for a `backlog:` field (e.g., `backlog: "0014"`).

If a backlog ID is found:
1. Use `blueprint backlog --archive` (JSON) to locate the file by ID — never parse files manually with grep/sed
2. Read the file, update its frontmatter: `status: archived`
3. If not already in `blueprint/expired/`, move it: `git mv blueprint/backlog/NNNN-*.md blueprint/expired/`

If no `backlog:` field exists, skip silently — not all plans come from the backlog.

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

## Step 5: Commit and Push

```bash
# 1. Commit plan file changes to project repo
git add blueprint/ && git commit -m "🧹 chore: finish NNNN-<description>"

# 2. Commit any remaining project code changes (if any)
/commit  # only if there are staged project code changes

# 3. Push project branch before merging
git push
```

## Step 5b: Trigger CI Tests Before Merge

Run: `echo "🏁 [finish:5b] triggering CI tests"`

Check if the project has an on-demand test workflow. The workflow must exist on the `main` branch (GitHub runs `issue_comment` workflows from the default branch).

```bash
# Check if tests.yml exists on the default branch (not just locally)
gh api "repos/{owner}/{repo}/contents/.github/workflows/tests.yml" --jq '.name' 2>/dev/null
```

If it exists, trigger the tests and wait:

```bash
# 1. Post @tests comment to trigger the workflow
gh pr comment "$PR_NUMBER" --body "@tests"

# 2. Wait for the workflow to start (15s — it takes a moment to queue)
sleep 15

# 3. Get the latest run ID for the tests workflow
RUN_ID=$(gh run list --workflow=tests.yml --limit=1 --json databaseId -q '.[0].databaseId')

# 4. Watch the run until it completes
if [ -n "$RUN_ID" ] && [ "$RUN_ID" != "null" ]; then
    gh run watch "$RUN_ID" --exit-status
    RESULT=$?
    if [ $RESULT -ne 0 ]; then
        echo "CI tests failed — cannot merge"
        # STOP. Use AskUserQuestion:
        # "CI tests failed on PR #XX. What do you want to do?"
        # Option 1: "Fix and retry" — investigate failures
        # Option 2: "Merge anyway" — skip CI gate
        # Option 3: "Abort" — don't merge
    else
        echo "CI tests passed"
    fi
fi
```

The bot will also post a result comment on the PR.

If the workflow doesn't exist on the default branch, skip this step silently — not all projects have CI tests configured.

## Step 5c: Pre-merge Conflict Check

Run: `echo "🏁 [finish:5c] checking for merge conflicts"`

Before merging, ensure the branch is up to date with the base branch. This prevents merge failures, especially in batch flows where `blueprint/` files diverge:

```bash
# Check if branch is behind base
BEHIND=$(git rev-list --count HEAD..origin/$BASE_BRANCH 2>/dev/null || echo "0")
if [ "$BEHIND" -gt 0 ]; then
    echo "Branch is $BEHIND commits behind $BASE_BRANCH — merging base into feature branch"
    git fetch origin "$BASE_BRANCH"
    git merge "origin/$BASE_BRANCH" --no-edit || {
        # Check if conflicts are only in blueprint/
        CONFLICTED=$(git diff --name-only --diff-filter=U)
        if echo "$CONFLICTED" | grep -qv "^blueprint/"; then
            echo "Code conflicts detected — manual resolution needed"
            # STOP. Use AskUserQuestion to inform the user.
        else
            # Plan-only conflicts: take theirs (base branch has the latest)
            echo "Resolving blueprint/ conflicts (taking base branch version)"
            git checkout --theirs blueprint/ && git add blueprint/
            git commit -m "🔀 merge: resolve plan file conflicts with $BASE_BRANCH"
        fi
    }
    git push
fi
```

## Step 6: Merge PR and Clean Up Branch

Run: `echo "🏁 [finish:6] merging PR and cleaning up branch"`

```bash
# Merge PR (deletes remote branch)
gh pr merge <PR_NUMBER> --merge --delete-branch
```

**Verify merge succeeded before continuing:**
```bash
MERGE_STATE=$(gh pr view "$PR_NUMBER" --json state --jq '.state')
if [ "$MERGE_STATE" != "MERGED" ]; then
    echo "PR merge failed — check for conflicts or failing checks"
    # STOP here. Use AskUserQuestion to inform the user and ask how to proceed.
    # Do NOT delete branches or continue if merge failed.
fi
```

Only after confirmed merge:
```bash
# Switch to base branch and delete local feature branch
FEATURE_BRANCH=$(git branch --show-current)
git checkout "$BASE_BRANCH"
git pull
git branch -d "$FEATURE_BRANCH"
```

## Step 7: Handle Main Branch — MANDATORY

Run: `echo "🏁 [finish:7] asking user about main merge"`

**STOP. You MUST use `AskUserQuestion` tool here. ALWAYS.**

If `HAS_STAGING` is true (merged to the staging branch):
- **Question:** "PR merged to staging branch. What about main?"
- **Option 1:** "Merge to main now" — Create PR and merge (see commands below)
- **Option 2:** "Create PR, I'll merge manually" — Create PR only
- **Option 3:** "I'll do it later"

**Merge to main commands (use `gh pr`, NEVER `gh api`):**
```bash
# Create PR from staging → main
gh pr create --base main --head "$STAGING_BRANCH" \
  --title "<emoji> <type>: <description>" \
  --body "Merges $STAGING_BRANCH → main. Contains PR #<NUMBER>: <brief>"

# Merge immediately
MAIN_PR=$(gh pr list --base main --head "$STAGING_BRANCH" --json number --jq '.[0].number')
gh pr merge "$MAIN_PR" --merge
```

If `HAS_STAGING` is false (merged directly to main):
- **Question:** "PR merged to main. Anything else needed?"
- **Option 1:** "All done"
- **Option 2:** "Deploy / run migrations"

## Step 8: Done

Report merges, Linear status (updated to Done if applicable), and post-deploy reminders.
If worktree: remind to run `/complete` from main repo terminal for cleanup.

Use $ARGUMENTS for any additional context.
