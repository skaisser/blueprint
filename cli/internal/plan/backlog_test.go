package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── helpers ────────────────────────────────────────────────────────────────

// writeBacklogFile creates a file at dir/name with the given content and
// returns the full path. The name is kept distinct from meta_test.go's
// writeFile helper which takes a full path rather than dir+name.
func writeBacklogFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeBacklogFile %s: %v", name, err)
	}
	return p
}

// ─── ParseBacklogFile ────────────────────────────────────────────────────────

func TestParseBacklogFile_YAMLFrontmatter(t *testing.T) {
	dir := t.TempDir()
	content := `---
id: "0001"
title: My Feature
type: feat
status: in-progress
priority: high
size: M
project: myproject
tags:
  - api
  - backend
created: "2025-01-15"
---

# My Feature

Some body content here.
`
	path := writeBacklogFile(t, dir, "0001-feat-my-feature.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ID != "0001" {
		t.Errorf("ID: got %q, want %q", item.ID, "0001")
	}
	if item.Title != "My Feature" {
		t.Errorf("Title: got %q, want %q", item.Title, "My Feature")
	}
	if item.Type != "feat" {
		t.Errorf("Type: got %q, want %q", item.Type, "feat")
	}
	if item.Status != "in-progress" {
		t.Errorf("Status: got %q, want %q", item.Status, "in-progress")
	}
	if item.Priority != "high" {
		t.Errorf("Priority: got %q, want %q", item.Priority, "high")
	}
	if item.Size != "M" {
		t.Errorf("Size: got %q, want %q", item.Size, "M")
	}
	if item.Project != "myproject" {
		t.Errorf("Project: got %q, want %q", item.Project, "myproject")
	}
	if len(item.Tags) != 2 || item.Tags[0] != "api" || item.Tags[1] != "backend" {
		t.Errorf("Tags: got %v, want [api backend]", item.Tags)
	}
	if item.Created != "2025-01-15" {
		t.Errorf("Created: got %q, want %q", item.Created, "2025-01-15")
	}
	if item.File != "0001-feat-my-feature.md" {
		t.Errorf("File: got %q, want %q", item.File, "0001-feat-my-feature.md")
	}
}

func TestParseBacklogFile_YAMLFrontmatter_TypeFallbackFromFilename(t *testing.T) {
	dir := t.TempDir()
	// YAML has no type field — should fall back to filename
	content := `---
id: "0002"
title: No Type Here
status: new
priority: medium
size: S
---
`
	path := writeBacklogFile(t, dir, "0002-refactor-no-type-here.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != "refactor" {
		t.Errorf("Type: got %q, want %q", item.Type, "refactor")
	}
}

func TestParseBacklogFile_BlockquoteFormat_FullMetadata(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** in-progress
> **Priority:** high
> **Size:** L
> **Created:** 2025-03-01
> **Plan:** temp/plans/0003-feat-login.md

# Login Feature

Body paragraph.
`
	path := writeBacklogFile(t, dir, "0003-feat-login.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ID != "0003" {
		t.Errorf("ID: got %q, want %q", item.ID, "0003")
	}
	if item.Type != "feat" {
		t.Errorf("Type: got %q, want %q", item.Type, "feat")
	}
	// bqTitle regex anchors at start of string (no (?m) flag), so when
	// blockquote metadata precedes the H1, the title falls back to the
	// filename slug: "0003-feat-login" → strip "0003-feat-" → "login".
	if item.Title != "login" {
		t.Errorf("Title: got %q, want %q", item.Title, "login")
	}
	if item.Status != "in-progress" {
		t.Errorf("Status: got %q, want %q", item.Status, "in-progress")
	}
	if item.Priority != "high" {
		t.Errorf("Priority: got %q, want %q", item.Priority, "high")
	}
	if item.Size != "L" {
		t.Errorf("Size: got %q, want %q", item.Size, "L")
	}
	if item.Created != "2025-03-01" {
		t.Errorf("Created: got %q, want %q", item.Created, "2025-03-01")
	}
	if item.Plan == nil || *item.Plan != "temp/plans/0003-feat-login.md" {
		t.Errorf("Plan: got %v, want %q", item.Plan, "temp/plans/0003-feat-login.md")
	}
}

func TestParseBacklogFile_BlockquoteFormat_DefaultStatus(t *testing.T) {
	dir := t.TempDir()
	// No Status line — should default to "new"
	content := `> **Priority:** low
> **Size:** XS

# Small Task
`
	path := writeBacklogFile(t, dir, "0004-chore-small-task.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Status != "new" {
		t.Errorf("Status: got %q, want %q", item.Status, "new")
	}
}

func TestParseBacklogFile_IDFromFilename(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** new

# Some Idea
`
	path := writeBacklogFile(t, dir, "0042-feat-some-idea.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != "0042" {
		t.Errorf("ID: got %q, want %q", item.ID, "0042")
	}
}

func TestParseBacklogFile_TypeFromFilename(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** new

# Perf Thing
`
	path := writeBacklogFile(t, dir, "0010-perf-something.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != "perf" {
		t.Errorf("Type: got %q, want %q", item.Type, "perf")
	}
}

func TestParseBacklogFile_TitleFromH1(t *testing.T) {
	dir := t.TempDir()
	// bqTitle uses ^#\s+ which anchors at the start of the WHOLE string (no
	// multiline flag). Title is only parsed from H1 when it appears first.
	content := `# My Explicit Title

> **Status:** new

Body.
`
	path := writeBacklogFile(t, dir, "0005-feat-my-explicit-title.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "My Explicit Title" {
		t.Errorf("Title: got %q, want %q", item.Title, "My Explicit Title")
	}
}

func TestParseBacklogFile_TitleDerivedFromFilename(t *testing.T) {
	dir := t.TempDir()
	// No H1 heading — title should be derived from filename slug
	content := `> **Status:** new
> **Priority:** medium
`
	path := writeBacklogFile(t, dir, "0006-feat-derived-from-slug.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "0006-feat-derived-from-slug" → strip prefix "0006-feat-" → "derived-from-slug" → "derived from slug"
	if item.Title != "derived from slug" {
		t.Errorf("Title: got %q, want %q", item.Title, "derived from slug")
	}
}

func TestParseBacklogFile_PlanFieldNotYetSkipped(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** new
> **Plan:** not yet planned

# Some Feature
`
	path := writeBacklogFile(t, dir, "0007-feat-some-feature.md", content)

	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Plan != nil {
		t.Errorf("Plan: expected nil for 'not yet planned', got %v", item.Plan)
	}
}

func TestParseBacklogFile_NonExistentFile(t *testing.T) {
	_, err := ParseBacklogFile("/tmp/does-not-exist-at-all.md")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// ─── ScanBacklogDir ──────────────────────────────────────────────────────────

func TestScanBacklogDir_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	items, err := ScanBacklogDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty slice, got %d items", len(items))
	}
}

func TestScanBacklogDir_NonExistentDirectory(t *testing.T) {
	items, err := ScanBacklogDir("/tmp/definitely-does-not-exist-xyz-abc")
	if err != nil {
		t.Fatalf("expected nil error for non-existent dir, got: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil slice, got %v", items)
	}
}

func TestScanBacklogDir_MultipleFilesSortedByID(t *testing.T) {
	dir := t.TempDir()

	// Write files out of order
	writeBacklogFile(t, dir, "0003-feat-third.md", `---
id: "0003"
title: Third
type: feat
status: new
priority: low
size: S
---
`)
	writeBacklogFile(t, dir, "0001-feat-first.md", `---
id: "0001"
title: First
type: feat
status: new
priority: high
size: L
---
`)
	writeBacklogFile(t, dir, "0002-fix-second.md", `---
id: "0002"
title: Second
type: fix
status: new
priority: medium
size: M
---
`)

	items, err := ScanBacklogDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].ID != "0001" || items[1].ID != "0002" || items[2].ID != "0003" {
		t.Errorf("items not sorted: got IDs %s, %s, %s", items[0].ID, items[1].ID, items[2].ID)
	}
}

func TestScanBacklogDir_SkipsHiddenFiles(t *testing.T) {
	dir := t.TempDir()

	writeBacklogFile(t, dir, ".hidden-file.md", `---
id: "9999"
title: Hidden
type: feat
status: new
priority: low
size: S
---
`)
	writeBacklogFile(t, dir, "0001-feat-visible.md", `---
id: "0001"
title: Visible
type: feat
status: new
priority: low
size: S
---
`)

	items, err := ScanBacklogDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item (hidden skipped), got %d", len(items))
	}
	if items[0].ID != "0001" {
		t.Errorf("expected visible item ID 0001, got %q", items[0].ID)
	}
}

func TestScanBacklogDir_SkipsNonMdFiles(t *testing.T) {
	dir := t.TempDir()

	writeBacklogFile(t, dir, "0001-feat-real.md", `---
id: "0001"
title: Real
type: feat
status: new
priority: low
size: S
---
`)
	writeBacklogFile(t, dir, "notes.txt", "just a text file")
	writeBacklogFile(t, dir, "0002-feat-ignored.json", `{"id":"0002"}`)

	items, err := ScanBacklogDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestScanBacklogDir_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()

	// Create a subdirectory with a .md file inside it
	subdir := filepath.Join(dir, "archive")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	writeBacklogFile(t, subdir, "0099-feat-archived.md", `---
id: "0099"
title: Archived
type: feat
status: done
priority: low
size: S
---
`)
	writeBacklogFile(t, dir, "0001-feat-active.md", `---
id: "0001"
title: Active
type: feat
status: new
priority: low
size: S
---
`)

	items, err := ScanBacklogDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only include the active item, not the one in the subdirectory
	if len(items) != 1 {
		t.Errorf("expected 1 item (subdirs skipped), got %d", len(items))
	}
}

// ─── ScanBacklog ─────────────────────────────────────────────────────────────

func setupPlanningDir(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	backlogDir := filepath.Join(base, "backlog")
	archiveDir := filepath.Join(backlogDir, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeBacklogFile(t, backlogDir, "0001-feat-active-one.md", `---
id: "0001"
title: Active One
type: feat
status: new
priority: high
size: L
---
`)
	writeBacklogFile(t, backlogDir, "0002-fix-active-two.md", `---
id: "0002"
title: Active Two
type: fix
status: in-progress
priority: medium
size: M
---
`)
	writeBacklogFile(t, archiveDir, "0010-feat-archived-one.md", `---
id: "0010"
title: Archived One
type: feat
status: done
priority: low
size: S
---
`)

	return base
}

func TestScanBacklog_ActiveItemsFound(t *testing.T) {
	base := setupPlanningDir(t)

	result, err := ScanBacklog(base, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Active) != 2 {
		t.Errorf("expected 2 active items, got %d", len(result.Active))
	}
	if result.Summary.ActiveCount != 2 {
		t.Errorf("summary active_count: got %d, want 2", result.Summary.ActiveCount)
	}
}

func TestScanBacklog_ArchiveExcludedByDefault(t *testing.T) {
	base := setupPlanningDir(t)

	result, err := ScanBacklog(base, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Archived) != 0 {
		t.Errorf("expected 0 archived items when includeArchive=false, got %d", len(result.Archived))
	}
	if result.Summary.ArchivedCount != 0 {
		t.Errorf("summary archived_count: got %d, want 0", result.Summary.ArchivedCount)
	}
}

func TestScanBacklog_ArchiveIncludedWhenFlagTrue(t *testing.T) {
	base := setupPlanningDir(t)

	result, err := ScanBacklog(base, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Archived) != 1 {
		t.Errorf("expected 1 archived item, got %d", len(result.Archived))
	}
	if result.Summary.ArchivedCount != 1 {
		t.Errorf("summary archived_count: got %d, want 1", result.Summary.ArchivedCount)
	}
}

func TestScanBacklog_SummaryCountsCorrect(t *testing.T) {
	base := setupPlanningDir(t)

	result, err := ScanBacklog(base, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary.ActiveCount != 2 {
		t.Errorf("active_count: got %d, want 2", result.Summary.ActiveCount)
	}
	if result.Summary.ArchivedCount != 1 {
		t.Errorf("archived_count: got %d, want 1", result.Summary.ArchivedCount)
	}
}

func TestScanBacklog_NonExistentPlanningDir(t *testing.T) {
	// backlog dir does not exist → active = nil/empty, no error
	result, err := ScanBacklog("/tmp/no-such-planning-dir-xyz", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Summary.ActiveCount != 0 {
		t.Errorf("expected 0 active count for missing dir, got %d", result.Summary.ActiveCount)
	}
}

// ─── IsOldFormat ─────────────────────────────────────────────────────────────

func TestIsOldFormat_YAMLFrontmatter_ReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	path := writeBacklogFile(t, dir, "0001-feat-new.md", `---
id: "0001"
title: New Format
type: feat
status: new
priority: low
size: S
---

# New Format
`)
	if IsOldFormat(path) {
		t.Error("expected IsOldFormat=false for YAML frontmatter file")
	}
}

func TestIsOldFormat_BlockquoteFormat_ReturnsTrue(t *testing.T) {
	dir := t.TempDir()
	path := writeBacklogFile(t, dir, "0002-feat-old.md", `> **Status:** new
> **Priority:** high
> **Size:** M

# Old Format Feature
`)
	if !IsOldFormat(path) {
		t.Error("expected IsOldFormat=true for blockquote format file")
	}
}

func TestIsOldFormat_NonExistentFile_ReturnsFalse(t *testing.T) {
	if IsOldFormat("/tmp/this-does-not-exist-at-all.md") {
		t.Error("expected IsOldFormat=false for non-existent file")
	}
}

func TestIsOldFormat_NoStatusLine_ReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	// Has blockquote but no Status line
	path := writeBacklogFile(t, dir, "0003-feat-no-status.md", `> **Priority:** medium

# Some Feature
`)
	if IsOldFormat(path) {
		t.Error("expected IsOldFormat=false for file without Status blockquote")
	}
}

// ─── BacklogResult.JSON ───────────────────────────────────────────────────────

func TestBacklogResult_JSON_Structure(t *testing.T) {
	result := &BacklogResult{
		Active: []*BacklogItem{
			{ID: "0001", Title: "Alpha", Type: "feat", Status: "new", Priority: "high", Size: "L"},
		},
		Summary: BacklogSummary{ActiveCount: 1, ArchivedCount: 0},
	}

	out := result.JSON()

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v\nOutput:\n%s", err, out)
	}

	if _, ok := parsed["active"]; !ok {
		t.Error("JSON missing 'active' key")
	}
	if _, ok := parsed["summary"]; !ok {
		t.Error("JSON missing 'summary' key")
	}

	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary is not a map")
	}
	if summary["active_count"] != float64(1) {
		t.Errorf("summary.active_count: got %v, want 1", summary["active_count"])
	}
}

func TestBacklogResult_JSON_EmptyResult(t *testing.T) {
	result := &BacklogResult{
		Active:  []*BacklogItem{},
		Summary: BacklogSummary{},
	}
	out := result.JSON()

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v", err)
	}
}

func TestBacklogResult_JSON_ArchiveOmittedWhenEmpty(t *testing.T) {
	result := &BacklogResult{
		Active:  []*BacklogItem{},
		Summary: BacklogSummary{},
	}
	out := result.JSON()

	if strings.Contains(out, `"archived"`) {
		t.Error("expected 'archived' to be omitted from JSON when nil/empty")
	}
}

// ─── BacklogResult.Table ──────────────────────────────────────────────────────

func TestBacklogResult_Table_WithActiveItems(t *testing.T) {
	result := &BacklogResult{
		Active: []*BacklogItem{
			{ID: "0001", Title: "Alpha Feature", Type: "feat", Status: "new", Priority: "high", Size: "L"},
			{ID: "0002", Title: "Beta Fix", Type: "fix", Status: "in-progress", Priority: "medium", Size: "M"},
		},
		Summary: BacklogSummary{ActiveCount: 2},
	}

	out := result.Table()

	if !strings.Contains(out, "Active Backlog") {
		t.Error("expected 'Active Backlog' header in table output")
	}
	if !strings.Contains(out, "0001") {
		t.Error("expected ID 0001 in table output")
	}
	if !strings.Contains(out, "Alpha Feature") {
		t.Error("expected title 'Alpha Feature' in table output")
	}
	if !strings.Contains(out, "Summary: 2 active") {
		t.Errorf("expected summary line, got:\n%s", out)
	}
}

func TestBacklogResult_Table_EmptyActiveItems(t *testing.T) {
	result := &BacklogResult{
		Active:  []*BacklogItem{},
		Summary: BacklogSummary{},
	}

	out := result.Table()

	if !strings.Contains(out, "No active backlog items.") {
		t.Errorf("expected empty message, got:\n%s", out)
	}
}

func TestBacklogResult_Table_WithArchivedItems(t *testing.T) {
	planStr := "temp/plans/0001.md"
	result := &BacklogResult{
		Active: []*BacklogItem{
			{ID: "0001", Title: "Active", Type: "feat", Status: "new", Priority: "high", Size: "L", Plan: &planStr},
		},
		Archived: []*BacklogItem{
			{ID: "0010", Title: "Done Feature", Type: "feat", Status: "done", Priority: "low", Size: "S"},
		},
		Summary: BacklogSummary{ActiveCount: 1, ArchivedCount: 1},
	}

	out := result.Table()

	if !strings.Contains(out, "Archived") {
		t.Error("expected 'Archived' section in table output")
	}
	if !strings.Contains(out, "Done Feature") {
		t.Error("expected archived item title in output")
	}
	if !strings.Contains(out, "1 archived") {
		t.Errorf("expected archived count in summary, got:\n%s", out)
	}
	// Plan value should appear
	if !strings.Contains(out, "temp/plans/0001.md") {
		t.Error("expected plan path in table output")
	}
}

func TestBacklogResult_Table_LongTitleTruncated(t *testing.T) {
	longTitle := strings.Repeat("a", 55) // > 50 chars → truncated to 47 + "..."
	result := &BacklogResult{
		Active: []*BacklogItem{
			{ID: "0001", Title: longTitle, Type: "feat", Status: "new", Priority: "low", Size: "S"},
		},
		Summary: BacklogSummary{ActiveCount: 1},
	}

	out := result.Table()

	if !strings.Contains(out, "...") {
		t.Error("expected '...' truncation for long title in table output")
	}
}

// ─── MigrateBacklogFile ───────────────────────────────────────────────────────

func TestMigrateBacklogFile_OldFormatConvertedToYAML(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** new
> **Priority:** high
> **Size:** L
> **Created:** 2025-01-10

# Login Feature

Some description of the feature.
`
	path := writeBacklogFile(t, dir, "0001-feat-login-feature.md", content)

	if err := MigrateBacklogFile(path, "myproject"); err != nil {
		t.Fatalf("MigrateBacklogFile: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading migrated file: %v", err)
	}
	migrated := string(result)

	if !strings.HasPrefix(migrated, "---") {
		t.Error("expected migrated file to start with YAML frontmatter '---'")
	}
	if !strings.Contains(migrated, "id:") {
		t.Error("expected 'id:' in YAML frontmatter")
	}
	if !strings.Contains(migrated, "title:") {
		t.Error("expected 'title:' in YAML frontmatter")
	}
	if !strings.Contains(migrated, "myproject") {
		t.Error("expected project name 'myproject' in YAML frontmatter")
	}
}

func TestMigrateBacklogFile_BodyContentPreserved(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** in-progress
> **Priority:** medium
> **Size:** M

# My Feature

This is the body text that must survive migration.

## Details

More details here.
`
	path := writeBacklogFile(t, dir, "0002-feat-my-feature.md", content)

	if err := MigrateBacklogFile(path, "proj"); err != nil {
		t.Fatalf("MigrateBacklogFile: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading migrated file: %v", err)
	}
	migrated := string(result)

	if !strings.Contains(migrated, "This is the body text that must survive migration.") {
		t.Error("expected body text to be preserved after migration")
	}
	if !strings.Contains(migrated, "## Details") {
		t.Error("expected '## Details' heading to be preserved after migration")
	}
	if !strings.Contains(migrated, "More details here.") {
		t.Error("expected 'More details here.' to be preserved after migration")
	}
}

func TestMigrateBacklogFile_ResultParsesAsYAML(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** new
> **Priority:** low
> **Size:** XS
> **Created:** 2025-02-20

# Small Task

Description.
`
	path := writeBacklogFile(t, dir, "0005-chore-small-task.md", content)

	if err := MigrateBacklogFile(path, "testproject"); err != nil {
		t.Fatalf("MigrateBacklogFile: %v", err)
	}

	// After migration the file should be parseable via YAML path
	item, err := ParseBacklogFile(path)
	if err != nil {
		t.Fatalf("ParseBacklogFile after migration: %v", err)
	}

	if item.ID != "0005" {
		t.Errorf("ID after migration: got %q, want %q", item.ID, "0005")
	}
	// MigrateBacklogFile uses ParseBacklogFile (blockquote path) to extract
	// the title, which falls back to filename slug when H1 is not at position 0.
	// "0005-chore-small-task" → strip "0005-chore-" → "small task".
	if item.Title != "small task" {
		t.Errorf("Title after migration: got %q, want %q", item.Title, "small task")
	}
	if item.Status != "new" {
		t.Errorf("Status after migration: got %q, want %q", item.Status, "new")
	}
}

func TestMigrateBacklogFile_WithPlanField(t *testing.T) {
	dir := t.TempDir()
	content := `> **Status:** in-progress
> **Priority:** high
> **Size:** L
> **Plan:** temp/plans/0006-feat-with-plan.md

# Feature With Plan

Details.
`
	path := writeBacklogFile(t, dir, "0006-feat-with-plan.md", content)

	if err := MigrateBacklogFile(path, "proj"); err != nil {
		t.Fatalf("MigrateBacklogFile: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading migrated file: %v", err)
	}

	if !strings.Contains(string(result), "temp/plans/0006-feat-with-plan.md") {
		t.Error("expected plan path to be preserved in YAML frontmatter")
	}
}

func TestMigrateBacklogFile_NonExistentFile_ReturnsError(t *testing.T) {
	err := MigrateBacklogFile("/tmp/does-not-exist-migrate.md", "proj")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// ─── MigrateBacklogDir ────────────────────────────────────────────────────────

func TestMigrateBacklogDir_MixedFiles_OldMigratedYAMLSkipped(t *testing.T) {
	dir := t.TempDir()

	// Old-format file (blockquote, no YAML frontmatter) — should be migrated.
	writeBacklogFile(t, dir, "0001-feat-old-one.md", `> **Status:** new
> **Priority:** high
> **Size:** L
> **Created:** 2025-01-01

# Old Feature One

Body.
`)
	writeBacklogFile(t, dir, "0002-fix-old-two.md", `> **Status:** in-progress
> **Priority:** medium
> **Size:** M

# Old Fix Two

Body.
`)

	// YAML-frontmatter file — already new format, should be skipped.
	writeBacklogFile(t, dir, "0003-feat-already-yaml.md", `---
id: "0003"
title: Already YAML
type: feat
status: new
priority: low
size: S
---

# Already YAML
`)

	result, err := MigrateBacklogDir(dir, "testproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Migrated) != 2 {
		t.Errorf("Migrated: got %d, want 2; list=%v", len(result.Migrated), result.Migrated)
	}
	if len(result.Skipped) != 1 {
		t.Errorf("Skipped: got %d, want 1; list=%v", len(result.Skipped), result.Skipped)
	}

	// Verify skipped entry mentions the YAML file
	if !strings.Contains(result.Skipped[0], "0003-feat-already-yaml.md") {
		t.Errorf("Skipped entry does not name the YAML file: %q", result.Skipped[0])
	}

	// Verify the two migrated files now have YAML frontmatter
	for _, name := range result.Migrated {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading migrated file %s: %v", name, err)
		}
		if !strings.HasPrefix(string(b), "---") {
			t.Errorf("migrated file %s does not start with YAML frontmatter", name)
		}
	}
}

func TestMigrateBacklogDir_NonExistentDir_ReturnsEmptyResult(t *testing.T) {
	result, err := MigrateBacklogDir("/tmp/no-such-dir-migrate-xyz-abc", "proj")
	if err != nil {
		t.Fatalf("unexpected error for non-existent dir: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated: expected empty, got %v", result.Migrated)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("Skipped: expected empty, got %v", result.Skipped)
	}
}

func TestMigrateBacklogDir_EmptyDir_ReturnsEmptyResult(t *testing.T) {
	dir := t.TempDir()

	result, err := MigrateBacklogDir(dir, "proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Migrated) != 0 {
		t.Errorf("Migrated: expected empty, got %v", result.Migrated)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("Skipped: expected empty, got %v", result.Skipped)
	}
}
