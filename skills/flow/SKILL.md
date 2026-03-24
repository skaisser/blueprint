---
name: flow
description: >
  Auto-chain the full SDLC workflow with checkpoints and pauses for review.
  Use this skill whenever the user says "/flow", "run the full workflow",
  "auto chain", "start to finish", "full pipeline", "plan and execute",
  or any request to run the complete plan-to-finish pipeline automatically.
  Also triggers on "chain skills", "workflow pipeline", "run everything",
  "take this from plan to PR", "full dev cycle", or "run plan through finish".
  NOT for autonomous/zero-intervention flows (use /flow-auto instead).
  Orchestrates: plan → review → approve → check → PR → finish.
---

# Flow: Auto-Chain SDLC Workflow

Orchestrate the full plan-to-finish pipeline with automatic skill chaining and review checkpoints.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

**Instead of 10 manual skill invocations, run one `/flow` and let it guide you through.**

## The Pipeline

```
/plan → /plan-review → ⏸ PAUSE → /plan-approved → /plan-check → ⏸ PAUSE → /pr → ⏸ PAUSE → /finish
                        (review)                                  (quality)        (review PR)
```

### Stages

| # | Skill | What it does | Pause after? |
|---|-------|-------------|--------------|
| 1 | `/plan` | Create the plan document | No — flows into review |
| 2 | `/plan-review` | Validate, mark complexity, execution strategy | **YES — user reviews plan** |
| 3 | `/plan-approved` | Execute all phases | No — flows into check |
| 4 | `/plan-check` | Verify all tasks complete, audit | **YES — user reviews check results** |
| 5 | `/pr` | Create pull request | **YES — user reviews PR** |
| 6 | `/finish` | Close plan, archive, GitHub issue update | End |

### Context Management (1M context)

With 1M context, most plans run the **full pipeline in a single session** — no mandatory clears needed. The coordinator only orchestrates (delegates to subagents), so it needs minimal context headroom.

```
Context < 30% after plan-review:   continue directly — full pipeline in ONE session
Context 30-50%:                    suggest clear, but continuing works fine
Context > 50%:                     recommend clear before plan-approved
```

The flow skill monitors actual context usage and suggests breaks only when genuinely needed — not at fixed pipeline stages or task counts.

## Usage

```bash
/flow <description>              # Start from scratch — runs /plan first
/flow --from plan-review         # Pick up from a specific stage
/flow --from plan-approved       # Skip to execution (plan already reviewed)
/flow "#42" <description>        # Start with GitHub issue linked
```

## Process

### Step 1: Determine Entry Point

```bash
echo "🔀 [flow:1] determining entry point"
~/.blueprint/bin/blueprint meta 2>/dev/null
```

**If `--from` flag:** Jump to that stage directly.

**If active plan exists** (blueprint meta returns plan_file with status):
- `awaiting-approval` → Start at Stage 2 (plan-review)
- `approved` → Start at Stage 3 (plan-approved)
- `in-progress` → Start at Stage 3 (plan-approved / resume)
- `completed` → Check if PR exists, if not start at Stage 5

**If no active plan:** Start at Stage 1 (plan).

### Step 2: Execute Current Stage

Run the appropriate skill for the current stage. Pass through any relevant arguments.

**Stage 1 — Plan:**
```
Running /plan with your description...
```
If `--wt` flag is present or user says "worktree": run `/plan-wt` instead of `/plan`.
→ When plan skill completes, automatically proceed to Stage 2.

**Stage 2 — Plan Review:**
```
Running /plan-review...
```
→ When review completes, hit **Checkpoint U**.

**Stage 3 — Plan Approved:**
```
Running /plan-approved...
```
→ When execution completes, automatically proceed to Stage 4.

**Stage 4 — Plan Check:**
```
Running /plan-check...
```
→ When check completes, hit **Checkpoint P**.

**Stage 5 — PR:**
```
Running /pr...
```
→ When PR is created, hit **Checkpoint R**.

**Stage 6 — Finish:**
```
Running /finish...
```
→ Pipeline complete.

### Step 3: Checkpoints

#### Checkpoint U: Post-Review (after Stage 2 — Unpack)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏸  CHECKPOINT: Plan Review Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Plan: NNNN-type-description
Phases: N phases, M tasks
Strategy: [execution mode summary]
Issue: #XX (if applicable)

Plan size determines next step:
    ≤15 tasks: continue directly to /plan-approved
    >15 tasks: recommend clearing context first
```

Use `AskUserQuestion` — options depend on plan size:

**Small/medium plan (≤15 tasks):**
- **"Continue — run /plan-approved now"** → Proceed directly to Stage 3 in same session
- **"I want to adjust the plan"** → Let user edit, then re-run `/plan-review`
- **"Stop here for now"** → Save progress, user can resume later

**Large plan (>15 tasks):**
- **"Clear context first"** → Tell user to run `/flow --from plan-approved` in fresh session
- **"Continue anyway"** → Proceed to Stage 3 in same session
- **"I want to adjust the plan"** → Let user edit, then re-run `/plan-review`
- **"Create a worktree first"** → Run `/plan-review -wt`, then tell user to cd to worktree

#### Checkpoint P: Post-Check (after Stage 4 — Preflight)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏸  CHECKPOINT: Plan Check Complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Plan: NNNN-type-description
Check result: [pass/issues found]
Issues: N issues found (if any)
```

**If plan-check found issues:**
Use `AskUserQuestion`:
- **"Fix issues and re-check"** → Address issues, re-run `/plan-check`
- **"Proceed to PR anyway"** → Continue to Stage 5 despite issues
- **"Stop here"** → Save progress, user fixes manually

**If plan-check passed clean:**
Use `AskUserQuestion`:
- **"Looks good — create PR"** → Proceed to Stage 5
- **"Wait, I want to review first"** → Pause, user reviews implementation
- **"Stop here for now"** → Save progress, user can resume later

#### Checkpoint R: Post-PR (after Stage 5 — Raise)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⏸  CHECKPOINT: PR Created
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PR: #NNN — title
URL: https://github.com/...
Branch: feat/description → staging branch
```

Use `AskUserQuestion`:
- **"PR looks good — run /finish"** → Proceed to Stage 6
- **"Wait for review first"** → Stop here, user can run `/finish` later
- **"Needs changes"** → User addresses feedback, then `/flow --from pr` to re-create

### Step 4: Context Management

Monitor context usage throughout the flow. At natural transition points:

**If context < 70%:** Continue to next stage in same session.

**If context 70-85%:** Warn user:
```
⚠️  Context at ~75%. Recommend clearing before next stage.
    Run: /flow --from <next-stage>
```

**If context > 85%:** Force checkpoint:
```
🔴 Context high. Clear before continuing.
    Next: /flow --from <next-stage>
```

### Step 5: Pipeline Complete

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅  FLOW COMPLETE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Plan: NNNN-type-description ✅
PR: #NNN merged
Issue: #XX → closed (if applicable)
Branch: cleaned up

Total pipeline: /plan → /plan-review → /plan-approved → /plan-check → /pr → /finish
```

## Variations

### Quick Flow (small tasks)
If the description sounds small (≤3 tasks), suggest `/quick` instead:
```
This sounds like a small task. Would you prefer:
1. /quick — Just do it, no plan overhead
2. /flow — Full pipeline (recommended for anything touching 4+ files)
```

### Hotfix Flow
If user says "urgent", "hotfix", or "emergency":
```
This sounds urgent. Would you prefer:
1. /hotfix — Emergency: commit → push → PR → merge (fastest)
2. /flow — Full pipeline with review checkpoints (safer)
```

### Resume Flow
If `/flow --from plan-approved` detects partial progress:
```
Plan has partial progress (Phase 2/4, 8/20 tasks).
Redirecting to /resume for efficient re-entry...
```
→ Hand off to `/resume` skill.

## Rules

- **Checkpoints are mandatory** — never skip review pauses (after plan-review, plan-check, and PR)
- **Context breaks are conditional** — only needed for large plans (>15 tasks); small/medium plans flow directly
- **Search memory at entry** — past plans inform the current one
- **Suggest /quick for small tasks** — don't over-engineer
- **Suggest /hotfix for emergencies** — don't slow down urgent fixes
- **Monitor context usage** — proactively warn about context limits
- **Each stage is idempotent** — safe to re-run with `--from`
- **GitHub issues flow through** — issue numbers propagate across all stages automatically

## Flags

- `--from <stage>`: Start from specific stage (plan, plan-review, plan-approved, plan-check, pr, finish)
- `--wt`: Use worktree mode (passed to plan-review)

Use $ARGUMENTS as the task description, GitHub issue number, or flags.
