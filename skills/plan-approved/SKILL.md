---
name: plan-approved
description: >
  Execute a reviewed plan by delegating work to subagents and coordinated teams.
  Reads the Execution Strategy from the plan, dispatches workers per round, tracks task completion
  with [x] marks and timestamps, commits after each phase, and manages context for long executions.
  Use this skill whenever the user says "/plan-approved", "execute the plan", "start implementation",
  "run the plan", or any request to begin coding based on an existing reviewed plan.
  Also triggers on "let's build it", "go ahead and implement", or "start coding".
  Does NOT trigger on: "resume", "continue the plan", "pick up where we left off", "where did we stop" — use /resume for those.
  ALWAYS run after /plan-review. With 1M context, if context usage is under 30% after plan-review,
  continue directly — the coordinator only orchestrates (delegates to subagents), so it needs minimal context.
  Only clear when context is already above 50% after plan-review.
---

# Plan Approved: Execute Implementation

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Execute a reviewed plan by delegating all work to subagents or coordinated teams. The leader orchestrates — it never writes code directly.

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /finish
                     ↑ context clear only if context usage > 50% after plan-review
```

## Critical Rules

1. **Read + validate in one pass.** Step 1 reads the plan via `blueprint meta`, validates review markers, detects TDD mode, and extracts the Execution Strategy — all from the same read. If `[H]/[S]/[O]` markers or `## Execution Strategy` are missing → STOP.
2. **Leader NEVER implements code.** Always delegate to subagents or teams. If you're reading source files, writing code, or running tests — STOP. You are violating this rule. Dispatch a worker instead.
3. **Skip team-execution.md** — the plan already has everything from /plan-review. Only read `plan-review/references/team-execution.md` if Execution Strategy is incomplete.
4. **Follow the Execution Strategy exactly** — mode, workers, model tiers per round.
5. **Parallel dispatch = ONE message.** For Parallel Subagents, Parallel Teams, or Mixed: dispatch ALL workers in a single message.
6. **One active team at a time.** Parallel Teams = N team-lead subagents (each autonomous), not 1 coordinator managing N teams.
7. **Commit plan marks + code together.** Never commit code without updating task marks. Never leave tasks unmarked.
8. **Every phase ends with /commit.** Workers commit their own phase code, coordinator commits plan updates to memory.

## Dispatch Enforcement — THE MOST IMPORTANT SECTION

**The #1 failure mode is the coordinator doing work itself instead of spawning agents.** This wastes context, runs sequentially, and defeats the entire execution strategy.

### Self-Check (ask yourself before EVERY action)

> "Am I about to read a source file, write code, or run a test?"
> → YES = STOP. Dispatch a worker via the `Agent` tool instead.
> → The ONLY exception is Mode G (Leader Direct, ≤3 trivial `[H]` tasks).

### Minimum Dispatch Rules

| Plan size | Minimum dispatch |
|-----------|-----------------|
| ≤3 tasks, all `[H]` | Leader Direct OK (Mode G) |
| 4-5 tasks | At least 1 subagent |
| 6+ tasks | At least 2 concurrent subagents per round |
| Any `[S]`/`[O]` | MUST be dispatched to agent — never Leader Direct |

### What the Agent Tool Call Looks Like

For each worker, call the `Agent` tool with `subagent_type: "general-purpose"`. The prompt MUST include:

1. **Phase details** — copy the exact tasks from the plan
2. **File paths** — every file the worker will touch
3. **Context** — relevant interfaces, contracts, patterns from Context Pack
4. **Test commands** — the project's test command with appropriate filters
5. **Worker Completion Protocol** — the full protocol from Step 4
6. **Working directory** — `cd /path/to/project` as the first instruction

Example dispatch (2 parallel subagents):
```
Agent call 1: "Implement Phase 1 — Service layer refactoring"
  prompt: "You are worker-1. cd /Users/.../project. Your phase: [paste phase details]. Files: [...]. When done: [paste Worker Completion Protocol]."

Agent call 2: "Implement Phase 2 — Frontend components"
  prompt: "You are worker-2. cd /Users/.../project. Your phase: [paste phase details]. Files: [...]. When done: [paste Worker Completion Protocol]."
```

**Both calls go in ONE message — they run simultaneously.** This is 2x faster than sequential.

**Model selection per worker:**
- If ALL tasks in the worker's assignment are `[H]` → dispatch with `model: "sonnet"` (faster output for simple tasks)
- If ANY task is `[S]` or `[O]` → use default model (Opus — maximum quality for anything with real logic)

Example with model selection:
```
Agent call 1: "Phase 1 — Config + migrations" (all [H])
  model: "sonnet"
  prompt: "You are worker-1..."

Agent call 2: "Phase 2 — Business logic service" (has [S] tasks)
  prompt: "You are worker-2..."  ← Opus (default) for maximum quality
```

### Failure Signs (stop and correct if you notice these)

- You've been running for 5+ minutes without dispatching any agent
- You're reading source files beyond the plan file and plan-review/references/team-execution.md
- You're writing or editing code files
- You're running tests (other than pre-dispatch baseline)
- You're investigating errors beyond a single pass to categorize them

---

## Step 1: Read Plan + Validate

Run these **in parallel** (one bash call each):

```bash
# Check 1: flow-auto lock
if [ -f "blueprint/.flow-auto-lock" ]; then
  echo "flow-auto is currently active (lock file found). Yielding."
  exit 0
fi

# Check 2: get plan metadata
echo "🔷 BP: plan-approved [1/4] reading plan file"
~/.blueprint/bin/blueprint meta
```

If the lock file exists, STOP and inform the user. Do not proceed — conflicting execution will cause merge conflicts.

Returns branch, base_branch, plan_file, next_num, project, team, git_remote as JSON.

**READ the full plan file**, then analyze it in ONE pass (no extra tool calls):

### 1A: Phase Status
- Count `[x]` vs `[ ]` tasks to determine status
- All `[x]` → **SKIP** | Mix of `[x]` and `[ ]` → **RESUME from first `[ ]`** | All `[ ]` → **EXECUTE**

### 1B: Validate Plan Was Reviewed
Check for BOTH — if EITHER is missing, STOP with `AskUserQuestion`:

| Check | What to look for |
|-------|-----------------|
| Complexity markers | `[H]`, `[S]`, or `[O]` on tasks in Phases section |
| Execution strategy | `## Execution Strategy` section in plan body |

Valid input statuses: `todo`, `awaiting-review`, or `approved` — all are fine. Status is set to `in-progress` at Step 2.

### 1C: Detect TDD Mode
Check for `mode: tdd` in frontmatter or `> **Mode:** TDD` in the plan. If active:

| Phase type | Constraint |
|------------|-----------|
| Phase 0 (RED) | Workers write ONLY tests — zero implementation code. ALL tests must fail before marking complete. |
| GREEN phases | Workers implement ONLY what failing tests require. Run targeted tests after each task. |
| REFACTOR phase | Run the project's test coverage command — must reach 100% coverage. |

TDD enforcement rules are appended to the Worker Completion Protocol (Step 4).

### 1D: Read Execution Strategy
Extract rounds, modes, model tiers, and worker assignments from the `## Execution Strategy` section. **Skip reading `plan-review/references/team-execution.md`** — the plan already has everything.

**Quick model reference (inline — no file read needed):**
- All `[H]` tasks → `model: "sonnet"` (fast output)
- Any `[S]`/`[O]` → default model (Opus, maximum quality)

## Step 2: Update Status

1. Update frontmatter `status: in-progress`
2. Set frontmatter `session: "${CLAUDE_SESSION_ID}"` (single field, not array) — this enables resume via `claude -r <session-id>`

## Step 3: Commit Baseline

```bash
echo "🔷 BP: plan-approved [2/4] committing baseline"

# Project code baseline (if staged changes):
blueprint commit "📋 plan: start execution NNNN"
# or on resume:
blueprint commit "📋 plan: resume execution NNNN"

# Plan file status update → project repo:
git add blueprint/ && git commit -m "📋 plan: start execution NNNN"
```

Creates the baseline snapshot: project code via `blueprint commit`, plan state via git. This is critical for the commit chain guarantee — every state is recoverable from git history.

## Step 4: Execute Rounds

```bash
echo "🔷 BP: plan-approved [3/4] starting execution — following strategy from plan"
```

**Context Protection:** Execute ONE round at a time. After each round: update plan → test → commit → assess context. With 1M context, most plans complete in a single session — suggest compacting only after 5+ rounds or when context visibly degrades.

Read the Execution Strategy section from the plan. Execute EXACTLY the mode it specifies.

### Mode Execution Reference

**Mode A — Parallel Subagents** (2+ independent all-`[H]` phases)
- Call `Agent` tool N times in **ONE message** — they run simultaneously
- Each prompt must be fully self-contained (phase details, files, test commands)

**Mode B — Parallel Teams** (2+ independent phases with `[S]`/`[O]`)
- Call `Agent` tool N times in **ONE message** — each spawned subagent IS a team-lead
- Each prompt instructs the subagent to: create a team (team-alpha, team-beta, etc.), assign workers (worker-1, worker-2, etc.), implement the phase
- Coordinator dispatches all and waits — does NOT manage individual teams mid-task

**Mode C — Single Team** (1 sequential `[S]`/`[O]` phase, coordinator actively manages)
- Use `TeamCreate` to create one team. Assign workers with sequential IDs (worker-1, worker-2, etc.)
- Coordinator assigns tasks, monitors, messages workers mid-task if blocked
- Only ONE active team at a time

**Mode D — Mixed Dispatch** (1 team + concurrent `[H]` subagents)
- In ONE message: create the team AND call `Agent` for the `[H]` subagent(s)
- Team runs under coordinator attention; subagents fire-and-forget simultaneously

**Mode E — Coordinated Team** (tasks within a team need sequential handoff)
- Same as Single Team, but worker-1 writes a handoff block in the plan before worker-2 starts
- Coordinator re-prompts worker-2 with the handoff pasted in

**Mode F — Single Subagent** (1 phase, all `[H]`, sequential)
- One `Agent` call. Self-contained prompt. Coordinator waits for completion.

**Mode G — Leader Direct** (≤ 3 `[H]` tasks total)
- Coordinator handles tasks directly — no spawn.

### Worker Completion Protocol

Every worker prompt (subagent or team) MUST include these instructions:

```
WHEN YOU FINISH YOUR PHASE:

1. Mark ALL your tasks [x] with timestamp in the plan file:
   - [x] [H] Task description ✅ DD/MM/YYYY HH:MM
   - Preserve the complexity marker ([H]/[S]/[O]) added by /plan-review
   - Get timestamp via: date "+%d/%m/%Y %H:%M" — NEVER guess the date

2. Run targeted tests and confirm they pass:
   - Run the project's test command with appropriate filter for your changes
   - On failure: categorize by root cause, report to coordinator — do NOT attempt multi-category fixes alone

3. Commit and push via /ship (includes formatting):
   - ALWAYS use /ship — NEVER use raw git commit
   - /ship runs: format → git add → git commit → git push

4. Commit plan marks:
   - git add blueprint/ && git commit -m "✨ feat: phase X complete NNNN"

HARD RULES:
- NEVER skip /ship — it handles formatting, commit, and push in one step
- NEVER use raw git commit — always /ship for code, git commit only for blueprint/ files
- NEVER leave tasks unmarked
- NEVER add AI signatures to commits
- NEVER stage .env, *-api-key, or credential files
```

### Worker Timeout and Partial Completion

- **Worker timeout:** If a worker has not returned after 10 minutes, check its status via `TaskOutput`. If stuck or hung, stop it via `TaskStop` and dispatch a replacement worker for the remaining tasks.
- **Partial completion:** If a worker returns having completed only some of its assigned tasks (e.g., 3 of 5 done), do NOT re-dispatch the entire phase. Instead:
  1. Verify which tasks are marked `[x]` in the plan file
  2. Dispatch a new worker for ONLY the remaining `[ ]` tasks
  3. Include in the new worker's prompt: "Tasks 1-3 are already complete. You are implementing tasks 4-5 only."

### Worker Prompt Template

Every worker prompt MUST include these context sections. The coordinator gathers this info BEFORE dispatching:

```
WORKER CONTEXT BLOCK (include in every worker prompt):

## Project
- Working directory: /path/to/project
- Phase: [paste exact phase details from plan]
- Files to touch: [list every file path]

## Tech Stack Versions
[Auto-detected from project config — see plan-review]
- Framework: X.x
- Key packages: [any relevant package versions]

## Code Patterns (read BEFORE implementing)
- Read one existing file of the same type for patterns
  Example: building a new Resource? Read an existing Resource first
- Follow the existing naming conventions, imports, and structure exactly

## Test Commands
- Run the project's test command with appropriate filter for your changes

## Commit Rules
- Use /ship for ALL commits (never raw git commit)
- Format: <emoji> <type>: <description>
- No AI signatures

[paste Worker Completion Protocol here]
```

The coordinator MUST:
1. Read project dependency files to extract framework versions
2. Read one existing file of the same type the worker will create (e.g., an existing Resource for a new Resource task)
3. Read relevant Model files to include fillable/casts/relationships in the prompt
4. Include the Tech Stack Versions section from the plan (added by /plan-review)

### Pre-Dispatch Validation (reduces errors)

Before dispatching each round, the coordinator MUST verify:

1. **File conflict check** — No two workers in the same round touch the same file. If they do, reassign to one worker or make sequential.
2. **Dependency check** — All phases in this round have their prerequisites completed (previous round's tasks all `[x]`).
3. **Context pack** — Each worker prompt includes: the specific phase details, file paths to touch, relevant interfaces/contracts from the Context Pack, and targeted test commands.
4. **Test baseline** — Run the targeted tests for the upcoming phase BEFORE dispatching. If tests already fail, fix first — workers should start from green.

Skip pre-dispatch validation only for Mode G (Leader Direct) where overhead exceeds value.

### Immediate Dispatch Rule — CRITICAL

**The leader is an orchestrator, not a debugger.** When a verification step reveals failures:

1. Read the diagnostic output (test results, error log) **ONCE** — a single pass
2. Categorize failures by distinct root cause (e.g., "3 auth failures, 2 missing factory fields, 1 route typo")
3. Dispatch **N agents in parallel** — one per failure category — in ONE message
4. **Do NOT spend more than 3 tool calls investigating before dispatching**

If a pre-existing diagnostic file exists (failing-tests.md, captured test output), read it ONCE, count distinct failure categories, and dispatch immediately. Investigation is a worker task, not a coordinator task.

### Context Budget Discipline

- **After 5+ rounds of execution**, or if context quality visibly degrades, suggest the user compact and resume. With 1M context, most plans complete without needing a compact.
- When context is heavy and failures remain, **prefer dispatching agents over inline investigation** — agents protect the leader's remaining context.
- The leader should NEVER spend more than 10% of its context on investigation. If you've read more than 3 files to diagnose an issue, you've already spent too much — dispatch a worker instead.

### Round Loop

For each incomplete round:

1. **Pre-validate** — Run the pre-dispatch checks above.
2. **Dispatch** — Execute the mode from the Execution Strategy. For parallel modes, send ALL workers in ONE message. Every worker prompt MUST include the Worker Completion Protocol.
3. **Wait** — Let all workers/teams complete. Confirm each worker has committed and marked tasks `[x]`.
4. **Verify** — `git diff`, run targeted tests for the completed phase.
4b. **Data validation** — If the completed phase involved data import/seeding, verify counts:
   - Run queries to check record counts against plan expectations
   - If counts are off by >20%: log warning, dispatch investigation agent
   - If counts are 0 (complete failure): dispatch fix agent before proceeding
   - This prevents cascading errors from silent data import failures
5. **On failures: apply Immediate Dispatch Rule** — categorize in ONE pass, dispatch N agents in parallel. Do NOT debug inline.
6. **Update plan** — If any worker missed marking tasks or committing, do it now.
7. **Update plan** — mark completed tasks with `[x]` and timestamps
8. **Commit** —
   ```bash
   /commit                                              # project code changes (if any coordinator fixes)
   git add blueprint/ && git commit -m "📋 plan: round X complete NNNN"       # plan progress
   ```
10. **Assess context** — After 5+ rounds or if context quality degrades: use `AskUserQuestion` to suggest compact + resume. With 1M context, most plans complete without needing this.

### Post-Round Error Recovery

If a worker reports test failures or incomplete work:

1. **Categorize in ONE pass** — Read the output once. Group by root cause.
2. **Dispatch N agents** — One per failure category, in ONE message. Include specific error context per agent.
3. **Verify fix** — Run targeted tests again before marking complete.
4. **Update plan** — Mark the fixed tasks, add a note about the recovery.

**NEVER investigate failures inline.** The coordinator reads results, categorizes, dispatches. Workers investigate and fix.

## Step 5: On Completion

```bash
echo "🔷 BP: plan-approved [4/4] all phases complete"
```

- Update frontmatter `status: completed` (v1 compat: also accepts `Completed`)
- Shutdown any active team
- Run final targeted tests
- `/commit` for final plan state (all `[x]` marks and timestamps visible)
- `git add blueprint/ && git commit -m "✨ feat: complete NNNN-<description>"`

Use `AskUserQuestion`:
- **Question:** "Implementation complete. What's next?"
- **Option 1:** "Run /plan-check" — Audit implementation vs plan
- **Option 2:** "Done for now" — I'll run /plan-check manually

## Rules

- Fix test failures before marking a phase complete
- Only commit via `/commit` (or `/ship` for workers)
- No AI signatures — hook rejects them
- For full test suite: ask the user to run in a separate terminal
- ALWAYS delegate work — leader orchestrates, never implements
- Keep all task marks and timestamps visible — never collapse during execution
- When in doubt about scope: check the plan, not memory. The plan is the single source of truth.
