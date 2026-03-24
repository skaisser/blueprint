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

**Read `references/stack-detection.md` (bundled with this skill) before starting.** It contains the full detection tables for languages, frameworks, test runners, asset pipelines, and databases (sections 1a–1f).

Follow sections 1a–1f from that reference, then echo the output summary as shown in section 1f.

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

Call `mcp__context7__resolve-library-id` with the framework name:
- Laravel → `"laravel"`
- Next.js → `"nextjs"`
- Django → `"django"`
- Rails → `"ruby on rails"`
- React → `"react"`
- Vue → `"vue"`
- FastAPI → `"fastapi"`
- Gin → `"gin golang"`
- etc.

### 3b. Query Documentation

Using the resolved library ID, call `mcp__context7__query-docs` for:

1. **Directory structure** — query: `"project directory structure conventions"`
2. **Key patterns** — query: `"best practices and common patterns"`
3. **Testing** — query: `"testing conventions and patterns"`

Use these results to inform:
- Which subdirectory CLAUDE.md files to generate
- What conventions to include in each file
- Framework-specific patterns and anti-patterns

### 3c. Context7 Fallback

If Context7 MCP tools are not available (tools not found, server not running):
1. Log: `echo "[context:3] Context7 unavailable — using built-in conventions"`
2. Fall back to common conventions for the detected framework (see Step 5 framework templates)
3. Continue without error — Context7 is optional enrichment, not required

## Step 4: Scan Project Structure

Run: `echo "[context:4] scanning project structure"`

### 4a. Find Existing CLAUDE.md Files

```bash
find . -name "CLAUDE.md" -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/.git/*" 2>/dev/null
```

### 4b. Map Project Directories

Identify directories that exist and would benefit from a CLAUDE.md:
- Directories with 5+ files that share patterns
- Directories with non-obvious rules or conventions
- Directories with complex flows or critical gotchas
- Directories with external API integrations

### 4c. Determine Mode

- **No existing CLAUDE.md files** → Full generation mode
- **Existing CLAUDE.md files found** → Audit mode (see Step 7)

## Step 5: Generate Root CLAUDE.md

Run: `echo "[context:5] generating root CLAUDE.md"`

The root CLAUDE.md should be customized with detected stack info. Structure:

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

Keep it concise — under 60 lines. The root file is a MAP, not an encyclopedia.

## Step 6: Generate Subdirectory CLAUDE.md Files

Run: `echo "[context:6] generating subdirectory CLAUDE.md files"`

Dispatch ALL workers in ONE message. Each worker generates CLAUDE.md files for its directory cluster. Workers use `general-purpose` subagent type.

### Rules for Subdirectory Files

- **10-20 lines max** per file — lean and focused
- **Only for directories that actually exist** — never create directories
- **Content from Context7 results OR built-in conventions** — prefer Context7 when available
- **Document the NON-OBVIOUS** — skip things any developer would know
- **No duplication** — don't repeat what's in root CLAUDE.md

### Framework-Specific Templates

Generate CLAUDE.md files based on the detected framework. Only create files for directories that **actually exist** in the project.

#### Laravel

| Directory | Key Content |
|-----------|-------------|
| `app/Models/` | Relationships, factories, casts, scopes, never raw SQL |
| `app/Http/Controllers/` | Single responsibility, use Form Requests, resource controllers |
| `app/Http/Requests/` | Validation rules, authorize method, custom messages |
| `app/Services/` | Business logic lives here, not in controllers |
| `app/Livewire/` | Component patterns, wire:model, events, lifecycle |
| `database/migrations/` | Never migrate:fresh, always add new migrations, never modify existing |
| `database/factories/` | Factory patterns, states, relationships |
| `tests/` | Pest patterns, factories over fixtures, no mocking DB |
| `resources/views/` | Blade/Livewire patterns, component library (DaisyUI if detected) |
| `routes/` | Route naming, middleware, group patterns |
| `config/` | Never hardcode — use env(), config caching |

#### Next.js

| Directory | Key Content |
|-----------|-------------|
| `app/` | App Router conventions, server vs client components, layouts, loading/error |
| `components/` | Component naming, props patterns, composition |
| `lib/` | Utility functions, API clients, shared logic |
| `public/` | Static assets only, no sensitive files |
| `tests/` or `__tests__/` | Testing framework, component testing patterns |
| `styles/` | CSS modules or Tailwind patterns |
| `types/` | TypeScript interfaces, shared types |

#### Django

| Directory | Key Content |
|-----------|-------------|
| `apps/` or app directories | App structure, models, views, serializers |
| `templates/` | Template inheritance, block patterns |
| `static/` | Static file handling, collectstatic |
| `tests/` | pytest fixtures, factory_boy, API test patterns |
| `core/` or `config/` | Settings structure, URL configuration |

#### Go

| Directory | Key Content |
|-----------|-------------|
| `cmd/` | Entry points, CLI structure, flag parsing |
| `internal/` | Package patterns, interfaces, dependency injection |
| `pkg/` | Public packages, API stability |
| `api/` | API definitions, proto files, OpenAPI specs |
| `tests/` or `*_test.go` | Table-driven tests, test helpers, testify patterns |

#### React Native / Expo

| Directory | Key Content |
|-----------|-------------|
| `app/` | Expo Router, screen patterns, layouts |
| `components/` | Component patterns, StyleSheet conventions |
| `hooks/` | Custom hooks, state management |
| `services/` or `lib/` | API clients, storage, utilities |
| `__tests__/` | Jest + React Native Testing Library patterns |

#### Rails

| Directory | Key Content |
|-----------|-------------|
| `app/models/` | ActiveRecord patterns, validations, scopes |
| `app/controllers/` | Strong params, before_actions, REST conventions |
| `app/views/` | Partials, helpers, Turbo/Hotwire if detected |
| `spec/` or `test/` | RSpec/Minitest patterns, FactoryBot |
| `db/migrations/` | Migration safety, never modify existing |

#### Generic (Unknown Framework)

For unrecognized frameworks, generate CLAUDE.md only for directories with:
- Complex patterns (5+ related files)
- Test directories
- Configuration directories
- API / integration directories

## Step 7: Audit Mode (Existing CLAUDE.md Files)

Run: `echo "[context:7] auditing existing CLAUDE.md files"`

If CLAUDE.md files already exist, run in audit mode:

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
- Incorrect testing patterns

### 7c. Report and Suggest

Use `AskUserQuestion` to present findings:
- List each stale/incorrect reference
- Suggest specific updates
- Let user approve or reject each change
- Do NOT overwrite without explicit confirmation

### 7d. Generate Missing Files

If new directories exist that should have CLAUDE.md but don't:
- Generate them following Step 6 rules
- Report them separately from audit findings

## Step 8: Audit README.md — MANDATORY

Run: `echo "[context:8] auditing README.md"`

Cross-reference README.md against what was learned during the scan:
- Framework/language versions — match detected versions
- Listed integrations — still present in dependencies?
- Directory structure tree — matches actual structure?
- Commands/scripts — still valid?
- Missing sections for major patterns discovered

Use `AskUserQuestion` to offer README.md updates if staleness is found.

## Step 9: Report

Run: `echo "[context:9] generation complete"`

Show summary:
- **Stack detected**: language, framework, test runner, assets, DB
- **Context7**: whether it was used, which docs were queried
- **Created**: list of new CLAUDE.md files generated
- **Updated**: list of existing CLAUDE.md files that were updated (audit mode)
- **Unchanged**: list of CLAUDE.md files that are still current
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
