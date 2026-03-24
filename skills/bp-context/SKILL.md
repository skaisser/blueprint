---
name: bp-context
description: >
  Scan project and generate/audit CLAUDE.md files with stack auto-detection and Context7 docs.
  Triggers on "/bp-context", "/context", "generate context", "audit CLAUDE.md",
  "scan project", or any request to create or update CLAUDE.md documentation files.
  Also triggers on "brownfield", "onboard project", "onboard this project",
  "onboard existing project", "scan existing project", "update context docs",
  "context scan", "check CLAUDE.md", "stale documentation", or "refresh project docs".
  Runs parallel workers per directory. Also audits README.md for staleness.
---

# Context: Generate & Audit CLAUDE.md Files

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Scan the project, auto-detect the tech stack, query framework documentation via Context7, and generate a tree of lean, focused CLAUDE.md files. Runs multiple subagents in parallel — one per directory cluster.

## When to Run

- After `/start` on an existing codebase (brownfield onboarding)
- When onboarding to a project for the first time
- After adding major new packages or refactoring a directory
- When a CLAUDE.md feels stale or has incorrect info
- When the user says "brownfield", "onboard project", or "scan existing project"

## Critical Rules

- NEVER overwrite existing CLAUDE.md without `--force` — but ALWAYS audit for staleness
- NEVER run full test suite — targeted tests only
- Read actual code — never guess patterns
- Dispatch ALL directory workers in ONE message — parallel, not sequential
- Leader NEVER writes CLAUDE.md files directly — always delegates to workers
- Use `AskUserQuestion` for ALL user interactions — never ask questions in plain text
- Do NOT auto-commit — let the user review generated files

## Step 1: Detect Stack — MANDATORY

Run: `echo "[context:1] detecting project stack"`

Use the CLI first — it reads the project and emits a structured summary:

```bash
blueprint detect-stack 2>/dev/null
```

If that succeeds, use its output directly. If it fails or is unavailable, fall back to manual detection using the tables in `references/stack-detection.md`.

After detection, echo a summary:

```bash
echo "[context:1] Stack detected:"
echo "  Language:    PHP 8.3"
echo "  Framework:   Laravel 11.x + Livewire 3.x"
echo "  Test runner: Pest PHP"
echo "  Assets:      Vite + Tailwind CSS"
echo "  Database:    MySQL + Redis"
```

## Step 2: Read Config — MANDATORY

Run: `echo "[context:2] reading blueprint config"`

Read `blueprint/.config.yml` for:
- `staging_branch` — used in root CLAUDE.md branch flow
- `language` — content language
- `stack` — compare against auto-detected values; flag discrepancies

## Step 3: Query Context7 for Framework Docs

Run: `echo "[context:3] querying Context7 for framework documentation"`

For the primary detected framework, use Context7 MCP tools:

### 3a. Resolve Library ID

Call `mcp__context7__resolve-library-id` with the framework name (e.g., `"laravel"`, `"nextjs"`, `"django"`).

### 3b. Query Documentation

Using the resolved library ID, call `mcp__context7__query-docs` for:

1. `"project directory structure conventions"`
2. `"best practices and common patterns"`
3. `"testing conventions and patterns"`

Use results to inform which CLAUDE.md files to generate and what conventions to include.

### 3c. Context7 Fallback

If Context7 MCP tools are unavailable:
1. Log: `echo "[context:3] Context7 unavailable — using built-in conventions"`
2. Fall back to framework templates in `references/stack-detection.md`
3. Continue without error — Context7 is optional enrichment, not required

## Step 4: Scan Project Structure

Run: `echo "[context:4] scanning project structure"`

```bash
blueprint context 2>/dev/null || find . -name "CLAUDE.md" -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/.git/*" 2>/dev/null
```

Identify directories that exist and would benefit from a CLAUDE.md:
- Directories with 5+ files that share patterns
- Directories with non-obvious rules or conventions
- Directories with complex flows or critical gotchas
- Directories with external API integrations

### Determine Mode

- **No existing CLAUDE.md files** → Full generation mode
- **Existing CLAUDE.md files found** → Audit mode (see Step 7)

## Step 5: Generate Root CLAUDE.md

Run: `echo "[context:5] generating root CLAUDE.md"`

Structure (keep under 60 lines — the root is a MAP, not an encyclopedia):

```markdown
# {Project Name}

## Tech Stack
{Auto-detected: language, framework + version, test runner, assets, database}

## BLUEPRINT Workflow
This project uses BLUEPRINT SDLC. Run `blueprint update` to update skills.

### Pipeline
/plan → /plan-review → /code → /test → /tdd-review → /commit → /push → /ship → /finish

### Branch Flow
feature branch → {staging_branch} → main

### Commit Format
<emoji> <type>: <description>

## Testing
- Runner: {detected test runner}
- NEVER mock what you can test — use real implementations
- {framework-specific testing notes}

## Key Conventions
{Framework-specific conventions from Context7 or built-in knowledge}

## Workspace
Plans, tasks, and context live in `blueprint/` — see README there.
```

## Step 6: Generate Subdirectory CLAUDE.md Files

Run: `echo "[context:6] generating subdirectory CLAUDE.md files"`

Dispatch ALL workers in ONE message. Each worker generates CLAUDE.md files for its directory cluster.

### Rules for Subdirectory Files

- **10-20 lines max** per file — lean and focused
- **Only for directories that actually exist** — never create directories
- **Content from Context7 results OR built-in conventions** — prefer Context7 when available
- **Document the NON-OBVIOUS** — skip things any developer would know
- **No duplication** — don't repeat what's in root CLAUDE.md

For framework-specific directory templates, read `references/stack-detection.md` → "Framework-Specific Directory Templates".

## Step 7: Audit Mode (Existing CLAUDE.md Files)

Run: `echo "[context:7] auditing existing CLAUDE.md files"`

### 7a. Compare Against Detected Stack

- Check framework version references — flag if outdated
- Check test runner references — flag if changed
- Check convention references — flag deprecated patterns
- Check directory references — flag if directories moved or removed

### 7b. Check for Staleness

- Wrong framework version mentioned
- Outdated patterns or deprecated APIs
- Missing conventions for newly added directories
- References to files/directories that no longer exist

### 7c. Report and Suggest

Use `AskUserQuestion` to present findings — list stale references, suggest specific updates, let user approve each change. Do NOT overwrite without explicit confirmation.

### 7d. Generate Missing Files

If new directories exist without CLAUDE.md, generate them following Step 6 rules. Report separately from audit findings.

## Step 8: Audit README.md — MANDATORY

Run: `echo "[context:8] auditing README.md"`

Cross-reference README.md against what was learned during the scan:
- Framework/language versions — match detected versions
- Listed integrations — still present in dependencies?
- Directory structure tree — matches actual structure?
- Commands/scripts — still valid?

Use `AskUserQuestion` to offer README.md updates if staleness is found.

## Step 9: Report

Run: `echo "[context:9] generation complete"`

Show summary:
- **Stack detected**: language, framework, test runner, assets, DB
- **Context7**: whether it was used, which docs were queried
- **Created**: list of new CLAUDE.md files generated
- **Updated**: list of existing CLAUDE.md files updated (audit mode)
- **Unchanged**: list of CLAUDE.md files still current
- **Skipped**: directories that exist but don't need CLAUDE.md
- **README.md**: staleness findings if any

Remind the user: "Review the generated files. Run `/bp-commit` when satisfied."

## Flags

- `--force` / `-f` — Regenerate all CLAUDE.md files (overwrite existing)
- `--dry-run` / `-d` — Show what would be created/updated without writing
- `--root` / `-r` — Root CLAUDE.md only
- `--audit` / `-a` — Audit only, no new generation
- `--no-context7` — Skip Context7 queries, use built-in conventions only

Use $ARGUMENTS as a specific directory path or flag.
