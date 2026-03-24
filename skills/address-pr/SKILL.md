---
name: address-pr
description: >
  Fetch PR review comments, categorize feedback, implement fixes, and push.
  Use this skill whenever the user says "/address-pr", "address review", "fix review comments",
  "handle PR feedback", or any request to implement fixes based on PR review feedback.
  Also triggers on "fix the review", "address feedback", "implement review changes",
  "implement the requested changes", or "handle review comments". Fetches real comments via blueprint pr-review.
---

# Address PR: Implement Review Fixes

## Language

Read `blueprint/.config.yml` → `language`. If `auto`, detect from the user's messages. All generated content MUST be in the detected language. Skill instructions stay in English — only output changes.

Address PR review feedback — fetch review comments, create a fix plan, and implement.

```
/plan → /plan-review → /plan-approved → /plan-check → /pr → /review → /address-pr → /finish
```

## Critical Rules

- You MUST read the actual review comments at Step 2. NEVER fabricate feedback.
- You MUST read affected files before modifying them. NEVER edit blind.
- You MUST use `AskUserQuestion` at Step 4 to confirm approach.
- You MUST run tests after fixes. NEVER skip testing.
- Follow steps in order.

## Step 1: Get PR Number

Run: `echo "🔷 BP: address-pr [1/4] fetching PR number"`

```bash
PR_NUMBER="${ARGUMENTS:-$(gh pr view --json number -q .number)}"
```

## Step 2: Fetch Review Data — MANDATORY

Run: `echo "🔷 BP: address-pr [2/4] reading review comments"`

```bash
~/.blueprint/bin/blueprint pr-review "$PR_NUMBER"
```

DO NOT fabricate review feedback. DO NOT proceed without reading real comments.

If the script fails or returns no comments: use `AskUserQuestion` to inform the user ("No review comments found for PR #XX — nothing to address. Want to run /review first, or provide a PR number?") and STOP.

## Step 3: Read Affected Files — MANDATORY

For each file mentioned in review comments, **READ the current code**.

DO NOT modify any file you haven't read first.

## Step 4: Present Fix Summary

Run: `echo "🔷 BP: address-pr [3/4] presenting fix summary"`

```
PR #XX Review Summary:
  - Blockers: N items (must fix)
  - Code issues: N items
  - Style fixes: N items
  - Suggestions: N items (optional)
  - Questions: N items (need response)
```

**STOP. Use `AskUserQuestion`:**
- **Fix all** — Address everything including suggestions
- **Blockers + issues only** — Skip optional suggestions
- **Let me choose** — User picks which items

## Step 5: Implement Fixes

For each approved fix:
1. Read the affected file(s)
2. Make the change — **all fixes MUST follow project conventions**
3. Run targeted tests

**Only modify files mentioned in review comments.** Do not include unrelated changes, refactors, or improvements not requested by reviewers.

For any reviewer **questions**: post a reply comment on the PR using `gh api` with the answer or clarification, then note it in the summary.

## Step 6: Update Plan (if exists)

If `blueprint/live/` has a matching plan: add "PR Review Fixes" section, mark fixes `[x]` with timestamp.

## Step 7: Commit & Push

Run: `echo "🔷 BP: address-pr [4/4] committing fixes"`

Use `/ship` with: `🐛 fix: address PR #XX review feedback` (include actual PR number)

After push, verify PR checks are passing:
```bash
gh pr checks "$PR_NUMBER" --watch --fail-fast 2>/dev/null || gh pr checks "$PR_NUMBER"
```

## Step 7b: Trigger Re-review

After pushing fixes, automatically trigger a new review to verify the fixes:

```bash
PR_NUM=$(gh pr view --json number -q '.number')
gh pr comment "$PR_NUM" --body "@claude review this PR and check if we are able to merge. Analyze the code changes for any issues, security concerns, or improvements needed."
```

This tightens the feedback loop — no need to manually run `/review` after addressing feedback.

## Step 8: Report & Next Action — MANDATORY

**STOP. Use `AskUserQuestion`:**
- **Question:** "All review fixes pushed and re-review triggered. What's next?"
- **Option 1:** "Wait for review results" — Check review status in a few minutes
- **Option 2:** "Finish & merge" — Run /finish (skip waiting for review)
- **Option 3:** "Done for now"

Use $ARGUMENTS as PR number if provided.
