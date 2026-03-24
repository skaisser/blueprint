#!/usr/bin/env bash
set -euo pipefail

# ─────────────────────────────────────────────
# worktree-install-deps.sh — Install dependencies based on detected stack
# Usage: worktree-install-deps.sh [path]
#
# Detects the project's package manager(s) from manifest files and installs:
#   composer.json     → composer install
#   package.json      → npm ci
#   go.mod            → go mod download
#   Gemfile           → bundle install
#   requirements.txt  → pip install -r requirements.txt
#   Pipfile           → pipenv install
#   pyproject.toml    → pip install -e . (or poetry install if pyproject has [tool.poetry])
# ─────────────────────────────────────────────

# ── Colors ───────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RESET='\033[0m'

info()    { echo -e "${CYAN}[install-deps]${RESET} $*"; }
success() { echo -e "${GREEN}[install-deps]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[install-deps]${RESET} $*"; }
error()   { echo -e "${RED}[install-deps]${RESET} $*" >&2; }

# ── Help ─────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") [path]

Install dependencies for a project based on its detected stack.

Arguments:
  path   Directory to install in (default: current working directory)

Supported stacks (auto-detected from manifest files):
  composer.json      PHP / Laravel    → composer install
  package.json       Node.js          → npm ci
  go.mod             Go               → go mod download
  Gemfile            Ruby             → bundle install
  requirements.txt   Python           → pip install -r requirements.txt
  Pipfile            Python (pipenv)  → pipenv install
  pyproject.toml     Python (poetry)  → poetry install  (or pip install -e .)

Options:
  -h, --help  Show this help message

Exit codes:
  0  All detected installers succeeded
  1  One or more installers failed

Examples:
  $(basename "$0")
  $(basename "$0") /path/to/worktree
  $(basename "$0") ~/Sites/my-project
EOF
}

# ── Argument parsing ─────────────────────────
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

DIR="${1:-.}"

if [[ ! -d "$DIR" ]]; then
  error "Directory not found: $DIR"
  exit 1
fi

# Resolve to absolute path
DIR=$(cd "$DIR" && pwd)
info "Installing dependencies in: ${DIR}"
echo ""

cd "$DIR"

# ── Tracking ─────────────────────────────────
INSTALLED=0
FAILED=0
SKIPPED=0

run_installer() {
  local label="$1"
  local check_cmd="$2"
  shift 2
  local cmd=("$@")

  if ! command -v "${cmd[0]}" &>/dev/null; then
    warn "  ${label}: skipping — '${cmd[0]}' not found in PATH"
    SKIPPED=$((SKIPPED + 1))
    return
  fi

  info "  ${label}: running '${cmd[*]}'..."
  if "${cmd[@]}"; then
    success "  ${label}: done."
    INSTALLED=$((INSTALLED + 1))
  else
    error "  ${label}: FAILED (exit $?)"
    FAILED=$((FAILED + 1))
  fi
}

# ── PHP / Composer ───────────────────────────
if [[ -f "composer.json" ]]; then
  info "Detected: composer.json (PHP/Laravel)"
  # Prefer herd php if available (per project conventions)
  if command -v herd &>/dev/null; then
    run_installer "Composer (herd)" "" herd composer install --no-interaction --prefer-dist
  else
    run_installer "Composer" "" composer install --no-interaction --prefer-dist
  fi
  echo ""
fi

# ── Node.js / npm ────────────────────────────
if [[ -f "package.json" ]]; then
  info "Detected: package.json (Node.js)"
  # Use npm ci if lock file exists (faster, reproducible)
  if [[ -f "package-lock.json" ]]; then
    run_installer "npm" "" npm ci
  elif [[ -f "yarn.lock" ]]; then
    run_installer "Yarn" "" yarn install --frozen-lockfile
  elif [[ -f "pnpm-lock.yaml" ]]; then
    run_installer "pnpm" "" pnpm install --frozen-lockfile
  else
    warn "  Node.js: no lock file found, using 'npm install'"
    run_installer "npm" "" npm install
  fi
  echo ""
fi

# ── Go ───────────────────────────────────────
if [[ -f "go.mod" ]]; then
  info "Detected: go.mod (Go)"
  run_installer "Go modules" "" go mod download
  echo ""
fi

# ── Ruby / Bundler ───────────────────────────
if [[ -f "Gemfile" ]]; then
  info "Detected: Gemfile (Ruby)"
  run_installer "Bundler" "" bundle install
  echo ""
fi

# ── Python / pip ─────────────────────────────
if [[ -f "pyproject.toml" ]]; then
  info "Detected: pyproject.toml (Python)"
  # Check for Poetry
  if grep -q '\[tool\.poetry\]' pyproject.toml 2>/dev/null; then
    run_installer "Poetry" "" poetry install
  else
    run_installer "pip (pyproject)" "" pip install -e .
  fi
  echo ""
elif [[ -f "Pipfile" ]]; then
  info "Detected: Pipfile (Python/pipenv)"
  run_installer "pipenv" "" pipenv install
  echo ""
elif [[ -f "requirements.txt" ]]; then
  info "Detected: requirements.txt (Python)"
  run_installer "pip" "" pip install -r requirements.txt
  echo ""
fi

# ── Summary ──────────────────────────────────
TOTAL=$((INSTALLED + FAILED + SKIPPED))

if [[ $TOTAL -eq 0 ]]; then
  warn "No supported manifest files detected in: ${DIR}"
  warn "Checked: composer.json, package.json, go.mod, Gemfile, requirements.txt, Pipfile, pyproject.toml"
  exit 0
fi

echo "─────────────────────────────────────────"
info "Summary: ${INSTALLED} installed, ${FAILED} failed, ${SKIPPED} skipped"

if [[ $FAILED -gt 0 ]]; then
  error "${FAILED} installer(s) failed. Check output above."
  exit 1
fi

success "All dependencies installed successfully."
exit 0
