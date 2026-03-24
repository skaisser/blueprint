---
name: quick
description: >
  Quick task execution without a full plan — just do it and commit.
  Use this skill whenever the user says "/quick", "quick fix", "just do it",
  "small fix", or any request for a small change that doesn't need planning.
  Also triggers on "quick task", "small change", "tweak this", "simple fix",
  "rename this", "fix this typo", "update the config", "one-liner",
  "trivial change", "swap this", "change this value", "toggle this",
  or any 1-3 file change that doesn't warrant a full /plan workflow.
  Supports optional Linear integration for issue tracking.
---

# Quick: Fast Task Execution

Quick task execution. No plan, no team. Just do it and commit.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

For small fixes, tweaks, and tasks that don't need a full planning workflow.

## Process

1. **Understand the task** from $ARGUMENTS
2. **Clarify if ambiguous** — AskUserQuestion ONLY if genuinely unclear
3. **Read relevant files**
4. **Confirm approach** — Brief one-liner (mention if referencing past plan)
5. **Implement** — Make the minimal change requested. Do not refactor, clean up, or improve surrounding code
6. **Verify** — Check the change is correct: no syntax errors, no broken references, run targeted tests if testable logic changed
7. **Commit** — Use `/commit`. Commit message must describe what changed and why, not just which file was touched
8. **Linear update** — Only if issue ID provided: push and update status
9. **Ask next** — AskUserQuestion:
   - "Yes, we're done"
   - "Create a PR for this" — Run /pr
   - "More changes needed" — Continue on branch

## Rules

- No plan file, no team, no worktree, no subagents
- No sequential thinking (keep it fast)
- Stay on current branch unless user requests a new branch
- Ask ONLY if genuinely ambiguous
- Always confirm approach in one line before changing
- Make only the change requested — do not refactor, clean up, or improve surrounding code
- Follow existing codebase patterns and conventions
- Run targeted tests — full suite: ask user to run separately
- If changes would span more than 3 files, suggest `/plan` instead — quick is for small scope
- Use `/ship` to commit and push

Use $ARGUMENTS as the task description.
