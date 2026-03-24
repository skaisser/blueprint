package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// GetNextNum
// ---------------------------------------------------------------------------

func TestGetNextNum_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	if got := GetNextNum(dir); got != "0001" {
		t.Errorf("expected %q, got %q", "0001", got)
	}
}

func TestGetNextNum_NonExistentDirectory(t *testing.T) {
	if got := GetNextNum("/tmp/does-not-exist-blueprint-test"); got != "0001" {
		t.Errorf("expected %q, got %q", "0001", got)
	}
}

func TestGetNextNum_ReturnsNextAfterHighest(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0001-feat-something.md")
	touch(t, dir, "0003-fix-other.md")

	if got := GetNextNum(dir); got != "0004" {
		t.Errorf("expected %q, got %q", "0004", got)
	}
}

func TestGetNextNum_IgnoresFilesWithoutNumberPrefix(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "README.md")
	touch(t, dir, "some-plan.md")
	touch(t, dir, "notes.txt")

	if got := GetNextNum(dir); got != "0001" {
		t.Errorf("expected %q, got %q", "0001", got)
	}
}

func TestGetNextNum_MixedFilesOnlyCountsNumbered(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "README.md")
	touch(t, dir, "0002-feat-login.md")
	touch(t, dir, "notes.txt")

	if got := GetNextNum(dir); got != "0003" {
		t.Errorf("expected %q, got %q", "0003", got)
	}
}

func TestGetNextNum_SingleFile(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0005-feat-payments.md")

	if got := GetNextNum(dir); got != "0006" {
		t.Errorf("expected %q, got %q", "0006", got)
	}
}

// ---------------------------------------------------------------------------
// FindPlanFile
// ---------------------------------------------------------------------------

func TestFindPlanFile_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	if got := FindPlanFile(dir, "feat/anything"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindPlanFile_NonExistentDirectory(t *testing.T) {
	if got := FindPlanFile("/tmp/does-not-exist-blueprint-test", "feat/anything"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindPlanFile_MatchesBranchSuffixAgainstFilename(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0001-feat-auth-flow-todo.md")
	touch(t, dir, "0002-fix-unrelated-todo.md")

	got := FindPlanFile(dir, "feat/auth-flow")
	want := filepath.Join(dir, "0001-feat-auth-flow-todo.md")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFindPlanFile_PrefersTodoOverOtherMd(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0001-feat-payments.md")
	touch(t, dir, "0002-feat-dashboard-todo.md")

	got := FindPlanFile(dir, "feat/no-match")
	// todo file should be preferred; it's the only candidate in todoFiles
	want := filepath.Join(dir, "0002-feat-dashboard-todo.md")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFindPlanFile_FallsBackToMostRecentlyModified(t *testing.T) {
	dir := t.TempDir()

	older := filepath.Join(dir, "0001-feat-alpha-todo.md")
	newer := filepath.Join(dir, "0002-feat-beta-todo.md")
	writeFile(t, older, "# alpha")
	writeFile(t, newer, "# beta")

	// Ensure newer has a later mtime.
	past := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(older, past, past); err != nil {
		t.Fatal(err)
	}

	got := FindPlanFile(dir, "feat/no-match-at-all")
	want := newer
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestFindPlanFile_IgnoresHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, ".hidden-todo.md")
	touch(t, dir, ".0001-secret-todo.md")

	if got := FindPlanFile(dir, "feat/anything"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindPlanFile_IgnoresNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0001-feat-something.txt")
	touch(t, dir, "0002-fix-thing.json")

	if got := FindPlanFile(dir, "feat/anything"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindPlanFile_BranchWithoutSlashStillMatches(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "0001-feat-auth-todo.md")

	// branch with no "/" — branchSuffix == branch
	got := FindPlanFile(dir, "auth")
	want := filepath.Join(dir, "0001-feat-auth-todo.md")
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// ---------------------------------------------------------------------------
// ParsePlanHeader
// ---------------------------------------------------------------------------

func TestParsePlanHeader_EmptyFilePath(t *testing.T) {
	h := ParsePlanHeader("")
	if h.PlanNum != "" || h.Status != "" || h.Progress != "" {
		t.Errorf("expected empty PlanHeader, got %+v", h)
	}
}

func TestParsePlanHeader_ExtractsPlanNumFromFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0042-feat-test.md")
	writeFile(t, path, "# Plan\n")

	h := ParsePlanHeader(path)
	if h.PlanNum != "0042" {
		t.Errorf("expected PlanNum %q, got %q", "0042", h.PlanNum)
	}
}

func TestParsePlanHeader_ParsesStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-feat-test.md")
	writeFile(t, path, "# Plan\n\n> **Status:** In Progress\n")

	h := ParsePlanHeader(path)
	if h.Status != "In Progress" {
		t.Errorf("expected Status %q, got %q", "In Progress", h.Status)
	}
}

func TestParsePlanHeader_ParsesProgress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-feat-test.md")
	writeFile(t, path, "# Plan\n\n> **Progress:** 3/5 tasks\n")

	h := ParsePlanHeader(path)
	if h.Progress != "3/5 tasks" {
		t.Errorf("expected Progress %q, got %q", "3/5 tasks", h.Progress)
	}
}

func TestParsePlanHeader_ParsesBothStatusAndProgress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0007-fix-auth.md")
	content := "# Plan\n\n> **Status:** Completed\n> **Progress:** 5/5 tasks\n"
	writeFile(t, path, content)

	h := ParsePlanHeader(path)
	if h.PlanNum != "0007" {
		t.Errorf("expected PlanNum %q, got %q", "0007", h.PlanNum)
	}
	if h.Status != "Completed" {
		t.Errorf("expected Status %q, got %q", "Completed", h.Status)
	}
	if h.Progress != "5/5 tasks" {
		t.Errorf("expected Progress %q, got %q", "5/5 tasks", h.Progress)
	}
}

func TestParsePlanHeader_OnlyReadsFirst30Lines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-feat-test.md")

	// Put status/progress beyond line 30.
	lines := ""
	for i := 0; i < 31; i++ {
		lines += "line content here\n"
	}
	lines += "> **Status:** Should Not Be Found\n"
	lines += "> **Progress:** 99/99 tasks\n"
	writeFile(t, path, lines)

	h := ParsePlanHeader(path)
	if h.Status != "" {
		t.Errorf("expected empty Status, got %q", h.Status)
	}
	if h.Progress != "" {
		t.Errorf("expected empty Progress, got %q", h.Progress)
	}
}

func TestParsePlanHeader_FilenameWithoutNumber(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "readme.md")
	writeFile(t, path, "> **Status:** Draft\n")

	h := ParsePlanHeader(path)
	if h.PlanNum != "" {
		t.Errorf("expected empty PlanNum, got %q", h.PlanNum)
	}
	if h.Status != "Draft" {
		t.Errorf("expected Status %q, got %q", "Draft", h.Status)
	}
}

func TestParsePlanHeader_NonExistentFile(t *testing.T) {
	h := ParsePlanHeader("/tmp/does-not-exist-blueprint-plan.md")
	// PlanNum extracted from filename even if file can't be opened.
	if h.Status != "" || h.Progress != "" {
		t.Errorf("expected empty Status/Progress for missing file, got %+v", h)
	}
}

// ---------------------------------------------------------------------------
// MetaResult.JSON
// ---------------------------------------------------------------------------

func TestMetaResultJSON_OutputFormat(t *testing.T) {
	m := &MetaResult{
		NextNum:    "0003",
		BaseBranch: "main",
		Branch:     "feat/payments",
		PlanFile:   "/plans/0002-feat-payments-todo.md",
		PlanNum:    "0002",
		Status:     "In Progress",
		Progress:   "2/4 tasks",
		Project:    "myapp",
		GitRemote:  "git@github.com:org/myapp.git",
		Today:      "2026-03-24",
	}

	got := m.JSON()

	// Must be valid JSON.
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("JSON() returned invalid JSON: %v\ngot: %s", err, got)
	}

	cases := map[string]string{
		"next_num":    "0003",
		"base_branch": "main",
		"branch":      "feat/payments",
		"plan_file":   "/plans/0002-feat-payments-todo.md",
		"plan_num":    "0002",
		"status":      "In Progress",
		"progress":    "2/4 tasks",
		"project":     "myapp",
		"git_remote":  "git@github.com:org/myapp.git",
		"today":       "2026-03-24",
	}
	for key, want := range cases {
		v, ok := decoded[key]
		if !ok {
			t.Errorf("key %q missing from JSON output", key)
			continue
		}
		if v != want {
			t.Errorf("key %q: expected %q, got %v", key, want, v)
		}
	}
}

func TestMetaResultJSON_EmptyStruct(t *testing.T) {
	m := &MetaResult{}
	got := m.JSON()

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("JSON() returned invalid JSON: %v\ngot: %s", err, got)
	}

	// All string fields should be present as empty strings.
	expectedKeys := []string{
		"next_num", "base_branch", "branch", "plan_file",
		"plan_num", "status", "progress", "project", "git_remote", "today",
	}
	for _, key := range expectedKeys {
		if _, ok := decoded[key]; !ok {
			t.Errorf("key %q missing from JSON output", key)
		}
	}
}

func TestMetaResultJSON_IsIndented(t *testing.T) {
	m := &MetaResult{NextNum: "0001"}
	got := m.JSON()

	// Indented JSON contains newlines.
	if len(got) == 0 {
		t.Fatal("JSON() returned empty string")
	}
	hasNewline := false
	for _, c := range got {
		if c == '\n' {
			hasNewline = true
			break
		}
	}
	if !hasNewline {
		t.Errorf("expected indented (multi-line) JSON, got: %s", got)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func touch(t *testing.T, dir, name string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, name), "")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
