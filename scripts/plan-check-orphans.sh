#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
# plan-check-orphans.sh — Detect orphaned test references
# Usage: plan-check-orphans.sh <COMMIT> <patterns...>
#
# Searches test files for references matching the given patterns,
# then checks if the referenced files/classes still exist.
# Reports any orphaned references found.
# ─────────────────────────────────────────────

# ── Colors ───────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

info()    { echo -e "${CYAN}[orphan-check]${RESET} $*"; }
success() { echo -e "${GREEN}[orphan-check]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[orphan-check]${RESET} $*"; }
error()   { echo -e "${RED}[orphan-check]${RESET} $*" >&2; }

# ── Help ─────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") <COMMIT> <pattern> [pattern...]

Detect orphaned test references — patterns used in test files that no longer
correspond to existing source files or classes.

Arguments:
  COMMIT      Git commit hash to check files against (HEAD, SHA, etc.)
  pattern     One or more file/class name patterns to search for in tests/
              Patterns are matched as literal substrings via git grep.

Options:
  -h, --help  Show this help message

Exit codes:
  0  No orphaned references found
  1  One or more orphaned references detected

Examples:
  $(basename "$0") HEAD UserService PaymentService
  $(basename "$0") abc1234 "app/Services/Billing"
  $(basename "$0") HEAD --patterns-from-diff
EOF
}

# ── Argument parsing ─────────────────────────
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 2 ]]; then
  error "Missing required arguments: COMMIT and at least one pattern"
  usage >&2
  exit 1
fi

COMMIT="$1"
shift
PATTERNS=("$@")

# ── Ensure we're in a git repo ───────────────
if ! git rev-parse --git-dir &>/dev/null; then
  error "Not in a git repository."
  exit 1
fi

# Validate commit
if ! git rev-parse --verify "$COMMIT" &>/dev/null; then
  error "Invalid commit reference: $COMMIT"
  exit 1
fi

# ── Determine test directories to search ────
# Search for common test directory patterns
TEST_DIRS=()
for candidate in tests/ test/ spec/ __tests__/; do
  if [[ -d "$candidate" ]]; then
    TEST_DIRS+=("$candidate")
  fi
done

if [[ ${#TEST_DIRS[@]} -eq 0 ]]; then
  warn "No test directories found (tests/, test/, spec/, __tests__/). Searching all files."
  TEST_SEARCH_ARGS=()
else
  TEST_SEARCH_ARGS=("${TEST_DIRS[@]}")
  info "Searching in: ${TEST_DIRS[*]}"
fi

# ── Process each pattern ─────────────────────
ORPHAN_COUNT=0
TOTAL_MATCHES=0

for pattern in "${PATTERNS[@]}"; do
  info "Checking pattern: '${pattern}'"

  # Find test files that reference this pattern
  MATCHING_FILES=()
  while IFS= read -r line; do
    [[ -n "$line" ]] && MATCHING_FILES+=("$line")
  done < <(git grep -l "$pattern" -- "${TEST_SEARCH_ARGS[@]+"${TEST_SEARCH_ARGS[@]}"}" 2>/dev/null || true)

  if [[ ${#MATCHING_FILES[@]} -eq 0 ]]; then
    info "  No test files reference '${pattern}'"
    continue
  fi

  TOTAL_MATCHES=$((TOTAL_MATCHES + ${#MATCHING_FILES[@]}))
  info "  Found in ${#MATCHING_FILES[@]} test file(s): ${MATCHING_FILES[*]}"

  # Now check if the referenced file/class actually exists at $COMMIT
  # Strategy: treat pattern as a possible file path fragment and class name.
  # Importantly: we EXCLUDE test directories from source existence checks to
  # avoid counting the test file itself as evidence the source exists.
  FILE_EXISTS=false

  # Check 1: Does any NON-TEST file path at $COMMIT contain this pattern?
  SOURCE_FILES_AT_COMMIT=$(git ls-tree -r --name-only "$COMMIT" 2>/dev/null \
    | grep -v -E '^(tests|test|spec|__tests__)/' \
    | grep "$pattern" || true)
  if [[ -n "$SOURCE_FILES_AT_COMMIT" ]]; then
    FILE_EXISTS=true
  fi

  # Check 2: For class-like patterns, check if any source file defines it
  # (searches committed content, not just file names)
  if ! $FILE_EXISTS; then
    # Look for class/function/interface definitions in source files (not tests)
    if git grep -l "class ${pattern}\b\|interface ${pattern}\b\|def ${pattern}\b\|func ${pattern}\b\|func.*${pattern}" \
       -- ':!tests/' ':!test/' ':!spec/' ':!__tests__/' &>/dev/null 2>/dev/null; then
      FILE_EXISTS=true
    fi
  fi

  if $FILE_EXISTS; then
    success "  '${pattern}' — referenced in tests and still exists. OK."
  else
    warn "  ORPHAN: '${pattern}' is referenced in tests but NOT found in source at ${COMMIT}"
    echo ""
    warn "  Test files with orphaned reference to '${pattern}':"
    for f in "${MATCHING_FILES[@]}"; do
      # Show the specific lines referencing the pattern
      echo -e "    ${YELLOW}${f}${RESET}"
      git grep -n "$pattern" -- "$f" 2>/dev/null | head -5 | sed 's/^/      /' || true
    done
    echo ""
    ORPHAN_COUNT=$((ORPHAN_COUNT + 1))
  fi
done

# ── Summary ──────────────────────────────────
echo ""
if [[ $ORPHAN_COUNT -eq 0 ]]; then
  success "No orphaned references found. All ${#PATTERNS[@]} pattern(s) checked."
  exit 0
else
  error "Found ${ORPHAN_COUNT} orphaned reference(s) across ${TOTAL_MATCHES} test file(s)."
  error "Review the above and remove or update stale test references."
  exit 1
fi
