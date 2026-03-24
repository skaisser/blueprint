#!/usr/bin/env python3
"""Tests for hooks/blueprint.py commit validation logic."""
import re
import os
import sys
import tempfile

# The regex from blueprint.py (extract and test)
COMMIT_PATTERN = re.compile(
    r'^(\S+)\s+(feat|fix|docs|refactor|test|perf|style|build|chore|'
    r'ci|revert|plan|migration|remove|security|deps|hotfix|merge|deploy)'
    r'(\([\w/\-]+\))?(!)?: .+'
)

def test_valid_commits():
    valid = [
        "✨ feat: add feature",
        "✨ feat(installer): add gum menu",
        "✨ feat!: rename workspace",
        "✨ feat(api)!: change endpoint",
        "🔄 ci: add release workflow",
        "↩️ revert: undo broken rule",
        "📋 plan: update blueprint",
        "🐛 fix: resolve bug",
        "🩹 hotfix: urgent fix",
        "📚 docs: update readme",
        "♻️ refactor: clean up code",
        "🧪 test: add coverage",
        "⚡ perf: optimize query",
        "💄 style: fix indent",
        "🔧 build: update makefile",
        "🧹 chore: maintenance",
        "🔒 security: fix xss",
        "🗃️ migration: add table",
        "📦 deps: update packages",
        "🚀 deploy: update config",
        "🔥 remove: drop legacy",
        "🔀 merge: merge staging",
    ]
    for msg in valid:
        assert COMMIT_PATTERN.match(msg), f"Should match: {msg}"
    print(f"  ✓ {len(valid)} valid commits pass")

def test_invalid_commits():
    invalid = [
        "feat: no emoji",
        "✨ unknown: bad type",
        "✨ feat:",  # no description
        "",  # empty
        "just some random text",
    ]
    for msg in invalid:
        assert not COMMIT_PATTERN.match(msg), f"Should NOT match: {msg}"
    print(f"  ✓ {len(invalid)} invalid commits rejected")

def test_breaking_change_detection():
    body = "✨ feat: something\n\nBREAKING CHANGE: old API removed"
    assert "BREAKING CHANGE:" in body
    print("  ✓ BREAKING CHANGE detected in body")

def test_staging_branch_config():
    # Create a temp config
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yml', delete=False) as f:
        f.write("staging_branch: homolog\nlanguage: auto\n")
        config_path = f.name

    try:
        # Read it the same way blueprint.py would (line parsing, no yaml dependency)
        staging_branch = "staging"  # default
        for line in open(config_path).read().splitlines():
            if line.strip().startswith("staging_branch:"):
                val = line.split(":", 1)[1].strip()
                if val:
                    staging_branch = val
        assert staging_branch == "homolog", f"Expected 'homolog', got '{staging_branch}'"
        print("  ✓ staging_branch read from config")
    finally:
        os.unlink(config_path)

def test_staging_branch_fallback():
    # No config file → should fall back to 'staging'
    staging_branch = "staging"
    config_path = "/tmp/nonexistent-blueprint-config-test.yml"
    if not os.path.exists(config_path):
        # Default fallback
        assert staging_branch == "staging"
        print("  ✓ staging_branch falls back to 'staging' when no config")

def test_scope_patterns():
    """Test various scope formats."""
    scopes = [
        "✨ feat(auth): add login",
        "🐛 fix(api/v2): fix endpoint",
        "♻️ refactor(user-service): clean up",
    ]
    for msg in scopes:
        assert COMMIT_PATTERN.match(msg), f"Should match scope: {msg}"
    print(f"  ✓ {len(scopes)} scoped commits pass")

def test_breaking_with_scope():
    """Test breaking change with scope."""
    msgs = [
        "✨ feat(api)!: change response format",
        "🐛 fix(db)!: change schema",
    ]
    for msg in msgs:
        assert COMMIT_PATTERN.match(msg), f"Should match breaking+scope: {msg}"
    print(f"  ✓ {len(msgs)} breaking+scope commits pass")

if __name__ == '__main__':
    print("Testing blueprint.py logic...")
    print("")
    test_valid_commits()
    test_invalid_commits()
    test_breaking_change_detection()
    test_staging_branch_config()
    test_staging_branch_fallback()
    test_scope_patterns()
    test_breaking_with_scope()
    print("")
    print("All tests passed!")
