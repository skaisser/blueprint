package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePRBody_NoInputs(t *testing.T) {
	body := generatePRBody("", "")

	if !strings.Contains(body, "## Summary") {
		t.Error("expected Summary section")
	}
	if !strings.Contains(body, "## Changes") {
		t.Error("expected Changes section")
	}
	if !strings.Contains(body, "## Test Plan") {
		t.Error("expected Test Plan section")
	}
	// Should NOT have Related section when no issue
	if strings.Contains(body, "## Related") {
		t.Error("expected no Related section without issue")
	}
	if strings.Contains(body, "Closes #") {
		t.Error("expected no Closes ref without issue")
	}
}

func TestGeneratePRBody_WithIssue(t *testing.T) {
	body := generatePRBody("42", "")

	if !strings.Contains(body, "## Related") {
		t.Error("expected Related section")
	}
	if !strings.Contains(body, "Closes #42") {
		t.Error("expected Closes #42")
	}
}

func TestGeneratePRBody_WithIssueHash(t *testing.T) {
	// Should handle "#42" input by stripping the hash
	body := generatePRBody("#42", "")

	if !strings.Contains(body, "Closes #42") {
		t.Error("expected Closes #42, got double hash or missing")
	}
	if strings.Contains(body, "Closes ##42") {
		t.Error("should not produce Closes ##42")
	}
}

func TestGeneratePRBody_WithPlanFile(t *testing.T) {
	// Create a temp plan file
	dir := t.TempDir()
	planFile := filepath.Join(dir, "0001-test-plan.md")

	content := `---
id: "0001"
title: Add authentication system
status: in-progress
issue: "99"
tags: [auth, security]
---

# Add Authentication System

## Goal

Implement OAuth2 authentication.

## Phase 1: Database setup [H]

- [x] Create users table
- [ ] Add OAuth columns

## Phase 2: API endpoints [S]

- [ ] Login endpoint
- [ ] Logout endpoint

## Phase 3: Frontend integration [S]

- [ ] Login form
`
	os.WriteFile(planFile, []byte(content), 0644)

	body := generatePRBody("", planFile)

	if !strings.Contains(body, "Add authentication system") {
		t.Errorf("expected plan title in summary, got:\n%s", body)
	}
	if !strings.Contains(body, "Database setup") {
		t.Errorf("expected phase 1 in changes, got:\n%s", body)
	}
	if !strings.Contains(body, "API endpoints") {
		t.Errorf("expected phase 2 in changes, got:\n%s", body)
	}
	if !strings.Contains(body, "Frontend integration") {
		t.Errorf("expected phase 3 in changes, got:\n%s", body)
	}
	// Should pick up issue from plan frontmatter
	if !strings.Contains(body, "Closes #99") {
		t.Errorf("expected Closes #99 from plan frontmatter, got:\n%s", body)
	}
}

func TestGeneratePRBody_IssueOverridesPlan(t *testing.T) {
	dir := t.TempDir()
	planFile := filepath.Join(dir, "0001-test.md")

	content := `---
id: "0001"
title: Some feature
issue: "50"
tags: [test]
---

# Some feature
`
	os.WriteFile(planFile, []byte(content), 0644)

	// Explicit issue flag should override plan issue
	body := generatePRBody("77", planFile)

	if !strings.Contains(body, "Closes #77") {
		t.Errorf("expected explicit issue 77, got:\n%s", body)
	}
	// Should not also have the plan issue
	if strings.Contains(body, "Closes #50") {
		t.Errorf("should not have plan issue when explicit issue given, got:\n%s", body)
	}
}

func TestGeneratePRBody_NonexistentPlanFile(t *testing.T) {
	body := generatePRBody("", "/nonexistent/plan.md")

	// Should still produce a valid template
	if !strings.Contains(body, "## Summary") {
		t.Error("expected Summary section even with missing plan file")
	}
}

func TestExtractPlanInfo_NoPlanFile(t *testing.T) {
	summary, changes, issue := extractPlanInfo("/nonexistent/file.md")
	if summary != "" {
		t.Error("expected empty summary")
	}
	if len(changes) != 0 {
		t.Error("expected no changes")
	}
	if issue != "" {
		t.Error("expected empty issue")
	}
}

func TestExtractPlanInfo_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "plan.md")
	os.WriteFile(f, []byte("# Just a heading\n\n## Phase 1: Setup [H]\n\nSome text\n"), 0644)

	summary, changes, issue := extractPlanInfo(f)
	if summary != "" {
		t.Error("expected empty summary with no frontmatter")
	}
	if len(changes) != 1 || changes[0] != "Setup" {
		t.Errorf("expected [Setup], got %v", changes)
	}
	if issue != "" {
		t.Error("expected empty issue")
	}
}

func TestExtractPlanInfo_PhaseComplexityStripping(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "plan.md")
	content := `---
id: "0001"
title: Test
tags: [t]
---

## Phase 1: Alpha [H]

## Phase 2: Beta [S]

## Phase 3: Gamma [O]
`
	os.WriteFile(f, []byte(content), 0644)

	_, changes, _ := extractPlanInfo(f)
	expected := []string{"Alpha", "Beta", "Gamma"}
	if len(changes) != len(expected) {
		t.Fatalf("expected %d changes, got %d: %v", len(expected), len(changes), changes)
	}
	for i, e := range expected {
		if changes[i] != e {
			t.Errorf("changes[%d] = %q, want %q", i, changes[i], e)
		}
	}
}

func TestIssueCmd_Registration(t *testing.T) {
	// Verify issue command is registered on root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "issue" {
			found = true
			// Check subcommands
			subCmds := cmd.Commands()
			hasFetch := false
			hasClose := false
			for _, sub := range subCmds {
				if sub.Use == "fetch <number>" {
					hasFetch = true
				}
				if sub.Use == "close <number>" {
					hasClose = true
				}
			}
			if !hasFetch {
				t.Error("issue command missing fetch subcommand")
			}
			if !hasClose {
				t.Error("issue command missing close subcommand")
			}
			break
		}
	}
	if !found {
		t.Error("issue command not registered on rootCmd")
	}
}

func TestPrBodyCmd_Registration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "pr-body" {
			found = true
			// Check flags
			issueFlag := cmd.Flags().Lookup("issue")
			if issueFlag == nil {
				t.Error("pr-body missing --issue flag")
			}
			planFlag := cmd.Flags().Lookup("plan")
			if planFlag == nil {
				t.Error("pr-body missing --plan flag")
			}
			break
		}
	}
	if !found {
		t.Error("pr-body command not registered on rootCmd")
	}
}
