---
id: "0001"
title: "refactor: migrate skills to blueprint CLI commands"
type: refactor
status: in-progress
complexity: S
project: blueprint
branch: refactor/skill-cli-migration
base: main
strategy: parallel-subagents
tags: [cli, skills, deduplication, v1.0]
backlog: "0001"
issue: null
created: "2026-03-24 05:22"
completed: null
pr: null
session: "d4c2df65-571d-4494-aa3a-68673c5220af"
reviews:
  - "Plan Created — 2026-03-24 05:22"
  - "2026-03-24 05:26 — Added meta rootCmd alias task (meta is under sdlcCmd, skills call blueprint meta without sdlc prefix)"
  - "2026-03-24 05:26 — Marked complexity: H:12, S:6, O:0"
  - "2026-03-24 05:26 — Phase 1+2 parallel (cli/ vs skills/), Phase 3+4+5 parallel after"
---

# refactor: Migrate skills to blueprint CLI commands

## Goal

Replace raw bash patterns (hardcoded paths, grep+awk, inline Python, manual branch checks) across 13+ skills with the `blueprint` CLI commands built in plan 0002. Skills should call `blueprint <command>` directly — the installer adds `~/.blueprint/bin` to PATH.

## Non-Goals

- Rewriting skill logic or restructuring skill flow
- Quality audit against skill-creator patterns (separate backlog item)
- Touching plan or plan-review skill thoroughness — those stay as-is
- Python parity in hooks/blueprint.py (separate backlog item)

## Context

- Plan 0002 built 17 CLI commands: config get, base-branch, branch safety, branch-name, meta, time, plan-tasks, plan-profile, plan-status, validate, issue fetch/close, pr-body, worktree create/remove, merge-chain, commit format, review-poll, detect-stack
- Audit found 7 priority patterns across 13+ skills (see backlog 0001 for full list)
- Skills install to `~/.claude/skills/`, CLI binary to `~/.blueprint/bin/blueprint`
- `blueprint meta` returns JSON — needs `blueprint meta <field>` for single-field extraction without jq/python
- `skills/plan/references/plan-template.md` line 153 still references hardcoded path
- **IMPORTANT:** `metaCmd` is registered under `sdlcCmd` (so actual path is `blueprint sdlc meta`). Skills call `blueprint meta` — Phase 1 must add a rootCmd alias.

## Tech Stack Versions

- Go: 1.24.11 (CLI source in cli/)
- cobra: CLI framework
- yaml.v3: YAML parsing
- Skills: Markdown (SKILL.md prompt files)

## Phases

### Phase 1: CLI enhancement — `blueprint meta <field>` + rootCmd alias [S]

**Touches:** `cli/cmd/sdlc.go`, `cli/cmd/root.go`, `cli/internal/plan/meta.go`

- [ ] [S] Register `metaCmd` on `rootCmd` as alias so `blueprint meta` works without `sdlc` prefix
- [ ] [S] Add optional positional arg to `blueprint meta` — `blueprint meta plan_file` returns just the value, no JSON wrapper
- [ ] [H] Support all MetaResult fields: next_num, base_branch, branch, plan_file, plan_num, status, progress, project, git_remote, today
- [ ] [H] Tests: meta single-field extraction + rootCmd alias in meta_test.go

**Verify:** `cd cli && go test ./internal/plan/ ./cmd/ -run "Meta" -v`

### Phase 2: Hardcoded path replacement [H]

**Touches:** `skills/` (13 skill SKILL.md files), `skills/plan/references/plan-template.md`

- [ ] [H] Replace all `~/.blueprint/bin/blueprint` with `blueprint` across 13 skills: address-pr, bp-status, finish, flow, flow-auto, flow-auto-wt, hotfix, plan, plan-approved, plan-check, plan-review, pr, resume
- [ ] [H] Fix plan-template.md line 153: `~/.blueprint/bin/blueprint meta` → `blueprint meta`
- [ ] [H] Verify zero hardcoded `~/.blueprint/bin/` remaining via grep

**Verify:** `grep -r "~/.blueprint/bin/" skills/ && echo "FAIL" || echo "PASS"`

### Phase 3: Branch operations migration [H]

**Touches:** `skills/flow-auto/SKILL.md`, `skills/flow-auto-wt/SKILL.md`, `skills/bp-push/SKILL.md`, `skills/bp-ship/SKILL.md`

- [ ] [H] flow-auto: replace `tr ' ' '-' | tr '[:upper:]' '[:lower:]' | head -c 50` with `blueprint branch-name "$DESCRIPTION"`
- [ ] [H] flow-auto-wt: same branch-name replacement
- [ ] [H] bp-push: replace manual `[ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]` with `blueprint branch safety`
- [ ] [H] bp-ship: same branch safety replacement

**Verify:** `grep -n "tr.*upper.*lower\|= \"main\"\|= \"master\"" skills/flow-auto/SKILL.md skills/flow-auto-wt/SKILL.md skills/bp-push/SKILL.md skills/bp-ship/SKILL.md && echo "FAIL" || echo "PASS"`

### Phase 4: Meta + frontmatter migration [S]

**Touches:** `skills/finish/SKILL.md`, `skills/hotfix/SKILL.md`, `skills/pr/SKILL.md`, `skills/plan-review/SKILL.md`

- [ ] [S] finish: replace `blueprint meta | python3 -c "import sys,json; ..."` with `blueprint meta plan_file`
- [ ] [S] finish: replace `grep '^issue:' "$PLAN_FILE" | sed` with `blueprint meta issue`
- [ ] [H] hotfix: replace inline Python meta parsing with `blueprint meta base_branch`
- [ ] [H] pr: replace `grep '^issue:' "$PLAN_FILE" | sed` with `blueprint meta issue`
- [ ] [S] plan-review: replace `cat blueprint/.config.yml | grep -A5 'stack:'` with `blueprint detect-stack`

**Verify:** `grep -n "python3 -c.*json\|grep.*issue.*sed\|cat.*config.yml.*grep" skills/finish/SKILL.md skills/hotfix/SKILL.md skills/pr/SKILL.md skills/plan-review/SKILL.md && echo "FAIL" || echo "PASS"`

### Phase 5: Plan-check modernization [S]

**Touches:** `skills/plan-check/SKILL.md`

- [ ] [S] Replace `git log --oneline | grep "plan: review" | head -1 | awk '{print $1}'` with `blueprint plan-tasks --baseline` for task comparison
- [ ] [H] Replace raw `grep -E "^- \[[ x]\]"` task counting with `blueprint plan-status`
- [ ] [S] Replace `git show "$COMMIT:$FILE" | grep` task diffing with `blueprint plan-tasks --baseline $COMMIT`

**Verify:** `grep -n "git log.*grep.*plan.*awk\|git show.*grep.*\\\\[" skills/plan-check/SKILL.md && echo "FAIL" || echo "PASS"`

## Acceptance

- [ ] Zero `~/.blueprint/bin/` hardcoded paths remaining in skills/
- [ ] Zero inline Python JSON parsing remaining in skills/
- [ ] Zero manual branch sanitization (tr+sed) remaining in skills/
- [ ] Zero manual main/master safety checks remaining in skills/
- [ ] `blueprint meta plan_file` returns raw value without JSON wrapper
- [ ] `blueprint meta` works without `sdlc` prefix
- [ ] All existing Go tests pass (`go test ./...`)
- [ ] make build-all produces 3 binaries

## Execution Strategy

> **Approach:** `/plan-approved` with Parallel Subagents (Mode A)
> **Total Tasks:** 19 (H: 12, S: 7, O: 0)
> **Estimated Rounds:** 3 (1 parallel foundation, 1 parallel migration, 1 verification)

### File-Touch Matrix

| Phase | Files/Dirs Touched | Depends On |
|-------|-------------------|------------|
| 1 | `cli/cmd/sdlc.go`, `cli/cmd/root.go`, `cli/internal/plan/meta.go` | — |
| 2 | `skills/` (13 SKILL.md files), `plan-template.md` | — |
| 3 | `skills/flow-auto/`, `skills/flow-auto-wt/`, `skills/bp-push/`, `skills/bp-ship/` | Phase 2 |
| 4 | `skills/finish/`, `skills/hotfix/`, `skills/pr/`, `skills/plan-review/` | Phase 1 + Phase 2 |
| 5 | `skills/plan-check/` | Phase 2 |

**Parallelism:** Phase 1 (cli/) and Phase 2 (skills/) have zero file overlap → parallel. Phase 3+4+5 touch different skill files but all depend on Phase 2 completing first.

### Round 1: Phase 1 + Phase 2 → Parallel (2 workers)

| Phase | Mode | Model | Tasks | Notes |
|-------|------|-------|-------|-------|
| Phase 1: CLI meta enhancement | Subagent | Opus | 4 (2x[S] + 2x[H]) | Go code changes |
| Phase 2: Hardcoded paths | Subagent | Sonnet | 3 (all [H]) | Global string replace |

### Round 2: Phase 3 + Phase 4 + Phase 5 → Parallel (3 workers)

| Phase | Mode | Model | Tasks | Notes |
|-------|------|-------|-------|-------|
| Phase 3: Branch ops | Subagent | Sonnet | 4 (all [H]) | Simple replacements |
| Phase 4: Meta migration | Subagent | Opus | 5 (3x[S] + 2x[H]) | Logic changes in finish/hotfix/pr |
| Phase 5: Plan-check | Subagent | Opus | 3 (2x[S] + 1x[H]) | Most complex rewrites |

### Round 3: Integration verification → Leader Direct

3 tasks: `make build-all`, `go test ./...`, verify grep checks pass.
