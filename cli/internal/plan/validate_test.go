package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePlan_Valid(t *testing.T) {
	content := `# Test Plan

## Acceptance Criteria
- [ ] Everything works [H]
- [ ] Tests pass [S]

## Phase 1
- [x] Task one [H]
- [ ] Task two [S]

## Execution Strategy
> Approach: sequential
`
	result := ValidatePlanContent(content)

	if !result.Valid {
		t.Errorf("expected valid=true, got false; errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidatePlan_MissingComplexityMarkers(t *testing.T) {
	content := `# Test Plan

## Acceptance Criteria
- [ ] Everything works [H]

## Phase 1
- [x] Task one
- [ ] Task two [S]
- [ ] Task three

## Execution Strategy
> Approach: sequential
`
	result := ValidatePlanContent(content)

	if result.Valid {
		t.Error("expected valid=false for missing markers")
	}

	foundMarkerError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "missing complexity markers") {
			foundMarkerError = true
			if !strings.Contains(e, "2 tasks") {
				t.Errorf("expected 2 tasks missing markers, got: %s", e)
			}
		}
	}
	if !foundMarkerError {
		t.Error("expected error about missing complexity markers")
	}
}

func TestValidatePlan_MissingExecutionStrategy(t *testing.T) {
	content := `# Test Plan

## Acceptance Criteria
- [ ] Everything works [H]

## Phase 1
- [x] Task one [H]
`
	result := ValidatePlanContent(content)

	if result.Valid {
		t.Error("expected valid=false for missing Execution Strategy")
	}

	foundError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "Execution Strategy") {
			foundError = true
		}
	}
	if !foundError {
		t.Error("expected error about missing Execution Strategy section")
	}
}

func TestValidatePlan_MissingAcceptanceCriteria(t *testing.T) {
	content := `# Test Plan

## Phase 1
- [x] Task one [H]

## Execution Strategy
> Approach: sequential
`
	result := ValidatePlanContent(content)

	if result.Valid {
		t.Error("expected valid=false for missing Acceptance Criteria")
	}

	foundError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "Acceptance Criteria") {
			foundError = true
		}
	}
	if !foundError {
		t.Error("expected error about missing Acceptance Criteria section")
	}
}

func TestValidatePlan_StaleFileReference(t *testing.T) {
	content := `# Test Plan

## Acceptance Criteria
- [ ] All tests pass [H]

## Phase 1
- [ ] Update ` + "`cli/cmd/nonexistent-file.go`" + ` [S]
- [ ] Fix ` + "`cli/internal/plan/tasks.go`" + ` [H]

## Execution Strategy
> Approach: sequential
`
	result := ValidatePlanContent(content)

	// Should have a warning about nonexistent file
	foundWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "nonexistent-file.go") && strings.Contains(w, "not found") {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Errorf("expected warning about nonexistent file reference, got warnings: %v", result.Warnings)
	}
}

func TestValidatePlan_ExistingFileReference(t *testing.T) {
	// Create a temp file that actually exists
	dir := t.TempDir()
	testFile := filepath.Join(dir, "existing.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	content := "# Test Plan\n\n## Acceptance Criteria\n- [ ] All tests pass [H]\n\n## Phase 1\n- [ ] Update `" + testFile + "` [S]\n\n## Execution Strategy\n> Approach\n"

	result := ValidatePlanContent(content)

	// Should NOT have a warning about existing file
	for _, w := range result.Warnings {
		if strings.Contains(w, "existing.go") {
			t.Errorf("unexpected warning about existing file: %s", w)
		}
	}
}

func TestValidatePlan_AllMissing(t *testing.T) {
	content := `# Bare Plan
- [ ] Task without marker
- [x] Another unmarked task
`
	result := ValidatePlanContent(content)

	if result.Valid {
		t.Error("expected valid=false")
	}

	// Should have errors for: missing markers, missing strategy, missing acceptance criteria
	if len(result.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestValidatePlan_FromFile(t *testing.T) {
	dir := t.TempDir()
	planFile := filepath.Join(dir, "plan.md")
	content := `# Test Plan

## Acceptance Criteria
- [ ] Works [H]

## Phase 1
- [x] Task [H]

## Execution Strategy
> Done
`
	if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ValidatePlan(planFile)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Valid {
		t.Errorf("expected valid plan, got errors: %v", result.Errors)
	}
}

func TestValidatePlan_FileNotFound(t *testing.T) {
	_, err := ValidatePlan("/nonexistent/path/plan.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidateResultJSON(t *testing.T) {
	result := &PlanValidateResult{
		Valid:    false,
		Warnings: []string{"test warning"},
		Errors:   []string{"test error"},
	}
	json := result.JSON()
	if json == "" {
		t.Error("expected non-empty JSON")
	}
	if !strings.Contains(json, "test warning") {
		t.Error("expected JSON to contain warning")
	}
	if !strings.Contains(json, "test error") {
		t.Error("expected JSON to contain error")
	}
}
