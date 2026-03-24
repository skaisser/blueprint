---
name: bp-commit
description: >
  Commit current changes following emoji+type conventions with code formatting.
  Triggers on "/bp-commit", "/commit", "commit this", "commit changes",
  "save my changes", or any request to create a git commit.
  Also triggers on "stage and commit", "commit with message", "make a commit",
  "let's commit", or "save this progress".
  Do NOT trigger when the user also mentions pushing — route to /bp-ship instead.
  Runs code formatting on changed files before committing.
---

# Commit: Stage and Commit Changes

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Commit the current changes following our conventions. **Commit only — do not push to remote.**

## Process

1. Review changes: `git status` and `git diff`
2. Stage only necessary files (never `git add -A` blindly)
3. NEVER stage `.env`, `.env.*`, `*-api-key`, `credentials.json`, or files containing API keys/tokens
4. Run code formatting if the project has a formatter configured (check project conventions)
5. Write a concise 1-2 sentence commit message focusing on **why** the change was made, not just what changed
6. Commit with emoji format: `<emoji> <type>: <description>`

## Emoji Types

| Emoji | Type | Use |
|-------|------|-----|
| ✨ | feat | New features |
| 🐛 | fix | Bug fixes |
| 📚 | docs | Documentation |
| 💄 | style | Formatting |
| ♻️ | refactor | Restructuring |
| ⚡ | perf | Performance |
| 🧪 | test | Testing |
| 🔧 | build | Build changes |
| 🧹 | chore | Maintenance |
| 📋 | plan | Planning |
| 🔒 | security | Security |
| 🗃️ | migration | DB migrations |
| 📦 | deps | Dependencies |
| 🔀 | merge | Branch merges |

Present tense, lowercase, atomic commits. NEVER add AI signatures.

Use $ARGUMENTS as context for the commit message if provided.
