---
name: flow-auto-wt
description: >
  Fully autonomous SDLC pipeline in an isolated git worktree — zero user intervention,
  zero git conflicts with other terminals. Creates a lightweight worktree (no Herd URL,
  no VS Code, no SSL), runs the full flow-auto pipeline inside it, and stops when the PR
  is ready for merge. Use this skill whenever the user says "/flow-auto-wt", "auto flow worktree",
  "parallel flow-auto", "flow-auto in worktree", or any request to run the autonomous pipeline
  in an isolated worktree. Also triggers on "isolated flow-auto", "worktree auto", "parallel auto",
  "run flow-auto in isolation", "don't touch my branch", "run in parallel", "parallel pipeline",
  or when running multiple flow-auto instances on the same repo. IMPORTANT: never use
  AskUserQuestion in this skill — all decisions are made autonomously.
---

# Flow Auto WT: Autonomous Pipeline in Isolated Worktree

Run the full `/flow-auto` pipeline inside an isolated git worktree. Same zero-touch autonomy,
but each instance gets its own working directory so multiple pipelines can run on the same repo
without git conflicts.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

```
/flow-auto-wt <description>
  → worktree → plan → review → execute → check → PR → cleanup → DONE
  Lightweight worktree (no Herd, no VS Code, no SSL). Just git isolation.
  Zero AskUserQuestion calls. Zero pauses. One command → PR ready for merge.
  Main repo working directory is NEVER modified.
```

## Why This Exists

Running 3+ `/flow-auto` instances on the same repo causes git conflicts — they share the same
working directory, index, and staged files. One agent's `git add` can pick up another's changes.

`/flow-auto-wt` solves this by creating a lightweight git worktree before running the pipeline.
Unlike `/plan-wt`, there's no Herd site, no VS Code window, no SSL setup — just git isolation.

| Variant | Isolation | Herd URL | VS Code | Use Case |
|---------|-----------|----------|---------|----------|
| `/flow-auto` | None | No | No | Single pipeline, exclusive repo access |
| `/flow-auto-wt` | Worktree | No | No | Parallel pipelines, shared repo |
| `/plan-wt` | Worktree | Yes | Yes | Interactive development with browser testing |

## Critical Rules

1. **NEVER use AskUserQuestion.** Every decision point is handled autonomously.
2. **NEVER merge the PR.** The pipeline stops after creating/fixing the PR. The user decides when to merge.
3. **NEVER modify the main repo's working directory or HEAD.** Use `git fetch` + remote refs only.
4. **Delegate all implementation.** Same as `/plan-approved` — the coordinator never writes code.
5. **Commit after every stage.** Every stage produces a commit.
6. **Loop review→fix max 3 times.** Prevent infinite loops.
7. **Skip GitHub Issues integration** unless $ARGUMENTS contains a GitHub issue ID.
8. **NEVER skip steps.** ALL 10 steps must execute in order.
9. **The checkpoint echo is NOT optional.** Every step MUST start with its echo command.
10. **Worktree is lightweight.** No Herd link/secure/open. No VS Code. No SSL. Just `git worktree add`.
11. **Auto-cleanup worktree after PR.** Once the PR is created and review loop finishes, remove the worktree (code is safely on the remote branch). Use `--keep-worktree` to skip cleanup.

---

## Step 1: Initialize

```bash
echo "🤖 [flow-auto-wt:1] initializing autonomous worktree pipeline"
```

```bash
# Guard: refuse to run if already inside a worktree
TOPLEVEL=$(git rev-parse --show-toplevel 2>/dev/null)
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  COMMON=$(git rev-parse --git-common-dir 2>/dev/null)
  GITDIR=$(git rev-parse --git-dir 2>/dev/null)
  if [ "$COMMON" != "$GITDIR" ]; then
    echo "ERROR: Already inside a worktree. Run from the main repo."
    # STOP. Do NOT proceed.
  fi
fi

~/.blueprint/bin/blueprint meta 2>/dev/null
```

**If active plan exists** with worktree field:
- Read the plan file. Check status.
- `awaiting-approval` or `approved` → Skip to Step 5 (execute)
- `in-progress` → Skip to Step 5 (resume)
- `completed` → Skip to Step 7 (PR)

**If no active plan:** Continue to Step 2.

Parse $ARGUMENTS for:
- GitHub issue ID (e.g., `#42`) → store for PR body
- `--from <stage>` → jump to that stage
- `--keep-worktree` → skip auto-cleanup after PR
- Everything else → task description

## Step 2: Create Worktree

```bash
echo "🤖 [flow-auto-wt:2] creating lightweight worktree"
```

**CRITICAL: Do NOT checkout or pull in the main repo. Use `git fetch` + remote refs only.**

```bash
REPO_NAME=$(basename "$PWD")
REPO_ROOT="$PWD"
PARENT_DIR=$(dirname "$PWD")

# Generate branch name from description
BRANCH="feat/$(echo "$DESCRIPTION" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g' | head -c 50)"

# Worktree path: parent dir / .worktrees / repo-branch-slug
WORKTREE_SLUG=$(echo "$BRANCH" | tr '/' '-')
WORKTREE_PATH="${PARENT_DIR}/.worktrees/${REPO_NAME}-${WORKTREE_SLUG}"

# Ensure parent exists
mkdir -p "${PARENT_DIR}/.worktrees"

# Fetch latest from remote WITHOUT modifying main repo's HEAD or working directory
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}
BASE_BRANCH=$(git rev-parse --verify "origin/$STAGING_BRANCH" 2>/dev/null && echo "$STAGING_BRANCH" || echo "main")
git fetch origin "$BASE_BRANCH"

# Create worktree from remote ref — main repo state stays untouched
git worktree add -b "$BRANCH" "$WORKTREE_PATH" "origin/$BASE_BRANCH"
```

**Worktree Gate — verify before continuing:**
```bash
CURRENT=$(git worktree list | grep "$WORKTREE_PATH")
echo "Worktree: $CURRENT"
if [ -z "$CURRENT" ]; then
    echo "ERROR: Worktree creation failed"
    # STOP. Report failure. Do NOT proceed.
fi
```

**Install dependencies (if applicable):**
```bash
cd "$WORKTREE_PATH"

# PHP projects
[ -f "composer.json" ] && composer install --no-interaction 2>/dev/null || true

# Node projects
[ -f "package.json" ] && npm install 2>/dev/null || true

# Build assets (non-interactive, no dev server)
[ -f "vite.config.js" ] || [ -f "vite.config.ts" ] && npm run build 2>/dev/null || true
```

**All subsequent steps run inside `$WORKTREE_PATH`.**

## Step 3: Plan

```bash
echo "🤖 [flow-auto-wt:3] creating plan"
```

1. Read `~/.claude/skills/plan/references/plan-template.md` for the plan format
2. Use `mcp__sequential-thinking__sequentialthinking` to analyze the task
3. Write the plan file to `blueprint/live/NNNN-feat-description.md`
   - Include `worktree: $WORKTREE_PATH` in frontmatter
   - Include `${CLAUDE_SESSION_ID}` in YAML `sessions:` and Work Sessions block
4. Commit:
   ```bash
   git add blueprint/ && git commit -m "📋 plan: create NNNN-description"
   ```

## Step 4: Review

```bash
echo "🤖 [flow-auto-wt:4] reviewing plan"
```

1. Read `~/.claude/skills/plan-review/references/team-execution.md`
2. Use `mcp__sequential-thinking__sequentialthinking` to validate
3. Mark ALL tasks with `[H]`/`[S]`/`[O]` complexity
4. Determine execution strategy (Mode A-G)
5. Add `## Execution Strategy` section
6. Update status to `Approved`, set progress counters
7. Commit:
   ```bash
   git add blueprint/ && git commit -m "📋 plan: review NNNN-description"
   ```

## Step 5: Execute

```bash
echo "🤖 [flow-auto-wt:5] executing plan"
```

Same as flow-auto Step 4:
1. Read full plan file. Identify completed/pending phases.
2. Update status to `In Progress`
3. Execute rounds following Execution Strategy
4. Workers commit their code directly in the worktree
5. On completion:
   ```bash
   git add blueprint/ && git commit -m "✨ feat: complete NNNN-description"
   ```

## Step 6: Plan Check

```bash
echo "🤖 [flow-auto-wt:6] auditing implementation"
```

Same as flow-auto Step 5:
1. `~/.blueprint/bin/blueprint context --diffs`
2. Compare plan vs implementation
3. Fix mismatches
4. Sync frontmatter
5. Commit:
   ```bash
   git add blueprint/ && git commit -m "🧹 chore: plan check NNNN"
   ```

## Step 7: Create PR

```bash
echo "🤖 [flow-auto-wt:7] creating pull request"
```

1. Determine base branch:
   ```bash
   STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
   STAGING_BRANCH=${STAGING_BRANCH:-staging}
   git rev-parse --verify "origin/$STAGING_BRANCH" 2>/dev/null && echo "$STAGING_BRANCH" || echo "main"
   ```
2. Push branch:
   ```bash
   git push -u origin "$BRANCH"
   ```
3. Gather context: `~/.blueprint/bin/blueprint context`
4. Create PR with `gh pr create`

## Step 8: Review Loop

```bash
echo "🤖 [flow-auto-wt:8] starting review loop"
```

### Pre-check: GitHub Action exists? (MANDATORY — do NOT skip)

**Do NOT rationalize skipping this step.** The pre-check below is the ONLY valid reason to skip. "The fix is small" or "tests already pass" are NOT valid skip reasons.

Valid skip paths:
1. `--no-review` is explicitly present in `$ARGUMENTS`
2. The GitHub Action check runs and finds no @claude workflow

```bash
if echo "$ARGUMENTS" | grep -q '\-\-no-review'; then
  echo "🤖 [flow-auto-wt:8] skipping review loop — --no-review flag set"
  # Jump directly to Step 9
fi

CLAUDE_ACTION=$(gh api repos/{owner}/{repo}/actions/workflows --jq '.workflows[] | select(.name | test("claude|Claude|CLAUDE")) | .id' 2>/dev/null)
if [ -z "$CLAUDE_ACTION" ]; then
  echo "🤖 [flow-auto-wt:8] skipping review loop — no @claude GitHub Action detected"
  # Jump directly to Step 9
fi
```

If the workflow exists, run at least 1 review cycle. Max 3 iterations:
1. Trigger review: `gh pr comment "$PR_NUM" --body "@claude review this PR and check if we are able to merge. Analyze the code changes for any issues, security concerns, or improvements needed."`
2. Wait for review (poll every 5 min, max 20 min)
3. If comments: fetch, categorize, dispatch fix agents, push
4. Loop until clean or max cycles reached

## Step 9: Cleanup & Report

```bash
echo "🤖 [flow-auto-wt:9] pipeline complete — cleaning up and posting report"
```

**Auto-cleanup worktree** (all code is safely on remote branch now):
```bash
cd "$REPO_ROOT"  # Return to main repo

# Remove worktree (code is safe on remote branch)
if [ -z "$KEEP_WORKTREE" ]; then
  git worktree remove "$WORKTREE_PATH" --force 2>/dev/null || true
  git worktree prune
  echo "Worktree cleaned up: $WORKTREE_PATH"
else
  echo "Worktree kept (--keep-worktree): $WORKTREE_PATH"
fi
```

Post PR comment with final report:

```bash
gh pr comment "$PR_NUM" --body "$(cat <<'EOF'
## 🤖 Autonomous Pipeline Complete — Opus 4.6 Final Report

### Summary
- **Plan:** NNNN-description
- **Phases:** N phases, M tasks — all completed
- **Review cycles:** N iterations
- **Commits:** X commits on this branch

### What was built
- [bullet summary of each phase's deliverables]

### Test status
- [targeted test results from execution]
- ⚠️ Run full test suite before merging

### Recommendations
- [any concerns, edge cases, or things to verify]

### Next steps
1. Review the PR
2. Run full test suite to verify
3. Run `/finish` to merge and rename plan

---
*Generated by /flow-auto-wt — Blueprint SDLC*
EOF
)"
```

## Step 10: Output to User

```bash
echo "🤖 [flow-auto-wt:10] done"
```

```
🤖 Flow Auto WT Complete!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Plan:      NNNN-description
Branch:    feat/description
PR:        #NNN — <title>
URL:       <pr-url>
Review:    N cycles
Worktree:  Cleaned up (code safe on remote)
Status:    Ready for human review

Next: /finish to merge PR and rename plan
Run full test suite to verify before merging.
```

**STOP. Pipeline complete. Do NOT merge.**

The user will run `/finish` to merge the PR and rename the plan file.

---

## Rules

- **NEVER use AskUserQuestion** — all decisions autonomous
- **NEVER merge** — leave PR open for human review
- **NEVER modify main repo HEAD or working directory** — use `git fetch` + remote refs only
- **NEVER skip tests** — workers run targeted tests
- **NEVER create Herd sites** — lightweight worktree only
- **Auto-cleanup worktree** after PR is created and review loop finishes (use `--keep-worktree` to skip)
- **Max 3 review cycles** — prevent infinite loops
- **Commit after every stage** — full git history
- **Same delegation rules as /plan-approved** — coordinator orchestrates, never implements
- **If context gets critical (>85%):** compact and continue
- **If a stage fails catastrophically:** commit what you have, create PR with failure note

## Flags

- `--from <stage>`: Start from specific stage
- `--no-review`: Skip the review loop
- `--max-cycles N`: Override max review cycle count (default: 3)
- `--no-install`: Skip composer/npm install in worktree
- `--keep-worktree`: Do not auto-cleanup worktree after PR creation

Use $ARGUMENTS as the task description, GitHub issue ID, or flags.
