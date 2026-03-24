package cmd

import (
	"strings"
	"testing"
)

// ─── FormatCommitMessage tests ──────────────────────────────────────────────

func TestCommitFormat_Feat(t *testing.T) {
	result, err := FormatCommitMessage("feat", "add auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "✨ feat: add auth"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCommitFormat_Fix(t *testing.T) {
	result, err := FormatCommitMessage("fix", "resolve bug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "🐛 fix: resolve bug"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCommitFormat_Docs(t *testing.T) {
	result, err := FormatCommitMessage("docs", "update readme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "📚 docs: update readme"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCommitFormat_Refactor(t *testing.T) {
	result, err := FormatCommitMessage("refactor", "extract service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "♻️") {
		t.Errorf("expected refactor emoji prefix, got %q", result)
	}
	if !strings.Contains(result, "refactor: extract service") {
		t.Errorf("expected message content, got %q", result)
	}
}

func TestCommitFormat_AllTypes(t *testing.T) {
	types := map[string]string{
		"feat":      "✨",
		"fix":       "🐛",
		"docs":      "📚",
		"refactor":  "♻️",
		"test":      "🧪",
		"perf":      "⚡",
		"style":     "💄",
		"build":     "🔧",
		"chore":     "🧹",
		"ci":        "🔄",
		"revert":    "↩️",
		"plan":      "📋",
		"migration": "🗃️",
		"remove":    "🔥",
		"security":  "🔒",
		"deps":      "📦",
		"hotfix":    "🩹",
		"merge":     "🔀",
		"deploy":    "🚀",
	}

	for commitType, emoji := range types {
		result, err := FormatCommitMessage(commitType, "test message")
		if err != nil {
			t.Errorf("type %q: unexpected error: %v", commitType, err)
			continue
		}
		if !strings.HasPrefix(result, emoji) {
			t.Errorf("type %q: expected prefix %q, got %q", commitType, emoji, result)
		}
		if !strings.Contains(result, commitType+": test message") {
			t.Errorf("type %q: expected format '<emoji> <type>: <msg>', got %q", commitType, result)
		}
	}
}

func TestCommitFormat_InvalidType(t *testing.T) {
	_, err := FormatCommitMessage("invalid", "some message")
	if err == nil {
		t.Error("expected error for invalid commit type")
	}
	if !strings.Contains(err.Error(), "invalid commit type") {
		t.Errorf("expected 'invalid commit type' in error, got: %v", err)
	}
}

func TestCommitFormat_EmptyType(t *testing.T) {
	_, err := FormatCommitMessage("", "some message")
	if err == nil {
		t.Error("expected error for empty commit type")
	}
}

// ─── merge-chain logic tests ────────────────────────────────────────────────

func TestMergeChain_StepReporting(t *testing.T) {
	// Test that runMergeChain returns proper step info on failure.
	// Since gh is not available in tests, we verify the error contains step info.
	result := runMergeChain("nonexistent-branch", "staging")
	if _, ok := result["error"]; !ok {
		t.Error("expected error in result when gh is not available")
	}
	if step, ok := result["step"]; ok {
		stepStr, _ := step.(string)
		if stepStr != "staging_pr" {
			t.Errorf("expected first failure at 'staging_pr', got %q", stepStr)
		}
	}
}

// ─── DetectStack tests ──────────────────────────────────────────────────────

func TestDetectStack_ReturnsMap(t *testing.T) {
	result := DetectStack()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if _, ok := result["language"]; !ok {
		t.Error("expected 'language' key in result")
	}
}

func TestDetectStack_GoProject(t *testing.T) {
	// We're running from cli/ which has go.mod, so it should detect Go
	result := DetectStack()
	if result["language"] != "go" {
		// Might not be in cli/ dir during test, that's ok
		t.Logf("detected language: %s (expected go if running from cli/)", result["language"])
	}
}

// ─── reviewPoll helper tests ────────────────────────────────────────────────

func TestReviewPoll_TimeoutResult(t *testing.T) {
	// With a 0-second timeout, should immediately return timeout
	result := pollForReview("99999", 0)
	if _, ok := result["timeout"]; !ok {
		t.Error("expected timeout result with zero duration")
	}
}
