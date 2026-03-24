# Plan Document Template v2

Use this exact structure when creating plan files in `blueprint/live/`.

## Filename Convention

`blueprint/live/NNNN-<type>-<short-description>.md`

Types: feat, fix, docs, style, refactor, perf, test, build, chore, hotfix

## YAML Frontmatter

### Fields set by `/plan` (creation)

```yaml
---
id: "0042"
title: "feat: Short Description"
type: feat                     # feat|fix|refactor|test|style|chore|docs|hotfix
status: todo                   # todo → awaiting-review → approved → in-progress → completed → canceled
project: my-project            # git repo name
branch: feat/short-description # feature branch name
base: staging                  # target branch for PR (read from blueprint/.config.yml staging_branch)
tags: [payment, webhook]       # topic tags
backlog: null                  # backlog item ID this was promoted from, or null
issue: null                    # GitHub issue number (e.g., 42), or null — /pr adds "Closes #N"
created: "23/03/2026 14:00"    # DD/MM/YYYY HH:MM from date command
completed: null                # filled by /finish
pr: null                       # filled by /pr
session: null                  # filled on first execution — single resume ID
---
```

### Fields added by `/plan-review` (never set by `/plan`)

```yaml
strategy: parallel-teams       # execution strategy (see /plan-review)
reviews:                       # corrections/findings from review
  - "T2: rollback already exists — convert to verify-only"
  - "T5: use Cache::lock not DB lock"
```

### Removed fields (and why)

| Removed | Reason |
|---------|--------|
| `plan_file` | Self-referential — file knows its own name |
| `phases_total` / `phases_done` | Computed by agent from `### Phase` count |
| `tasks_total` / `tasks_done` | Computed by agent from `[x]` vs `[ ]` count |
| `sessions` (array) | Single `session` field for resume. History lives in git log |
| `base_branch` | Renamed to `base` — shorter, same meaning |
| `name: plan` | Unnecessary — it's a plan file by convention |

## Body Structure

The body has exactly 4 sections. No blockquote status bar, no rules block, no session log.

```markdown
# {title}

## Goal
{1-3 sentences: what this plan achieves and why it matters}

## Non-Goals
{What is explicitly OUT of scope — prevents scope creep during execution}

## Context

{Only what the executing agent NEEDS. Key file paths, constraints, interfaces.
Bullet list, not prose. No essays.}

- `path/to/file` — what it does and why it matters
- `path/other/file:L40-87` — specific lines if relevant
- Constraint: must be idempotent / backward-compatible / etc.
- Queue: `queue-name` (existing|new)
- No API changes / New route: POST /api/...

## Phases

### Phase 1: {Descriptive Name}

**Touches:** `app/Path/To/`, `resources/views/`

- [ ] Task description — specific and implementable
- [ ] Another task — one action per checkbox
- [ ] Create test for X scenario

**Verify:** Run the project's test command with appropriate filter

### Phase 2: {Descriptive Name}

**Touches:** `app/Other/Path/`

- [ ] Task description
- [ ] Task description

**Verify:** Run the project's test command with appropriate filter

### Phase N: Tests

**Touches:** `tests/`

- [ ] Test scenario A
- [ ] Test scenario B
- [ ] Regression: all existing related tests pass

**Verify:** Run the project's test command with full scope filter

## Acceptance
- [ ] Criterion that proves the plan succeeded
- [ ] Another criterion
- [ ] All existing related tests pass
```

**Important:**
- Tasks are left **unmarked** (no `[H]`/`[S]`/`[O]`) — `/plan-review` adds those
- No `## Execution Strategy` section — `/plan-review` adds that
- No blockquote status bar — frontmatter is the single source of truth
- No rules block — execution rules live in `/plan-approved` skill
- Status starts as `todo`

## After `/plan-review`

`/plan-review` modifies the plan by:
1. Adding `[H]`/`[S]`/`[O]` markers to every task
2. Adding `## Execution Strategy` section after Phases
3. Updating frontmatter with `strategy` and `reviews`
4. Setting status to `approved`
5. Does NOT change Goal, Non-Goals, Context, or Acceptance (unless a review correction demands it)

## Task Completion Format

When executing, tasks are marked complete with a timestamp. This format is enforced by `/plan-approved`, not by the plan file itself.

```markdown
- [x] [H] Create WebhookController ✅ 23/03/2026 14:32
- [ ] [S] Map payload → ProcessPayment input format
```

Rules (enforced by execution skills, NOT written in plan):
- Timestamp from `date "+%d/%m/%Y %H:%M"` — never guessed
- `✅` emoji marks completion visually
- `/commit` after each phase
- **Never collapse phases** — keep all tasks visible

## Phase Collapsing — REMOVED

**Do NOT collapse phases.** Keep all tasks visible in the plan file at all times. The full task list with `[x]` marks is the audit trail — git history is not a substitute for readable plan state.

## Getting Plan Metadata

Use `~/.blueprint/bin/blueprint meta` to get plan metadata as JSON.
Returns: `next_num`, `base_branch`, `plan_file`, `project`, `team`, `git_remote`.

Note: `blueprint meta` returns `base_branch` (the script's field name). In plan frontmatter, use `base` instead.
