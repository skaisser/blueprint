#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
# ci-trigger-wait.sh — Trigger CI and wait for completion
# Usage: ci-trigger-wait.sh <PR_NUMBER> [workflow]
#
# Posts a @tests comment on the PR to trigger CI, then polls
# for the workflow run to complete. Exits 0 on pass, 1 on fail.
# Timeout: 10 minutes.
# ─────────────────────────────────────────────

TIMEOUT_SECONDS=600
POLL_INTERVAL=30
TRIGGER_COMMENT="@tests"

# ── Colors ───────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

info()    { echo -e "${CYAN}[ci-trigger]${RESET} $*"; }
success() { echo -e "${GREEN}[ci-trigger]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[ci-trigger]${RESET} $*"; }
error()   { echo -e "${RED}[ci-trigger]${RESET} $*" >&2; }

# ── Help ─────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") <PR_NUMBER> [workflow]

Trigger CI by posting a comment on a PR, then wait for the workflow to complete.

Arguments:
  PR_NUMBER   GitHub pull request number (required)
  workflow    Workflow filename to watch, e.g. tests.yml (optional)
              If omitted, watches the most recent workflow run on the PR branch.

Options:
  -h, --help  Show this help message

Examples:
  $(basename "$0") 42
  $(basename "$0") 42 tests.yml

Exit codes:
  0  Workflow completed successfully
  1  Workflow failed, was cancelled, or timed out
EOF
}

# ── Argument parsing ─────────────────────────
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 1 ]]; then
  error "Missing required argument: PR_NUMBER"
  usage >&2
  exit 1
fi

PR_NUM="$1"
WORKFLOW="${2:-}"

# Validate PR_NUM is numeric
if ! [[ "$PR_NUM" =~ ^[0-9]+$ ]]; then
  error "PR_NUMBER must be a positive integer, got: $PR_NUM"
  exit 1
fi

# ── Dependency check ─────────────────────────
if ! command -v gh &>/dev/null; then
  error "gh CLI is required but not installed. Install from https://cli.github.com"
  exit 1
fi

# ── Fetch PR branch ──────────────────────────
info "Fetching PR #${PR_NUM} details..."
PR_BRANCH=$(gh pr view "$PR_NUM" --json headRefName --jq '.headRefName' 2>/dev/null) || {
  error "Could not fetch PR #${PR_NUM}. Check the PR number and your gh auth."
  exit 1
}
info "PR branch: ${PR_BRANCH}"

# ── Post trigger comment ─────────────────────
info "Posting trigger comment '${TRIGGER_COMMENT}' on PR #${PR_NUM}..."
gh pr comment "$PR_NUM" --body "$TRIGGER_COMMENT"
success "Comment posted."

# ── Wait for workflow to start ───────────────
info "Waiting 10s for workflow to start..."
sleep 10

# ── Poll for run completion ──────────────────
ELAPSED=0
RUN_ID=""

info "Polling for workflow run (timeout: ${TIMEOUT_SECONDS}s, interval: ${POLL_INTERVAL}s)..."

while [[ $ELAPSED -lt $TIMEOUT_SECONDS ]]; do
  # Build the list command
  LIST_ARGS=(run list --branch "$PR_BRANCH" --limit 5 --json databaseId,status,conclusion,workflowName,headBranch)
  if [[ -n "$WORKFLOW" ]]; then
    LIST_ARGS+=(--workflow "$WORKFLOW")
  fi

  RUNS_JSON=$(gh "${LIST_ARGS[@]}" 2>/dev/null) || {
    warn "Could not fetch runs, retrying in ${POLL_INTERVAL}s..."
    sleep "$POLL_INTERVAL"
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
    continue
  }

  # Pick the most recent run
  RUN_STATUS=$(echo "$RUNS_JSON" | command -p python3 -c "
import sys, json
runs = json.load(sys.stdin)
if not runs:
    print('none')
else:
    r = runs[0]
    print(r.get('status','unknown') + '|' + str(r.get('databaseId','')) + '|' + (r.get('conclusion') or ''))
" 2>/dev/null) || RUN_STATUS="none"

  if [[ "$RUN_STATUS" == "none" ]]; then
    info "No runs found yet (${ELAPSED}s elapsed)..."
    sleep "$POLL_INTERVAL"
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
    continue
  fi

  STATUS=$(echo "$RUN_STATUS" | cut -d'|' -f1)
  RUN_ID=$(echo "$RUN_STATUS" | cut -d'|' -f2)
  CONCLUSION=$(echo "$RUN_STATUS" | cut -d'|' -f3)

  info "Run #${RUN_ID} — status: ${STATUS}, conclusion: ${CONCLUSION:-pending} (${ELAPSED}s elapsed)"

  if [[ "$STATUS" == "completed" ]]; then
    if [[ "$CONCLUSION" == "success" ]]; then
      success "CI passed! Run #${RUN_ID} completed successfully."
      exit 0
    else
      error "CI failed. Run #${RUN_ID} concluded: ${CONCLUSION}"
      exit 1
    fi
  fi

  sleep "$POLL_INTERVAL"
  ELAPSED=$((ELAPSED + POLL_INTERVAL))
done

error "Timeout after ${TIMEOUT_SECONDS}s. Last run status: ${STATUS:-unknown}"
exit 1
