---
name: bp-test
description: >
  Create tests following project conventions with proper factories and real implementations.
  Triggers on "/bp-test", "/test", "write tests", "create tests",
  "add tests for", or any request to create new test files.
  Also triggers on "test this", "write a test for", "feature test",
  "unit test", or any request to generate test code.
  Uses the project's test runner and conventions detected from configuration.
---

# Test: Create Tests

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Create tests for: $ARGUMENTS

## Core Philosophy

**NEVER mock what you can test.** Use real implementations, real database, real factories. Mocks hide bugs — real integrations catch them.

Only mock when you literally cannot use the real thing:
- External HTTP APIs (use HTTP faking/stubbing)
- Third-party services with no sandbox (payment gateways in production mode)
- Time-dependent behavior (time travel helpers)

Everything else — models, services, repositories, jobs, events, notifications, components — test with real implementations.

## Coverage Goal

- **New projects:** 100% test coverage is the target. Every public method, every branch, every edge case.
- **Older projects:** Improve coverage incrementally. New code must have tests. Existing untested code gets covered as we touch it.
- **Every PR should increase or maintain coverage** — never decrease it.

## Environment
- **Database:** In-memory or test database as configured by the project
- **Framework:** Detect from `blueprint/.config.yml` or project files (Pest, Jest, pytest, etc.)
- **Create tests** using the project's scaffolding commands if available

## Syntax Rules
- Follow the project's established test conventions
- Write test descriptions in the language detected from `blueprint/.config.yml` → `language`
- Use the simplest, most readable syntax available in the test framework

## Structure

Follow the project's test organization patterns. Tests should be:
- Clearly named describing the behavior being tested
- Focused on one behavior per test
- Using real factories/fixtures with explicit values for business-logic fields

## Factory / Fixture Best Practices

```
// BAD: Random values cause flaky tests
create a record with all defaults

// GOOD: Explicit values for business-logic fields
create a record with specific status, type, and relationships set explicitly
```

## What to Test (Checklist)

For every piece of code, consider:

1. **Happy path** — does it work as intended?
2. **Validation** — does it reject bad input?
3. **Authorization** — does it block unauthorized users?
4. **Edge cases** — null values, empty collections, boundary conditions
5. **Error handling** — what happens when things fail?
6. **Side effects** — does it dispatch jobs, send notifications, fire events?
7. **Database state** — does it create/update/delete the right records?

## Don't Mock These — Test Them For Real

| Instead of mocking... | Do this |
|----------------------|---------|
| Database models | Use factories with real persistence |
| Input validation | Submit real request, assert validation errors |
| UI components | Use component testing with real props |
| Jobs & queues | Use queue faking only to assert dispatch, test job logic directly |
| Events | Use event faking only to assert dispatch, test listener logic directly |
| Notifications | Use notification faking to assert sent, test content directly |
| Services/repositories | Use real instances with real database |

## After Creating

1. **Run targeted test:** Execute the specific test file to verify it works
2. **Full suite:** Ask user to run the full test suite in a separate terminal
3. **Coverage check:** Ask user to run full coverage in a separate terminal

## Rules

- NEVER mock what you can test with real implementations
- NEVER run the full test suite — ask user to run it in a separate terminal
- NEVER run full coverage — ask user to run it in a separate terminal
- ALWAYS set explicit values for business-logic factory fields
- ALWAYS aim for 100% coverage on new projects
- ALWAYS test validation rules, authorization, and edge cases
- All user interactions MUST use `AskUserQuestion` tool, never plain text questions

Use $ARGUMENTS as the code path, class name, or feature to test.
