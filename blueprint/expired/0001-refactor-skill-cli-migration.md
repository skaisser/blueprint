---
id: "0001"
title: "refactor: migrate skills to blueprint CLI commands"
type: refactor
status: planned
priority: high
size: large
project: blueprint
tags: [cli, skills, deduplication, quality]
created: "2026-03-24 05:14"
plan: "0001"
issue: null
depends: null
---

# refactor: Migrate skills to blueprint CLI commands

## What
Replace raw bash patterns (grep, awk, sed, inline Python) across 13+ skills with the blueprint CLI commands built in plan 0002. Also audit skill quality against skill-creator patterns — explain *why* instead of heavy MUSTs, use progressive disclosure.

## Why
Skills currently duplicate logic the CLI already handles — hardcoded binary paths, manual branch sanitization, raw config parsing. This creates maintenance burden and fragile behavior when paths or formats change. Using the CLI makes skills shorter, more reliable, and easier to maintain.

## Context
Plan 0002 built 17 CLI commands. The audit found 7 priority patterns to fix:

1. **Hardcoded binary paths** (13 skills) — `~/.blueprint/bin/blueprint` should be just `blueprint`
   - address-pr, bp-status, finish, flow, flow-auto, flow-auto-wt, hotfix, plan, plan-approved, plan-check, plan-review, pr, resume
2. **Inline Python JSON parsing** (finish, hotfix) — `blueprint meta | python3 -c "..."` should be a direct key arg
3. **Branch name sanitization** (flow-auto, flow-auto-wt) — use `blueprint branch-name`
4. **Manual main/master checks** (bp-push, bp-ship) — use `blueprint branch safety`
5. **Raw frontmatter parsing** (finish, hotfix, pr) — use `blueprint meta` or `blueprint frontmatter`
6. **Plan-check task diffing** (plan-check) — use `blueprint plan-tasks --baseline`
7. **Config cat+grep** (plan-review) — use `blueprint config get` or `blueprint detect-stack`

Secondary: review skill writing quality against skill-creator SKILL.md patterns. Plan and plan-review are already high quality — focus quality improvements on simpler skills.

## Notes
- Priority 1 (hardcoded paths) is a simple global search-and-replace
- Priority 2 may need a `blueprint meta --key plan_file` flag added to the CLI
- Plan and plan-review should keep Opus-level thoroughness — don't simplify those
- The `## CLI Acceleration Opportunities` section in plan-review already documents known gaps
