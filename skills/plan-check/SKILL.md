---
name: plan-check
description: >
  Audit implementation against the plan — compare planned vs actual changes, detect orphaned test
  references, fix checkbox mismatches, and sync frontmatter counts. Use this skill whenever the
  user says "/plan-check", "check the plan", "audit the plan", "verify implementation", "compare
  plan vs code", or any request to validate that what was built matches what was planned.
  Also triggers on "orphaned tests", "plan audit", "check task marks", "did we implement everything",
  or "sync frontmatter counts".
  ALWAYS run after /plan-approved and before /pr — this is the quality gate.
---

# Plan Check: Audit Plan vs Implementation

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Analyze and update the plan file for the current branch, comparing planned vs actual implementation.

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Critical Rules

- You MUST read the plan file first. NEVER fabricate implementation status.
- Follow steps in order. DO NOT skip or reorder steps.
- **AUDIT ONLY** — Do NOT modify application or feature code. Only update the plan file, fix orphaned test references, and sync metadata.
- **Do NOT re-implement any incomplete tasks.** Flag them as missing in the report and leave as `[ ]`. The user decides whether to implement or descope.
- **False-positive prevention:** Before unmarking a completed `[x]` task as `[ ]`, verify via `git diff` that the planned file was truly NOT modified. If a file was modified but differently than planned, keep `[x]` and note the deviation — do not revert to `[ ]`.

## Step 1: Gather Plan + Diffs (parallel)

Run: `echo "🔷 BP: plan-check [1/3] reading plan context + diffs"`

Run these **simultaneously in a single step** (do not wait between them):

1. `~/.blueprint/bin/blueprint meta` — returns branch, base_branch, plan_file, project as JSON
2. `~/.blueprint/bin/blueprint context --diffs` — returns commits, changed files, diff stat, AND per-file diffs

If $ARGUMENTS is a plan path or number, locate that file directly instead of using plan_file from meta.

**READ the full plan file** (after meta returns the path) — DO NOT proceed without reading it.

## Step 2: Compare Plan vs Implementation

- Was every planned task implemented?
- Are `[x]`/`[ ]` marks accurate?
- Unplanned files modified? (document why)
- Planned files NOT modified? (document why)

## Step 2a: Detect Deleted Tasks — CRITICAL

**Agents sometimes delete `[ ]` tasks they couldn't solve instead of reporting failure.** Compare the plan at plan-review time vs now to catch any removed tasks.

### 1. Get the plan file at plan-review commit

```bash
PLAN_REVIEW_COMMIT=$(git log --oneline | grep "plan: review" | head -1 | awk '{print $1}')
# PLAN_FILE already available from Step 1's blueprint meta output
echo "Comparing plan at $PLAN_REVIEW_COMMIT vs current"
```

### 2. Extract tasks from both versions

```bash
# Tasks at plan-review time (the approved baseline)
git show "$PLAN_REVIEW_COMMIT:$PLAN_FILE" 2>/dev/null | grep -E "^- \[[ x]\]" > /tmp/plan-review-tasks.txt

# Tasks now (after execution)
grep -E "^- \[[ x]\]" "$PLAN_FILE" > /tmp/plan-current-tasks.txt
```

### 3. Diff the task lists

```bash
diff /tmp/plan-review-tasks.txt /tmp/plan-current-tasks.txt
```

Look for lines only in the review version (prefixed with `<`) — these are tasks that were **deleted during execution**.

### 4. Flag deleted tasks

If any tasks were removed:
- **List each deleted task** in the report (Step 6)
- **Re-add them** to the plan as `[ ]` with a note: `(restored by plan-check — removed during execution)`
- These must be implemented or explicitly marked as descoped by the user

This is a hard failure — deleted tasks indicate an agent tried to hide incomplete work.

## Step 2b: Grep Test Suite for Orphaned References — CRITICAL

**This step catches test files that existed at plan-review time, referenced removed behaviors, but were NOT modified during implementation.**

### 1. Find the plan-review baseline commit

```bash
PLAN_REVIEW_COMMIT=$(git log --oneline | grep "plan: review" | head -1 | awk '{print $1}')
echo "Plan-review baseline: $PLAN_REVIEW_COMMIT"
```

### 2. Extract removed patterns from the diffs

From the diffs gathered in Step 1, identify significant removed patterns (lines starting with `-`):
- Old route params, removed properties/fields, old method calls, old URL structures, old validation rule names
- **Minimum pattern length: 4+ characters.** Exclude generic words (`name`, `id`, `type`, `data`, `test`, `user`). Use specific identifiers only (e.g., `old_field_name`, `legacyEndpoint`, `validateOldRule`).

Use **specific, non-generic patterns** to avoid false positives.

### 3. Grep the plan-review commit for each pattern

```bash
git grep -l "REMOVED_PATTERN" $PLAN_REVIEW_COMMIT -- "tests/"
```

### 4. Cross-check against modified files

Flag any test file that contained the removed pattern at plan-review time AND was NOT modified during implementation. These are **orphaned test references**.

Fix all orphaned references before proceeding to the audit commit.

## Step 3: Update Plan File

- Fix `[x]`/`[ ]` mismatches
- Add timestamps: `date "+%d/%m/%Y %H:%M"`
- Add "Plan vs Implementation" comparison table
- Keep existing content — add, don't delete
- Do NOT add session entries or blockquote status bars — v2 plans use single `session` field set by /plan-approved

### Verify Acceptance Criteria

If the plan has an `## Acceptance Criteria` section, verify each item against the actual implementation:
- Test each criterion (run commands, check files, confirm behavior)
- Mark verified items `[x]` with timestamp: `- [x] Criterion ✅ DD/MM/YYYY HH:MM`
- Leave items `[ ]` if they fail — note what failed
- Report any failed AC items in Step 6

## Step 4: Audit Commit — MANDATORY

Run: `echo "🔷 BP: plan-check [2/3] audit commit — full task list with all marks visible"`

Count mismatches fixed, deleted tasks restored, and orphaned refs fixed during this audit to build a descriptive commit message:

```bash
# If orphaned test fixes or other code changes were made, run /ship first
# Then always commit the updated plan to the project repo:
git add blueprint/ && git commit -m "🧹 chore: plan check NNNN — fixed N mismatches, restored N deleted tasks, fixed N orphaned refs"
```

The plan file MUST have the full task list with all `[x]` marks and timestamps before pushing.

## Step 5: Report

Run: `echo "🔷 BP: plan-check [3/3] plan check complete"`

```
Plan Check Complete:
  - Planned items: X/Y implemented
  - Deleted tasks: N (restored — agents removed instead of implementing)
  - Additional items: N (beyond plan)
  - Missing items: N (not implemented)
  - Acceptance Criteria: X/Y verified
  - Files planned: X | Files modified: Y
  - Status: [All matched / Discrepancies found]
```

**STOP. You MUST use `AskUserQuestion` tool here.**

- **Question:** "Plan audited. What's next?"
- **Option 1:** "Run /pr" — Push and create the pull request
- **Option 2:** "I have more changes" — Continue working, run /plan-check again when done

Use $ARGUMENTS as plan file path if provided, otherwise auto-detect from branch.
