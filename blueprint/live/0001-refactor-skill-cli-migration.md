---
id: "0001"
title: "refactor: migrate skills to blueprint CLI commands"
type: refactor
status: todo
project: blueprint
branch: refactor/skill-cli-migration
base: main
tags: [cli, skills, deduplication, v1.0]
backlog: "0001"
issue: null
created: "2026-03-24 05:22"
completed: null
pr: null
session: null
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

## Phases

### Phase 1: CLI enhancement — `blueprint meta <field>`

**Touches:** `cli/cmd/sdlc.go`, `cli/internal/plan/meta.go`

- [ ] Add optional positional arg to `blueprint meta` — `blueprint meta plan_file` returns just the value, no JSON wrapper
- [ ] Support all MetaResult fields: next_num, base_branch, branch, plan_file, plan_num, status, progress, project, git_remote, today
- [ ] Tests: meta single-field extraction in meta_test.go

**Verify:** `cd cli && go test ./internal/plan/ -run "Meta" -v`

### Phase 2: Hardcoded path replacement

**Touches:** `skills/` (13 skill SKILL.md files), `skills/plan/references/plan-template.md`

- [ ] Replace all `~/.blueprint/bin/blueprint` with `blueprint` across 13 skills: address-pr, bp-status, finish, flow, flow-auto, flow-auto-wt, hotfix, plan, plan-approved, plan-check, plan-review, pr, resume
- [ ] Fix plan-template.md line 153: `~/.blueprint/bin/blueprint meta` → `blueprint meta`
- [ ] Verify zero hardcoded `~/.blueprint/bin/` remaining via grep

**Verify:** `grep -r "~/.blueprint/bin/" skills/ && echo "FAIL" || echo "PASS"`

### Phase 3: Branch operations migration

**Touches:** `skills/flow-auto/SKILL.md`, `skills/flow-auto-wt/SKILL.md`, `skills/bp-push/SKILL.md`, `skills/bp-ship/SKILL.md`

- [ ] flow-auto: replace `tr ' ' '-' | tr '[:upper:]' '[:lower:]' | head -c 50` with `blueprint branch-name "$DESCRIPTION"`
- [ ] flow-auto-wt: same branch-name replacement
- [ ] bp-push: replace manual `[ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]` with `blueprint branch safety`
- [ ] bp-ship: same branch safety replacement

**Verify:** `grep -n "tr.*upper.*lower\|= \"main\"\|= \"master\"" skills/flow-auto/SKILL.md skills/flow-auto-wt/SKILL.md skills/bp-push/SKILL.md skills/bp-ship/SKILL.md && echo "FAIL" || echo "PASS"`

### Phase 4: Meta + frontmatter migration

**Touches:** `skills/finish/SKILL.md`, `skills/hotfix/SKILL.md`, `skills/pr/SKILL.md`, `skills/plan-review/SKILL.md`

- [ ] finish: replace `blueprint meta | python3 -c "import sys,json; ..."` with `blueprint meta plan_file` (depends on Phase 1)
- [ ] finish: replace `grep '^issue:' "$PLAN_FILE" | sed` with `blueprint meta issue` or `blueprint frontmatter validate`
- [ ] hotfix: replace inline Python meta parsing with `blueprint meta base_branch`
- [ ] pr: replace `grep '^issue:' "$PLAN_FILE" | sed` with `blueprint meta issue`
- [ ] plan-review: replace `cat blueprint/.config.yml | grep -A5 'stack:'` with `blueprint detect-stack`

**Verify:** `grep -n "python3 -c.*json\|grep.*issue.*sed\|cat.*config.yml.*grep" skills/finish/SKILL.md skills/hotfix/SKILL.md skills/pr/SKILL.md skills/plan-review/SKILL.md && echo "FAIL" || echo "PASS"`

### Phase 5: Plan-check modernization

**Touches:** `skills/plan-check/SKILL.md`

- [ ] Replace `git log --oneline | grep "plan: review" | head -1 | awk '{print $1}'` with `blueprint plan-tasks --baseline` for task comparison
- [ ] Replace raw `grep -E "^- \[[ x]\]"` task counting with `blueprint plan-status`
- [ ] Replace `git show "$COMMIT:$FILE" | grep` task diffing with `blueprint plan-tasks --baseline $COMMIT`

**Verify:** `grep -n "git log.*grep.*plan.*awk\|git show.*grep.*\\\\[" skills/plan-check/SKILL.md && echo "FAIL" || echo "PASS"`

## Acceptance

- [ ] Zero `~/.blueprint/bin/` hardcoded paths remaining in skills/
- [ ] Zero inline Python JSON parsing remaining in skills/
- [ ] Zero manual branch sanitization (tr+sed) remaining in skills/
- [ ] Zero manual main/master safety checks remaining in skills/
- [ ] `blueprint meta plan_file` returns raw value without JSON wrapper
- [ ] All existing Go tests pass (`go test ./...`)
- [ ] make build-all produces 3 binaries
