#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
# test_scripts.sh — Test runner for scripts/
# Tests argument parsing, help flags, and logic with mock/minimal data.
# Does NOT make real gh API calls or real git merges.
# ─────────────────────────────────────────────

SCRIPT_FILE="${BASH_SOURCE[0]}"
# Resolve to absolute path if relative
if [[ "$SCRIPT_FILE" != /* ]]; then
  SCRIPT_FILE="$(pwd)/${SCRIPT_FILE}"
fi
SCRIPTS_DIR="$(cd "$(dirname "$SCRIPT_FILE")/../../scripts" && pwd)"
PASS=0
FAIL=0
SKIP=0

# Temp file for cross-subshell failure tracking
FAIL_FILE=$(mktemp)
trap 'rm -f "$FAIL_FILE"' EXIT

# ── Colors ───────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Test helpers ─────────────────────────────
pass() { echo -e "  ${GREEN}PASS${RESET} $*"; PASS=$((PASS + 1)); }
fail() {
  echo -e "  ${RED}FAIL${RESET} $*"
  FAIL=$((FAIL + 1))
  echo "1" >> "$FAIL_FILE"
}
skip() { echo -e "  ${YELLOW}SKIP${RESET} $*"; SKIP=$((SKIP + 1)); }
section() { echo ""; echo -e "${BOLD}${CYAN}── $* ──${RESET}"; }

assert_exit() {
  local expected="$1"
  local desc="$2"
  shift 2
  local actual=0
  "$@" &>/dev/null || actual=$?
  if [[ "$actual" -eq "$expected" ]]; then
    pass "$desc (exit $actual)"
  else
    fail "$desc (expected exit $expected, got $actual)"
  fi
}

assert_output_contains() {
  local pattern="$1"
  local desc="$2"
  shift 2
  local output=""
  output=$("$@" 2>&1) || true
  if echo "$output" | grep -q "$pattern"; then
    pass "$desc"
  else
    fail "$desc (pattern '${pattern}' not found in output)"
    echo "    Output was: $(echo "$output" | head -5)"
  fi
}

# ────────────────────────────────────────────────────────────────
section "ci-trigger-wait.sh"
# ────────────────────────────────────────────────────────────────

CI_SCRIPT="${SCRIPTS_DIR}/ci-trigger-wait.sh"

# Script is executable
if [[ -x "$CI_SCRIPT" ]]; then
  pass "ci-trigger-wait.sh is executable"
else
  fail "ci-trigger-wait.sh is not executable"
fi

# Help flag exits 0 and prints usage
assert_exit 0 "-h flag exits 0" bash "$CI_SCRIPT" -h
assert_exit 0 "--help flag exits 0" bash "$CI_SCRIPT" --help
assert_output_contains "Usage:" "-h prints usage" bash "$CI_SCRIPT" -h
assert_output_contains "PR_NUMBER" "-h mentions PR_NUMBER" bash "$CI_SCRIPT" -h
assert_output_contains "Exit codes:" "-h shows exit codes" bash "$CI_SCRIPT" -h

# Missing argument exits 1
assert_exit 1 "no args exits 1" bash "$CI_SCRIPT"

# Non-numeric PR_NUMBER exits 1
assert_exit 1 "non-numeric PR exits 1" bash "$CI_SCRIPT" "abc"
assert_exit 1 "empty PR exits 1" bash "$CI_SCRIPT" ""

# gh missing: simulate by overriding PATH
assert_exit 1 "missing gh exits 1" env PATH="/usr/bin:/bin" bash "$CI_SCRIPT" "42"

# Valid PR number (but gh not authenticated — expect non-zero)
# This tests that the script reaches the gh call and fails gracefully
if ! command -v gh &>/dev/null; then
  skip "gh not installed — skipping live PR validation test"
else
  # We expect exit 1 because PR #0 doesn't exist
  assert_exit 1 "invalid PR number fails gracefully" bash "$CI_SCRIPT" "0"
fi

# ────────────────────────────────────────────────────────────────
section "pre-merge-check.sh"
# ────────────────────────────────────────────────────────────────

MERGE_SCRIPT="${SCRIPTS_DIR}/pre-merge-check.sh"
TMPDIR_MERGE=$(mktemp -d)
trap 'rm -rf "$TMPDIR_MERGE"' EXIT

# Script is executable
if [[ -x "$MERGE_SCRIPT" ]]; then
  pass "pre-merge-check.sh is executable"
else
  fail "pre-merge-check.sh is not executable"
fi

# Help flag
assert_exit 0 "-h flag exits 0" bash "$MERGE_SCRIPT" -h
assert_exit 0 "--help flag exits 0" bash "$MERGE_SCRIPT" --help
assert_output_contains "BASE_BRANCH" "-h mentions BASE_BRANCH" bash "$MERGE_SCRIPT" -h
assert_output_contains "Exit codes:" "-h shows exit codes" bash "$MERGE_SCRIPT" -h

# Missing argument
assert_exit 1 "no args exits 1" bash "$MERGE_SCRIPT"

# Not in a git repo exits 1
(
  cd "$TMPDIR_MERGE"
  assert_exit 1 "non-git dir exits 1" bash "$MERGE_SCRIPT" "main"
)

# In a git repo with no remote — fetch should fail with exit 1
TMPGIT=$(mktemp -d)
trap 'rm -rf "$TMPGIT" "$TMPDIR_MERGE"' EXIT
(
  cd "$TMPGIT"
  git init -q
  git config user.email "test@test.com"
  git config user.name "Test"
  echo "initial" > README.md
  git add README.md
  git commit -q -m "init"
  # No remote: fetch should fail
  assert_exit 1 "fetch failure exits 1" bash "$MERGE_SCRIPT" "main"
)

# In a git repo with uncommitted changes exits 1
TMPGIT2=$(mktemp -d)
trap 'rm -rf "$TMPGIT2" "$TMPDIR_MERGE" "$TMPGIT"' EXIT
(
  cd "$TMPGIT2"
  git init -q
  git config user.email "test@test.com"
  git config user.name "Test"
  echo "initial" > README.md
  git add README.md
  git commit -q -m "init"
  echo "dirty" > dirty.txt
  # Uncommitted changes — should exit 1
  assert_exit 1 "dirty working tree exits 1" bash "$MERGE_SCRIPT" "main"
)

# ────────────────────────────────────────────────────────────────
section "plan-check-orphans.sh"
# ────────────────────────────────────────────────────────────────

ORPHAN_SCRIPT="${SCRIPTS_DIR}/plan-check-orphans.sh"
TMPDIR_ORPHAN=$(mktemp -d)
trap 'rm -rf "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT

# Script is executable
if [[ -x "$ORPHAN_SCRIPT" ]]; then
  pass "plan-check-orphans.sh is executable"
else
  fail "plan-check-orphans.sh is not executable"
fi

# Help flag
assert_exit 0 "-h flag exits 0" bash "$ORPHAN_SCRIPT" -h
assert_exit 0 "--help flag exits 0" bash "$ORPHAN_SCRIPT" --help
assert_output_contains "COMMIT" "-h mentions COMMIT" bash "$ORPHAN_SCRIPT" -h
assert_output_contains "pattern" "-h mentions pattern" bash "$ORPHAN_SCRIPT" -h
assert_output_contains "Exit codes:" "-h shows exit codes" bash "$ORPHAN_SCRIPT" -h

# Missing arguments
assert_exit 1 "no args exits 1" bash "$ORPHAN_SCRIPT"
assert_exit 1 "only commit, no patterns exits 1" bash "$ORPHAN_SCRIPT" "HEAD"

# Not in a git repo
(
  cd "$TMPDIR_ORPHAN"
  assert_exit 1 "non-git dir exits 1" bash "$ORPHAN_SCRIPT" "HEAD" "SomeClass"
)

# In a real git repo with a pattern not in tests → exit 0 (nothing found = not orphaned)
REPO_ROOT="$(cd "$(dirname "$SCRIPT_FILE")/../../" && pwd)"
(
  cd "$REPO_ROOT"
  # Use a pattern highly unlikely to be in tests
  assert_exit 0 "nonexistent pattern exits 0 (nothing orphaned)" \
    bash "$ORPHAN_SCRIPT" "HEAD" "ZZZNOTAREALCLASSZZZ_9999"
)

# Create a temp git repo to test orphan detection
TMPGIT3=$(mktemp -d)
trap 'rm -rf "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
(
  cd "$TMPGIT3"
  git init -q
  git config user.email "test@test.com"
  git config user.name "Test"

  # Create a source file
  mkdir -p app tests
  echo "class MyService {}" > app/MyService.php

  # Create test referencing the source file
  echo "use App\MyService; // test" > tests/MyServiceTest.php

  git add .
  git commit -q -m "initial"

  # Pattern exists in tests AND source → no orphan → exit 0
  assert_exit 0 "existing class not orphaned" bash "$ORPHAN_SCRIPT" "HEAD" "MyService"

  # Delete the source file and commit
  rm app/MyService.php
  git add app/MyService.php
  git commit -q -m "remove service"

  # Now MyService is referenced in tests but source is gone → orphan → exit 1
  assert_exit 1 "deleted source is detected as orphan" bash "$ORPHAN_SCRIPT" "HEAD" "MyService"
)

# ────────────────────────────────────────────────────────────────
section "worktree-install-deps.sh"
# ────────────────────────────────────────────────────────────────

DEPS_SCRIPT="${SCRIPTS_DIR}/worktree-install-deps.sh"

# Script is executable
if [[ -x "$DEPS_SCRIPT" ]]; then
  pass "worktree-install-deps.sh is executable"
else
  fail "worktree-install-deps.sh is not executable"
fi

# Help flag
assert_exit 0 "-h flag exits 0" bash "$DEPS_SCRIPT" -h
assert_exit 0 "--help flag exits 0" bash "$DEPS_SCRIPT" --help
assert_output_contains "path" "-h mentions path" bash "$DEPS_SCRIPT" -h
assert_output_contains "composer.json" "-h mentions composer.json" bash "$DEPS_SCRIPT" -h
assert_output_contains "package.json" "-h mentions package.json" bash "$DEPS_SCRIPT" -h
assert_output_contains "go.mod" "-h mentions go.mod" bash "$DEPS_SCRIPT" -h
assert_output_contains "Gemfile" "-h mentions Gemfile" bash "$DEPS_SCRIPT" -h
assert_output_contains "requirements.txt" "-h mentions requirements.txt" bash "$DEPS_SCRIPT" -h
assert_output_contains "Exit codes:" "-h shows exit codes" bash "$DEPS_SCRIPT" -h

# Non-existent path exits 1
assert_exit 1 "nonexistent path exits 1" bash "$DEPS_SCRIPT" "/nonexistent/path/zzzz"

# Empty dir (no manifests) exits 0 with warning
TMPDIR_EMPTY=$(mktemp -d)
trap 'rm -rf "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
assert_exit 0 "empty dir exits 0 (no manifests)" bash "$DEPS_SCRIPT" "$TMPDIR_EMPTY"
assert_output_contains "No supported manifest" "empty dir warns no manifests" bash "$DEPS_SCRIPT" "$TMPDIR_EMPTY"

# go.mod detection — go should be available in PATH
TMPDIR_GO=$(mktemp -d)
trap 'rm -rf "$TMPDIR_GO" "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
cat > "$TMPDIR_GO/go.mod" <<'EOF'
module example.com/test
go 1.22
EOF
if command -v go &>/dev/null; then
  assert_exit 0 "go.mod detected and go mod download runs" bash "$DEPS_SCRIPT" "$TMPDIR_GO"
else
  skip "go not installed — skipping go.mod install test"
fi

# package.json detection — no lock file, use npm install
TMPDIR_NODE=$(mktemp -d)
trap 'rm -rf "$TMPDIR_NODE" "$TMPDIR_GO" "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
echo '{"name":"test","version":"1.0.0"}' > "$TMPDIR_NODE/package.json"
if command -v npm &>/dev/null; then
  assert_output_contains "package.json" "package.json detected" bash "$DEPS_SCRIPT" "$TMPDIR_NODE"
else
  skip "npm not installed — skipping package.json detection test"
fi

# composer.json detection — composer may not be installed
TMPDIR_COMPOSER=$(mktemp -d)
trap 'rm -rf "$TMPDIR_COMPOSER" "$TMPDIR_NODE" "$TMPDIR_GO" "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
echo '{"name":"test/project","require":{}}' > "$TMPDIR_COMPOSER/composer.json"
assert_output_contains "composer.json" "composer.json detected" bash "$DEPS_SCRIPT" "$TMPDIR_COMPOSER"

# requirements.txt detection
TMPDIR_PY=$(mktemp -d)
trap 'rm -rf "$TMPDIR_PY" "$TMPDIR_COMPOSER" "$TMPDIR_NODE" "$TMPDIR_GO" "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"' EXIT
echo "requests>=2.0" > "$TMPDIR_PY/requirements.txt"
assert_output_contains "requirements.txt" "requirements.txt detected" bash "$DEPS_SCRIPT" "$TMPDIR_PY"

# Gemfile detection
TMPDIR_RUBY=$(mktemp -d)
trap 'rm -rf "$TMPDIR_RUBY" "$TMPDIR_PY" "$TMPDIR_COMPOSER" "$TMPDIR_NODE" "$TMPDIR_GO" "$TMPDIR_EMPTY" "$TMPGIT3" "$TMPDIR_ORPHAN" "$TMPDIR_MERGE" "$TMPGIT" "$TMPGIT2"; rm -f "$FAIL_FILE"' EXIT
echo "source 'https://rubygems.org'" > "$TMPDIR_RUBY/Gemfile"
assert_output_contains "Gemfile" "Gemfile detected" bash "$DEPS_SCRIPT" "$TMPDIR_RUBY"

# ────────────────────────────────────────────────────────────────
section "General: all scripts have correct shebang + set -euo pipefail"
# ────────────────────────────────────────────────────────────────

for script in "$SCRIPTS_DIR"/*.sh; do
  name=$(basename "$script")
  first_line=$(head -1 "$script")
  if [[ "$first_line" == "#!/usr/bin/env bash" ]]; then
    pass "${name}: correct shebang"
  else
    fail "${name}: missing or wrong shebang (got: $first_line)"
  fi

  if grep -q 'set -euo pipefail' "$script"; then
    pass "${name}: has set -euo pipefail"
  else
    fail "${name}: missing set -euo pipefail"
  fi

  if [[ -x "$script" ]]; then
    pass "${name}: is executable"
  else
    fail "${name}: not executable"
  fi
done

# ── Final summary ─────────────────────────────
# Count failures written from subshells
SUBSHELL_FAILS=$(wc -l < "$FAIL_FILE" 2>/dev/null | tr -d ' ') || SUBSHELL_FAILS=0
FAIL=$((FAIL + SUBSHELL_FAILS))

echo ""
echo "─────────────────────────────────────────────────"
TOTAL=$((PASS + FAIL + SKIP))
echo -e "${BOLD}Results: ${GREEN}${PASS} passed${RESET}, ${RED}${FAIL} failed${RESET}, ${YELLOW}${SKIP} skipped${RESET} (${TOTAL} total)"

if [[ $FAIL -gt 0 ]]; then
  exit 1
fi
exit 0
