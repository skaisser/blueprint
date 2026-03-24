---
name: flow-auto
description: >
  Fully autonomous SDLC pipeline — zero user intervention. Opus 4.6 runs the entire workflow
  from plan through PR, handles review feedback loops, and stops only when the PR is ready
  for human merge. Use this skill whenever the user says "/flow-auto", "auto flow",
  "autonomous flow", "just do everything", "full auto", "hands off", "run it all",
  or any request to run the complete pipeline without checkpoints or user decisions.
  Also triggers on "no intervention", "auto pipeline", "unattended flow", "fire and forget flow",
  "do everything and leave me a PR", or "I'm going to sleep just build it".
  This is the zero-touch version of /flow — same pipeline, no pauses. IMPORTANT: never use AskUserQuestion in this skill — all decisions are made autonomously.
---

# Flow Auto: Fully Autonomous SDLC Pipeline

Run the entire SDLC pipeline with zero user intervention. Opus 4.6 makes all decisions autonomously — planning, reviewing, executing, creating the PR, addressing review feedback, and looping until the PR is clean. Stops only when done, leaving a final recommendation comment on the PR.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

```
/flow-auto <description>
  → plan → review → execute → check → PR → review → fix loop → DONE
  Zero AskUserQuestion calls. Zero pauses. One command → PR ready for merge.
```

## Why This Exists

The normal `/flow` has 2 mandatory checkpoints where it asks the user what to do. With 1M context and Opus 4.6's reasoning, those decisions are predictable 95% of the time:
- "Continue to plan-approved?" → Yes, if context is under 50%
- "PR looks good?" → Run review, fix issues, repeat

`/flow-auto` makes those decisions itself, saving 5-10 minutes of human wait time per feature.

## Critical Rules

1. **NEVER use AskUserQuestion.** Every decision point is handled autonomously. If you catch yourself about to ask the user something, STOP — make the decision yourself based on context and plan state.
2. **NEVER merge the PR.** The pipeline stops after creating/fixing the PR. The user decides when to merge.
3. **Delegate all implementation.** Same as `/plan-approved` — the coordinator never writes code. Dispatch subagents.
4. **Commit after every stage.** Every stage produces a commit — the full git history is recoverable.
5. **Loop review→fix max 3 times.** If review issues persist after 3 cycles, stop and leave a comment explaining what remains.
6. **Skip GitHub Issues integration** unless $ARGUMENTS contains a GitHub issue number (e.g., #42).
7. **NEVER skip steps.** ALL 8 steps must execute in order. The audit hook BLOCKS PR creation if Step 5 (plan-check) was not run, and WARNS if Step 7 (review loop) was not attempted. Each step MUST echo its checkpoint: `echo "🤖 [flow-auto:N] description"` — this is how the audit hook tracks progress. Skipping a step is a pipeline violation.
8. **The checkpoint echo is NOT optional.** Every step MUST start with its echo command. Without it, the audit hook cannot verify the step ran and will block subsequent steps.
9. **Effort budget per phase.** If a single phase takes more than 2 fix rounds (dispatch→fail→dispatch→fail), log a warning, mark the phase as "partially complete" with a note, and move on. Perfectionism on one phase should not block the entire pipeline. Data import tasks have a "good enough" threshold — if counts are within 80% of expected, proceed.
10. **PR must be merge-ready on first attempt.** Step 5 (Quality Sweep + Green Gate) is the primary quality bar. The pipeline MUST NOT create a PR until Step 5's Green Gate passes (all targeted tests green) and the Quality Sweep is clean. If tests fail after 2 fix rounds in Step 5c, do NOT proceed to Step 6 — dispatch one final comprehensive fix agent targeting all failures, then re-run tests. Only proceed to PR creation when tests pass or after 3 total fix rounds (hard ceiling). The review loop (Step 7) exists as validation, not as the primary quality mechanism.

---

## Step 1: Initialize

```bash
echo "🤖 [flow-auto:1] initializing autonomous pipeline"
~/.blueprint/bin/blueprint meta 2>/dev/null
```

**If active plan exists** (blueprint meta returns plan_file):
- Read the plan file. Check status.
- `awaiting-approval` or `approved` → Skip to Step 4 (execute)
- `in-progress` → Skip to Step 4 (resume from last `[ ]`)
- `completed` → Skip to Step 6 (PR)
- No plan but branch exists → Skip to Step 6 (PR)

**If no active plan:** Continue to Step 2.

Parse $ARGUMENTS for:
- GitHub issue number (e.g., `#42`) → store for PR body
- `--from <stage>` → jump to that stage
- Everything else → task description

## Step 2: Plan (replaces /plan)

```bash
echo "🤖 [flow-auto:2] creating plan"
```

1. Read `~/.claude/skills/plan/references/plan-template.md` for the plan format
2. Use `mcp__sequential-thinking__sequentialthinking` to analyze the task:
   - Break into phases with clear, actionable tasks
   - Identify file dependencies and parallel opportunities
   - Add acceptance criteria
3. Create the feature branch:
   ```bash
   BRANCH="feat/$(echo "$DESCRIPTION" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | head -c 50)"
   git checkout -b "$BRANCH"
   ```
4. Write the plan file to `blueprint/live/NNNN-feat-description.md` using the template format
   - Include `${CLAUDE_SESSION_ID}` in YAML `session:` and Work Sessions block (see plan-template.md)
5. Commit:
   ```bash
   git add blueprint/ && git commit -m "📋 plan: create NNNN-description"
   ```

**Autonomous decision:** Always proceed to Step 3. No pause needed.

## Step 3: Review (replaces /plan-review)

```bash
echo "🤖 [flow-auto:3] reviewing plan"
```

1. Read `~/.claude/skills/plan-review/references/team-execution.md` for execution strategy
2. Use `mcp__sequential-thinking__sequentialthinking` to validate:
   - Completeness, dependencies, file conflicts, missing tests
   - Fix issues directly — don't flag them
3. Mark ALL tasks with `[H]`/`[S]`/`[O]` complexity
4. Determine execution strategy (Mode A-G) based on parallelism potential
5. Add `## Execution Strategy` section to plan
6. Update status to `Approved`, set progress counters
7. Commit:
   ```bash
   git add blueprint/ && git commit -m "📋 plan: review NNNN-description"
   ```

**Autonomous decision:** Check context usage.
- Context < 50% → Continue to Step 4 directly
- Context ≥ 50% → Log a warning but continue anyway (the coordinator only orchestrates, it needs minimal context)

## Step 4: Execute (replaces /plan-approved)

```bash
echo "🤖 [flow-auto:4] executing plan"
```

1. Read the full plan file. Identify completed/pending phases.
2. Read `references/team-execution.md` for delegation strategy
3. Update status to `In Progress`
4. Commit baseline:
   ```bash
   git add blueprint/ && git commit -m "📋 plan: start execution NNNN"
   ```

5. **Execute rounds** following the Execution Strategy:
   - For each round: pre-validate → dispatch workers → wait → verify → update plan
   - Workers get the full Worker Completion Protocol (same as /plan-approved)
   - Workers use `/ship` to commit their code
   - After each round: sync frontmatter, commit plan progress

6. **On worker failures:** Apply Immediate Dispatch Rule — categorize in ONE pass, dispatch N fix agents in parallel. Max 2 fix rounds per phase before moving on and noting the issue.

6b. **Post-phase data validation** — After any phase that imports, seeds, or creates data, run a validation check:
   ```bash
   # Verify record counts match plan expectations using the project's test command or REPL
   ```
   - If counts are off by >20% from plan expectations, log a warning but continue
   - If counts are 0 (complete failure), dispatch a fix agent before proceeding
   - This prevents cascading errors from undetected data import failures

7. When all phases complete:
   ```bash
   git add blueprint/ && git commit -m "✨ feat: complete NNNN-description"
   ```

**Autonomous decision:** Always proceed to Step 5. No pause.

## Step 5: Plan Check + Quality Sweep + Green Gate (replaces /plan-check)

```bash
echo "🤖 [flow-auto:5] auditing implementation, quality sweep, and verifying tests"
```

### 5a: Audit plan vs implementation

1. Run `~/.blueprint/bin/blueprint context --diffs` to get all changes
2. Compare plan vs implementation:
   - Check all `[x]`/`[ ]` marks are accurate
   - Detect deleted tasks (compare plan-review baseline vs current)
   - Grep for orphaned test references
   - Verify acceptance criteria
3. Fix any mismatches directly (re-add deleted tasks, fix marks)
4. If discrepancies found, dispatch fix agents immediately

### 5b: Coordinator self-review (Quality Sweep)

Before any external review, the coordinator reviews its own diff to catch issues proactively — this is what makes the PR merge-ready on first attempt:

1. Read the full diff: `git diff $(git merge-base HEAD "${BASE_BRANCH:-main}") HEAD`
2. Scan for common review issues:
   - **Incomplete implementations:** TODO/FIXME/HACK comments, placeholder values, empty method bodies
   - **Import/namespace issues:** unused imports, missing use statements
   - **Code quality:** duplicated blocks (>5 lines), overly complex methods, missing return types on public methods
   - **Security:** hardcoded credentials, raw SQL without bindings
   - **Convention violations:** follow project conventions as defined in project config
3. For each category with issues found, dispatch a fix agent with the specific file:line references
4. After fix agents complete, re-read the diff and verify fixes landed
5. Max 2 self-review rounds — catch the obvious, don't chase perfection

This sweep catches 80%+ of what an external reviewer would flag, eliminating most review loop iterations.

### 5c: Integration test sweep (Green Gate)

Run ALL test files created or modified during execution as a single batch to catch cross-phase regressions:

```bash
# Collect all test files touched by this branch
TEST_FILES=$(git diff --name-only "$(git merge-base HEAD "${BASE_BRANCH:-main}")" HEAD | grep -iE '(Test|test|spec)\.' | tr '\n' ' ')
if [ -n "$TEST_FILES" ]; then
  # Run using the project's test command
  $TEST_FILES
fi
```

**If tests fail:** Dispatch fix agents targeting the failures. Re-run. Max 2 fix rounds.
**If tests still fail after 2 rounds:** Dispatch one final comprehensive fix agent with ALL failure output and full file context. Re-run tests (round 3 — hard ceiling).
**If tests still fail after 3 rounds:** Mark specific failures in the plan as blockers. The PR will be created but marked BLOCKED in the final report.
**If all tests pass:** The PR is certified merge-ready. This is the target outcome.

### 5d: Finalize

1. Set `MERGE_READY` flag based on:
   - Green Gate passed (all tests green) AND
   - Quality Sweep clean (no issues remaining after fix rounds)
   - Both must be true for MERGE_READY = true
3. Commit:
   ```bash
   git add blueprint/ && git commit -m "🧹 chore: plan check NNNN"
   ```

**Autonomous decision:** Proceed to Step 6 only after Green Gate + Quality Sweep complete. If MERGE_READY is false, the PR will be created but explicitly marked as BLOCKED — this is a last resort, not the normal path. The pipeline targets MERGE_READY = true before PR creation.

## Step 6: Create PR (replaces /pr)

```bash
echo "🤖 [flow-auto:6] creating pull request"
```

1. Determine base branch:
   ```bash
   STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
   STAGING_BRANCH=${STAGING_BRANCH:-staging}
   git rev-parse --verify "$STAGING_BRANCH" 2>/dev/null && echo "$STAGING_BRANCH" || echo "main"
   ```
2. Push branch:
   ```bash
   git push -u origin "$BRANCH"
   ```
3. Gather context: `~/.blueprint/bin/blueprint context`
4. Create PR with `gh pr create`:
   - Title: emoji + type + short description (under 70 chars)
   - Body: summary bullets, test plan, GitHub issue link if provided
   - Base branch: staging branch (if exists) or main

### Auto-Merge Chain (feat → staging → main)

When the PR is created to the staging branch, execute the full merge chain automatically:

```bash
# Step 1: Merge feat → staging
PR_NUM=$(gh pr view --json number -q '.number')
gh pr merge "$PR_NUM" --merge -m "🔀 merge: $BRANCH into $STAGING_BRANCH"

# Step 2: Create staging → main PR
git checkout "$STAGING_BRANCH" && git pull origin "$STAGING_BRANCH"
MAIN_PR_URL=$(gh pr create --base main --head "$STAGING_BRANCH" \
  --title "🔀 merge: $STAGING_BRANCH into main" \
  --body "Auto-merge from flow-auto pipeline. Source branch: $BRANCH")

# Step 3: Merge staging → main
MAIN_PR_NUM=$(gh pr view --json number -q '.number')
gh pr merge "$MAIN_PR_NUM" --merge -m "🔀 merge: $STAGING_BRANCH into main"

# Step 4: Return to feature branch for review loop
git checkout "$BRANCH"
```

**Error handling:**
- If PR already exists (exit code 1, "already exists"): find existing PR with `gh pr list --base $STAGING_BRANCH --head $BRANCH` and merge it
- If merge conflicts: stop and report — do NOT force merge
- If staging→main PR already merged: skip silently

**Autonomous decision:** Always proceed to Step 7.

## Step 7: Review Loop (replaces /review + /address-pr)

```bash
echo "🤖 [flow-auto:7] starting review loop"
```

This is the autonomous review→fix cycle. Max 3 iterations.

### Pre-check: GitHub Action exists?

Before starting the review loop, check if the @claude GitHub Action is configured:

```bash
CLAUDE_ACTION=$(gh api repos/{owner}/{repo}/actions/workflows --jq '.workflows[] | select(.name | test("claude|Claude|CLAUDE")) | .id' 2>/dev/null)
if [ -z "$CLAUDE_ACTION" ]; then
  echo "🤖 [flow-auto:7] skipping review loop — no @claude GitHub Action detected"
  # Jump directly to Step 8
fi
```

If no Claude workflow is found, skip the entire review loop and proceed to Step 8. This saves up to 20 minutes of wasted polling.

### For each iteration:

1. **Trigger review:**
   ```bash
   PR_NUM=$(gh pr view --json number -q '.number')
   gh pr comment "$PR_NUM" --body "@claude review this PR and check if we are able to merge. Analyze the code changes for any issues, security concerns, or improvements needed."
   ```

2. **Wait for review** (poll every 5 minutes, max 20 minutes — reviews can take 10-15 min):
   ```bash
   # Check for new review comments — poll every 5 minutes (300s), max 4 checks
   for i in $(seq 1 4); do
     sleep 300
     gh api repos/{owner}/{repo}/pulls/$PR_NUM/reviews --jq '.[].state' | grep -q "CHANGES_REQUESTED" && break
     # Also check if bot commented (approval or info)
     BOT_COMMENTS=$(gh api repos/{owner}/{repo}/issues/$PR_NUM/comments --jq '[.[] | select(.user.login == "claude[bot]")] | length')
     [ "$BOT_COMMENTS" -gt 0 ] && break
     echo "Waiting for review... ($i/4, next check in 5 min)"
   done
   ```

3. **If no review comments after 20 min:** The @claude GitHub Action may not be set up or is slow. Skip the review loop and go to Step 8.

4. **If review comments exist:**
   - Fetch all comments: `~/.blueprint/bin/blueprint pr-review $PR_NUM`
   - Categorize feedback by type (bugs, style, missing tests, etc.)
   - Dispatch fix agents (one per category) — same pattern as /address-pr
   - Push fixes: `git push`

5. **Check if issues remain:**
   - If all feedback addressed → exit loop
   - If iteration < 3 → loop back to trigger another review
   - If iteration = 3 → exit loop, note remaining issues

### Review loop exit:
```bash
echo "🤖 [flow-auto:7] review loop complete after N iterations"
```

## Step 8: Final Report (THE ONLY OUTPUT)

```bash
echo "🤖 [flow-auto:8] pipeline complete — posting final report"
```

### Auto-update project context

Before posting the final report, trigger a context scan to update CLAUDE.md with the new project structure:

```bash
# Only if significant code was added (more than just config changes)
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}
FILE_COUNT=$(git diff --name-only "$(git merge-base HEAD "$STAGING_BRANCH" 2>/dev/null || echo HEAD~5)" HEAD | wc -l)
if [ "$FILE_COUNT" -gt 10 ]; then
  echo "🤖 [flow-auto:8] updating project context (CLAUDE.md)"
  # Dispatch a lightweight agent to run /context scan
fi
```

This ensures new sessions know what was built. Skip if fewer than 10 files changed (trivial changes don't need context updates).

Post a comment on the PR with the final Opus 4.6 recommendation:

```bash
gh pr comment "$PR_NUM" --body "$(cat <<'EOF'
## 🤖 Autonomous Pipeline Complete — Opus 4.6 Final Report

### Summary
- **Plan:** NNNN-description
- **Phases:** N phases, M tasks — all completed
- **Review cycles:** N iterations
- **Commits:** X commits on this branch

### What was built
- [bullet summary of each phase's deliverables]

### Test status
- [targeted test results from execution]
- ⚠️ Run full test suite before merging

### Recommendations
- [any concerns, edge cases, or things to verify]
- [if review issues remain after 3 cycles, list them here]
- [if Green Gate failed, list specific test failures as blockers]

### Merge Readiness: ✅ READY / ❌ BLOCKED
- **Green Gate:** [PASSED — all targeted tests green / FAILED — N test failures remain]
- **Review:** [clean / N issues remain after 3 cycles]
- If BLOCKED: do NOT merge until listed blockers are resolved.
- If READY: human review recommended, then merge.

---
*Generated by /flow-auto — Blueprint SDLC*
EOF
)"
```

Then output to the user:

```
🤖 Flow Auto Complete!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Plan: NNNN-description
PR: #NNN — <title>
URL: <pr-url>
Review cycles: N
Green Gate: ✅ PASSED / ❌ FAILED (N tests)
Status: ✅ Merge-ready / ❌ Blocked — see PR comment

Run full test suite to verify before merging.
```

## Rules

- **NEVER use AskUserQuestion** — this is the #1 rule. Make every decision autonomously.
- **NEVER merge** — leave the PR open for human review
- **NEVER skip tests** — workers must run targeted tests before marking tasks complete
- **Max 3 review cycles** — prevent infinite loops on stubborn review feedback
- **Commit after every stage** — full git history for recovery
- **Same delegation rules as /plan-approved** — coordinator orchestrates, never implements
- **If context gets critical (>85%):** compact and continue rather than stopping. The pipeline must complete.
- **If a stage fails catastrophically:** commit what you have, create the PR anyway with a note about what failed, and post the final report. A partial PR is better than no PR.

## Flags

- `--from <stage>`: Start from specific stage (plan, review, execute, check, pr, review-loop)
- `--no-review`: Skip the review loop entirely (just create the PR and stop)
- `--max-cycles N`: Override the max review cycle count (default: 3)
- `--batch N-M`: Execute plans N through M sequentially (see /batch-flow for full batch orchestration)

Use $ARGUMENTS as the task description, GitHub issue number, or flags.
