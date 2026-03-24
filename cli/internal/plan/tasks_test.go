package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTasks_MixedCheckboxes(t *testing.T) {
	content := `## Phase 1
- [x] Task one [H]
- [ ] Task two [S]
- [x] Task three [O]
- [ ] Task four
`
	tasks := ParseTasks(content)
	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasks))
	}

	// Task 1: done, H
	if !tasks[0].Done || tasks[0].Complexity != "H" {
		t.Errorf("task 0: expected done=true, complexity=H, got done=%v, complexity=%s", tasks[0].Done, tasks[0].Complexity)
	}

	// Task 2: not done, S
	if tasks[1].Done || tasks[1].Complexity != "S" {
		t.Errorf("task 1: expected done=false, complexity=S, got done=%v, complexity=%s", tasks[1].Done, tasks[1].Complexity)
	}

	// Task 3: done, O
	if !tasks[2].Done || tasks[2].Complexity != "O" {
		t.Errorf("task 2: expected done=true, complexity=O, got done=%v, complexity=%s", tasks[2].Done, tasks[2].Complexity)
	}

	// Task 4: not done, no complexity
	if tasks[3].Done || tasks[3].Complexity != "" {
		t.Errorf("task 3: expected done=false, complexity='', got done=%v, complexity=%s", tasks[3].Done, tasks[3].Complexity)
	}
}

func TestParseTasks_Empty(t *testing.T) {
	tasks := ParseTasks("No tasks here\nJust some text\n")
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseTasks_NestedTasks(t *testing.T) {
	content := `- [x] Parent task [S]
  - [ ] Sub-task one [H]
  - [x] Sub-task two [H]
`
	tasks := ParseTasks(content)
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestParseTasks_CaseInsensitiveCheckmark(t *testing.T) {
	content := `- [X] Done with uppercase X [H]
- [x] Done with lowercase x [S]
- [ ] Not done [H]
`
	tasks := ParseTasks(content)
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	if !tasks[0].Done {
		t.Error("task 0 should be done (uppercase X)")
	}
	if !tasks[1].Done {
		t.Error("task 1 should be done (lowercase x)")
	}
	if tasks[2].Done {
		t.Error("task 2 should not be done")
	}
}

func TestExtractTasks(t *testing.T) {
	dir := t.TempDir()
	planFile := filepath.Join(dir, "plan.md")
	content := `# Test Plan

## Phase 1
- [x] First task [H]
- [x] Second task [S]

## Phase 2
- [ ] Third task [S]
- [ ] Fourth task [O]
`
	if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ExtractTasks(planFile)
	if err != nil {
		t.Fatal(err)
	}

	if result.Total != 4 {
		t.Errorf("expected total=4, got %d", result.Total)
	}
	if result.Done != 2 {
		t.Errorf("expected done=2, got %d", result.Done)
	}
	if result.PlanFile != planFile {
		t.Errorf("expected plan_file=%s, got %s", planFile, result.PlanFile)
	}
}

func TestExtractTasks_FileNotFound(t *testing.T) {
	_, err := ExtractTasks("/nonexistent/path/plan.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPlanProfile(t *testing.T) {
	dir := t.TempDir()
	planFile := filepath.Join(dir, "plan.md")
	content := `# Test Plan

## Phase 1: Foundation [S]
- [x] Task one [H]
- [x] Task two [H]
- [ ] Task three [S]

## Phase 2: Advanced [S]
- [ ] Task four [S]
- [ ] Task five [O]
- [x] Task six [H]

## Phase 3: Final [H]
- [ ] Task seven [H]
- [x] Task eight [S]
`
	if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := PlanProfile(planFile)
	if err != nil {
		t.Fatal(err)
	}

	if result.TotalTasks != 8 {
		t.Errorf("expected total_tasks=8, got %d", result.TotalTasks)
	}
	if result.DoneTasks != 4 {
		t.Errorf("expected done_tasks=4, got %d", result.DoneTasks)
	}
	if result.OpenTasks != 4 {
		t.Errorf("expected open_tasks=4, got %d", result.OpenTasks)
	}
	if result.Phases != 3 {
		t.Errorf("expected phases=3, got %d", result.Phases)
	}
	if result.H != 4 {
		t.Errorf("expected h=4, got %d", result.H)
	}
	if result.S != 3 {
		t.Errorf("expected s=3, got %d", result.S)
	}
	if result.O != 1 {
		t.Errorf("expected o=1, got %d", result.O)
	}
	if result.MaxComplexity != "O" {
		t.Errorf("expected max_complexity=O, got %s", result.MaxComplexity)
	}
}

func TestPlanProfile_NoOComplexity(t *testing.T) {
	content := `## Phase 1
- [x] Task one [H]
- [ ] Task two [S]
`
	result := PlanProfileFromContent(content)
	if result.MaxComplexity != "S" {
		t.Errorf("expected max_complexity=S, got %s", result.MaxComplexity)
	}
}

func TestPlanProfile_OnlyH(t *testing.T) {
	content := `## Phase 1
- [x] Task one [H]
- [ ] Task two [H]
`
	result := PlanProfileFromContent(content)
	if result.MaxComplexity != "H" {
		t.Errorf("expected max_complexity=H, got %s", result.MaxComplexity)
	}
}

func TestPlanProfile_NoTasks(t *testing.T) {
	content := `# Empty plan
Some description.
`
	result := PlanProfileFromContent(content)
	if result.TotalTasks != 0 {
		t.Errorf("expected total_tasks=0, got %d", result.TotalTasks)
	}
	if result.MaxComplexity != "" {
		t.Errorf("expected max_complexity='', got %s", result.MaxComplexity)
	}
}

func TestPlanStatus_Completed(t *testing.T) {
	content := `# Plan

## Phase 1
- [x] Task one [H]
- [x] Task two [S]

## Phase 2
- [x] Task three [H]
`
	result := PlanStatusFromContent(content, "test-plan.md")

	if result.Total != 3 {
		t.Errorf("expected total=3, got %d", result.Total)
	}
	if result.Done != 3 {
		t.Errorf("expected done=3, got %d", result.Done)
	}
	if result.Phases != 2 {
		t.Errorf("expected phases=2, got %d", result.Phases)
	}
	if result.PhasesDone != 2 {
		t.Errorf("expected phases_done=2, got %d", result.PhasesDone)
	}
	if result.Status != "completed" {
		t.Errorf("expected status=completed, got %s", result.Status)
	}
}

func TestPlanStatus_InProgress(t *testing.T) {
	content := `# Plan

## Phase 1
- [x] Task one [H]
- [x] Task two [S]

## Phase 2
- [ ] Task three [H]
- [x] Task four [S]
`
	result := PlanStatusFromContent(content, "test-plan.md")

	if result.Total != 4 {
		t.Errorf("expected total=4, got %d", result.Total)
	}
	if result.Done != 3 {
		t.Errorf("expected done=3, got %d", result.Done)
	}
	if result.PhasesDone != 1 {
		t.Errorf("expected phases_done=1, got %d", result.PhasesDone)
	}
	if result.Status != "in-progress" {
		t.Errorf("expected status=in-progress, got %s", result.Status)
	}
}

func TestPlanStatus_NotStarted(t *testing.T) {
	content := `# Plan

## Phase 1
- [ ] Task one [H]
- [ ] Task two [S]
`
	result := PlanStatusFromContent(content, "test-plan.md")

	if result.Done != 0 {
		t.Errorf("expected done=0, got %d", result.Done)
	}
	if result.Status != "not-started" {
		t.Errorf("expected status=not-started, got %s", result.Status)
	}
}

func TestBaselineDiff(t *testing.T) {
	// Test the diff logic directly using ParseTasks
	oldContent := `## Phase 1
- [x] Keep this task [H]
- [ ] Remove this task [S]
- [ ] Another kept task [H]
`
	newContent := `## Phase 1
- [x] Keep this task [H]
- [ ] Another kept task [H]
- [ ] Brand new task [O]
`
	oldTasks := ParseTasks(oldContent)
	newTasks := ParseTasks(newContent)

	// Build maps
	oldMap := make(map[string]Task)
	for _, t := range oldTasks {
		oldMap[t.Text] = t
	}
	newMap := make(map[string]Task)
	for _, t := range newTasks {
		newMap[t.Text] = t
	}

	// Added
	var added []Task
	for _, task := range newTasks {
		if _, exists := oldMap[task.Text]; !exists {
			added = append(added, task)
		}
	}

	// Removed
	var removed []Task
	for _, task := range oldTasks {
		if _, exists := newMap[task.Text]; !exists {
			removed = append(removed, task)
		}
	}

	if len(added) != 1 {
		t.Fatalf("expected 1 added task, got %d", len(added))
	}
	if added[0].Text != "Brand new task [O]" {
		t.Errorf("expected added task text 'Brand new task [O]', got '%s'", added[0].Text)
	}

	if len(removed) != 1 {
		t.Fatalf("expected 1 removed task, got %d", len(removed))
	}
	if removed[0].Text != "Remove this task [S]" {
		t.Errorf("expected removed task text 'Remove this task [S]', got '%s'", removed[0].Text)
	}
}

func TestTasksResultJSON(t *testing.T) {
	result := &TasksResult{
		PlanFile: "test.md",
		Tasks:    []Task{{Text: "test task [H]", Done: true, Complexity: "H"}},
		Total:    1,
		Done:     1,
	}
	json := result.JSON()
	if json == "" {
		t.Error("expected non-empty JSON")
	}
}

func TestProfileResultJSON(t *testing.T) {
	result := &ProfileResult{
		OpenTasks:     2,
		DoneTasks:     3,
		TotalTasks:    5,
		Phases:        2,
		MaxComplexity: "S",
		H:             2,
		S:             3,
	}
	json := result.JSON()
	if json == "" {
		t.Error("expected non-empty JSON")
	}
}

func TestStatusResultJSON(t *testing.T) {
	result := &StatusResult{
		PlanFile:   "test.md",
		Total:      10,
		Done:       5,
		Phases:     3,
		PhasesDone: 1,
		Status:     "in-progress",
	}
	json := result.JSON()
	if json == "" {
		t.Error("expected non-empty JSON")
	}
}
