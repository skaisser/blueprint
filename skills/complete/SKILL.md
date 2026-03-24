---
name: complete
description: >
  Clean up worktree after PR merge — only needed for /plan-wt flows.
  Use this skill whenever the user says "/complete", "clean up worktree", "remove worktree",
  "delete the worktree", "worktree cleanup",
  or any request to clean up after a worktree-based feature is merged.
  For standard /plan flows, /finish handles everything — /complete is a no-op.
  Detects environment automatically and skips if nothing to clean up.
---

# Complete: Worktree Cleanup

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Clean up after a PR has been merged. **Only needed for worktree flows (`/plan-wt`).**

For standard `/plan` flows, `/finish` already handles everything. This command detects the environment and skips if there's nothing to clean up.

```
Standard: /plan → ... → /finish ✅ (done)
Worktree: /plan-wt → ... → /finish → /complete (run from main repo terminal)
```

## Step 1: Detect + Validate Environment

Single bash call to detect worktree, check PR status, and run safety checks:

```bash
IS_WORKTREE=$(git rev-parse --git-dir 2>/dev/null | grep -q "worktrees" && echo "yes" || echo "no")
CURRENT_BRANCH=$(git branch --show-current)
echo "worktree=$IS_WORKTREE branch=$CURRENT_BRANCH"
# If in worktree: check PR merged + safety
if [ "$IS_WORKTREE" = "yes" ]; then
  gh pr list --head "$CURRENT_BRANCH" --state merged --json number --jq '.[0].number' || echo "NO_MERGED_PR"
  git diff --quiet && git diff --cached --quiet && echo "clean" || echo "ERROR: uncommitted changes"
  STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
  [[ "$CURRENT_BRANCH" =~ ^(main|master|develop)$ ]] || [[ "$CURRENT_BRANCH" == "$STAGING_BRANCH" ]] && echo "ERROR: protected branch" || echo "branch_safe"
fi
```

### If NOT in a worktree:
Check for leftover branches. If none, report "Nothing to clean up" and STOP.

### If in a worktree but PR NOT merged:
STOP. NEVER delete branches of unmerged PRs.

### If in a worktree with uncommitted changes or protected branch:
STOP with error.

### If all checks pass:
Use AskUserQuestion to confirm: "Will remove worktree at [path], delete branch [name] locally and remotely. Proceed?"

## Step 2: Execute Cleanup

```bash
MAIN_REPO=$(git worktree list | head -1 | awk '{print $1}')
WORKTREE_PATH=$(pwd)
cd "$MAIN_REPO"
git fetch --all --prune
git checkout "$BASE_BRANCH" && git pull origin "$BASE_BRANCH"
git worktree remove "$WORKTREE_PATH"
git worktree prune
git branch -d "$BRANCH" 2>/dev/null || git branch -D "$BRANCH"
git push origin --delete "$BRANCH" 2>/dev/null || true
```

## Step 3: Verify + Report

Run `git worktree list` and confirm the removed worktree path no longer appears in the output. If it still appears, run `git worktree prune` again and re-check. Only proceed to report once the worktree is confirmed gone.

```bash
REMAINING=$(git worktree list)
echo "$REMAINING"
echo "$REMAINING" | grep -q "$WORKTREE_PATH" && echo "ERROR: worktree still present" || echo "VERIFIED: worktree removed"
```

If verification fails, report the error and STOP — do not claim success.

If verified:

```
Cleanup complete!
  Removed: [worktree path], branch (local + remote)
  Current: [main repo path] on $BASE_BRANCH
  Remaining worktrees: [list from verification output]
```

## Safety

- Verify PR is merged before deleting branch
- Use `-d` (safe delete) when possible
- Never delete main, master, the staging branch (from `blueprint/.config.yml`), or develop
- All user interactions MUST use `AskUserQuestion` tool, never plain text questions

Use $ARGUMENTS for any additional context.
