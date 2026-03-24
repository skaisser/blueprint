---
name: bp:run-tests
description: >
  Run targeted tests, handle failures, and report results.
  Triggers on "/bp:run-tests", "/run-tests", "run tests", "test this",
  "run the tests", or any request to execute existing tests.
  Also triggers on "check tests", "are tests passing", "test suite", or "verify tests".
  Agent runs targeted tests only — full suite is always run by the user in a separate terminal.
---

# Run Tests: Execute and Report

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Run targeted tests and report results.

## Critical Rules

- **Agent runs targeted tests only** — specific file or filter
- **Full suite: ask user to run it in a separate terminal** — never run the full suite from an agent
- **Full coverage: ask user to run it in a separate terminal** — never run full coverage from an agent
- **NEVER mock what you can test** — if a fix involves adding a mock, reconsider

## Running Tests

Use the project's test runner as detected from the project configuration. Run targeted tests only:

```bash
# Targeted (agent runs these) — adapt commands to project's test runner
<test-runner> --filter="TestName"
<test-runner> tests/Feature/SomeTest.php
<test-runner> tests/Feature/SomeTest.php --filter="specific test name"

# Targeted coverage (agent can run these)
<coverage-tool> --filter="TestName"

# Full suite (USER runs in separate terminal)
# Tell user to run the full suite in their terminal

# Full coverage (USER runs in separate terminal)
# Tell user to run full coverage in their terminal
```

## On Failures

1. Read the error message carefully
2. Check common flaky patterns:
   - Random factory values → use explicit values for business-logic fields
   - Random/short generated values too short for validation → hardcode values
   - Random IDs causing skips → set explicit values
   - Missing seed data → add to setup
3. **Fix the root cause** — never skip tests, never add mocks to make it pass
4. Re-run the targeted test to confirm fix
5. Ask user to run full suite to verify nothing else broke

## Reporting

After running, show a concise summary (do NOT dump raw test output):
- Pass/fail count
- Failures with file:line and error message
- Suggested fix for each failure
- If all pass, use `AskUserQuestion`: "All N tests pass. Run the full suite?" with options: "Yes, I'll run it" / "No, we're good"

## Interaction Rules

- All user interactions MUST use `AskUserQuestion` tool, never plain text questions

Use $ARGUMENTS as test filter, file path, or class name. If no arguments provided, detect recently changed files with `git diff --name-only HEAD` and run tests for those files. If no changes detected either, use `AskUserQuestion` to ask what to test.
