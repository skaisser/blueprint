#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
# pre-merge-check.sh — Conflict detection + auto-resolve for blueprint/ directory
# Usage: pre-merge-check.sh <BASE_BRANCH>
#
# Fetches base branch, counts commits behind, attempts merge,
# auto-resolves blueprint/ conflicts with "ours" strategy.
# Exit 0 if clean or auto-resolved, exit 1 if unresolvable conflicts remain.
# ─────────────────────────────────────────────

# ── Colors ───────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

info()    { echo -e "${CYAN}[pre-merge]${RESET} $*"; }
success() { echo -e "${GREEN}[pre-merge]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[pre-merge]${RESET} $*"; }
error()   { echo -e "${RED}[pre-merge]${RESET} $*" >&2; }

# ── Help ─────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") <BASE_BRANCH>

Check for merge conflicts before merging into a base branch.
Automatically resolves conflicts in the blueprint/ directory (accepts ours).

Arguments:
  BASE_BRANCH  The branch to merge from (e.g. staging, main)

Options:
  -h, --help   Show this help message

Exit codes:
  0  No conflicts, or all conflicts auto-resolved
  1  Unresolvable conflicts remain (manual intervention required)

Examples:
  $(basename "$0") staging
  $(basename "$0") main
EOF
}

# ── Argument parsing ─────────────────────────
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 1 ]]; then
  error "Missing required argument: BASE_BRANCH"
  usage >&2
  exit 1
fi

BASE="$1"
REMOTE="origin"

# ── Ensure we're in a git repo ───────────────
if ! git rev-parse --git-dir &>/dev/null; then
  error "Not in a git repository."
  exit 1
fi

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
info "Current branch: ${CURRENT_BRANCH}"
info "Base branch:    ${REMOTE}/${BASE}"

# ── Stash check ──────────────────────────────
if ! git diff --quiet || ! git diff --cached --quiet; then
  error "Working tree has uncommitted changes. Please commit or stash before running pre-merge-check."
  exit 1
fi

# ── Fetch base branch ────────────────────────
info "Fetching ${REMOTE}/${BASE}..."
git fetch "$REMOTE" "$BASE" 2>&1 | sed "s/^/  /" || {
  error "Failed to fetch ${REMOTE}/${BASE}. Check remote name and branch existence."
  exit 1
}

# ── Count commits behind ─────────────────────
BEHIND=$(git rev-list --count HEAD.."${REMOTE}/${BASE}" 2>/dev/null) || BEHIND=0
AHEAD=$(git rev-list --count "${REMOTE}/${BASE}"..HEAD 2>/dev/null) || AHEAD=0

info "Branch is ${AHEAD} commit(s) ahead, ${BEHIND} commit(s) behind ${REMOTE}/${BASE}."

if [[ "$BEHIND" -eq 0 ]]; then
  success "Already up to date with ${REMOTE}/${BASE}. No merge needed."
  exit 0
fi

# ── Attempt merge ────────────────────────────
info "Attempting merge of ${REMOTE}/${BASE} into ${CURRENT_BRANCH}..."

MERGE_OUTPUT=""
MERGE_EXIT=0
MERGE_OUTPUT=$(git merge "origin/${BASE}" --no-edit --no-commit 2>&1) || MERGE_EXIT=$?

# ── Check for conflicts ──────────────────────
CONFLICTED_FILES=$(git diff --name-only --diff-filter=U 2>/dev/null) || CONFLICTED_FILES=""

if [[ -z "$CONFLICTED_FILES" && $MERGE_EXIT -eq 0 ]]; then
  # Clean merge — commit it
  if git diff --cached --quiet; then
    success "Merge resulted in no changes."
  else
    git merge --continue --no-edit 2>/dev/null || git commit --no-edit
    success "Merge completed cleanly."
  fi
  exit 0
fi

if [[ -n "$CONFLICTED_FILES" ]]; then
  info "Conflicts detected in:"
  echo "$CONFLICTED_FILES" | sed 's/^/  /'

  # ── Auto-resolve blueprint/ conflicts ────────
  BLUEPRINT_CONFLICTS=$(echo "$CONFLICTED_FILES" | grep '^blueprint/' || true)
  OTHER_CONFLICTS=$(echo "$CONFLICTED_FILES" | grep -v '^blueprint/' || true)

  if [[ -n "$BLUEPRINT_CONFLICTS" ]]; then
    info "Auto-resolving blueprint/ conflicts (keeping ours)..."
    while IFS= read -r file; do
      if [[ -n "$file" ]]; then
        git checkout --ours -- "$file"
        git add -- "$file"
        success "  Auto-resolved (ours): $file"
      fi
    done <<< "$BLUEPRINT_CONFLICTS"
  fi

  if [[ -n "$OTHER_CONFLICTS" ]]; then
    error "Unresolvable conflicts remain (manual intervention required):"
    echo "$OTHER_CONFLICTS" | sed 's/^/  /'
    warn "Aborting merge. Resolve the above conflicts manually, then re-run."
    git merge --abort 2>/dev/null || true
    exit 1
  fi

  # All conflicts resolved — commit
  git merge --continue --no-edit 2>/dev/null || git commit --no-edit
  success "All conflicts auto-resolved. Merge complete."
  exit 0
fi

# If merge_exit != 0 but no conflicted files, something else went wrong
error "Merge failed unexpectedly. Output:"
echo "$MERGE_OUTPUT" | sed 's/^/  /'
git merge --abort 2>/dev/null || true
exit 1
