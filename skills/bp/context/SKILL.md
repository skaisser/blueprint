---
name: bp:context
description: >
  Scan project and generate/audit CLAUDE.md files in directories with complex patterns.
  Triggers on "/bp:context", "/context", "generate context", "audit CLAUDE.md",
  "scan project", or any request to create or update CLAUDE.md documentation files.
  Also triggers on "update context docs", "context scan", "check CLAUDE.md", "onboard project",
  "stale documentation", or "refresh project docs". Runs parallel workers per directory.
  Also audits README.md for staleness.
---

# Context: Generate & Audit CLAUDE.md Files

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Scan the project and generate/audit CLAUDE.md files in directories with complex patterns. Runs multiple subagents in parallel — one per directory cluster.

## When to Run

- After `/bp:start` on an existing codebase
- After adding major new packages or refactoring a directory
- When a CLAUDE.md feels stale or has incorrect info
- When onboarding to a project for the first time

## Critical Rules

- NEVER overwrite existing CLAUDE.md without `--force` — but ALWAYS audit for staleness
- NEVER run full test suite — targeted tests only
- Read actual code — never guess patterns
- Dispatch ALL directory workers in ONE message — parallel, not sequential
- Leader NEVER writes CLAUDE.md files directly — always delegates to workers

## Step 1: Detect Project Type + Search Memory — MANDATORY

Run: `echo "[context:1] detecting project type + searching memory"`

**Detect framework** markers (Laravel, Expo, React Native, Next.js, React, etc.).

## Step 2: Framework-Specific Setup Check

Check for framework-specific tools and helpers. If available, run update commands.

## Step 3: Scan Project Structure

Run: `echo "[context:3] scanning project structure"`

Identify directories that benefit from CLAUDE.md (5+ files with shared patterns, non-obvious rules, complex flows, critical gotchas, external API integrations).

## Step 4: Audit Existing CLAUDE.md Files — MANDATORY

```bash
find . -name "CLAUDE.md" -not -path "*/vendor/*" -not -path "*/node_modules/*" 2>/dev/null
```

Check for staleness: wrong testing restrictions, incorrect versions, deprecated patterns, missing conventions.

## Step 5: Dispatch Parallel Workers

Dispatch ALL workers in ONE message. Group directories into logical clusters.

Each worker: reads files, audits existing CLAUDE.md, generates new ones where needed. Workers use `general-purpose` subagent type. Max 200 lines per CLAUDE.md. Document the NON-OBVIOUS.

## Step 6: Root CLAUDE.md

Generate or audit root CLAUDE.md based on project structure and conventions found during scanning.
- **Testing Patterns**: How tests are organized across the project
- **Architecture Decisions**: Patterns that emerged across plans
- **Known Gotchas**: Problems discovered in past plans that future sessions should know about

This gives future sessions instant context — not just what the project IS, but what was LEARNED building it.

## Step 7: Report

Show created, updated, unchanged, and skipped directories.

## Step 8: Audit README.md — MANDATORY

Cross-reference README.md against what was learned during the scan. Check versions, integrations, structure tree, commands, missing patterns. Use `AskUserQuestion` to offer updates.

## Flags

- `--force` / `-f` — Regenerate all CLAUDE.md files
- `--dry-run` / `-d` — Show what would be created/updated
- `--root` / `-r` — Root CLAUDE.md only
- `--audit` / `-a` — Audit only, no new generation

Use $ARGUMENTS as a specific directory path or flag.
