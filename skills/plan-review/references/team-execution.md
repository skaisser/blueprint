# Team Execution Guide

## Core Principle

**Leader NEVER implements code.** Leader orchestrates: read plan → delegate → verify → update plan → commit. This keeps the main context lean — with 1M context, even large plans (10+ phases, 40+ tasks) complete in a single session without restarts.

## Philosophy: Quality × Speed

BLUEPRINT optimizes for **shipping correct code fast**, not for minimizing API cost.

- **Sonnet for volume**: 10 parallel Sonnet workers finish faster than 2 Opus workers doing the same mechanical tasks
- **Opus for quality**: Business logic, test design, and architecture decisions get Opus — the cost of a retry from a wrong implementation is higher than the cost of using a better model
- **Never Haiku**: The quality gap causes retries that cost more than using Sonnet directly

The coordinator (leader) always runs on Opus — it makes delegation decisions that affect the entire plan's execution quality.

## Model Selection

Assign models based on task complexity. Optimize for **quality and speed** — cost is secondary.

| Marker | Model | When to Use |
|--------|-------|-------------|
| `[H]` | Sonnet | Fast output for mechanical tasks: CRUD, config, styling, renaming, migrations, simple tests |
| `[S]` | Opus | Maximum quality for real logic: business logic, services, API integration, complex tests, components |
| `[O]` | Opus | Deep reasoning: architecture decisions, multi-system coordination, sequential-only (rare) |

**Default to `[H]` / Sonnet.** Escalate to `[S]` / Opus only when the task needs reasoning. `[O]` is almost never needed — it's the same model as `[S]` but signals "this task is sequential and needs full context."

Speed comes from **parallelism** (multiple Sonnet workers simultaneously), not from using weaker models. One Opus worker thinking deeply is faster than three retries from a model that wasn't smart enough.

## Delegation Strategy

| Phase Shape | Strategy | How |
|-------------|----------|-----|
| 2+ parallel phases with `[S]`/`[O]` | **Parallel Teams** | Spawn N team-lead subagents in ONE message — each runs its own team internally |
| 2+ parallel phases, all `[H]` | **Parallel Subagents** | Spawn all subagents in ONE message — fire and forget |
| 1 sequential phase with `[S]`/`[O]` | **Single Team** | Coordinator directly creates one team, assigns workers, manages mid-task |
| Sequential `[S]`/`[O]` + concurrent `[H]` | **Mixed dispatch** | 1 team (coordinator manages) + subagents (fire-and-forget), all dispatched in ONE message |
| Tasks within a team need handoff | **Coordinated Team** | Worker A writes handoff block → coordinator re-prompts Worker B |
| 1 phase all `[H]`, sequential | **Single Subagent** | One `Agent` call, handles tasks in order |
| ≤3 `[H]` tasks total | **Leader Direct** | No spawn — coordinator handles directly |

**The constraint:** Each team-lead can manage ONE team at a time.
- Parallel Teams = N team-lead subagents running independently (NOT one coordinator managing N teams)
- Single Team = coordinator IS the team-lead, gives full attention to one team

### Why Always Delegate?

Real-world performance data shows consistent patterns:
- Leader implementing directly: 8 tasks, ~13 min, ~80% context consumed
- Coordinated team (3 workers): 6 tasks, ~4 min, ~20% context consumed
- Coordinated team (2 workers): 4 tasks, ~6 min, ~20% context consumed

Delegating everything keeps the leader at ~20% context usage vs ~80% when implementing directly.

## Delegating via Parallel Teams (independent [S]/[O] phases)

Spawn N "team-lead subagents" in ONE message — each runs independently as its own team coordinator:

1. Spawn subagent A with a full prompt: "You are the team lead for Phase 1. Create a team named `team-alpha`, assign workers (worker-1, worker-2, etc.), implement Phase 1..."
2. Spawn subagent B with a full prompt: "You are the team lead for Phase 2. Create a team named `team-beta`, assign workers (worker-1, worker-2, etc.), implement Phase 2..."
3. Both subagents run in parallel — each creates its own team and manages its own workers
4. Plan coordinator waits for both to report back, then commits and continues

**Key:** The plan coordinator does NOT create the teams directly. It spawns team-lead subagents that each manage their own team. One team-lead = one team.

**When to use over parallel subagents:** Any phase with `[S]`/`[O]` work benefits from a full team internally — smarter, workers can coordinate within the phase. Pure `[H]` phases stay as regular subagents.

## Delegating via Parallel Subagents (independent all-[H] phases)

Spawn all workers in ONE message — they run simultaneously with zero team overhead:

The key is: launch multiple workers concurrently and keep prompts self-contained.

For a single phase, use one worker (no overhead either).

Model selection guideline: choose the worker model tier based on the hardest task it owns (H/S/O).

Include in every worker prompt:
- Specific files and logic to implement
- Relevant context from the plan
- Read `blueprint/.config.yml` for project config (staging branch, language, stack)
- "Run the project's formatter before finishing"
- "Run targeted tests for your changes"
- "Commit format: `<emoji> <type>[(scope)][!]: <description>` — present tense, lowercase"
- "NEVER add AI signatures to commits (Co-Authored-By, Generated by Claude, etc.)"
- "NEVER stage `.env`, `*-api-key`, or credential files"

## Coordinated Team (handoff required)

If worker-to-worker messaging exists in your tool, use it.

If it does NOT exist (common), emulate coordination in a tool-agnostic way:

1. Worker A: produce the intermediate artifact (contract/type/schema decision) and write it into the plan file under a "Handoff" heading
2. Leader: sanity-check the handoff and re-prompt Worker B with the handoff pasted in
3. Worker B: implement against the handoff and report back
4. Leader: integrate and run targeted tests

## Worker Prompt Template

Include in every worker prompt:
- Specific files and logic to implement
- Relevant context from the plan
- Read `blueprint/.config.yml` for project config (staging branch, language, stack)
- "Run the project's formatter before reporting done"
- "Run targeted tests for your changes"
- "Commit format: `<emoji> <type>[(scope)][!]: <description>` — present tense, lowercase"
- "NEVER add AI signatures to commits (Co-Authored-By, Generated by Claude, etc.)"
- "NEVER stage `.env`, `*-api-key`, or credential files"
- "When done, report summary of changes + test results"

## Worker Naming Convention

Use sequential IDs within each team for clarity in logs and plan tracking:

- Single team: `worker-1`, `worker-2`, `worker-3`, ...
- Parallel teams: prefix with team name — `alpha-worker-1`, `beta-worker-1`, etc.
- Team names: `team-alpha`, `team-beta`, `team-gamma`, `team-delta`, ...

## Execution Performance Log

Track execution times to build evidence for team vs subagent decisions.

| Plan | Phase | Strategy | Tasks | Time | Context |
|------|-------|----------|-------|------|---------|
| *Add rows after completing plans* | | | | | |
