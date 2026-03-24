---
name: start
description: >
  Initialize a new project with GitHub Actions, git hooks, blueprint workspace, CLAUDE.md, and branches.
  Use this skill whenever the user says "/start", "initialize project", "setup project",
  "start new project", or any request to set up the standard development workflow in a repo.
  Also triggers on "project init", "bootstrap project", "setup hooks", or "initialize repo".
  Sets up everything: claude-pr-reviewer.yml + tests.yml actions, commit hooks, blueprint/ workspace,
  CLAUDE.md, configurable staging branch. Updates are handled by `blueprint update` — no separate sync skills needed.
---

# Start: Initialize Project

Initialize a new project with GitHub Actions, git hooks, BLUEPRINT workspace, CLAUDE.md, and branch setup.

## Language

Read `blueprint/.config.yml` → `language` field. If set to `auto` (default), detect the language from the user's messages. All generated content (CLAUDE.md, blueprint files, status output, instructions) MUST be in the detected language. Skill instructions stay in English — only output changes. The emoji + type prefix in commits stays English always (`✨ feat:`), only the description adapts.

## Pre-requisite: GitHub App (one-time per GitHub account)

The `@claude` and `@tests` GitHub Actions require the Claude Code GitHub App to be installed. This is a **one-time setup per GitHub account** — not per-project.

### First-time setup (once ever)

Tell the user:

> **One-time GitHub setup required:**
>
> 1. Run `/install-github-app` in Claude Code — this opens your browser to authorize the Claude Code GitHub App for your GitHub account
> 2. After authorizing, GitHub generates a `CLAUDE_CODE_OAUTH_TOKEN` secret
> 3. Go to your repo → Settings → Secrets and variables → Actions → New repository secret
> 4. Name: `CLAUDE_CODE_OAUTH_TOKEN`, Value: the token from step 2
>
> You only do step 1 once. Step 3-4 is needed per repository.
>
> **Video walkthrough:** https://blueprint.skaisser.dev/setup (coming soon)

### Already installed the app? (new repo only)

If the user has already installed the GitHub App on a previous project, they only need to add the secret to the new repo:

> **Add the secret to this repo:**
>
> Go to repo → Settings → Secrets and variables → Actions → New repository secret
> Name: `CLAUDE_CODE_OAUTH_TOKEN`, Value: your existing OAuth token
>
> If you lost the token, run `/install-github-app` again to regenerate it.

Use `AskUserQuestion` to ask: "Have you already installed the Claude Code GitHub App on your GitHub account? (yes/no)" — then show the appropriate instructions above.

Continue with the rest of `/start` regardless — hooks, workspace, and CLAUDE.md don't depend on the GitHub App.

## What It Sets Up

1. **`blueprint/.config.yml`** — Project config (staging branch name, detected stack)
2. **`.github/workflows/claude-pr-reviewer.yml`** — GitHub Action for `@claude` PR review mentions
3. **`.github/workflows/tests.yml`** — On-demand test runner triggered by `@tests` comment on PRs
4. **`.githooks/`** — commit-msg (validates emoji format, blocks AI signatures) + pre-push (smart main protection)
5. **`blueprint/`** — BLUE workspace: `backlog/`, `live/`, `upstream/`, `expired/`
6. **`CLAUDE.md`** — Project-level instructions (stack-agnostic, customized with detected stack)
7. **`.gitignore` entries** — blueprint-specific entries
8. **Staging branch** — Configurable name (default: `staging`)

## Process

### 1. Verify Git Repository
```bash
git rev-parse --is-inside-work-tree || { echo "Not a git repository"; exit 1; }
```

### 2. Detect Project Stack

Before generating anything, read the project to understand the stack. This informs the CLAUDE.md template and the `tests.yml` workflow.

```bash
# Language detection
HAS_COMPOSER=$(test -f composer.json && echo true || echo false)
HAS_PACKAGE_JSON=$(test -f package.json && echo true || echo false)
HAS_GEMFILE=$(test -f Gemfile && echo true || echo false)
HAS_REQUIREMENTS=$(test -f requirements.txt -o -f pyproject.toml && echo true || echo false)
HAS_GO_MOD=$(test -f go.mod && echo true || echo false)

# PHP specifics
if [ "$HAS_COMPOSER" = true ]; then
    PHP_VERSION=$(grep -oP '"php":\s*"\^?\K[0-9.]+' composer.json 2>/dev/null || echo "8.3")
    HAS_PEST=$(test -f vendor/bin/pest && echo true || echo false)
    HAS_LARAVEL=$(grep -q laravel/framework composer.json 2>/dev/null && echo true || echo false)
fi

# Node specifics
if [ "$HAS_PACKAGE_JSON" = true ]; then
    NODE_VERSION=$(cat .nvmrc 2>/dev/null || echo "22")
    HAS_VITE=$(test -f vite.config.js -o -f vite.config.ts && echo true || echo false)
fi
```

### 3. Ask Staging Branch Name

Use `AskUserQuestion` to ask the user what they call their staging/pre-production branch:

> What do you call your staging branch? This is the branch between feature branches and `main`.
>
> Common names:
> - **staging** (most common — default)
> - **develop** / **dev** (GitFlow convention)
> - **homolog** (common in Brazil)
> - **qa**
>
> Press Enter for `staging`, or type your preferred name:

Store the answer. Default to `staging` if the user presses Enter or says nothing.

### 4. Create BLUEPRINT Config

Create `blueprint/.config.yml` — this is the project-level config that all BLUEPRINT skills read from:

```yaml
# BLUEPRINT project configuration
# All skills read from this file — edit freely.

# The branch between feature branches and main.
# Your hooks and workflows use this name automatically.
staging_branch: staging

# Language for generated content (plans, PRs, CLAUDE.md, messages).
# auto = detect from user's messages. Or set explicitly: en, pt-BR, es, fr, de, ja, zh, etc.
# Commit types (feat, fix) and frontmatter keys always stay in English.
language: auto

# Detected project stack (informational — used by /start and tests.yml)
stack:
  language: php          # php | node | python | go | ruby | unknown
  framework: laravel     # laravel | express | django | gin | rails | none
  test_runner: pest      # pest | phpunit | jest | vitest | pytest | go-test | none
  has_assets: true       # true if Vite/webpack detected
```

Fill in the detected values from step 2. The `staging_branch` field uses whatever the user chose in step 3.

### 5. Copy Git Hooks

Copy hooks from BLUEPRINT templates. The pre-push hook reads `staging_branch` from `blueprint/.config.yml`.

```bash
TEMPLATE_DIR=""
if [ -d "$HOME/.blueprint/templates" ]; then
    TEMPLATE_DIR="$HOME/.blueprint/templates"
elif [ -d "$HOME/.claude/templates" ]; then
    TEMPLATE_DIR="$HOME/.claude/templates"
fi

mkdir -p .githooks
if [ -n "$TEMPLATE_DIR" ] && [ -d "$TEMPLATE_DIR/.githooks" ]; then
    cp -n "$TEMPLATE_DIR/.githooks/commit-msg" .githooks/ 2>/dev/null || true
    cp -n "$TEMPLATE_DIR/.githooks/pre-push" .githooks/ 2>/dev/null || true
else
    # Write hooks inline (see hook content in templates/)
fi
chmod +x .githooks/commit-msg .githooks/pre-push 2>/dev/null || true
```

The **pre-push hook** must:
1. Read `staging_branch` from `blueprint/.config.yml` (fall back to `staging` if missing)
2. Block direct pushes to `main` ONLY if the staging branch exists (local or remote)
3. If no staging branch exists → allow pushes to `main` (single-branch projects)
4. Always allow pushes to the staging branch and feature branches

### 6. Configure Git Hooks Path
```bash
git config core.hooksPath .githooks
```

### 7. Create BLUEPRINT Workspace (BLUE folders)

```bash
mkdir -p blueprint/backlog blueprint/live blueprint/upstream blueprint/expired
```

Add `.gitkeep` to each empty folder so git tracks them:
```bash
for dir in blueprint/backlog blueprint/live blueprint/upstream blueprint/expired; do
    [ -z "$(ls -A "$dir" 2>/dev/null)" ] && touch "$dir/.gitkeep"
done
```

### 8. Create CLAUDE.md

If `CLAUDE.md` does not exist, generate one customized for the detected stack.

**The CLAUDE.md template MUST include these sections:**

```markdown
# {Project Name} — Claude Code Instructions

## Tech Stack

{Detected from step 2 — list framework, language, DB, test runner, etc.}

## Development

### Local URL

{If Laravel+Herd: `project-folder.test`, otherwise detect or leave placeholder}

### Common Commands

{Stack-appropriate commands — e.g., migrations, dev server, test commands}

## Workflow

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

**Branch flow**: `feat/feature-name → {staging_branch} → main`

- Feature branches always target `{staging_branch}`
- `{staging_branch} → main` happens via PR after QA
- Direct pushes to `main` are blocked by hooks

## Testing Rules

**Always use targeted tests** — fast feedback:

{Stack-appropriate test commands with --filter or equivalent}

Full suite → ask the user to run it in their own terminal.

## BLUEPRINT Workspace

```
blueprint/
├── backlog/    ← Ideas not yet planned (/backlog)
├── live/       ← Currently in development (/plan)
├── upstream/   ← Shipped and merged (/finish)
└── expired/    ← Cancelled or deferred
```

Blueprint config: `blueprint/.config.yml`

## Commits

Format: `<emoji> <type>: <description>` (present tense, lowercase)

```
✨ feat: add payment processing
🐛 fix: resolve duplicate email validation
🧪 test: add coverage for order service
```

**Never** add AI signatures (`Co-Authored-By`, `Generated by`, etc.)

## Updates

Run `blueprint update` to get the latest skills, hooks, and CLI binary.

## Project-Specific Notes

<!-- Add project-specific patterns, architecture decisions, or gotchas here -->
```

Replace `{Project Name}` with `$(basename "$(pwd)")` and `{staging_branch}` with the value from `blueprint/.config.yml`. Fill in detected stack details. If stack can't be detected, use generic placeholders the user can fill in.

### 9. Create GitHub Actions (only if staging branch exists or was just created)

Only create these workflows if the project uses a staging branch. If the project is single-branch (no staging), skip entirely — no workflows needed.

#### 9a. claude-pr-reviewer.yml — @claude Review Action

Do NOT use `cp`. Write the file directly to ensure it is always functional.
If `.github/workflows/claude-pr-reviewer.yml` already exists, skip (no clobber).

```yaml
name: Claude PR Reviewer

on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]
  issues:
    types: [opened, assigned]
  pull_request_review:
    types: [submitted]

jobs:
  claude:
    if: |
      (github.event_name == 'issue_comment' && contains(github.event.comment.body, '@claude')) ||
      (github.event_name == 'pull_request_review_comment' && contains(github.event.comment.body, '@claude')) ||
      (github.event_name == 'pull_request_review' && contains(github.event.review.body, '@claude')) ||
      (github.event_name == 'issues' && (contains(github.event.issue.body, '@claude') || contains(github.event.issue.title, '@claude')))
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: read
      issues: read
      id-token: write
      actions: read
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 1

      - name: Run Claude Code
        id: claude
        uses: anthropics/claude-code-action@v1
        with:
          claude_code_oauth_token: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
          additional_permissions: |
            actions: read
```

**How it works:** Comment `@claude review this PR` on a PR, and Claude Code runs in GitHub Actions to review the code.

**Requires:** `CLAUDE_CODE_OAUTH_TOKEN` secret (see Pre-requisite section above).

#### 9b. tests.yml — @tests On-Demand Test Runner

This workflow runs the project's test suite on demand — triggered by commenting `@tests` on a PR. It must mirror the project's local test setup exactly so CI matches local results.

If `.github/workflows/tests.yml` already exists, skip (no clobber).

**IMPORTANT:**
- The workflow file MUST exist on the `main` branch to trigger via `issue_comment` — GitHub runs `issue_comment` workflows from the default branch, not the PR branch
- The `permissions` block is REQUIRED — without `issues: write` and `pull-requests: write`, the report step fails
- PR checkout MUST use `refs/pull/${{ github.event.issue.number }}/head`
- If the project uses Vite/webpack, you MUST run `npm ci && npm run build` before tests

**Generate the workflow dynamically based on the detected stack from step 2:**

The trigger, permissions, PR checkout, and report step are always the same. Only the middle steps (setup, install, build, test) change per project.

```yaml
name: Tests

on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  tests:
    if: |
      (github.event_name == 'issue_comment' && contains(github.event.comment.body, '@tests')) ||
      (github.event_name == 'pull_request_review_comment' && contains(github.event.comment.body, '@tests'))
    runs-on: ubuntu-latest
    steps:
      - name: Checkout PR branch
        uses: actions/checkout@v4
        with:
          ref: refs/pull/${{ github.event.issue.number }}/head
          fetch-depth: 1

      # === DYNAMIC STEPS — generated from stack detection ===
      #
      # PHP/Laravel example:
      #   - Setup PHP (shivammathur/setup-php)
      #   - composer install
      #   - Setup Node + npm ci + npm run build (if Vite)
      #   - cp .env.example .env && php artisan key:generate
      #   - Run: ./vendor/bin/pest --parallel --processes=10
      #
      # Node/JS example:
      #   - Setup Node (actions/setup-node)
      #   - npm ci
      #   - npm run build (if build script exists)
      #   - Run: npx jest / npx vitest
      #
      # Python example:
      #   - Setup Python (actions/setup-python)
      #   - pip install -r requirements.txt
      #   - Run: pytest
      #
      # Go example:
      #   - Setup Go (actions/setup-go)
      #   - go test ./...

      - name: Report result
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const status = '${{ job.status }}' === 'success' ? '✅' : '❌';
            const message = `${status} **Tests ${('${{ job.status }}').toUpperCase()}** — <TEST_COMMAND>\n\n<RUNTIME_INFO> · ${new Date().toISOString()}`;
            const issueNumber = context.issue.number;
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: issueNumber,
              body: message
            });
```

Replace `<TEST_COMMAND>` with the actual test command and `<RUNTIME_INFO>` with detected runtime details.

**How it works:** Comment `@tests` on a PR, and the full test suite runs in CI against the PR branch. Results are posted back as a comment with pass/fail status.

### 10. Update .gitignore

Add blueprint-specific entries if missing:

```bash
for entry in ".db-sync.env" "database/backups/" ".firecrawl/"; do
    grep -qxF "$entry" .gitignore 2>/dev/null || echo "$entry" >> .gitignore
done
```

### 11. Create Staging Branch (if missing)

Use the `staging_branch` name from `blueprint/.config.yml`:

```bash
STAGING_BRANCH=$(grep 'staging_branch:' blueprint/.config.yml | awk '{print $2}')
STAGING_BRANCH=${STAGING_BRANCH:-staging}

if ! git show-ref --verify --quiet "refs/heads/$STAGING_BRANCH" && ! git show-ref --verify --quiet "refs/remotes/origin/$STAGING_BRANCH"; then
    git checkout -b "$STAGING_BRANCH"
    git push -u origin "$STAGING_BRANCH" 2>/dev/null || echo "No origin — local only"
    git checkout main 2>/dev/null || git checkout master
fi
```

### 12. Output

Report what was created/skipped, then remind the user:

```
✓ blueprint/.config.yml — staging branch: {staging_branch}
✓ .githooks/commit-msg — emoji format validation
✓ .githooks/pre-push — main branch protection (requires {staging_branch} → main PR)
✓ blueprint/ — BLUE workspace (backlog, live, upstream, expired)
✓ CLAUDE.md — project instructions ({detected_stack})
✓ .github/workflows/claude-pr-reviewer.yml — @claude PR review
✓ .github/workflows/tests.yml — @tests CI runner ({detected_test_command})
✓ {staging_branch} branch created

Next steps:
  1. Add CLAUDE_CODE_OAUTH_TOKEN secret to this repo (see instructions above)
  2. Review and customize CLAUDE.md for your project
  3. Commit this setup: ✨ feat: initialize BLUEPRINT workspace
  4. Start planning with /backlog or /plan
  5. Keep BLUEPRINT up to date: blueprint update
```

## Safety

- Use `cp -n` (no clobber) to avoid overwriting existing files
- Don't commit automatically — let user review first
- Only create staging branch if it doesn't exist
- All user interactions MUST use `AskUserQuestion` tool, never plain text questions
- Never delete existing workflow files — only add missing ones

## Keeping Up to Date

There are no separate `/update-hooks`, `/workflow-sync`, or `/sync` skills. All updates go through a single command:

```bash
blueprint update
```

This updates the CLI binary, skills, hooks, and templates to the latest release. Run it anytime — it's safe and idempotent.

Use $ARGUMENTS for any additional context.
