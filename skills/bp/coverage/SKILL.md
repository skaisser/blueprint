---
name: bp:coverage
description: >
  Analyze test coverage gaps and suggest tests to reach 100% coverage.
  Triggers on "/bp:coverage", "/coverage", "check coverage", "test coverage",
  "coverage analysis", or any request to analyze or improve test coverage.
  Also triggers on "coverage gaps", "untested code", "missing tests", or "coverage report".
  Agent runs targeted coverage only — full suite must be run by the user in a separate terminal.
---

# Coverage: Test Coverage Analysis

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Analyze test coverage gaps and create tests to close them.

## Coverage Goals

- **New projects:** 100% coverage — no exceptions, no excuses
- **Older projects:** Improve incrementally — every PR should increase or maintain coverage
- **Every new file must have tests** — no untested code enters the codebase

## Running Coverage

**Agent runs targeted coverage only:**
Use the project's coverage tool with a filter for the specific area being analyzed.

**Full coverage — user runs in separate terminal:**
Tell the user to run full coverage in their terminal and share the results.

**NEVER run the full coverage suite from an agent.**

## After Receiving Results

1. **Identify uncovered files** — sort by lowest coverage first
2. **Identify uncovered methods** — which public methods have no test hitting them?
3. **Identify missing branches** — if/else, switch, try/catch, early returns
4. **Prioritize by risk:**
   - Business logic (payments, orders, auth) → cover first
   - Services and repositories → cover second
   - Controllers and UI components → cover third
   - Helpers and utilities → cover last

## Creating Tests to Close Gaps

Follow the `/bp:test` skill rules:
- **NEVER mock what you can test** — real factories, real database, real implementations
- Explicit factory values for business-logic fields
- Test happy path, validation, authorization, edge cases, error handling

## What "100% Coverage" Actually Means

It's not just line coverage. For each class, verify:
- Every public method is called in at least one test
- Every conditional branch (if/else, ternary, null coalesce) has both paths tested
- Every exception/error path is tested
- Every validation rule is tested (both pass and fail)
- Every authorization check is tested (both allowed and denied)

## Rules

- NEVER run full coverage suite — ask user to run it in a separate terminal
- NEVER mock what you can test
- NEVER suggest skipping tests or lowering coverage targets
- ALWAYS create tests for uncovered code (don't just report gaps)
- ALWAYS aim for 100% on new projects
- ALWAYS prioritize business-critical code first

## Interaction Rules

- All user interactions MUST use `AskUserQuestion` tool, never plain text questions

Use $ARGUMENTS as filter, file path, or class name to analyze.
