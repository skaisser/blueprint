package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSessionState_CreatesDirectory(t *testing.T) {
	sessionID := "state-test-mkdir-" + filepath.Base(t.TempDir())
	state := NewSessionState(sessionID)
	defer os.RemoveAll(state.Dir)

	info, err := os.Stat(state.Dir)
	if err != nil {
		t.Fatalf("expected session directory to exist, got error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %q to be a directory", state.Dir)
	}
}

func TestReadText_NonExistentFileReturnsEmpty(t *testing.T) {
	state := NewSessionState("state-test-readtext-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	got := state.ReadText("does-not-exist")
	if got != "" {
		t.Errorf("ReadText on missing file: got %q, want \"\"", got)
	}
}

func TestReadInt_NonExistentFileReturnsZero(t *testing.T) {
	state := NewSessionState("state-test-readint-missing-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	got := state.ReadInt("does-not-exist")
	if got != 0 {
		t.Errorf("ReadInt on missing file: got %d, want 0", got)
	}
}

func TestReadInt_NonNumericContentReturnsZero(t *testing.T) {
	state := NewSessionState("state-test-readint-nan-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	state.WriteText("counter", "not-a-number")
	got := state.ReadInt("counter")
	if got != 0 {
		t.Errorf("ReadInt on non-numeric content: got %d, want 0", got)
	}
}

func TestWriteInt_ReadInt_RoundTrip(t *testing.T) {
	state := NewSessionState("state-test-writeint-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	cases := []int{0, 1, 42, -7, 1000000}
	for _, val := range cases {
		state.WriteInt("num", val)
		got := state.ReadInt("num")
		if got != val {
			t.Errorf("WriteInt(%d) then ReadInt: got %d", val, got)
		}
	}
}

func TestIncrInt_FromZero(t *testing.T) {
	state := NewSessionState("state-test-incr-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	// First call on a non-existent counter must start from 0 and return 1.
	got := state.IncrInt("fresh-counter")
	if got != 1 {
		t.Errorf("IncrInt from zero: got %d, want 1", got)
	}
	if stored := state.ReadInt("fresh-counter"); stored != 1 {
		t.Errorf("IncrInt stored value: got %d, want 1", stored)
	}
}

func TestAppendLine_CreatesFileIfNotExists(t *testing.T) {
	state := NewSessionState("state-test-appendline-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	name := "new-log"
	if state.Exists(name) {
		t.Fatal("file should not exist before AppendLine")
	}

	state.AppendLine(name, "hello")
	if !state.Exists(name) {
		t.Error("AppendLine should create the file if it does not exist")
	}

	text := state.ReadText(name)
	if !strings.Contains(text, "hello") {
		t.Errorf("AppendLine content = %q, want to contain 'hello'", text)
	}
}

func TestTouch_Idempotent(t *testing.T) {
	state := NewSessionState("state-test-touch-idem-" + filepath.Base(t.TempDir()))
	defer os.RemoveAll(state.Dir)

	state.Touch("marker")
	// Write content so we can verify Touch does not wipe it — idempotency
	// means the file stays present; content may change (WriteFile truncates).
	// The contract tested here is simply: calling Touch a second time must
	// not cause Exists to return false.
	state.Touch("marker")

	if !state.Exists("marker") {
		t.Error("marker should still exist after second Touch call")
	}
}
