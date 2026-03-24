---
name: bp-branch
description: >
  Create a new feature branch following type/kebab-case conventions.
  Triggers on "/bp-branch", "/branch", "create branch", "new branch",
  "make a branch", or any request to create a git branch.
  Also triggers on "checkout new branch", "start a branch", "feature branch",
  "I need a new branch", or "branch for [something]".
  Uses type/description format: feat/feature-name, fix/bug-name, etc.
---

# Branch: Create Feature Branch

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Create a new feature branch following our conventions.

## Format
```
<type>/<description>
```

## Types
- feat: New features
- fix: Bug fixes
- docs: Documentation
- style: Formatting
- refactor: Restructuring
- perf: Performance
- test: Testing
- build: Build changes
- chore: Maintenance
- hotfix: Urgent fixes
- plan: Planning/spec work
- security: Security fixes
- migration: Database migrations
- deps: Dependency updates
- deploy: Deployment/CI changes
- remove: Removing code/features

## Rules
- Use kebab-case for description
- Keep it short and descriptive
- Match the commit type you'll use

## Process

1. Verify the branch doesn't already exist:
   ```bash
   git branch --list "<type>/<branch-name>" && echo "ERROR: branch already exists" && exit 1
   ```

2. Start from latest base branch:
   ```bash
   # Determine base: use staging branch if it exists, otherwise main
   STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml 2>/dev/null | awk '{print $2}')
   STAGING_BRANCH=${STAGING_BRANCH:-staging}
   if git show-ref --verify --quiet refs/heads/$STAGING_BRANCH; then
       BASE="$STAGING_BRANCH"
   else
       BASE="main"
   fi
   git checkout "$BASE" && git pull origin "$BASE"
   ```

3. Create the branch:
   ```bash
   git checkout -b <type>/<branch-name>
   ```

Do NOT commit or push anything — branch creation only.

Use $ARGUMENTS as the branch name/description.
