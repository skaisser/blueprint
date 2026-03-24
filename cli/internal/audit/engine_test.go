package audit

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// ParsePayload
// ---------------------------------------------------------------------------

func TestParsePayload_Standard(t *testing.T) {
	raw := `{"tool_name":"Write","session_id":"abc123","tool_input":{"file_path":"main.go","command":"run"}}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolName != "write" {
		t.Errorf("ToolName = %q, want %q", p.ToolName, "write")
	}
	if p.SessionID != "abc123" {
		t.Errorf("SessionID = %q, want %q", p.SessionID, "abc123")
	}
	if p.FilePath != "main.go" {
		t.Errorf("FilePath = %q, want %q", p.FilePath, "main.go")
	}
	if p.Command != "run" {
		t.Errorf("Command = %q, want %q", p.Command, "run")
	}
}

func TestParsePayload_AlternateKeyNames(t *testing.T) {
	raw := `{"toolName":"Edit","sessionId":"xyz789","toolInput":{"filePath":"src/app.go"}}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolName != "edit" {
		t.Errorf("ToolName = %q, want %q", p.ToolName, "edit")
	}
	if p.SessionID != "xyz789" {
		t.Errorf("SessionID = %q, want %q", p.SessionID, "xyz789")
	}
	if p.FilePath != "src/app.go" {
		t.Errorf("FilePath = %q, want %q", p.FilePath, "src/app.go")
	}
}

func TestParsePayload_MissingSessionIDDefaultsToUnknown(t *testing.T) {
	raw := `{"tool_name":"read","tool_input":{}}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.SessionID != "unknown" {
		t.Errorf("SessionID = %q, want %q", p.SessionID, "unknown")
	}
}

func TestParsePayload_MissingToolInputDefaultsToEmptyMap(t *testing.T) {
	raw := `{"tool_name":"bash","session_id":"s1"}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Input == nil {
		t.Fatal("Input should not be nil")
	}
	if len(p.Input) != 0 {
		t.Errorf("Input should be empty map, got %v", p.Input)
	}
}

func TestParsePayload_NestedToolNameMap(t *testing.T) {
	// tool_name is a nested map with a "name" key
	raw := `{"tool_name":{"name":"apply_patch"},"session_id":"s2","tool_input":{}}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolName != "apply_patch" {
		t.Errorf("ToolName = %q, want %q", p.ToolName, "apply_patch")
	}
}

func TestParsePayload_PatchTextExtracted(t *testing.T) {
	// JSON \n sequences are decoded to real newlines by the JSON parser.
	raw := `{"tool_name":"apply_patch","session_id":"s3","tool_input":{"patchText":"*** Begin Patch\n*** Add File: foo/bar.go\n+package foo\n*** End Patch"}}`
	p, err := ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "*** Begin Patch\n*** Add File: foo/bar.go\n+package foo\n*** End Patch"
	if p.PatchText != want {
		t.Errorf("PatchText = %q, want %q", p.PatchText, want)
	}
	// Verify PatchPaths is usable from a parsed payload
	paths := p.PatchPaths()
	if len(paths) != 1 || paths[0] != "foo/bar.go" {
		t.Errorf("PatchPaths() = %v, want [foo/bar.go]", paths)
	}
}

func TestParsePayload_InvalidJSONReturnsError(t *testing.T) {
	_, err := ParsePayload(strings.NewReader(`not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParsePayload_EmptyJSONReturnsError(t *testing.T) {
	_, err := ParsePayload(strings.NewReader(``))
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

// ---------------------------------------------------------------------------
// PatchPaths
// ---------------------------------------------------------------------------

func TestPatchPaths_EmptyReturnsNil(t *testing.T) {
	p := &Payload{}
	if paths := p.PatchPaths(); paths != nil {
		t.Errorf("PatchPaths() = %v, want nil", paths)
	}
}

func TestPatchPaths_SingleFile(t *testing.T) {
	p := &Payload{
		PatchText: "*** Add File: internal/foo/bar.go\n+package foo\n",
	}
	paths := p.PatchPaths()
	if len(paths) != 1 {
		t.Fatalf("PatchPaths() len = %d, want 1", len(paths))
	}
	if paths[0] != "internal/foo/bar.go" {
		t.Errorf("paths[0] = %q, want %q", paths[0], "internal/foo/bar.go")
	}
}

func TestPatchPaths_MultipleFiles(t *testing.T) {
	p := &Payload{
		PatchText: "*** Add File: a/one.go\n+x\n*** Update File: b/two.go\n~y\n*** Delete File: c/three.go\n-z\n",
	}
	paths := p.PatchPaths()
	if len(paths) != 3 {
		t.Fatalf("PatchPaths() len = %d, want 3", len(paths))
	}
	want := []string{"a/one.go", "b/two.go", "c/three.go"}
	for i, w := range want {
		if paths[i] != w {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], w)
		}
	}
}

func TestPatchPaths_NoMatchReturnsEmpty(t *testing.T) {
	p := &Payload{PatchText: "some random patch text with no markers"}
	paths := p.PatchPaths()
	if len(paths) != 0 {
		t.Errorf("PatchPaths() = %v, want empty slice", paths)
	}
}

// ---------------------------------------------------------------------------
// AllPaths
// ---------------------------------------------------------------------------

func TestAllPaths_OnlyFilePath(t *testing.T) {
	p := &Payload{FilePath: "main.go"}
	paths := p.AllPaths()
	if len(paths) != 1 || paths[0] != "main.go" {
		t.Errorf("AllPaths() = %v, want [main.go]", paths)
	}
}

func TestAllPaths_OnlyPatchPaths(t *testing.T) {
	p := &Payload{
		PatchText: "*** Add File: pkg/util.go\n+x\n",
	}
	paths := p.AllPaths()
	if len(paths) != 1 || paths[0] != "pkg/util.go" {
		t.Errorf("AllPaths() = %v, want [pkg/util.go]", paths)
	}
}

func TestAllPaths_BothFilePathAndPatch(t *testing.T) {
	p := &Payload{
		FilePath:  "cmd/root.go",
		PatchText: "*** Add File: internal/engine.go\n+x\n*** Update File: internal/rules.go\n~y\n",
	}
	paths := p.AllPaths()
	if len(paths) != 3 {
		t.Fatalf("AllPaths() len = %d, want 3", len(paths))
	}
	if paths[0] != "cmd/root.go" {
		t.Errorf("paths[0] = %q, want %q", paths[0], "cmd/root.go")
	}
	if paths[1] != "internal/engine.go" {
		t.Errorf("paths[1] = %q, want %q", paths[1], "internal/engine.go")
	}
	if paths[2] != "internal/rules.go" {
		t.Errorf("paths[2] = %q, want %q", paths[2], "internal/rules.go")
	}
}

func TestAllPaths_NeitherReturnsNil(t *testing.T) {
	p := &Payload{}
	paths := p.AllPaths()
	if len(paths) != 0 {
		t.Errorf("AllPaths() = %v, want empty", paths)
	}
}

// ---------------------------------------------------------------------------
// IsWriteLike
// ---------------------------------------------------------------------------

func TestIsWriteLike(t *testing.T) {
	tests := []struct {
		toolName string
		want     bool
	}{
		{"write", true},
		{"edit", true},
		{"apply_patch", true},
		{"applypatch", true},
		{"bash", false},
		{"read", false},
		{"", false},
		{"glob", false},
	}
	for _, tc := range tests {
		p := &Payload{ToolName: tc.toolName}
		if got := p.IsWriteLike(); got != tc.want {
			t.Errorf("IsWriteLike() for %q = %v, want %v", tc.toolName, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// normalizeToolName
// ---------------------------------------------------------------------------

func TestNormalizeToolName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Write", "write"},
		{"BASH", "bash"},
		{"computer_use.bash", "bash"},
		{"some.dotted.tool_name", "tool_name"},
		{"", ""},
		{"  ", ""},
		{" Edit ", "edit"},
		{"apply_patch", "apply_patch"},
	}
	for _, tc := range tests {
		if got := normalizeToolName(tc.input); got != tc.want {
			t.Errorf("normalizeToolName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// firstString
// ---------------------------------------------------------------------------

func TestFirstString_StringValue(t *testing.T) {
	m := map[string]interface{}{
		"tool_name": "bash",
	}
	if got := firstString(m, "tool_name"); got != "bash" {
		t.Errorf("firstString = %q, want %q", got, "bash")
	}
}

func TestFirstString_NestedMapWithName(t *testing.T) {
	m := map[string]interface{}{
		"tool_name": map[string]interface{}{
			"name": "write",
		},
	}
	if got := firstString(m, "tool_name"); got != "write" {
		t.Errorf("firstString = %q, want %q", got, "write")
	}
}

func TestFirstString_NestedMapWithToolName(t *testing.T) {
	m := map[string]interface{}{
		"tool_name": map[string]interface{}{
			"tool_name": "edit",
		},
	}
	if got := firstString(m, "tool_name"); got != "edit" {
		t.Errorf("firstString = %q, want %q", got, "edit")
	}
}

func TestFirstString_MissingKeyReturnsEmpty(t *testing.T) {
	m := map[string]interface{}{
		"other": "value",
	}
	if got := firstString(m, "tool_name", "toolName"); got != "" {
		t.Errorf("firstString = %q, want empty", got)
	}
}

func TestFirstString_FirstMatchWins(t *testing.T) {
	m := map[string]interface{}{
		"tool_name": "first",
		"toolName":  "second",
	}
	if got := firstString(m, "tool_name", "toolName"); got != "first" {
		t.Errorf("firstString = %q, want %q", got, "first")
	}
}

func TestFirstString_NilValueSkipped(t *testing.T) {
	m := map[string]interface{}{
		"tool_name": nil,
		"toolName":  "fallback",
	}
	if got := firstString(m, "tool_name", "toolName"); got != "fallback" {
		t.Errorf("firstString = %q, want %q", got, "fallback")
	}
}

// ---------------------------------------------------------------------------
// Logger helpers
// ---------------------------------------------------------------------------

// newTestLogger creates a Logger that writes to a file in dir.
func newTestLogger(t *testing.T, dir, session, tool string) (*Logger, string) {
	t.Helper()
	logFile := filepath.Join(dir, "test.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("could not create log file: %v", err)
	}
	l := &Logger{
		file:      f,
		timestamp: "2024-01-01 00:00:00",
		session:   session,
		tool:      tool,
	}
	return l, logFile
}

// readLines reads all lines from a file.
func readLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open log file: %v", err)
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// ---------------------------------------------------------------------------
// NewLogger
// ---------------------------------------------------------------------------

func TestNewLogger_CreatesLogger(t *testing.T) {
	// Override HOME so NewLogger writes into our temp dir.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	l := NewLogger("session-abc", "bash")
	if l == nil {
		t.Fatal("NewLogger returned nil")
	}
	defer l.Close()

	if l.session != "session-" {
		t.Errorf("session = %q, want %q", l.session, "session-")
	}
	if l.tool != "bash" {
		t.Errorf("tool = %q, want %q", l.tool, "bash")
	}
	if l.file == nil {
		t.Error("Logger file handle should not be nil")
	}
}

func TestNewLogger_ShortSessionNotTruncated(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	l := NewLogger("abc", "write")
	defer l.Close()

	if l.session != "abc" {
		t.Errorf("session = %q, want %q", l.session, "abc")
	}
}

func TestNewLogger_LogDirCreated(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	l := NewLogger("s", "tool")
	defer l.Close()

	logDir := filepath.Join(tmp, ".claude", "hooks", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("log directory was not created: %s", logDir)
	}
}

// ---------------------------------------------------------------------------
// Logger.Log
// ---------------------------------------------------------------------------

func TestLogger_Log_WritesEntry(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "sess1", "bash")
	defer l.Close()

	l.Log("hello world")
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	line := lines[0]
	if !strings.Contains(line, "SESSION:sess1") {
		t.Errorf("log line missing session: %q", line)
	}
	if !strings.Contains(line, "bash") {
		t.Errorf("log line missing tool name: %q", line)
	}
	if !strings.Contains(line, "hello world") {
		t.Errorf("log line missing message: %q", line)
	}
}

func TestLogger_Log_NilFileNoPanic(t *testing.T) {
	l := &Logger{file: nil, timestamp: "ts", session: "s", tool: "t"}
	// Should not panic when file is nil
	l.Log("should not panic")
}

// ---------------------------------------------------------------------------
// Logger.Warn
// ---------------------------------------------------------------------------

func TestLogger_Warn_WritesWarningEntry(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "sess2", "write")
	defer l.Close()

	l.Warn("suspicious path")
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "WARNING") {
		t.Errorf("warn log missing WARNING: %q", lines[0])
	}
	if !strings.Contains(lines[0], "suspicious path") {
		t.Errorf("warn log missing message: %q", lines[0])
	}
}

// ---------------------------------------------------------------------------
// Logger.LogCall
// ---------------------------------------------------------------------------

func TestLogger_LogCall_WithDescription(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "s3", "bash")
	defer l.Close()

	l.LogCall(map[string]interface{}{
		"description": "run unit tests",
		"command":     "go test ./...",
	})
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	line := lines[0]
	if !strings.Contains(line, "CALL") {
		t.Errorf("LogCall line missing CALL: %q", line)
	}
	if !strings.Contains(line, "description=run unit tests") {
		t.Errorf("LogCall line missing description: %q", line)
	}
}

func TestLogger_LogCall_WithFilePath(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "s4", "write")
	defer l.Close()

	l.LogCall(map[string]interface{}{
		"file_path": "internal/engine.go",
	})
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "file_path=internal/engine.go") {
		t.Errorf("LogCall line missing file_path: %q", lines[0])
	}
}

func TestLogger_LogCall_EmptyInputWritesNoSummary(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "s5", "read")
	defer l.Close()

	l.LogCall(map[string]interface{}{})
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "no-summary") {
		t.Errorf("LogCall with empty input should log 'no-summary': %q", lines[0])
	}
}

func TestLogger_LogCall_LongValueTruncatedAt80(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "s6", "bash")
	defer l.Close()

	longVal := strings.Repeat("x", 120)
	l.LogCall(map[string]interface{}{
		"command": longVal,
	})
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	// The stored value portion should be max 80 chars
	if strings.Contains(lines[0], longVal) {
		t.Errorf("LogCall should truncate values longer than 80 chars")
	}
	truncated := strings.Repeat("x", 80)
	if !strings.Contains(lines[0], truncated) {
		t.Errorf("LogCall line should contain the first 80 chars of value: %q", lines[0])
	}
}

func TestLogger_LogCall_MaxThreeParts(t *testing.T) {
	dir := t.TempDir()
	l, logFile := newTestLogger(t, dir, "s7", "bash")
	defer l.Close()

	l.LogCall(map[string]interface{}{
		"description": "desc",
		"command":     "cmd",
		"file_path":   "fp",
		"name":        "nm",
	})
	l.Close()

	lines := readLines(t, logFile)
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}
	// At most 3 key=value parts separated by " | "
	parts := strings.Split(lines[0], " | ")
	// The CALL prefix itself is not a part, so count separators
	// line format: [...] CALL | k=v | k=v | k=v
	// Split on " | " gives: ["[...] CALL", "k=v", "k=v", "k=v"]
	if len(parts) > 4 {
		t.Errorf("LogCall should log at most 3 summary parts, got %d: %q", len(parts)-1, lines[0])
	}
}

// ---------------------------------------------------------------------------
// Logger.Close
// ---------------------------------------------------------------------------

func TestLogger_Close_NilFileNoPanic(t *testing.T) {
	l := &Logger{file: nil}
	// Should not panic
	l.Close()
}

func TestLogger_Close_ClosesFile(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLogger(t, dir, "s8", "tool")
	// Should not panic and the file should be closed
	l.Close()
	// Writing after close should not cause a panic in the logger
	l.file = nil
	l.Log("after close") // exercises nil guard
}
