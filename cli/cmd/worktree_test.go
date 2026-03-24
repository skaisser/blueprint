package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

// ─── computeWorktreePath tests ──────────────────────────────────────────────

func TestWorktreePath_WithPlanNumber(t *testing.T) {
	path := computeWorktreePath("feat/0002-cli-enhancements")
	base := filepath.Base(path)
	if !strings.Contains(base, "0002") {
		t.Errorf("expected path to contain plan number 0002, got %q", base)
	}
}

func TestWorktreePath_WithoutPlanNumber(t *testing.T) {
	path := computeWorktreePath("feat/my-feature")
	base := filepath.Base(path)
	if !strings.Contains(base, "feat-my-feature") {
		t.Errorf("expected path to contain sanitized branch name, got %q", base)
	}
}

func TestWorktreePath_SlashSanitization(t *testing.T) {
	path := computeWorktreePath("hotfix/urgent-fix")
	base := filepath.Base(path)
	// Should not contain filesystem separators in the name
	if strings.Contains(base, "/") {
		t.Errorf("expected no slashes in directory name, got %q", base)
	}
}

func TestWorktreePath_SpecialChars(t *testing.T) {
	path := computeWorktreePath("feat/add@auth!")
	base := filepath.Base(path)
	if strings.Contains(base, "@") || strings.Contains(base, "!") {
		t.Errorf("expected special chars removed, got %q", base)
	}
}

func TestWorktreePath_PlanNumberExtraction(t *testing.T) {
	tests := []struct {
		branch   string
		contains string
	}{
		{"feat/0001-init", "0001"},
		{"fix/0042-bug-fix", "0042"},
		{"feat/no-number", "feat-no-number"},
	}

	for _, tt := range tests {
		path := computeWorktreePath(tt.branch)
		base := filepath.Base(path)
		if !strings.Contains(base, tt.contains) {
			t.Errorf("branch %q: expected path to contain %q, got %q", tt.branch, tt.contains, base)
		}
	}
}

// ─── isWorktree tests ───────────────────────────────────────────────────────

func TestIsWorktree_NonExistentPath(t *testing.T) {
	result := isWorktree("/tmp/nonexistent-path-12345")
	if result {
		t.Error("expected false for non-existent path")
	}
}

func TestIsWorktree_RegularDirectory(t *testing.T) {
	result := isWorktree(t.TempDir())
	if result {
		t.Error("expected false for regular directory without .git")
	}
}
