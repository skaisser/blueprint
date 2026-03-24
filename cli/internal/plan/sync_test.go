package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writePlanFile writes content to a file named filename inside dir and returns the full path.
func writePlanFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writePlanFile: %v", err)
	}
	return path
}

// basicPlan returns a plan body matching the canonical shape described in the task.
func basicPlan() string {
	return `---
id: "0001"
title: "test plan"
status: in-progress
---

# Test Plan

## Phases

### Phase 1: Test
- [x] Task 1
- [ ] Task 2

### Phase 2: More
- [x] Task 3
- [x] Task 4

## Acceptance
- [ ] All done
`
}

// readFrontmatter parses the raw YAML block out of a written plan file so tests
// can assert on individual keys without importing the yaml package.
func readFrontmatter(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFrontmatter ReadFile: %v", err)
	}
	parts := strings.SplitN(string(b), "---", 3)
	if len(parts) < 3 {
		t.Fatalf("readFrontmatter: no valid frontmatter in %s", path)
	}
	return parts[1]
}

// ---- tests ------------------------------------------------------------------

func TestSyncPlanFile_BasicTaskCounts(t *testing.T) {
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TasksDone != 3 {
		t.Errorf("TasksDone: got %d, want 3", res.TasksDone)
	}
	if res.TasksTotal != 4 {
		t.Errorf("TasksTotal: got %d, want 4", res.TasksTotal)
	}
}

func TestSyncPlanFile_PhaseCounts(t *testing.T) {
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Phase 1 has one open task → not done. Phase 2 is fully checked → done.
	if res.PhasesTotal != 2 {
		t.Errorf("PhasesTotal: got %d, want 2", res.PhasesTotal)
	}
	if res.PhasesDone != 1 {
		t.Errorf("PhasesDone: got %d, want 1 (only Phase 2 is fully checked)", res.PhasesDone)
	}
}

func TestSyncPlanFile_AllPhasesComplete(t *testing.T) {
	body := `---
id: "0002"
title: "all done plan"
status: in-progress
---

# All Done

## Phases

### Phase 1: Alpha
- [x] Task A
- [x] Task B

### Phase 2: Beta
- [x] Task C

## Notes
Nothing here.
`
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0002-all-todo.md", body)

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TasksDone != 3 {
		t.Errorf("TasksDone: got %d, want 3", res.TasksDone)
	}
	if res.TasksTotal != 3 {
		t.Errorf("TasksTotal: got %d, want 3", res.TasksTotal)
	}
	if res.PhasesDone != 2 {
		t.Errorf("PhasesDone: got %d, want 2", res.PhasesDone)
	}
	if res.PhasesTotal != 2 {
		t.Errorf("PhasesTotal: got %d, want 2", res.PhasesTotal)
	}
}

func TestSyncPlanFile_AcceptanceSectionExcluded(t *testing.T) {
	// Tasks under ## Acceptance must NOT be counted in the task totals.
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// basicPlan has 4 tasks in Phases and 1 unchecked in Acceptance.
	// The sync must only count the 4 phase tasks.
	if res.TasksTotal != 4 {
		t.Errorf("TasksTotal should exclude Acceptance tasks: got %d, want 4", res.TasksTotal)
	}
}

func TestSyncPlanFile_FrontmatterUpdated(t *testing.T) {
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	_, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fm := readFrontmatter(t, path)
	if !strings.Contains(fm, "tasks_done: 3") {
		t.Errorf("frontmatter missing tasks_done: 3\n%s", fm)
	}
	if !strings.Contains(fm, "tasks_total: 4") {
		t.Errorf("frontmatter missing tasks_total: 4\n%s", fm)
	}
	if !strings.Contains(fm, "phases_done: 1") {
		t.Errorf("frontmatter missing phases_done: 1\n%s", fm)
	}
	if !strings.Contains(fm, "phases_total: 2") {
		t.Errorf("frontmatter missing phases_total: 2\n%s", fm)
	}
}

func TestSyncPlanFile_FinishMode_SetsCompletedStatus(t *testing.T) {
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	today := time.Now().Format("2006-01-02")

	res, err := SyncPlanFile(path, true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Finished {
		t.Error("SyncResult.Finished should be true")
	}

	fm := readFrontmatter(t, path)
	if !strings.Contains(fm, "status: completed") {
		t.Errorf("frontmatter missing status: completed\n%s", fm)
	}
	// The YAML marshaller may quote the date string (e.g. `"2026-03-24"`),
	// so check for the date value regardless of surrounding quotes.
	if !strings.Contains(fm, today) {
		t.Errorf("frontmatter missing completed date %s\n%s", today, fm)
	}
}

func TestSyncPlanFile_FinishMode_WithPRNumber(t *testing.T) {
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0001-test-todo.md", basicPlan())

	res, err := SyncPlanFile(path, true, "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.PR != "42" {
		t.Errorf("SyncResult.PR: got %q, want %q", res.PR, "42")
	}

	fm := readFrontmatter(t, path)
	if !strings.Contains(fm, "pr: 42") {
		t.Errorf("frontmatter missing pr: 42\n%s", fm)
	}
	if !strings.Contains(fm, "status: completed") {
		t.Errorf("frontmatter missing status: completed\n%s", fm)
	}
}

func TestSyncPlanFile_ErrorMissingFile(t *testing.T) {
	_, err := SyncPlanFile("/nonexistent/path/plan.md", false, "")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSyncPlanFile_ErrorInvalidFrontmatter(t *testing.T) {
	dir := t.TempDir()
	// No --- delimiters — invalid frontmatter
	body := "# Just a plain markdown file\n\n- [x] Task 1\n"
	path := writePlanFile(t, dir, "bad-plan.md", body)

	_, err := SyncPlanFile(path, false, "")
	if err == nil {
		t.Fatal("expected error for invalid frontmatter, got nil")
	}
}

func TestSyncPlanFile_AutoDetect_FindsTodoPlanInBlueprint(t *testing.T) {
	// Create a temp dir that will act as the working directory.
	// SyncPlanFile("", ...) globs blueprint/*-todo.md relative to cwd.
	tmpCwd := t.TempDir()

	// Create the blueprint sub-dir and a -todo.md plan file inside it.
	blueprintDir := filepath.Join(tmpCwd, "blueprint")
	if err := os.MkdirAll(blueprintDir, 0755); err != nil {
		t.Fatal(err)
	}

	planContent := `---
id: "0010"
title: "auto detect plan"
status: in-progress
---

# Auto Detect Plan

### Phase 1: Test
- [x] Task A
- [ ] Task B
`
	writePlanFile(t, blueprintDir, "0010-auto-todo.md", planContent)

	// chdir into the temp root so the glob resolves correctly.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(tmpCwd); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Call with empty planFile — auto-detect mode.
	res, err := SyncPlanFile("", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TasksTotal != 2 {
		t.Errorf("TasksTotal: got %d, want 2", res.TasksTotal)
	}
	if res.TasksDone != 1 {
		t.Errorf("TasksDone: got %d, want 1", res.TasksDone)
	}
}

func TestSyncPlanFile_SessionsBlockquoteFormat(t *testing.T) {
	// Old-style sessions: `> - Session N: note`
	body := `---
id: "0020"
title: "session old plan"
status: in-progress
---

# Session Old Plan

### Phase 1: Work
- [x] Task 1

## Sessions

> - Session 1: initial implementation
> - Session 2: bug fixes and review
`
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0020-session-todo.md", body)

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Sessions != 2 {
		t.Errorf("Sessions: got %d, want 2", res.Sessions)
	}

	// Verify the sessions list was written back to frontmatter.
	fm := readFrontmatter(t, path)
	if !strings.Contains(fm, "sessions:") {
		t.Errorf("frontmatter missing sessions key\n%s", fm)
	}
	if !strings.Contains(fm, "initial implementation") {
		t.Errorf("frontmatter missing first session note\n%s", fm)
	}
}

func TestSyncPlanFile_SessionsNewBacktickFormat(t *testing.T) {
	// New-style sessions: `> - \`uuid\` 2026-03-24 13:00 - note`
	body := `---
id: "0030"
title: "session new plan"
status: in-progress
---

# Session New Plan

### Phase 1: Work
- [x] Task 1
- [x] Task 2

## Sessions

> - ` + "`" + `abc123de-f456-7890-abcd-ef1234567890` + "`" + ` 2026-03-24 14:00 - first session note
> - ` + "`" + `bbbbbbbb-cccc-dddd-eeee-ffffffffffff` + "`" + ` 2026-03-24 16:30 - second session note
`
	dir := t.TempDir()
	path := writePlanFile(t, dir, "0030-session-new-todo.md", body)

	res, err := SyncPlanFile(path, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Sessions != 2 {
		t.Errorf("Sessions: got %d, want 2", res.Sessions)
	}

	fm := readFrontmatter(t, path)
	if !strings.Contains(fm, "sessions:") {
		t.Errorf("frontmatter missing sessions key\n%s", fm)
	}
	if !strings.Contains(fm, "first session note") {
		t.Errorf("frontmatter missing first session note\n%s", fm)
	}
	if !strings.Contains(fm, "second session note") {
		t.Errorf("frontmatter missing second session note\n%s", fm)
	}
}
