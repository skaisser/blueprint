package plan

import (
	"testing"
)

// ─── Commit ───────────────────────────────────────────────────────────────────

func TestCommit_EmptyMessage_ReturnsError(t *testing.T) {
	err := Commit("", nil)
	if err == nil {
		t.Fatal("expected error for empty commit message, got nil")
	}
	if err.Error() != "commit message required" {
		t.Errorf("error message: got %q, want %q", err.Error(), "commit message required")
	}
}
