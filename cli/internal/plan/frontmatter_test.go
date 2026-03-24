package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateFrontmatter_ValidBacklog(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "backlog")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0001"
title: "Add dark mode"
status: backlog
tags: [ui, feature]
---
# Add dark mode
`
	path := createTestFile(t, dir, "0001-feat-dark-mode.md", content)
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFrontmatter_ValidLive(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0002"
title: "Implement auth"
status: in-progress
complexity: S
branch: feat/auth
tags: [backend]
---
# Implement auth
`
	path := createTestFile(t, dir, "0002-feat-auth-todo.md", content)
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFrontmatter_ValidUpstream(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "upstream")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0003"
title: "Setup CI"
status: completed
branch: feat/ci
tags: [devops]
---
# Setup CI
`
	path := createTestFile(t, dir, "0003-feat-ci.md", content)
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFrontmatter_ValidExpired(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "expired")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0004"
title: "Old feature"
status: cancelled
tags: [deprecated]
---
# Old feature
`
	path := createTestFile(t, dir, "0004-feat-old.md", content)
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFrontmatter_MissingFields(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0005"
---
# Incomplete plan
`
	path := createTestFile(t, dir, "0005-feat-incomplete.md", content)
	result := ValidateFrontmatter(path)

	if result.Valid {
		t.Error("expected invalid, got valid")
	}

	// Should report missing title, status, tags, complexity, branch
	errStr := strings.Join(result.Errors, " | ")
	for _, field := range []string{"title", "status", "tags", "complexity", "branch"} {
		if !strings.Contains(errStr, field) {
			t.Errorf("expected error about missing '%s', got: %v", field, result.Errors)
		}
	}
}

func TestValidateFrontmatter_StatusMismatch(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0006"
title: "Wrong status"
status: backlog
complexity: S
branch: feat/wrong
tags: [test]
---
# Wrong status
`
	path := createTestFile(t, dir, "0006-feat-wrong.md", content)
	result := ValidateFrontmatter(path)

	if result.Valid {
		t.Error("expected invalid due to status mismatch")
	}

	errStr := strings.Join(result.Errors, " | ")
	if !strings.Contains(errStr, "status 'backlog' but file is in live/") {
		t.Errorf("expected status mismatch error, got: %v", result.Errors)
	}
}

func TestValidateFrontmatter_NoFrontmatter(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `# Just a heading
Some content without frontmatter.
`
	path := createTestFile(t, dir, "0007-feat-nofm.md", content)
	result := ValidateFrontmatter(path)

	if result.Valid {
		t.Error("expected invalid for missing frontmatter")
	}
}

func TestFixFrontmatter_AddsMissingFields(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0008"
---
# My feature plan
`
	path := createTestFile(t, dir, "0008-feat-myplan.md", content)

	changes, err := FixFrontmatter(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("expected changes, got none")
	}

	// Verify the file was updated
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("file should be valid after fix, got errors: %v", result.Errors)
	}
}

func TestFixFrontmatter_DryRunDoesNotWrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "backlog")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0009"
---
# Backlog item
`
	path := createTestFile(t, dir, "0009-feat-item.md", content)

	changes, err := FixFrontmatter(path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("expected changes reported, got none")
	}

	// File should NOT have been modified
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "tags:") {
		t.Error("dry-run should not have modified the file")
	}
}

func TestFixFrontmatter_FixesStatusMismatch(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "live")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0010"
title: "Some feature"
status: backlog
complexity: S
branch: feat/test
tags: [test]
---
# Some feature
`
	path := createTestFile(t, dir, "0010-feat-test.md", content)

	changes, err := FixFrontmatter(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have fixed the status
	foundStatusFix := false
	for _, c := range changes {
		if strings.Contains(c, "fix status") {
			foundStatusFix = true
		}
	}
	if !foundStatusFix {
		t.Error("expected status fix change")
	}

	// Verify file is now valid
	result := ValidateFrontmatter(path)
	if !result.Valid {
		t.Errorf("file should be valid after fix, got errors: %v", result.Errors)
	}
}

func TestFixFrontmatter_NoFrontmatter(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "backlog")
	os.MkdirAll(dir, 0755)

	content := `# A heading
Some content
`
	path := createTestFile(t, dir, "0011-feat-nofm.md", content)

	changes, err := FixFrontmatter(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("expected changes")
	}

	// Should have added frontmatter
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "---\n") {
		t.Error("expected YAML frontmatter to be added")
	}
}

func TestFixFrontmatter_NoChangesNeeded(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "backlog")
	os.MkdirAll(dir, 0755)

	content := `---
id: "0012"
title: "Complete item"
status: backlog
tags: [test]
created: "2025-01-01"
---
# Complete item
`
	path := createTestFile(t, dir, "0012-feat-complete.md", content)

	changes, err := FixFrontmatter(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if changes != nil {
		t.Errorf("expected no changes, got: %v", changes)
	}
}
