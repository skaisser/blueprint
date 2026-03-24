---
name: batch-flow
description: >
  Execute multiple plans sequentially in a single session — loops flow-auto for plans N through M
  with auto context compaction, merge chain automation, and crash recovery.
  Use this skill whenever the user says "/batch-flow", "batch flow", "run plans 2 to 6",
  "execute all plans", "sequential plans", or any request to run multiple plans in one session.
  Also triggers on "multi-plan", "batch execute", "run remaining plans", "plans N through M",
  "batch pipeline", "run plans 3 4 and 5", "I have N plans to execute",
  or "execute plan N through plan M". This is the multi-plan wrapper around /flow-auto.
  IMPORTANT: never use AskUserQuestion in this skill — all decisions are made autonomously.
---

# Batch Flow: Multi-Plan Sequential Execution

Run multiple plans sequentially in a single session. Wraps `/flow-auto-wt` (worktree-isolated) in a loop with context management, merge chain automation, and crash recovery.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

```
/batch-flow N-M [--auto-merge] [--effort-budget Nm]
  → For each plan N to M:
    → flow-auto pipeline → merge chain → context check → next plan
  Zero AskUserQuestion calls. One command → N plans executed.
```

## Why This Exists

Building an MVP often requires 5-10 sequential plans. Running `/flow-auto` manually for each one means:
- Re-entering the command 5-10 times
- Manual branch switching between plans
- No context management between plans
- No crash recovery

`/batch-flow` automates the entire loop.

## Critical Rules

1. **NEVER use AskUserQuestion.** Fully autonomous — same as flow-auto.
2. **NEVER merge to main without PR.** The merge chain creates PRs for every merge.
3. **Context compaction between plans.** If context usage exceeds 60% after a plan, compact before starting the next.
4. **Skip completed plans.** If a plan file has status `Completed`, skip it.
5. **Resume from last incomplete.** On crash/restart, detect the last incomplete plan and resume from there.
6. **Each plan gets its own branch.** Never reuse branches across plans.
7. **Effort budget applies per plan.** If `--effort-budget` is set, each plan gets that budget — not the total.

---

## Step 1: Parse Arguments and Discover Plans

```bash
echo "🤖 [batch-flow:1] initializing batch pipeline"
```

Parse $ARGUMENTS for:
- `N-M` or `N M` → plan range (e.g., `2-6` means plans 0002 through 0006)
- `--auto-merge` → execute the full merge chain after each plan (feat→staging→main)
- `--effort-budget Nm` → max effort per plan (e.g., `30m`)
- `--skip-completed` → skip plans with status `Completed` (default: true)

Discover plan files:
```bash
ls blueprint/live/[0-9]*-*.md | sort
```

Filter to the requested range. Build the execution queue.

## Step 2: Pre-flight Checks

```bash
echo "🤖 [batch-flow:2] pre-flight checks"
```

1. Verify all plan files in range exist
2. Check git status is clean (no uncommitted changes)
3. Verify on the staging branch (or the correct base branch):
   ```bash
   STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
   STAGING_BRANCH=${STAGING_BRANCH:-staging}
   ```
4. Check that plans are in correct sequence (no gaps)
5. Log the batch plan:

```
📋 Batch Pipeline — Plans NNNN to MMMM
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Plan NNNN: description — status
Plan NNNN+1: description — status
...
Total: X plans, Y already completed, Z to execute
```

## Step 3: Execute Loop

For each plan in the queue:

### 3A: Check if already completed
```bash
# Read plan status from YAML frontmatter
STATUS=$(grep '^status:' "$PLAN_FILE" | head -1 | awk '{print $2}')
if [ "$STATUS" = "Completed" ]; then
  echo "🤖 [batch-flow:3] skipping plan $PLAN_NUM — already completed"
  continue
fi
```

### 3B: Prepare for plan execution
```bash
echo "🤖 [batch-flow:3] starting plan $PLAN_NUM ($CURRENT/$TOTAL)"

# Ensure on staging branch
git checkout "$STAGING_BRANCH" && git pull origin "$STAGING_BRANCH"

# Create feature branch (flow-auto Step 2 handles this if plan is new)
# If plan already has a branch, check it out
```

### 3C: Execute flow-auto-wt for this plan

**ALWAYS use `/flow-auto-wt` (worktree isolation), NEVER `/flow-auto`.** Each plan runs in its own worktree to prevent `blueprint/` merge conflicts between sequential plans. Without worktree isolation, every `/finish` creates conflicts because all branches share the same `blueprint/` directory.

Dispatch flow-auto-wt as a subagent with the plan file as argument:

```
Agent({
  description: "flow-auto-wt plan NNNN",
  prompt: "Run /flow-auto-wt for plan file: $PLAN_FILE. This is plan $CURRENT of $TOTAL in a batch run. [paste full flow-auto-wt skill instructions]",
  mode: "auto"
})
```

Wait for completion. Capture result (PR URL, status, any issues).

### 3D: Auto-merge chain (if --auto-merge)

After flow-auto creates the PR:

```bash
# Merge feat → staging
PR_NUM=$(gh pr list --base "$STAGING_BRANCH" --head "$BRANCH" --json number -q '.[0].number')
if [ -n "$PR_NUM" ]; then
  gh pr merge "$PR_NUM" --merge -m "🔀 merge: $BRANCH into $STAGING_BRANCH"
fi

# Create and merge staging → main
git checkout "$STAGING_BRANCH" && git pull origin "$STAGING_BRANCH"
MAIN_PR=$(gh pr create --base main --head "$STAGING_BRANCH" \
  --title "🔀 merge: $STAGING_BRANCH into main (plan $PLAN_NUM)" \
  --body "Batch pipeline auto-merge. Plan: $PLAN_NUM" 2>&1)

if echo "$MAIN_PR" | grep -q "already exists"; then
  MAIN_PR_NUM=$(gh pr list --base main --head "$STAGING_BRANCH" --json number -q '.[0].number')
else
  MAIN_PR_NUM=$(echo "$MAIN_PR" | grep -oP '\d+$')
fi

gh pr merge "$MAIN_PR_NUM" --merge -m "🔀 merge: $STAGING_BRANCH into main"
```

**Error handling:**
- If PR already merged: skip silently
- If merge conflicts: STOP the batch, report which plan caused the conflict
- If staging→main has no diff: skip the PR creation

### 3E: Context management

```bash
echo "🤖 [batch-flow:3e] checking context usage"
```

After each plan completes:
- If context usage > 60%: compact context before next plan
- If context usage > 85%: force compact — the next plan needs headroom
- Log context state for debugging

### 3F: Update progress

```bash
echo "🤖 [batch-flow:3f] plan $PLAN_NUM complete ($CURRENT/$TOTAL)"
```

## Step 4: Final Report

```bash
echo "🤖 [batch-flow:4] batch pipeline complete"
```

Output summary:

```
🤖 Batch Flow Complete!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Plans executed: X of Y
Skipped (already complete): Z

Results:
  Plan NNNN: ✅ PR #XX — merged to main
  Plan NNNN+1: ✅ PR #XX — merged to main
  Plan NNNN+2: ⚠️ PR #XX — review issues remain (3 cycles exhausted)
  ...

Total PRs: X (Y merged, Z pending review)
Total commits: X

Run full test suite to verify before deploying.
```

## Crash Recovery

If the batch is interrupted (context limit, crash, user intervention):

1. On restart with `/batch-flow N-M`:
   - Scan all plans in range for status
   - Skip `Completed` plans
   - Resume from first non-completed plan
   - If a plan is `In Progress`, use `/resume` logic for that plan

2. Progress is always recoverable from:
   - Plan file statuses (Completed / In Progress / Awaiting Approval)
   - Git branch state
   - PR state on GitHub

## Flags

- `N-M`: Plan range (required). E.g., `2-6` for plans 0002-0006
- `--auto-merge`: Execute full merge chain after each plan (default: off)
- `--effort-budget Nm`: Max effort per plan. E.g., `30m` for 30 minutes
- `--skip-completed`: Skip completed plans (default: on)
- `--from N`: Start from plan N (skip earlier plans regardless of status)

## Rules

- **NEVER use AskUserQuestion** — fully autonomous
- **NEVER merge without PRs** — every merge goes through a PR
- **NEVER skip context checks** — compaction between plans prevents context exhaustion
- **NEVER continue after merge conflict** — stop and report
- Same delegation rules as flow-auto — coordinator orchestrates, never implements
- Each plan is independent — failure in one plan does not skip subsequent plans (unless merge conflict)

Use $ARGUMENTS as plan range and flags.
