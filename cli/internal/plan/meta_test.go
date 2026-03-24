package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPlanFile_SearchesLiveFirst(t *testing.T) {
	// Create temp directory structure: planDir/live/
	planDir := t.TempDir()
	liveDir := filepath.Join(planDir, "live")
	if err := os.MkdirAll(liveDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a plan file in live/
	liveFile := "0001-feat-my-feature-todo.md"
	if err := os.WriteFile(filepath.Join(liveDir, liveFile), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a plan file in root (older number)
	rootFile := "0000-feat-old-plan.md"
	if err := os.WriteFile(filepath.Join(planDir, rootFile), []byte("# Old Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	result := FindPlanFile(planDir, "feat/my-feature")
	if result == "" {
		t.Fatal("FindPlanFile returned empty string")
	}

	expected := filepath.Join(planDir, "live", liveFile)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindPlanFile_FallsBackToRoot(t *testing.T) {
	planDir := t.TempDir()
	liveDir := filepath.Join(planDir, "live")
	if err := os.MkdirAll(liveDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Only create a plan file in root, live/ is empty
	rootFile := "0001-feat-root-feature-todo.md"
	if err := os.WriteFile(filepath.Join(planDir, rootFile), []byte("# Root Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	result := FindPlanFile(planDir, "feat/root-feature")
	if result == "" {
		t.Fatal("FindPlanFile returned empty string")
	}

	expected := filepath.Join(planDir, rootFile)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestFindPlanFile_EmptyDirs(t *testing.T) {
	planDir := t.TempDir()
	liveDir := filepath.Join(planDir, "live")
	if err := os.MkdirAll(liveDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := FindPlanFile(planDir, "feat/something")
	if result != "" {
		t.Errorf("expected empty string for empty dirs, got %s", result)
	}
}

func TestFindPlanFile_NoLiveDir(t *testing.T) {
	planDir := t.TempDir()

	// Only root file, no live/ dir at all
	rootFile := "0001-feat-thing-todo.md"
	if err := os.WriteFile(filepath.Join(planDir, rootFile), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	result := FindPlanFile(planDir, "feat/thing")
	if result == "" {
		t.Fatal("FindPlanFile returned empty string")
	}

	expected := filepath.Join(planDir, rootFile)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestGetNextNum_ScansLiveAndRoot(t *testing.T) {
	planDir := t.TempDir()
	liveDir := filepath.Join(planDir, "live")
	if err := os.MkdirAll(liveDir, 0755); err != nil {
		t.Fatal(err)
	}

	// File in root with num 0003
	if err := os.WriteFile(filepath.Join(planDir, "0003-feat-old.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// File in live/ with num 0005
	if err := os.WriteFile(filepath.Join(liveDir, "0005-feat-new-todo.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result := GetNextNum(planDir)
	if result != "0006" {
		t.Errorf("expected 0006, got %s", result)
	}
}

func TestGetNextNum_EmptyDirs(t *testing.T) {
	planDir := t.TempDir()
	result := GetNextNum(planDir)
	if result != "0001" {
		t.Errorf("expected 0001, got %s", result)
	}
}

func TestMetaResult_GetField_AllFields(t *testing.T) {
	m := &MetaResult{
		NextNum:    "0002",
		BaseBranch: "staging",
		Branch:     "feat/my-feature",
		PlanFile:   "blueprint/live/0001-feat-my-feature-todo.md",
		PlanNum:    "0001",
		Status:     "in-progress",
		Progress:   "3/10 tasks",
		Project:    "blueprint",
		GitRemote:  "git@github.com:skaisser/blueprint.git",
		Today:      "2026-03-24 10:00",
	}

	tests := []struct {
		field    string
		expected string
	}{
		{"next_num", "0002"},
		{"base_branch", "staging"},
		{"branch", "feat/my-feature"},
		{"plan_file", "blueprint/live/0001-feat-my-feature-todo.md"},
		{"plan_num", "0001"},
		{"status", "in-progress"},
		{"progress", "3/10 tasks"},
		{"project", "blueprint"},
		{"git_remote", "git@github.com:skaisser/blueprint.git"},
		{"today", "2026-03-24 10:00"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := m.GetField(tt.field)
			if got != tt.expected {
				t.Errorf("GetField(%q) = %q, want %q", tt.field, got, tt.expected)
			}
		})
	}
}

func TestMetaResult_GetField_Unknown(t *testing.T) {
	m := &MetaResult{
		NextNum: "0001",
		Branch:  "main",
	}

	got := m.GetField("nonexistent")
	if got != "" {
		t.Errorf("GetField(\"nonexistent\") = %q, want empty string", got)
	}

	got = m.GetField("")
	if got != "" {
		t.Errorf("GetField(\"\") = %q, want empty string", got)
	}
}
