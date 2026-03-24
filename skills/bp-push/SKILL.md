---
name: bp-push
description: >
  Push the current branch to remote with branch safety checks.
  Triggers on "/bp-push", "/push", "push this", "push to remote",
  "push branch", or any request to push code to the remote repository.
  Also triggers on "push my code", "push the branch", or "send it to remote".
  Does NOT trigger on "push my changes" — use /bp-ship if uncommitted changes exist.
  Blocks pushes to main/master — only staging and feature branches are allowed.
---

# Push: Push Branch to Remote

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Push the current branch to remote.

## Pre-flight Checks

```bash
BRANCH=$(git branch --show-current)
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml 2>/dev/null | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}

# Block pushes to main/master
if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
    echo "ERROR: Cannot push directly to '$BRANCH'. Use a PR from $STAGING_BRANCH → $BRANCH."
    exit 1
fi

# Warn if working tree is dirty
if [ -n "$(git status --porcelain)" ]; then
    echo "Working tree has uncommitted changes. Use /bp-ship to commit and push, or /bp-commit first."
    exit 1
fi
```

The staging branch is pushable directly (plan commits, hotfixes). Only `main` requires a PR.

## Push

```bash
git push -u origin $(git branch --show-current)
```

If push is rejected due to remote changes:
```
Push rejected — remote has new commits. Pull or rebase first:
git pull --rebase origin <branch>
```

## Rules
- Do NOT commit anything — push only.
- Do NOT create a PR — push only.
- Do NOT force push unless the user explicitly requests it.
- Keep output minimal — just confirm the push result:
  ```
  Pushed to origin/<branch>
  ```
