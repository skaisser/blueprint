---
id: 0002
title: CLI enhancements and script extraction from skill audit
status: backlog
complexity: S
tags: [cli, go, scripts, performance, deduplication]
---

# CLI Enhancements & Script Extraction

Consolidated from parallel audit of all 27 skills. These are bash patterns that should be extracted to the Go CLI or standalone scripts.

## Priority 1 — High Impact CLI Commands (eliminates duplication + prevents bugs)

- [ ] `blueprint config get <key>` — read single value from `blueprint/.config.yml`
  - **Reuse:** 10+ skills (pr, finish, hotfix, bp-ship, bp-push, bp-branch, bp-context, bp-status, flow-auto, flow-auto-wt, batch-flow)
  - **Replaces:** `grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}'`
  - **Why:** most-repeated pattern in entire codebase, grep+awk is fragile vs Go YAML parsing

- [ ] `blueprint base-branch` — resolve correct PR target (staging if exists, else main)
  - **Reuse:** 5+ skills (pr, finish, flow-auto, flow-auto-wt, batch-flow)
  - **Replaces:** 8-line inline bash with config read + git ref check + fallback
  - **Why:** critical branch targeting — getting it wrong means PRs target wrong branch

- [ ] `blueprint plan-tasks [--baseline COMMIT]` — extract task lines, optionally diff vs baseline
  - **Reuse:** plan-check
  - **Replaces:** 12-line bash (find review commit, extract tasks from 2 versions, diff)
  - **Why:** deleted task detection is a hard failure mode, Go does this without temp files

- [ ] `blueprint branch safety` — check if current branch allows push
  - **Reuse:** bp-ship, bp-push
  - **Replaces:** duplicated 7-10 line safety check
  - **Why:** safety-critical, duplicated with slight variations

## Priority 2 — Speed Up LLM Operations

- [ ] `blueprint validate <plan-file>` — automated plan completeness check
  - **Reuse:** plan-review, plan-approved
  - **Replaces:** LLM manually checking for markers, test tasks, stale refs
  - **Why:** already identified as candidate in plan-review itself

- [ ] `blueprint plan-profile <plan-file>` — Quick Plan Profile metrics as JSON
  - **Reuse:** plan-review (Step 4)
  - **Replaces:** LLM counting tasks, phases, complexity markers
  - **Returns:** `{"open_tasks": 12, "phases": 4, "max_complexity": "S", "h": 8, "s": 3, "o": 1}`

- [ ] `blueprint plan-status` — plan phase/task completion as JSON
  - **Reuse:** plan-approved, plan-check, flow
  - **Replaces:** LLM counting [x] vs [ ] manually
  - **Returns:** `{"total": 20, "done": 15, "phases": 4, "phases_done": 3, "status": "in-progress"}`

- [ ] `blueprint detect-stack` — auto-detect language, framework, test runner, assets, DB
  - **Reuse:** start, bp-context, plan-review
  - **Replaces:** 15-100 lines of file-existence checks and lock file parsing
  - **Returns:** `{"language": "php", "framework": "laravel", "version": "12.x", "test_runner": "pest"}`

## Priority 3 — Complex Operations (error-prone in bash)

- [ ] `blueprint worktree create <branch> [--base <ref>]` — create lightweight worktree
  - **Reuse:** flow-auto-wt
  - **Replaces:** 15+ lines (compute path, fetch, worktree add, verify)
  - **Why:** most complex git operation, fragile in bash

- [ ] `blueprint worktree remove <path>` — clean up worktree safely
  - **Reuse:** flow-auto-wt, complete
  - **Replaces:** 5-6 lines (cd, remove --force, prune)

- [ ] `blueprint merge-chain <branch> [--staging <branch>]` — feat→staging→main merge
  - **Reuse:** flow-auto, batch-flow
  - **Replaces:** 15+ lines (gh pr create, merge, error handling)
  - **Why:** most error-prone bash in the system

- [ ] `blueprint branch-name <description>` — generate sanitized branch name
  - **Reuse:** flow-auto, flow-auto-wt
  - **Replaces:** `tr + sed + head -c 50` chain

- [ ] `blueprint commit format <type> <message>` — format with emoji + validate
  - **Reuse:** bp-commit, bp-ship, bp-tdd-review
  - **Replaces:** emoji-type lookup table (14 entries)

- [ ] `blueprint review-poll <pr-number> [--max-checks N]` — poll for review completion
  - **Reuse:** flow-auto, flow-auto-wt
  - **Replaces:** 8-line for loop with sleep + gh api

## Script Candidates (standalone .sh files)

- [ ] `scripts/ci-trigger-wait.sh <PR_NUMBER> [workflow]` — trigger @tests and wait
  - **Used in:** finish
  - **Why:** multi-step (gh api check, comment, sleep, poll), needs retry logic

- [ ] `scripts/pre-merge-check.sh <BASE_BRANCH>` — conflict detection + auto-resolve blueprint/
  - **Used in:** finish
  - **Why:** complex git merge + conditional conflict resolution

- [ ] `scripts/plan-check-orphans.sh <COMMIT> <patterns...>` — detect orphaned test refs
  - **Used in:** plan-check
  - **Why:** semi-contextual (LLM identifies patterns, script does mechanical grep)

- [ ] `scripts/worktree-install-deps.sh` — install deps in worktree based on detected stack
  - **Used in:** flow-auto-wt
  - **Why:** multi-tool (composer, npm, vite) with error suppression

## Bug Fix — `blueprint meta` Path

- [ ] Fix `blueprint meta` to search `blueprint/live/` first, then fall back to `blueprint/`
  - **Impact:** eliminates fallback code in finish, plan-review, plan-check, resume
  - **Current:** searches old flat `blueprint/` layout, misses nested `blueprint/live/`

## GitHub Issue Integration (added from audit)

- [ ] `blueprint issue fetch <number>` — fetch GitHub issue as JSON (title, body, labels, state)
  - **Used in:** plan, quick, hotfix
  - **Replaces:** `gh issue view N --json ...`
  - **Why CLI:** standardizes issue fetching, could cache, adds error handling

- [ ] `blueprint issue close <number>` — close issue with verification
  - **Used in:** finish
  - **Replaces:** `gh issue view + gh issue close` sequence

- [ ] `blueprint pr-body [--issue N] [--plan FILE]` — generate PR body with issue refs
  - **Used in:** pr
  - **Replaces:** manual "Closes #N" insertion
  - **Why CLI:** deterministic template, reads plan frontmatter for issue/backlog refs

## Metrics

| Category | Count | Impact |
|----------|-------|--------|
| CLI commands (Priority 1) | 4 | Eliminates 10+ duplicated patterns, prevents bugs |
| CLI commands (Priority 2) | 4 | Speeds up LLM operations, reduces counting errors |
| CLI commands (Priority 3) | 6 | Replaces complex/error-prone bash |
| GitHub Issue Integration | 3 | Standardizes issue fetch/close/PR-body across skills |
| Scripts | 4 | Encapsulates multi-tool sequences |
| Bug fix | 1 | Eliminates fallback code in 4 skills |
| **Total** | **22** | |
