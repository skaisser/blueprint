---
name: backlog
description: >
  Capture, manage, and promote pre-planning ideas before they become full plans.
  Use this skill whenever the user says "/backlog", "add to backlog", "backlog idea",
  "capture this idea", "save this for later", "I had an idea", or any request to manage
  the idea backlog. Also triggers on "list ideas", "promote idea", "backlog list",
  "drop idea", "mark done", "what's in the backlog", "show me the ideas", "pending ideas",
  or any request to add/view/promote/done/drop items in blueprint/backlog/.
  The backlog is the pre-planning stage — ideas live here until promoted to /plan.
  Even if the user doesn't explicitly say "backlog", trigger when they describe wanting
  to capture a feature idea, task, or improvement for later without implementing it now.
  This skill is the gateway to the SDLC — nothing gets planned or built without first
  passing through the backlog.
---

# Backlog: Pre-Planning Idea Manager

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

The backlog is the first stage of the development lifecycle. It exists because good ideas deserve to be captured immediately — but implemented thoughtfully. When someone says "we should add notifications" during a bug fix, the backlog catches that idea so it doesn't get lost, without derailing the current work.

Think of it as a parking lot for ideas: lightweight enough to add in seconds, structured enough to evaluate later.

## Git Policy

**Read-only commands (add, list, view, archive):** These MUST NOT run any git commands — no `git add`, no `git commit`, no `git status`, no `git mv`. They only create or read `blueprint/backlog/` files. The backlog is a scratch pad; ideas are uncommitted working-directory files until they graduate.

**State-changing commands (promote, done, drop):** These DO use git (`git mv`, `git add`, `git commit`) because they represent decisions that should be recorded in history.

This separation is intentional: capturing an idea should be instant and zero-friction. Git operations only happen when the idea's lifecycle changes.

## Commands

```
/backlog add <description>     → Capture a new idea (NO git)
/backlog                       → List all active ideas (NO git)
/backlog <number>              → View details of a specific idea (NO git)
/backlog promote <number>      → Graduate idea into a full /plan (uses git)
/backlog done <number>         → Mark as completed and archive (uses git)
/backlog drop <number>         → Discard idea and archive with reason (uses git)
/backlog archive               → List archived ideas (NO git)
```

The `$ARGUMENTS` variable contains everything after `/backlog`. Parse it to determine which command the user wants: no args = list, a number = view, "add ..." = add, "promote N" = promote, "done N" = done, "drop N" = drop, "archive" = archive.

---

## Adding an Idea (`/backlog add`)

This is the most common operation. The goal is to capture just enough context that someone (including future-you) can evaluate the idea later without needing the original conversation for context.

**IMPORTANT: `/backlog add` MUST NOT run any git commands. It writes a file to the working directory and nothing else. The file stays uncommitted — it will be committed later when promoted, or during the next `/commit`.**

### Step 1: Scan existing ideas and determine next ID

**MANDATORY: Use the CLI to list existing ideas and check for duplicates.**

```bash
mkdir -p blueprint/backlog blueprint/expired
# List all existing ideas via CLI (handles both YAML and legacy formats)
blueprint backlog --archive --format table
```

Check the output for potential duplicates with the user's idea. The CLI also provides IDs — the next ID is one higher than the highest existing ID.

If a potential duplicate exists, surface it before proceeding. If no duplicates, continue.

**NEVER use grep/sed/awk/cat to parse backlog files. The CLI handles all formats correctly.**

### Step 2: Classify with the user

Use `AskUserQuestion` to quickly classify the idea. Ask a single multi-part question covering:

- **Type**: `feat` (new feature), `fix` (bug), `refactor`, `perf`, `chore`, `docs`, `test`
- **Size**: `small` (< half day), `medium` (1-3 days), `large` (> 3 days)
- **Priority**: `high` (blocking or urgent), `medium` (important but not urgent), `low` (nice to have)

If the user's original message already makes these obvious (e.g., "urgent bug: login is broken"), infer what you can and only ask about what's ambiguous.

### Step 3: Write the idea file

Filename: `blueprint/backlog/NNNN-<type>-<short-slug>.md`

The slug should be 2-4 words, kebab-case, descriptive enough to identify the idea from a file listing.

```markdown
---
id: "NNNN"
title: "<type>: Short Description"
type: feat
status: new
priority: high
size: medium
project: <project-name>
tags: [relevant, tags]
created: "DD/MM/YYYY HH:MM"
plan: null
issue: null
depends: null
---

# <type>: Short Description

## What
One paragraph describing what needs to happen.

## Why
Why this matters — the user impact or technical motivation.

## Context
Any relevant context: related features, technical constraints, conversations that sparked the idea.
Bullet list format — key file paths, API docs, root cause analysis, affected data.

- `path/to/file` — what it does
- Affected: N orders / N users
- Pattern to follow: `path/to/existing/pattern`

## Notes
Optional: links, references, edge cases to consider.
```

### Field Reference

| Field | Type | Purpose |
|-------|------|---------|
| `id` | string | Sequential, zero-padded 4 digits |
| `title` | string | `type: short description` |
| `type` | enum | feat, fix, refactor, test, chore, docs |
| `status` | enum | `new` → `ready` → `planned` → `archived` (or `on-hold`) |
| `priority` | enum | `critical`, `high`, `medium`, `low` |
| `size` | enum | `small` (≤5 tasks), `medium` (6-15), `large` (16+, may need multiple plans) |
| `project` | string | Project slug |
| `tags` | array | Topic tags — same vocabulary as plan tags |
| `created` | string | `DD/MM/YYYY HH:MM` |
| `plan` | string\|null | Plan ID when promoted (e.g., `"0177"`) |
| `issue` | number\|null | GitHub issue number, or null |
| `depends` | array\|null | Other backlog IDs this depends on |

Keep it lightweight. The backlog is for capturing intent, not writing specs. A few sentences per section is ideal. If the user gave a one-liner ("add notifications"), expand it just enough to be useful later, but don't over-document.

### Step 4: Confirm (NO GIT — just confirm)

Show the user a brief summary: the idea number, title, classification, and a reminder they can promote it later with `/backlog promote NNNN`.

**Do NOT run `git add` or `git commit`.** The file is an uncommitted working-directory file. It will be picked up by the next `/commit` or when the idea is promoted/done/dropped.

---

## Listing Active Ideas (`/backlog`)

**MANDATORY: Use the CLI — never parse backlog files manually.**

```bash
blueprint backlog --format table
```

This outputs a clean formatted table with all active backlog items (ID, type, title, priority, size, status, plan link). The CLI handles both YAML frontmatter and legacy blockquote formats correctly.

For JSON output (useful for programmatic checks):
```bash
blueprint backlog
```

If the backlog is empty, say so and remind the user how to add ideas.

**No git commands. No file modifications. Read-only.**

---

## Listing Archived Ideas (`/backlog archive`)

**MANDATORY: Use the CLI.**

```bash
blueprint backlog --archive --format table
```

This shows both active AND archived items. The archive section includes status (Done / Dropped / Planned) and plan links.

**No git commands. No file modifications. Read-only.**

---

## Viewing Details (`/backlog <number>`)

Search both `blueprint/backlog/` and `blueprint/expired/` for a file matching the number prefix (e.g., `0003-*`). Read and display the full contents. If not found, say so and suggest listing active ideas.

**No git commands. No file modifications. Read-only.**

---

## Promoting to Plan (`/backlog promote <number>`)

Promotion is the bridge between "idea" and "implementation." It means the idea has been evaluated and is worth investing planning time into.

1. Find and read the idea file
2. Update its frontmatter: `status: planned`
3. Check if the backlog item has an `issue:` field with a GitHub issue number
4. Ask the user: "Worktree or local?" — this determines whether to run `/plan` (works on current branch) or `/plan-wt` (creates an isolated git worktree)
5. Hand off to the chosen plan skill with the idea's context. If the backlog item has an `issue:` number, pass it so the plan sets `issue: <NUMBER>` in its frontmatter
6. Once the plan is created, update the backlog frontmatter: `plan: "PLAN_ID"` (e.g., `plan: "0177"`)
7. Move the idea file to `blueprint/expired/` using `git mv` (preserves history)
8. Commit: `📋 plan: promote backlog NNNN to plan`

The `/plan` skill will also set `backlog: "NNNN"` in the plan frontmatter, creating a bidirectional link.

---

## Marking Done (`/backlog done <number>`)

For ideas that were completed outside the normal plan flow (e.g., it was a quick fix that didn't need a full plan, or it was resolved by another change).

1. Find the idea file
2. Update frontmatter: `status: archived`
3. Move to `blueprint/expired/` using `git mv`
4. Commit: `📋 plan: mark backlog NNNN as done`

---

## Dropping an Idea (`/backlog drop <number>`)

For ideas that are no longer relevant — priorities changed, the feature was descoped, or it turned out to be unnecessary.

1. Find the idea file
2. Ask the user for a brief reason (or accept one from args: `/backlog drop 3 superseded by new auth system`)
3. Update frontmatter: `status: archived`
4. Move to `blueprint/expired/` using `git mv`
5. Commit: `📋 plan: drop backlog NNNN — <reason>`

---

## Critical Rules

**The backlog never implements.** This is the single most important rule. When a user says `/backlog add notifications`, you create a markdown file describing the idea — you do NOT write code, create migrations, build components, or make any implementation changes. The backlog captures intent. Implementation happens only after the idea is promoted (`/backlog promote`) → planned (`/plan`) → approved (`/plan-approved`).

The reason this matters: jumping straight to code skips the thinking that prevents wasted work. The backlog→plan→implement pipeline ensures ideas are evaluated, scoped, and broken down before anyone writes a line of code.

**Other rules:**
- **CLI-first: ALWAYS use `blueprint backlog` for listing/reading backlog items** — never parse files with grep/sed/awk/cat. The CLI handles both YAML frontmatter and legacy blockquote formats correctly. The audit hook (rule 15) will block manual parsing attempts.
- **Old format migration: If you detect old blockquote-format files, suggest running `blueprint backlog migrate` to convert them to YAML frontmatter before proceeding.**
- One idea per file — keep them atomic and independently promotable
- IDs are global and never reused across active and expired
- Always use `git mv` when moving files (promote/done/drop only) to preserve history
- The only output of `/backlog add` is: a `blueprint/backlog/NNNN-*.md` file written to disk. No git commands.
- List, view, and archive commands are strictly read-only — no git commands, no file writes
- Never scan the entire codebase — backlog is file-based, only read `blueprint/backlog/`
- When in doubt about type/size/priority, ask — don't guess

Use $ARGUMENTS as: description (to add), number (to view), or subcommand + args.
