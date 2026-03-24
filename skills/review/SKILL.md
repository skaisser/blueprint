---
name: review
description: >
  Trigger @claude PR review on a Pull Request via GitHub Actions.
  Use this skill whenever the user says "/review", "review the PR", "code review",
  "trigger review", or any request to get an automated code review on a PR.
  Also triggers on "claude review", "run review", or "check the PR".
  Posts a comment mentioning @claude which triggers the GitHub Action.
---

# Review: Trigger PR Code Review

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Trigger mention-based PR review on a Pull Request.

**CRITICAL GUARDRAILS:**
- Do NOT perform the review yourself — this skill only posts a comment to trigger the CI-based review
- Do NOT modify any code files, project files, or configuration
- Do NOT merge, approve, or close the PR
- Do NOT run tests or make any changes to the codebase

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Step 1: Get PR Number and Trigger Review

- If `$ARGUMENTS` provided, use it as PR number
- Otherwise: `gh pr view --json number -q .number`

If no PR found, STOP: "No open PR found. Create one with /pr first."

Then trigger the review:

```bash
gh pr comment <PR_NUMBER> --body "@claude review this PR and check if we are able to merge. Analyze the code changes for any issues, security concerns, or improvements needed."
```

If the comment fails (non-zero exit), STOP and report the error.

## Step 2: Verify Action Triggered

Check that the GitHub Action workflow was triggered:

```bash
gh run list --workflow=claude-pr-reviewer.yml --limit=1 --json status,createdAt,event -q '.[0]'
```

Report the run status. If no run found, note that the action may take a moment to start or may not be configured.

## Step 3: Confirm and Ask

Confirm review was triggered and provide the PR link. Review usually appears within 2-5 minutes.

**STOP. You MUST use `AskUserQuestion` tool here.**

- **Question:** "Review triggered on PR #N. Ready to address feedback?"
- **Option 1:** "Address review now" — Run /address-pr
- **Option 2:** "I'll check the review first"

Use $ARGUMENTS as PR number if provided.
