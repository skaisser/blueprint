---
name: plan
description: >
  Create structured implementation plans with git branches and memory search.
  Use this skill whenever the user wants to plan a feature, fix, refactor, or any development task.
  Triggers on: "/plan", "plan this", "create a plan for", "let's plan", "I need to plan",
  "break this into phases", "how should we approach", "let's think about how to build",
  or any request that involves planning, phasing, or strategizing before coding.
  For worktree-based plans (isolated directory), use /plan-wt instead.
  ALWAYS use this skill before starting implementation — planning first, coding second.
---

# Plan: Create Implementation Plan & Branch

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Create a plan and branch in the current repository. The plan document lives in `blueprint/live/` and follows a strict template with phases, tasks, and acceptance criteria.

**Need an isolated worktree instead?** Use `/plan-wt` — same plan, but creates a separate directory.

## Critical Rules

These are non-negotiable — violating them produces broken plans:

1. **Read the template first.** Read `references/plan-template.md` (bundled with this skill) at Step 1. Never create a plan without it. This relative path works from both `${CLAUDE_PLUGIN_ROOT}/skills/plan/` (plugin install) and `~/.claude/skills/plan/` (traditional install).
2. **Use sequential thinking.** Call `mcp__sequential-thinking__sequentialthinking` at Step 2. This forces structured analysis before writing.
3. **Ask when ambiguous.** Use `AskUserQuestion` at Step 2 if anything is unclear. Don't assume scope, priority, or approach.
4. **Follow step order.** Steps are designed to build on each other. Skipping or reordering breaks the data flow.
5. **Do NOT mark `[H]`/`[S]`/`[O]` complexity tiers.** That is `/plan-review`'s responsibility. Leave tasks unmarked.

## Task Quality Rules

Every task in the plan MUST follow these rules:

- **Specific and actionable** — never write vague tasks like "implement feature", "set up backend", "add functionality", or "handle edge cases". Every task must name the file(s) it touches and the concrete deliverable (e.g., "Add `status` enum column to `orders` migration" not "update database").
- **Atomic** — each task must be independently executable by a subagent with no implicit dependencies on other tasks in the same phase. If task B requires task A's output, they belong in different phases.
- **Right-sized** — match task granularity to scope:
  - A 1-phase config/migration change: 2–4 tasks
  - A multi-phase feature: 4–8 tasks per phase
  - If a phase has more than 10 tasks, split it into two phases

## GitHub Issue Integration

If `$ARGUMENTS` contains a GitHub issue number (e.g., `/plan 42`, `/plan #42`, `/plan issue 42`), fetch the issue and use it as context:

1. Parse the issue number from `$ARGUMENTS` (strip `#` prefix or `issue` keyword if present)
2. Fetch issue details:
   ```bash
   gh issue view <NUMBER> --json title,body,labels,assignees
   ```
3. Use the issue title as the plan title, and the issue body as context for the Goal/Context sections
4. Set `issue: <NUMBER>` in the plan's frontmatter
5. If `$ARGUMENTS` is a plain description (not a number), set `issue: null` as default

**How to distinguish from backlog:** If the number matches a file in `blueprint/backlog/` or `blueprint/expired/`, treat it as a backlog ID (see below). If no backlog file matches, treat it as a GitHub issue number. The user can also be explicit: `/plan issue 42` always means GitHub issue, `/plan backlog 3` always means backlog.

## Backlog Integration

If `$ARGUMENTS` is a number (e.g., `/plan 3`) and matches a backlog item:

1. Check `blueprint/backlog/` and `blueprint/expired/` for that idea number
2. Use its content as the basis for the plan (Context section feeds the plan's Context)
3. Add `backlog: "NNNN"` to the plan's frontmatter (bidirectional link)
4. After creating the plan, update the backlog item:
   - Set frontmatter `status: planned`
   - Set frontmatter `plan: "PLAN_ID"` (the new plan's ID)
   - Move the file to `blueprint/expired/` via `git mv`


## Step 1: Parallel Data Gather

```bash
echo "🔷 BP: plan [1/3] parallel data gather — blueprint meta + template + pull"
```

Run these **simultaneously in a single step** (do not wait between them):

1. `~/.blueprint/bin/blueprint meta` — returns next_num, base_branch, project, team, git_remote as JSON. Use `next_num` for NEXT_NUM, `base_branch` for BASE_BRANCH.
2. Read `references/plan-template.md` (bundled with this skill) — plan FORMAT and phase structure reference. Use blueprint meta for the number, not this file.
3. `git checkout "$BASE_BRANCH" && git pull origin "$BASE_BRANCH"`

Do NOT create a plan from memory. Do NOT proceed without reading the template.

## Step 2: Explore + Sequential Thinking

```bash
echo "🔷 BP: plan [2/3] code exploration + sequential thinking"
```

**First — launch Explore agents in parallel** (before sequential thinking, not after):
- Launch one Explore agent per affected domain (models, services, controllers, tests, etc.)
- Multiple agents run simultaneously — don't wait for one before launching the next
- Use $ARGUMENTS to guide what to explore
- **MAX 3 Explore agents** — more adds time without proportional value
- **Skip Explore entirely** when: $ARGUMENTS is < 20 words AND mentions a single file/component, OR task is config/migration/styling only
- **1 agent is enough** when: task touches 1-2 files in the same domain (e.g., "add a field to User model")

**Then — sequential thinking** with all gathered data:

You MUST use `mcp__sequential-thinking__sequentialthinking`. This is where the plan takes shape:

- Analyze $ARGUMENTS, GitHub issue details, and Explore findings together
- Break into phases with clear acceptance criteria and phase dependencies
- Ensure the plan is runnable from a clean context:
  - Body has exactly 4 sections: Goal, Non-Goals, Context (bullet list), Phases, Acceptance
  - For each phase: Touches + Tasks + Verify (3 fields only — no Goal, no Done-when)
  - If a coordinated team will be required, add an explicit `## Handoffs` section
- With 1M context, there is no need to split plans — the coordinator only orchestrates (delegates to subagents), so even large plans (10+ phases, 40+ tasks) are fine. Keep everything in a single plan for simplicity.
- Every plan MUST include test/verification tasks — testing is part of planning, not an afterthought
- **Do NOT assign `[H]`/`[S]`/`[O]` complexity tiers** — `/plan-review` handles that
- Keep focused: 3-6 thoughts is enough for most plans
- Use `AskUserQuestion` for genuine ambiguities ONLY — do NOT assume

## Step 3: Create Plan and Branch

```bash
echo "🔷 BP: plan [3/3] creating plan file and branch"
```

Write the plan file using the v2 format. The frontmatter has NO `plan_file`, `phases_total/done`, `tasks_total/done`, `sessions` array, or `name` field. Status starts as `todo`.

```bash
# Create blueprint/live directory if needed
mkdir -p blueprint/live

# Write plan file using v2 frontmatter format
cat > "blueprint/live/${NEXT_NUM}-<type>-<slug>.md" <<'EOF'
---
id: "NNNN"
title: "<type>: Short Description"
type: <type>
status: todo
project: <project>
branch: <type>/<short-description>
base: <BASE_BRANCH>
tags: [relevant, tags]
backlog: null
issue: null
created: "DD/MM/YYYY HH:MM"
completed: null
pr: null
session: null
---

# <type>: Short Description

## Goal
{1-3 sentences}

## Non-Goals
{What is explicitly out of scope}

## Context
- `path/to/file` — what and why
- Constraint: ...

## Phases

### Phase 1: {Name}
**Touches:** ...
- [ ] Task (unmarked — no [H]/[S]/[O])
**Verify:** Run the project's test command with appropriate filter

## Acceptance
- [ ] Criterion
EOF

# Create and checkout feature branch
git checkout -b "<type>/<short-description>"

# Commit the plan file + verify branch in one step
git add blueprint/ && git commit -m "📋 plan: add ${NEXT_NUM}-<type>-<short-description>" && echo "Branch: $(git branch --show-current)"
```

If the output branch name does not match the intended branch — **STOP**. Do NOT proceed to implementation. Report the failure and ask the user to resolve it.

## Step 4: Present and Ask Next Step

```bash
code "$PWD" && code "$PWD/$PLAN_FILE"
```

Open the plan in VS Code so the user can review it before `/plan-review`.

Output a summary:
```
Plan created and branch ready!
Plan:   blueprint/live/NNNN-type-description.md
Branch: type/description
Phases: N phases, X tasks
```

**STOP. Use `AskUserQuestion` here.**

- **Question:** "Plan ready. Run /plan-review now?"
- **Option 1:** "I'll review in VS Code (recommended)" — User reviews and requests edits; do NOT run /plan-review yet
- **Option 2:** "Make edits now" — Apply requested changes, then re-open in VS Code
- **Option 3:** "Looks good — run /plan-review" — Only run after user confirms

## Completion Flow

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Rules

- Create plan in base branch FIRST, then create feature branch
- Break into logical phases with acceptance criteria
- Use $ARGUMENTS as the feature/fix description to plan
