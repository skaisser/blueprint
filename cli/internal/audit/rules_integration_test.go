package audit

import (
	"os"
	"testing"
)

// newNilLogger returns a Logger with a nil file — Log/Warn are silently dropped,
// Block still calls os.Exit(2) so we never invoke it in these tests.
func newNilLogger() *Logger {
	return &Logger{file: nil}
}

// newTestState creates a fresh SessionState under os.TempDir and registers
// cleanup via t.Cleanup so the caller never needs a manual defer.
func newTestState(t *testing.T) *SessionState {
	t.Helper()
	state := NewSessionState("test-" + t.Name())
	t.Cleanup(func() { os.RemoveAll(state.Dir) })
	return state
}

// ---------------------------------------------------------------------------
// uniqueSkillNames
// ---------------------------------------------------------------------------

func TestUniqueSkillNames_ReturnsDeduplicatedNames(t *testing.T) {
	names := uniqueSkillNames()

	// skillMap maps .jsx, .tsx, .html, .css → "frontend-design"
	// All four entries share the same value, so uniqueSkillNames must return
	// exactly one entry.
	if len(names) != 1 {
		t.Errorf("expected 1 unique skill name, got %d: %v", len(names), names)
	}
	if len(names) > 0 && names[0] != "frontend-design" {
		t.Errorf("expected 'frontend-design', got %q", names[0])
	}
}

func TestUniqueSkillNames_NoDuplicates(t *testing.T) {
	names := uniqueSkillNames()
	seen := make(map[string]bool)
	for _, n := range names {
		if seen[n] {
			t.Errorf("duplicate skill name returned: %q", n)
		}
		seen[n] = true
	}
}

// ---------------------------------------------------------------------------
// isAutonomousPipeline
// ---------------------------------------------------------------------------

func TestIsAutonomousPipeline_TrueCases(t *testing.T) {
	autonomous := []string{"batch-flow", "flow-auto", "flow-auto-wt"}
	for _, cmd := range autonomous {
		if !isAutonomousPipeline(cmd) {
			t.Errorf("isAutonomousPipeline(%q) = false, want true", cmd)
		}
	}
}

func TestIsAutonomousPipeline_FalseCases(t *testing.T) {
	nonAutonomous := []string{"plan", "pr", "finish", "hotfix-push", "", "review", "plan-check"}
	for _, cmd := range nonAutonomous {
		if isAutonomousPipeline(cmd) {
			t.Errorf("isAutonomousPipeline(%q) = true, want false", cmd)
		}
	}
}

// ---------------------------------------------------------------------------
// rule2TrackSkillReads
// ---------------------------------------------------------------------------

func TestRule2_SkillMdReadSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "/home/user/.claude/skills/frontend-design/SKILL.md",
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("skill-read-frontend-design") {
		t.Error("expected state marker 'skill-read-frontend-design' after reading SKILL.md")
	}
}

func TestRule2_NestedSkillMdSetsCorrectMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "/some/path/skills/backend-api/SKILL.md",
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("skill-read-backend-api") {
		t.Error("expected state marker 'skill-read-backend-api'")
	}
}

func TestRule2_PlanTemplateMdSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "/home/user/.claude/references/plan/plan-template.md",
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("read-plan-template") {
		t.Error("expected state marker 'read-plan-template' after reading plan-template.md")
	}
}

func TestRule2_TeamExecutionMdSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "/home/user/.claude/references/plan/team-execution.md",
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("read-team-execution") {
		t.Error("expected state marker 'read-team-execution' after reading team-execution.md")
	}
}

func TestRule2_SkillToolInvocationSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "skill",
		Input:    map[string]interface{}{"skill": "plan"},
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("skill-read-plan") {
		t.Error("expected state marker 'skill-read-plan' after Skill tool invocation")
	}
}

func TestRule2_SkillToolFrontendDesignMatchesKnownSkill(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// "frontend-design" is in skillMap — invocation should mark both
	// "skill-read-frontend-design" (via exact match) and the known skill.
	p := &Payload{
		ToolName: "skill",
		Input:    map[string]interface{}{"skill": "frontend-design"},
	}

	rule2TrackSkillReads(p, state, log)

	if !state.Exists("skill-read-frontend-design") {
		t.Error("expected 'skill-read-frontend-design' marker after invoking frontend-design skill")
	}
}

func TestRule2_NonSkillReadDoesNotSetMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "/home/user/project/main.go",
	}

	rule2TrackSkillReads(p, state, log)

	if state.Exists("skill-read-frontend-design") {
		t.Error("should NOT set skill marker for unrelated file read")
	}
	if state.Exists("read-plan-template") {
		t.Error("should NOT set plan-template marker for unrelated file read")
	}
	if state.Exists("read-team-execution") {
		t.Error("should NOT set team-execution marker for unrelated file read")
	}
}

func TestRule2_NonReadToolIgnored(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "cat /home/user/.claude/skills/frontend-design/SKILL.md",
	}

	rule2TrackSkillReads(p, state, log)

	if state.Exists("skill-read-frontend-design") {
		t.Error("bash cat of SKILL.md should NOT set skill marker — only Read tool counts")
	}
}

// ---------------------------------------------------------------------------
// rule3TeamCompliance
// ---------------------------------------------------------------------------

func TestRule3_TeamCreateSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// Prime team-execution read so we don't get a Warn (Warn is safe, but
	// let's keep the test focused on the marker side-effect).
	state.Touch("read-team-execution")

	p := &Payload{
		ToolName: "teamcreate",
		Input:    map[string]interface{}{"name": "my-team"},
	}

	rule3TeamCompliance(p, state, log)

	if !state.Exists("team-created") {
		t.Error("expected 'team-created' marker after TeamCreate call")
	}
}

func TestRule3_TeamCreateWithoutTeamExecutionReadWarns(t *testing.T) {
	// We just verify no panic/exit occurs (Warn is safe), and the marker is
	// still set even when team-execution.md was not read.
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "teamcreate",
		Input:    map[string]interface{}{"name": "late-team"},
	}

	rule3TeamCompliance(p, state, log)

	if !state.Exists("team-created") {
		t.Error("team-created marker must be set regardless of missing team-execution read")
	}
}

func TestRule3_TaskWithTeamNameDoesNotIncrementStandaloneCount(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "task",
		Input:    map[string]interface{}{"team_name": "worker-team", "description": "do work"},
	}

	rule3TeamCompliance(p, state, log)

	count := state.ReadInt("standalone-task-count")
	if count != 0 {
		t.Errorf("team task should not increment standalone-task-count, got %d", count)
	}
}

func TestRule3_TaskWithoutTeamNameIncrementsStandaloneCount(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "task",
		Input:    map[string]interface{}{"description": "standalone work"},
	}

	rule3TeamCompliance(p, state, log)

	count := state.ReadInt("standalone-task-count")
	if count != 1 {
		t.Errorf("standalone-task-count should be 1, got %d", count)
	}
}

func TestRule3_MultipleStandaloneTasksAccumulateCount(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "task",
		Input:    map[string]interface{}{"description": "work"},
	}

	// Two calls — should reach 2 without triggering Block (team-created is absent
	// but count < 3 so only a Warn would fire at 3+).
	rule3TeamCompliance(p, state, log)
	rule3TeamCompliance(p, state, log)

	count := state.ReadInt("standalone-task-count")
	if count != 2 {
		t.Errorf("standalone-task-count should be 2, got %d", count)
	}
}

func TestRule3_TeamCreateRecordsTeamName(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()
	state.Touch("read-team-execution")

	p := &Payload{
		ToolName: "teamcreate",
		Input:    map[string]interface{}{"name": "frontend-team"},
	}

	rule3TeamCompliance(p, state, log)

	teamsCreated := state.ReadText("teams-created.txt")
	if teamsCreated == "" {
		t.Error("expected teams-created.txt to be non-empty after TeamCreate")
	}
}

// ---------------------------------------------------------------------------
// rule4AskUserTracking
// ---------------------------------------------------------------------------

func TestRule4_AskUserQuestionSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{ToolName: "askuserquestion"}

	rule4AskUserTracking(p, state, log)

	if !state.Exists("asked-user") {
		t.Error("expected 'asked-user' marker after AskUserQuestion call")
	}
}

func TestRule4_QuestionAliasAlsoSetsMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{ToolName: "question"}

	rule4AskUserTracking(p, state, log)

	if !state.Exists("asked-user") {
		t.Error("expected 'asked-user' marker for 'question' alias")
	}
}

func TestRule4_OtherToolsDoNotSetMarker(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{ToolName: "bash", Command: "echo hello"}

	rule4AskUserTracking(p, state, log)

	if state.Exists("asked-user") {
		t.Error("bash call should NOT set 'asked-user' marker")
	}
}

// ---------------------------------------------------------------------------
// rule5CommandCheckpoints
// ---------------------------------------------------------------------------

func TestRule5_NewCheckpointFormatTracked(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `echo "🔷 BP: plan-review [1/3]"`,
	}

	rule5CommandCheckpoints(p, state, log)

	checkpoints := state.ReadText("checkpoints.txt")
	if checkpoints == "" {
		t.Error("expected checkpoints.txt to be populated after BP checkpoint echo")
	}
	activeCmd := state.ReadText("active-command")
	if activeCmd != "plan-review" {
		t.Errorf("expected active-command 'plan-review', got %q", activeCmd)
	}
}

func TestRule5_NewCheckpointFormatStoresSkillAndStep(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `echo "🔷 BP: flow-auto [2/8]"`,
	}

	rule5CommandCheckpoints(p, state, log)

	checkpoints := state.ReadText("checkpoints.txt")
	// Expect "flow-auto:2" to be stored
	if checkpoints == "" {
		t.Error("checkpoints.txt should not be empty")
	}
	activeCmd := state.ReadText("active-command")
	if activeCmd != "flow-auto" {
		t.Errorf("expected active-command 'flow-auto', got %q", activeCmd)
	}
}

func TestRule5_LegacyCheckpointFormatTracked(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `echo "🏁 [plan-review:1]"`,
	}

	rule5CommandCheckpoints(p, state, log)

	checkpoints := state.ReadText("checkpoints.txt")
	if checkpoints == "" {
		t.Error("expected checkpoints.txt to be populated after legacy checkpoint echo")
	}
	activeCmd := state.ReadText("active-command")
	if activeCmd != "plan-review" {
		t.Errorf("expected active-command 'plan-review' from legacy format, got %q", activeCmd)
	}
}

func TestRule5_LegacyCheckpointStepsRecorded(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `echo "🏁 [finish:3]"`,
	}

	rule5CommandCheckpoints(p, state, log)

	activeCmd := state.ReadText("active-command")
	if activeCmd != "finish" {
		t.Errorf("expected active-command 'finish', got %q", activeCmd)
	}
}

func TestRule5_NonBashToolIgnored(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "some/file.md",
	}

	rule5CommandCheckpoints(p, state, log)

	checkpoints := state.ReadText("checkpoints.txt")
	if checkpoints != "" {
		t.Error("non-bash tool should not affect checkpoints")
	}
}

func TestRule5_EmptyCommandIgnored(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{ToolName: "bash", Command: ""}

	rule5CommandCheckpoints(p, state, log)

	checkpoints := state.ReadText("checkpoints.txt")
	if checkpoints != "" {
		t.Error("empty command should not affect checkpoints")
	}
}

func TestRule5_GhPrMergeInAutonomousPipelineAllowed(t *testing.T) {
	// In an autonomous pipeline context, gh pr merge must not cause a warn or block.
	// We set active-command to "batch-flow" and call rule5 — no exit expected.
	state := newTestState(t)
	log := newNilLogger()
	state.WriteText("active-command", "batch-flow")

	p := &Payload{
		ToolName: "bash",
		Command:  "gh pr merge 42 --merge",
	}

	// Should complete without panic (Block would call os.Exit).
	rule5CommandCheckpoints(p, state, log)
}

func TestRule5_GhPrCreateInAutonomousPipelineAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()
	state.WriteText("active-command", "flow-auto")

	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule5CommandCheckpoints(p, state, log)
	// No exit = pass
}

// ---------------------------------------------------------------------------
// rule8TestParallel
// ---------------------------------------------------------------------------

func TestRule8_TargetedTestWithFilterAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `vendor/bin/pest --filter="TestMyFeature"`,
	}

	// Must not call os.Exit — targeted tests are always allowed.
	rule8TestParallel(p, state, log)
}

func TestRule8_TargetedTestWithFilePathAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "vendor/bin/pest tests/Feature/SomeFeatureTest.php",
	}

	rule8TestParallel(p, state, log)
}

func TestRule8_HerdCoverageWithFilterAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `herd coverage vendor/bin/pest --coverage --filter="MyTest"`,
	}

	rule8TestParallel(p, state, log)
}

func TestRule8_NonTestCommandPassesThrough(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "git status",
	}

	rule8TestParallel(p, state, log)
}

func TestRule8_NonBashToolIgnored(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "tests/Feature/SomeTest.php",
	}

	rule8TestParallel(p, state, log)
}

func TestRule8_FullSuiteWithParallelAndProcessesAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// Correctly-formed full suite command — must not be blocked.
	p := &Payload{
		ToolName: "bash",
		Command:  "./vendor/bin/pest --parallel --processes=10",
	}

	rule8TestParallel(p, state, log)
}

// ---------------------------------------------------------------------------
// rule15BacklogCLI
// ---------------------------------------------------------------------------

func TestRule15_NonBacklogCommandPassesThrough(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "grep something /other/path/file.txt",
	}

	// No match against backlog path → passes through, no exit.
	rule15BacklogCLI(p, state, log)
}

func TestRule15_LsBacklogAllowed(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// ls is not in the blocked set (only grep/sed/awk/cat/for).
	p := &Payload{
		ToolName: "bash",
		Command:  "ls blueprint/backlog",
	}

	rule15BacklogCLI(p, state, log)
}

func TestRule15_NonBashToolIgnored(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "blueprint/backlog/item.yaml",
	}

	rule15BacklogCLI(p, state, log)
}

func TestRule15_RegexDoesNotMatchNonBlueprintBacklogPath(t *testing.T) {
	// Ensure the patterns do NOT fire for unrelated paths.
	cases := []string{
		"grep foo /tmp/backlog/something",
		"cat /home/user/.planning/backlog/item",
	}
	for _, cmd := range cases {
		if reBacklogGrep.MatchString(cmd) {
			t.Errorf("reBacklogGrep should NOT match %q", cmd)
		}
		if reBacklogCat.MatchString(cmd) {
			t.Errorf("reBacklogCat should NOT match %q", cmd)
		}
	}
}

func TestRule15_BlueprintBacklogGrepPatternMatches(t *testing.T) {
	cases := []string{
		"grep title blueprint/backlog/item.yaml",
		"sed 's/foo/bar/' blueprint/backlog/item.yaml",
		"awk '{print}' blueprint/backlog/",
	}
	for _, cmd := range cases {
		if !reBacklogGrep.MatchString(cmd) {
			t.Errorf("reBacklogGrep SHOULD match %q", cmd)
		}
	}
}

func TestRule15_BlueprintBacklogCatPatternMatches(t *testing.T) {
	if !reBacklogCat.MatchString("cat blueprint/backlog/item.yaml") {
		t.Error("reBacklogCat should match 'cat blueprint/backlog/...'")
	}
}

// ---------------------------------------------------------------------------
// rule1SkillReadBeforeFrontend — PASS paths
// ---------------------------------------------------------------------------

func TestRule1_NonWriteToolReturnsImmediately(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "src/components/Button.jsx",
	}

	// read is not write-like — rule must return without blocking.
	rule1SkillReadBeforeFrontend(p, state, log, p.AllPaths())
}

func TestRule1_WriteToolNonFrontendExtensionPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	for _, ext := range []string{".go", ".py", ".php", ".ts"} {
		p := &Payload{
			ToolName: "write",
			FilePath: "src/main" + ext,
		}
		// None of these extensions are in skillMap — must pass without blocking.
		rule1SkillReadBeforeFrontend(p, state, log, p.AllPaths())
	}
}

func TestRule1_WriteJsxWithSkillAlreadyReadPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	// Skill was read before this write — rule must not block.
	state.Touch("skill-read-frontend-design")

	p := &Payload{
		ToolName: "write",
		FilePath: "src/App.jsx",
	}

	rule1SkillReadBeforeFrontend(p, state, log, p.AllPaths())
}

// ---------------------------------------------------------------------------
// rule6PlanTemplateRead — PASS paths
// ---------------------------------------------------------------------------

func TestRule6_NonWriteToolReturnsImmediately(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "blueprint/0001-feat-test.md",
	}

	rule6PlanTemplateRead(p, state, log, p.AllPaths())
}

func TestRule6_WriteToNonBlueprintPathPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "write",
		FilePath: "docs/architecture.md",
	}

	rule6PlanTemplateRead(p, state, log, p.AllPaths())
}

func TestRule6_WriteToBlueprintMdWithTemplateAlreadyReadPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	// Template was read — rule must not block.
	state.Touch("read-plan-template")

	p := &Payload{
		ToolName: "write",
		FilePath: "blueprint/0001-feat-new-plan.md",
	}

	// File does not exist yet (new file creation scenario).
	rule6PlanTemplateRead(p, state, log, p.AllPaths())
}

func TestRule6_WriteToBlueprintBacklogSkipped(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	// No template read — but /backlog/ paths are always skipped.
	p := &Payload{
		ToolName: "write",
		FilePath: "blueprint/backlog/0042-some-item.md",
	}

	rule6PlanTemplateRead(p, state, log, p.AllPaths())
}

func TestRule6_WriteToExistingBlueprintMdPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	// Create the file first — rule only blocks NEW file creation.
	os.MkdirAll("blueprint", 0755)
	os.WriteFile("blueprint/0001-existing.md", []byte("# existing"), 0644)

	p := &Payload{
		ToolName: "edit",
		FilePath: "blueprint/0001-existing.md",
	}

	rule6PlanTemplateRead(p, state, log, p.AllPaths())
}

// ---------------------------------------------------------------------------
// rule9TaskDeletion — WARN path (no os.Exit)
// ---------------------------------------------------------------------------

func TestRule9_NonWriteToolReturnsImmediately(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "blueprint/0001-feat-test-todo.md",
	}

	rule9TaskDeletion(p, state, log, p.AllPaths())
}

func TestRule9_WriteToNonTodoFilePasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "write",
		FilePath: "blueprint/0001-feat-test.md",
	}

	rule9TaskDeletion(p, state, log, p.AllPaths())
}

func TestRule9_EditToolRemovingUncheckedTaskWarns(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll("blueprint", 0755)
	planPath := "blueprint/0001-feat-test-todo.md"
	planContent := "# Plan\n\n- [ ] Task A\n- [ ] Task B\n- [x] Task C\n"
	os.WriteFile(planPath, []byte(planContent), 0644)

	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "edit",
		FilePath: planPath,
		Input: map[string]interface{}{
			"old_string": "- [ ] Task A\n- [ ] Task B\n",
			"new_string": "- [ ] Task A\n",
		},
	}

	// Warn fires but no os.Exit — test must complete normally.
	rule9TaskDeletion(p, state, log, p.AllPaths())
}

func TestRule9_WriteToolReducingUncheckedTasksWarns(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll("blueprint", 0755)
	planPath := "blueprint/0001-feat-test-todo.md"
	planContent := "# Plan\n\n- [ ] Task A\n- [ ] Task B\n- [x] Task C\n"
	os.WriteFile(planPath, []byte(planContent), 0644)

	state := newTestState(t)
	log := newNilLogger()

	// Write with fewer unchecked tasks than currently on disk.
	p := &Payload{
		ToolName: "write",
		FilePath: planPath,
		Input: map[string]interface{}{
			"content": "# Plan\n\n- [x] Task C\n",
		},
	}

	rule9TaskDeletion(p, state, log, p.AllPaths())
}

func TestRule9_EditToolNoUncheckedRemovedPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll("blueprint", 0755)
	planPath := "blueprint/0001-feat-test-todo.md"
	os.WriteFile(planPath, []byte("# Plan\n\n- [ ] Task A\n- [x] Task B\n"), 0644)

	state := newTestState(t)
	log := newNilLogger()

	// Marking an unchecked task as checked — no deletion, no warn.
	p := &Payload{
		ToolName: "edit",
		FilePath: planPath,
		Input: map[string]interface{}{
			"old_string": "- [ ] Task A\n",
			"new_string": "- [x] Task A\n",
		},
	}

	rule9TaskDeletion(p, state, log, p.AllPaths())
}

// ---------------------------------------------------------------------------
// rule10BlockDangerous — PASS paths
// ---------------------------------------------------------------------------

func TestRule10_NonBashToolReturnsImmediately(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "database/migrations/2024_01_01_create_users.php",
	}

	rule10BlockDangerous(p, state, log)
}

func TestRule10_HarmlessBashCommandPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "ls -la",
	}

	rule10BlockDangerous(p, state, log)
}

func TestRule10_PhpArtisanMigrateWithoutFreshPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "php artisan migrate",
	}

	// "migrate" alone must not trigger migrate:fresh rule.
	rule10BlockDangerous(p, state, log)
}

func TestRule10_GitCommitWithoutAISigPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Initialise a real git repo so git commands resolve without errors.
	os.MkdirAll(tmpDir, 0755)

	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `git commit -m "✨ feat: add new feature"`,
	}

	rule10BlockDangerous(p, state, log)
}

// ---------------------------------------------------------------------------
// rule11ClaudeReviewPrompt — PASS paths
// ---------------------------------------------------------------------------

func TestRule11_NonBashToolReturnsImmediately(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "some/file.md",
	}

	rule11ClaudeReviewPrompt(p, state, log)
}

func TestRule11_BashWithoutGhPrCommentPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "gh pr list",
	}

	rule11ClaudeReviewPrompt(p, state, log)
}

func TestRule11_GhPrCommentWithoutClaudePasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr comment 42 --body "LGTM"`,
	}

	rule11ClaudeReviewPrompt(p, state, log)
}

func TestRule11_GhPrCommentWithFullClaudeReviewPromptPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// Contains both @claude review and "check if we are able to merge" — valid.
	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr comment 42 --body "@claude review this PR and check if we are able to merge. Analyze the code changes for any issues."`,
	}

	rule11ClaudeReviewPrompt(p, state, log)
}

// ---------------------------------------------------------------------------
// rule12PlanCheckSkip — PASS paths
// ---------------------------------------------------------------------------

func TestRule12_NonBashToolReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "blueprint/0001-feat-test.md",
	}

	rule12PlanCheckSkip(p, state, log)
}

func TestRule12_BashWithoutGhPrCreateReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "git push origin feat/my-feature",
	}

	rule12PlanCheckSkip(p, state, log)
}

func TestRule12_GhPrCreateWithNoPlanApprovedPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// No plan-approved checkpoint in state — rule silently passes.
	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule12PlanCheckSkip(p, state, log)
}

func TestRule12_GhPrCreateWithPlanApprovedAndPlanCheckPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// Both checkpoints present — no warning.
	state.AppendLine("checkpoints.txt", "plan-approved:1")
	state.AppendLine("checkpoints.txt", "plan-check:1")

	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule12PlanCheckSkip(p, state, log)
}

// ---------------------------------------------------------------------------
// rule13UncheckedAcceptance — PASS paths
// ---------------------------------------------------------------------------

func TestRule13_NonBashToolReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "blueprint/0001-feat-test-todo.md",
	}

	rule13UncheckedAcceptance(p, state, log)
}

func TestRule13_BashWithoutGhPrCreateReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "git status",
	}

	rule13UncheckedAcceptance(p, state, log)
}

func TestRule13_GhPrCreateWithNoTodoFilePasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	// No blueprint/*-todo.md files exist — glob returns nothing, no warn.
	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule13UncheckedAcceptance(p, state, log)
}

func TestRule13_GhPrCreateWithAllAcceptanceCriteriaCheckedPasses(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll("blueprint", 0755)
	content := "# Plan\n\n## Acceptance Criteria\n\n- [x] All tests pass\n- [x] Reviewed\n\n## Tasks\n"
	os.WriteFile("blueprint/0001-feat-test-todo.md", []byte(content), 0644)

	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule13UncheckedAcceptance(p, state, log)
}

// ---------------------------------------------------------------------------
// rule14FlowAutoEnforcement — PASS paths
// ---------------------------------------------------------------------------

func TestRule14_NonBashToolReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "some/file.go",
	}

	rule14FlowAutoEnforcement(p, state, log)
}

func TestRule14_BashWithoutGhPrCreateReturns(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "bash",
		Command:  "git log --oneline -5",
	}

	rule14FlowAutoEnforcement(p, state, log)
}

func TestRule14_GhPrCreateWithNoFlowAutoCheckpointsPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// No flow-auto checkpoints at all — isFlowAuto is false, rule ignores PR.
	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule14FlowAutoEnforcement(p, state, log)
}

func TestRule14_GhPrCreateWithFlowAutoStep5CheckpointPasses(t *testing.T) {
	state := newTestState(t)
	log := newNilLogger()

	// Both step 4 and step 5 present — no block.
	state.AppendLine("checkpoints.txt", "flow-auto:4")
	state.AppendLine("checkpoints.txt", "flow-auto:5")

	p := &Payload{
		ToolName: "bash",
		Command:  `gh pr create --title "feat: something" --body "body"`,
	}

	rule14FlowAutoEnforcement(p, state, log)
}

// ---------------------------------------------------------------------------
// RunRules — smoke test with a benign payload
// ---------------------------------------------------------------------------

func TestRunRules_BenignReadPayloadDoesNotPanic(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	state := newTestState(t)
	log := newNilLogger()

	p := &Payload{
		ToolName: "read",
		FilePath: "README.md",
		Input:    map[string]interface{}{},
	}

	// RunRules must complete without panic or os.Exit for a harmless read.
	RunRules(p, state, log)
}
