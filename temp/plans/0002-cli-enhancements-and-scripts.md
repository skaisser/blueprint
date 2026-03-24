---
id: 0002
title: CLI enhancements and script extraction from skill audit
status: in-progress
session: "d4c2df65-571d-4494-aa3a-68673c5220af"
complexity: S
branch: feat/cli-enhancements
strategy: mixed-dispatch
tags: [cli, go, scripts, performance, deduplication, v1.0]
reviews:
  - "Restructured flat list into 6 executable phases with dependencies"
  - "Added test tasks for every CLI command and script"
  - "Added acceptance criteria"
  - "Added frontmatter validate/fix from final-review-recommendations.md"
  - "Phase 2+3+4+6 can run in parallel (zero file overlap)"
---

# CLI Enhancements & Script Extraction

Consolidated from parallel audit of all 27 skills. Bash patterns extracted to Go CLI or standalone scripts.

## Goal

Eliminate duplicated bash patterns across skills by implementing them as `blueprint` CLI subcommands (Go, compiled, fast) or standalone scripts (.sh, reusable). Every CLI command returns structured JSON for easy consumption by skills.

**CRITICAL: Python parity requirement.** Every CLI command implemented in Go MUST have an equivalent in `hooks/blueprint.py`. The Python file is the fallback when the Go binary isn't installed. Commands use `blueprint <command>` (no `sdlc` subcommand) — e.g., `blueprint backlog`, `blueprint config get`, `blueprint base-branch`.

## Acceptance Criteria

- [ ] All 17 CLI commands compile and pass Go tests
- [ ] All 4 scripts are executable and tested
- [ ] `blueprint meta` searches `blueprint/live/` first (bug fix)
- [ ] `make build-all` produces 3 binaries
- [ ] `go test ./...` passes all tests (existing + new)
- [ ] Zero `grep 'staging_branch:' blueprint/.config.yml | awk` remaining in skills (replaced by `blueprint config get`)

## Tech Stack Versions

- Go: 1.22+ (CLI source in cli/)
- cobra: CLI framework (already in go.mod)
- yaml.v3: YAML parsing (already in go.mod)
- Shell: bash 5+ (scripts)

---

## Phase 1: Foundation — meta fix + config command [S]

**Touches:** `cli/internal/plan/meta.go`, `cli/cmd/sdlc.go`

- [ ] Fix `blueprint meta` to search `blueprint/live/` first, fall back to `blueprint/` [S]
  - Update the glob pattern in meta.go: `blueprint/live/[0-9]*-*.md` before `blueprint/[0-9]*-*.md`
  - Impact: eliminates fallback code in finish, plan-review, plan-check, resume
- [ ] Implement `blueprint config get <key>` — read single YAML value from `blueprint/.config.yml` [S]
  - Parse YAML properly (not grep+awk), handle nested keys with dot notation
  - `blueprint config get staging_branch` → `staging`
  - `blueprint config get language` → `auto`
  - `blueprint config get stack.language` → `php`
  - Returns raw value on stdout, exit 1 if key missing
- [ ] Implement `blueprint frontmatter validate <file>` — check frontmatter against schema [S]
  - Detect file type from location (backlog/, live/, upstream/, expired/)
  - Report missing fields, wrong types, status/location mismatches
  - JSON output: `{"valid": false, "errors": ["missing field: issue", "status 'backlog' but file is in live/"]}`
- [ ] Implement `blueprint frontmatter fix <file> [--dry-run]` — auto-correct frontmatter [S]
  - Add missing fields with defaults, normalize dates, fix status/location mismatches
  - `--dry-run` reports what would change without writing
- [ ] Tests: `cli/internal/plan/meta_test.go` — test live/ first search [H]
- [ ] Tests: `cli/cmd/config_test.go` — test config get with nested keys, missing file, missing key [S]
- [ ] Tests: `cli/internal/plan/frontmatter_test.go` — validate + fix with all file types [S]

**Verify:** `cd cli && go test ./internal/plan/ ./cmd/ -run "Config|Meta|Frontmatter"`

---

## Phase 2: Branch operations [S]

**Touches:** `cli/internal/git/repo.go`, `cli/cmd/sdlc.go`

- [ ] Implement `blueprint base-branch` — resolve PR target (staging if exists, else main) [S]
  - Read staging_branch from config, check local + remote refs, return branch name
  - JSON: `{"base": "staging", "has_staging": true, "current": "feat/my-feature"}`
- [ ] Implement `blueprint branch safety` — check if current branch allows push [H]
  - Blocks main/master, returns exit 0 (safe) or exit 1 (blocked) with message
- [ ] Implement `blueprint branch-name <description>` — generate sanitized branch name [H]
  - `blueprint branch-name "Add user auth"` → `feat/add-user-auth`
  - Sanitize: lowercase, replace spaces with hyphens, remove special chars, truncate to 50
- [ ] Tests: `cli/internal/git/repo_test.go` — test base-branch with/without staging, branch safety, branch-name edge cases [S]

**Verify:** `cd cli && go test ./internal/git/ -run "BaseBranch|BranchSafety|BranchName"`

---

## Phase 3: Plan operations [S]

**Touches:** `cli/internal/plan/tasks.go` (new), `cli/internal/plan/validate.go` (new), `cli/cmd/sdlc.go`

- [ ] Implement `blueprint plan-tasks [--baseline COMMIT]` — extract tasks, diff vs baseline [S]
  - Parse `- [ ]` and `- [x]` lines from plan file
  - `--baseline COMMIT`: compare current tasks vs tasks at that commit, report added/removed/changed
  - JSON: `{"tasks": [...], "added": [...], "removed": [...]}`
- [ ] Implement `blueprint plan-profile <plan-file>` — Quick Plan Profile as JSON [H]
  - Count tasks, phases, complexity markers
  - JSON: `{"open_tasks": 12, "phases": 4, "max_complexity": "S", "h": 8, "s": 3, "o": 1}`
- [ ] Implement `blueprint plan-status` — plan completion status [H]
  - JSON: `{"total": 20, "done": 15, "phases": 4, "phases_done": 3, "status": "in-progress"}`
- [ ] Implement `blueprint validate <plan-file>` — full plan completeness check [S]
  - Check: all tasks have [H]/[S]/[O], Execution Strategy section exists, acceptance criteria present
  - Check: referenced files exist (Glob), no stale paths
  - JSON: `{"valid": true, "warnings": [], "errors": []}`
- [ ] Tests: `cli/internal/plan/tasks_test.go` — task parsing, baseline diff, edge cases [S]
- [ ] Tests: `cli/internal/plan/validate_test.go` — plan validation with valid/invalid fixtures [S]

**Verify:** `cd cli && go test ./internal/plan/ -run "Tasks|Profile|Status|Validate"`

---

## Phase 4: GitHub Issue integration [S]

**Touches:** `cli/cmd/issue.go` (new), `cli/internal/github/client.go`

- [ ] Implement `blueprint issue fetch <number>` — fetch GH issue as JSON [S]
  - Calls `gh issue view N --json title,body,labels,assignees,state`
  - Returns parsed JSON, handles errors (not found, no gh, no auth)
- [ ] Implement `blueprint issue close <number>` — close with verification [H]
  - Check state first, close if open, confirm closed
- [ ] Implement `blueprint pr-body [--issue N] [--plan FILE]` — generate PR body template [S]
  - Reads plan frontmatter for issue/backlog refs
  - Generates markdown with "Closes #N" refs, summary from plan, changes section
- [ ] Tests: `cli/cmd/issue_test.go` — mock gh responses, test fetch/close/pr-body [S]

**Verify:** `cd cli && go test ./cmd/ -run "Issue|PrBody"`

---

## Phase 5: Complex operations [S]

**Touches:** `cli/cmd/worktree.go` (new), `cli/cmd/merge.go` (new), `cli/internal/audit/rules.go`

- [ ] Implement `blueprint worktree create <branch> [--base <ref>]` — create lightweight worktree [S]
  - Compute path, fetch, git worktree add, verify, return path
  - JSON: `{"path": "/Users/.../project2", "branch": "feat/x", "base": "staging"}`
- [ ] Implement `blueprint worktree remove <path>` — clean up safely [H]
  - Safety checks, git worktree remove --force, prune
- [ ] Implement `blueprint merge-chain <branch> [--staging <branch>]` — feat→staging→main [S]
  - Full PR create + merge chain with error handling
  - JSON: `{"staging_pr": 42, "main_pr": 43, "merged": true}`
- [ ] Implement `blueprint commit format <type> <message>` — format with emoji [H]
  - Lookup emoji for type, validate format, return formatted message
  - `blueprint commit format feat "add auth"` → `✨ feat: add auth`
- [ ] Implement `blueprint review-poll <pr-number> [--timeout 20m]` — poll for review [S]
  - Check for review comments, bot responses, timeout handling
- [ ] Implement `blueprint detect-stack` — auto-detect project stack [S]
  - Check composer.json, package.json, go.mod, Gemfile, requirements.txt
  - Parse lock files for framework versions
  - JSON: `{"language": "php", "framework": "laravel", "version": "12.x", "test_runner": "pest"}`
- [ ] Tests: `cli/cmd/worktree_test.go` — create/remove with mock git [S]
- [ ] Tests: `cli/cmd/merge_test.go` — merge-chain with mock gh [S]
- [ ] Tests: `cli/internal/audit/commit_format_test.go` — all emoji types, validation [H]

**Verify:** `cd cli && go test ./cmd/ ./internal/audit/ -run "Worktree|Merge|CommitFormat|DetectStack|ReviewPoll"`

---

## Phase 6: Standalone scripts [H]

**Touches:** `scripts/` (new directory)

- [ ] `scripts/ci-trigger-wait.sh <PR_NUMBER> [workflow]` — trigger @tests and wait [H]
  - Check workflow exists, post comment, poll for run, wait for completion
  - Exit 0 on pass, exit 1 on fail, timeout after 10 min
- [ ] `scripts/pre-merge-check.sh <BASE_BRANCH>` — conflict detection + auto-resolve blueprint/ [H]
  - Fetch base, check behind count, merge, auto-resolve blueprint/ conflicts
- [ ] `scripts/plan-check-orphans.sh <COMMIT> <patterns...>` — detect orphaned test refs [H]
  - Accept patterns as args, git grep for matches in tests/, report findings
- [ ] `scripts/worktree-install-deps.sh [path]` — install deps based on detected stack [H]
  - Check composer.json → composer install, package.json → npm ci, etc.
- [ ] All scripts: `chmod +x`, add shebang, basic error handling [H]
- [ ] Tests: `tests/scripts/test_scripts.sh` — run each script with mock data [H]

**Verify:** `bash tests/scripts/test_scripts.sh`

---

## Execution Strategy

> **Approach:** `/plan-approved` with Mixed Dispatch (Mode D)
> **Total Tasks:** 34 (H: 16, S: 18, O: 0)
> **Estimated Rounds:** 4 (1 sequential foundation, 1 parallel burst, 1 sequential complex, 1 verification)

### File-Touch Matrix

| Phase | Files/Dirs Touched | Depends On |
|-------|-------------------|------------|
| 1 | `cli/internal/plan/meta.go`, `cli/cmd/sdlc.go` | — |
| 2 | `cli/internal/git/repo.go`, `cli/cmd/sdlc.go` | Phase 1 (config get) |
| 3 | `cli/internal/plan/tasks.go`, `cli/internal/plan/validate.go`, `cli/cmd/sdlc.go` | Phase 1 (meta fix) |
| 4 | `cli/cmd/issue.go` (new), `cli/internal/github/client.go` | — |
| 5 | `cli/cmd/worktree.go`, `cli/cmd/merge.go`, `cli/internal/audit/` | Phase 2 (base-branch) |
| 6 | `scripts/*.sh` (new) | — |

**Parallelism:** Phases 3, 4, 6 have zero file overlap with each other AND with Phase 2. Phase 2 shares `cli/cmd/sdlc.go` with Phase 3 — but they add different subcommands (no conflict if workers add, not modify). Safe to run in parallel.

### Round 1: Phase 1 → Single Subagent (foundation — everything depends on config get + meta fix)

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 1: Foundation | Opus | 7 (4x[S] + 3x[H]) | Config get + meta fix + frontmatter validate/fix |

### Round 2: Phase 2 + Phase 3 + Phase 4 + Phase 6 → Parallel (4 workers)

| Phase | Mode | Model | Tasks | Notes |
|-------|------|-------|-------|-------|
| Phase 2: Branch ops | Subagent | Opus | 4 (1x[S] + 3x[H]) | Depends on Phase 1 config get |
| Phase 3: Plan ops | Subagent | Opus | 6 (3x[S] + 3x[H]) | Depends on Phase 1 meta fix |
| Phase 4: GH Issue | Subagent | Opus | 4 (2x[S] + 2x[H]) | Independent |
| Phase 6: Scripts | Subagent | Sonnet | 6 (all [H]) | Independent, no Go |

### Round 3: Phase 5 → Single Subagent (depends on Phase 2)

| Phase | Model | Tasks | Notes |
|-------|-------|-------|-------|
| Phase 5: Complex ops | Opus | 9 (4x[S] + 5x[H]) | Worktree, merge-chain, detect-stack |

### Round 4: Integration verification → Leader Direct

3 tasks: `make build-all`, `go test ./...`, verify scripts executable.
