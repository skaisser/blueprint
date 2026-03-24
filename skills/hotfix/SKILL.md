---
name: hotfix
description: >
  Emergency hotfix deployment: commit, push to staging branch, create PR to main, and merge.
  Use this skill when the user says "/hotfix", "emergency push", "hotfix to production",
  "push to staging and merge to main", "urgent deploy", "fast push to main",
  "deploy hotfix", or any request for an emergency/urgent deployment bypassing the normal
  /plan → /pr → /finish workflow. This is the fast lane: commit → push → PR → merge.
---

# Hotfix: Emergency Deploy to Main

Fast-lane deployment for urgent fixes. Bypasses normal plan workflow.

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

**Flow:** commit → push to staging branch → PR staging→main → merge → GitHub Issue update

## When to Use

- Production bugs that need immediate fix
- Urgent security patches
- Quick important changes that can't wait for full plan cycle
- User explicitly asks for emergency/hotfix deploy

## Process

### 1. Pre-flight Checks

```bash
BRANCH=$(git branch --show-current)
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}
BASE=$(~/.blueprint/bin/blueprint meta 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('base_branch','$STAGING_BRANCH'))")
```

**Verify:**
- Current branch is NOT main or master (block if so)
- There are changes to commit (staged or unstaged)
- Tests pass for changed files (run targeted tests only)

### 2. Commit Changes

Stage and commit:

```bash
# Stage all changes (review what's being added)
git add -A
# Verify no unrelated files are staged — if unrelated files appear, unstage them

# Commit with hotfix emoji
git commit -m "🩹 hotfix: $DESCRIPTION"
```

### 3. Push to Staging Branch

```bash
# If on feature branch, push there first
git push origin "$BRANCH"

# If branch is not staging, merge into staging
if [ "$BRANCH" != "$STAGING_BRANCH" ]; then
    git checkout "$STAGING_BRANCH"
    git pull origin "$STAGING_BRANCH"
    git merge "$BRANCH" -m "🔀 merge: $BRANCH into $STAGING_BRANCH (hotfix)"
    git push origin "$STAGING_BRANCH"
    git checkout "$BRANCH"
fi
```

### 4. Create PR: staging → main

**GitHub Issue Detection:** If `$ARGUMENTS` contains a GitHub issue reference (e.g., `/hotfix #42 fix login timeout`, `/hotfix issue 42 fix crash`), parse the issue number and include `Closes #N` in the PR body.

```bash
# Parse issue number from $ARGUMENTS if present
ISSUE_REF=""
# Check for patterns: #42, issue 42, or leading number
ISSUE_NUM=$(echo "$ARGUMENTS" | grep -oE '(#|issue ?)([0-9]+)' | grep -oE '[0-9]+' | head -1)
if [ -n "$ISSUE_NUM" ]; then
    ISSUE_REF="Closes #$ISSUE_NUM"
fi

# Create PR
gh pr create \
    --base main \
    --head "$STAGING_BRANCH" \
    --title "🩹 hotfix: $DESCRIPTION" \
    --body "## Emergency Hotfix

**What:** $DESCRIPTION
**Why:** Urgent fix requiring immediate deployment
**Branch:** $BRANCH → $STAGING_BRANCH → main

### Changes
$(git log main..$STAGING_BRANCH --oneline)

### Verification
- [ ] Targeted tests pass
- [ ] Manual smoke test completed

### References
${ISSUE_REF}
"
```

### 5. Merge PR

```bash
PR_NUMBER=$(gh pr view "$STAGING_BRANCH" --json number -q .number)
gh pr merge "$PR_NUMBER" --merge --delete-branch=false
```

**IMPORTANT:** Do NOT delete the staging branch — it's permanent.

### 6. Commit Plan Update (if applicable)

```bash
git add blueprint/ && git commit -m "🩹 hotfix: plan update for $DESCRIPTION" 2>/dev/null || true
```

## Rules

- **ALWAYS review staged files** — unstage anything unrelated to the hotfix before committing
- **ALWAYS run targeted tests before pushing** — even hotfixes get tested
- **NEVER skip the PR** — main is protected, always go through staging → main PR
- **NEVER delete the staging branch** after merge
- **FULLY AUTOMATIC** — push → PR → merge, NO questions asked. This is the emergency lane.
- If tests fail, STOP and use `AskUserQuestion` — don't push broken code
- Use `🩹 hotfix:` commit format
- All user interactions MUST use `AskUserQuestion` tool, never plain text questions

## Arguments

Use $ARGUMENTS as the hotfix description. If empty, ask the user what the hotfix is for.
