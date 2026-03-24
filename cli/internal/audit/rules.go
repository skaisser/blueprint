package audit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// BlueprintConfig holds project-level config from blueprint/.config.yml.
type BlueprintConfig struct {
	StagingBranch string `yaml:"staging_branch"`
}

// LoadBlueprintConfig reads blueprint/.config.yml and returns the config.
// Returns defaults if the file doesn't exist.
func LoadBlueprintConfig() *BlueprintConfig {
	cfg := &BlueprintConfig{StagingBranch: "staging"} // default

	data, err := os.ReadFile("blueprint/.config.yml")
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg
	}

	if cfg.StagingBranch == "" {
		cfg.StagingBranch = "staging"
	}

	return cfg
}

// Pre-compiled regex patterns for performance (hot path).
var (
	reCheckpoint     = regexp.MustCompile(`\[([a-z-]+:[0-9a-z-]+)\]`)               // legacy: [skill:step]
	reCheckpointBP   = regexp.MustCompile(`BP:\s+([a-z-]+)\s+\[(\d+)/(\d+)\]`)    // new: BP: skill [N/TOTAL]
	reMigrateFresh   = regexp.MustCompile(`migrate:fresh`)
	reGitPushMain    = regexp.MustCompile(`git\s+push(?:\s+-\S+)*\s+\S+\s+main\b`)
	reBarePush       = regexp.MustCompile(`(?:^|&&|;)\s*git\s+push\s*$`)
	reCdPath         = regexp.MustCompile(`(?:^|&&|;)\s*cd\s+(\S+)`)
	reGitCommit      = regexp.MustCompile(`git\s+commit`)
	reAISig          = regexp.MustCompile(`(?i)Co-Authored-By|Generated.by.AI|Generated.by.Claude`)
	reGhPrCreateEdit = regexp.MustCompile(`gh\s+pr\s+(create|edit)`)
	reGhPrComment    = regexp.MustCompile(`gh\s+pr\s+comment`)
	reClaudeReview   = regexp.MustCompile(`(?i)@claude\s+review`)
	reCheckMerge     = regexp.MustCompile(`(?i)check if we are able to merge`)
	reGhPrMerge      = regexp.MustCompile(`gh pr merge`)
	reGhPrCreate     = regexp.MustCompile(`gh pr create`)
	reGhApiPulls     = regexp.MustCompile(`gh api repos/\S+/pulls\b.*-X POST`)
	reGhPrEdit       = regexp.MustCompile(`gh\s+pr\s+edit`)
	reTestCmd        = regexp.MustCompile(`(vendor/bin/pest|php\s+artisan\s+test|artisan\s+test|herd\s+coverage)`)
	reFilter         = regexp.MustCompile(`--filter`)
	reTestFile       = regexp.MustCompile(`tests/\S+`)
	reParallel       = regexp.MustCompile(`--parallel`)
	reProcesses10    = regexp.MustCompile(`--processes[=\s]+10`)
	reTaskChecked    = regexp.MustCompile(`(?i)- \[x\]`)
	reTaskUnchecked  = regexp.MustCompile(`- \[ \]`)
	reTaskAny        = regexp.MustCompile(`(?i)- \[[x ]\]`)
	rePlanningTodo   = regexp.MustCompile(`blueprint/.*-todo\.md$`)
	rePlanningMd     = regexp.MustCompile(`blueprint/.*\.md$`)
	reAcceptCriteria = regexp.MustCompile(`(?is)##\s*Acceptance\s*Criteria\s*\n(.*?)(?:\n##\s|\z)`)

	// Commit message validation regex — supports optional (scope) and ! for breaking changes.
	reCommitMsg = regexp.MustCompile(`^(\S+)\s+(feat|fix|docs|refactor|test|perf|style|build|chore|ci|revert|plan|migration|remove|security|deps|hotfix|merge|deploy)(\([\w/\-]+\))?(!)?: .+`)

	aiSigPatternsPR = []*regexp.Regexp{
		regexp.MustCompile(`(?i)Generated\s+with\s+\[?Claude`),
		regexp.MustCompile(`(?i)Generated\s+with\s+Claude\s+Code`),
		regexp.MustCompile(`(?i)Co-Authored-By.*claude`),
		regexp.MustCompile(`(?i)Co-Authored-By.*anthropic`),
		regexp.MustCompile(`(?i)Co-Authored-By.*noreply@anthropic`),
		regexp.MustCompile(`🤖\s*Generated`),
		regexp.MustCompile(`(?i)Generated\s+by\s+AI`),
		regexp.MustCompile(`(?i)Generated\s+by\s+Claude`),
		regexp.MustCompile(`(?i)claude\.com/claude-code`),
	}

	aiSigPatternsComment = []*regexp.Regexp{
		regexp.MustCompile(`(?i)Generated\s+with\s+\[?Claude`),
		regexp.MustCompile(`🤖\s*Generated`),
		regexp.MustCompile(`(?i)Generated\s+by\s+AI`),
		regexp.MustCompile(`(?i)Generated\s+by\s+Claude`),
		regexp.MustCompile(`(?i)claude\.com/claude-code`),
	}
)

// AllowedCommitTypes maps type keywords to their expected emoji.
var AllowedCommitTypes = map[string]string{
	"feat":      "✨",
	"fix":       "🐛",
	"docs":      "📚",
	"refactor":  "♻️",
	"test":      "🧪",
	"perf":      "⚡",
	"style":     "💄",
	"build":     "🔧",
	"chore":     "🧹",
	"ci":        "🔄",
	"revert":    "↩️",
	"plan":      "📋",
	"migration": "🗃️",
	"remove":    "🔥",
	"security":  "🔒",
	"deps":      "📦",
	"hotfix":    "🩹",
	"merge":     "🔀",
	"deploy":    "🚀",
}

// ValidateCommitMessage validates a commit message against the convention.
// Returns (valid bool, errorMessage string).
func ValidateCommitMessage(msg string) (bool, string) {
	if msg == "" {
		return false, "empty commit message"
	}

	// Check for AI signatures
	if reAISig.MatchString(msg) {
		return false, "AI signatures are blocked in commit messages"
	}

	if !reCommitMsg.MatchString(msg) {
		return false, "commit message does not match format: <emoji> <type>[(scope)][!]: <description>"
	}

	return true, ""
}

// DetectBreakingChange checks if a commit body contains BREAKING CHANGE footer.
func DetectBreakingChange(body string) bool {
	return strings.Contains(body, "BREAKING CHANGE:")
}

// skillMap maps file extensions to required skill names.
var skillMap = map[string]string{
	".jsx":  "frontend-design",
	".tsx":  "frontend-design",
	".html": "frontend-design",
	".css":  "frontend-design",
}

// Pre-compiled patterns for rule 15 (backlog CLI enforcement).
var (
	reBacklogGrep = regexp.MustCompile(`(grep|sed|awk).*blueprint/backlog`)
	reBacklogLoop = regexp.MustCompile(`for\s+\w+\s+in\s+.*blueprint/backlog`)
	reBacklogCat  = regexp.MustCompile(`cat\s+.*blueprint/backlog`)
)

// RunRules executes all 15 enforcement rules.
func RunRules(p *Payload, state *SessionState, log *Logger) {
	paths := p.AllPaths()

	// ENFORCEMENT 1: Skill read before frontend edits
	rule1SkillReadBeforeFrontend(p, state, log, paths)

	// ENFORCEMENT 2: Track SKILL.md and reference file reads
	rule2TrackSkillReads(p, state, log)

	// ENFORCEMENT 3: Team vs Subagent compliance
	rule3TeamCompliance(p, state, log)

	// ENFORCEMENT 4: Track AskUserQuestion calls
	rule4AskUserTracking(p, state, log)

	// ENFORCEMENT 5: Command checkpoint tracking & prerequisites
	rule5CommandCheckpoints(p, state, log)

	// ENFORCEMENT 6: Plan file writes require template read
	rule6PlanTemplateRead(p, state, log, paths)

	// ENFORCEMENT 7: Block GitHub workflow creation without staging branch
	rule7BlockWorkflowWithoutStagingBranch(p, state, log, paths)

	// ENFORCEMENT 8: Full test suite must run in parallel
	rule8TestParallel(p, state, log)

	// ENFORCEMENT 9: Plan task deletion detection
	rule9TaskDeletion(p, state, log, paths)

	// ENFORCEMENT 10: Block dangerous commands
	rule10BlockDangerous(p, state, log)

	// ENFORCEMENT 11: @claude review prompt must be comprehensive
	rule11ClaudeReviewPrompt(p, state, log)

	// ENFORCEMENT 12: Plan-check skip detection
	rule12PlanCheckSkip(p, state, log)

	// ENFORCEMENT 13: Unchecked acceptance criteria on PR creation
	rule13UncheckedAcceptance(p, state, log)

	// ENFORCEMENT 14: flow-auto step enforcement
	rule14FlowAutoEnforcement(p, state, log)

	// ENFORCEMENT 15: Backlog CLI enforcement — block manual parsing
	rule15BacklogCLI(p, state, log)
}

func rule1SkillReadBeforeFrontend(p *Payload, state *SessionState, log *Logger, paths []string) {
	if !p.IsWriteLike() || len(paths) == 0 {
		return
	}
	for _, path := range paths {
		ext := filepath.Ext(path)
		skillName, ok := skillMap[ext]
		if !ok {
			continue
		}
		if !state.Exists("skill-read-" + skillName) {
			log.Block(fmt.Sprintf(
				"Tried to modify '%s' without reading SKILL.md first. Read ~/.claude/skills/%s/SKILL.md before proceeding.",
				path, skillName,
			))
		}
	}
}

func rule2TrackSkillReads(p *Payload, state *SessionState, log *Logger) {
	// Track Skill tool invocations
	if p.ToolName == "skill" {
		skillArg := strings.TrimSpace(firstInputString(p.Input, "skill"))
		if skillArg != "" {
			state.Touch("skill-read-" + skillArg)
			// Also check if this matches a known skill
			for _, known := range uniqueSkillNames() {
				if strings.HasPrefix(skillArg, known) || strings.HasPrefix(known, skillArg) {
					state.Touch("skill-read-" + known)
				}
			}
			log.Log("✅ SKILL INVOKED: " + skillArg)
		}
	}

	// Track Read tool for SKILL.md and reference files
	if p.ToolName == "read" && p.FilePath != "" {
		if strings.Contains(p.FilePath, "SKILL.md") {
			skillName := filepath.Base(filepath.Dir(p.FilePath))
			state.Touch("skill-read-" + skillName)
			log.Log(fmt.Sprintf("✅ SKILL READ: %s (%s)", skillName, p.FilePath))
		}
		if strings.Contains(p.FilePath, "plan-template.md") {
			state.Touch("read-plan-template")
			log.Log("✅ REF READ: plan-template.md")
		}
		if strings.Contains(p.FilePath, "team-execution.md") {
			state.Touch("read-team-execution")
			log.Log("✅ REF READ: team-execution.md")
		}
	}
}

func rule3TeamCompliance(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName == "teamcreate" {
		teamName := firstInputString(p.Input, "name", "team_name")
		if teamName == "" {
			teamName = "unnamed"
		}
		if !state.Exists("read-team-execution") {
			log.Warn(fmt.Sprintf(
				"Creating team '%s' without reading team-execution.md first. Read ~/.claude/references/plan/team-execution.md for delegation strategy.",
				teamName,
			))
		}
		state.Touch("team-created")
		state.AppendLine("teams-created.txt", teamName)
		log.Log("✅ TEAM CREATED: " + teamName)
	}

	if p.ToolName == "task" {
		hasTeam := firstInputString(p.Input, "team_name") != ""
		if hasTeam {
			log.Log("✅ TEAM WORKER TASK")
		} else {
			count := state.IncrInt("standalone-task-count")
			if count >= 3 && !state.Exists("team-created") {
				log.Warn(fmt.Sprintf(
					"Standalone Task call #%d in this session — NO TeamCreate detected. team-execution.md requires TeamCreate for 3+ tasks. You said you'd use teams. Did you lie?",
					count,
				))
				log.Log(fmt.Sprintf("🚨 POTENTIAL LIE: %d standalone Task calls, 0 TeamCreate calls", count))
			} else {
				log.Log(fmt.Sprintf("✅ STANDALONE SUBAGENT #%d (1-2 tasks, acceptable)", count))
			}
		}
	}
}

func rule4AskUserTracking(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName == "askuserquestion" || p.ToolName == "question" {
		state.Touch("asked-user")
		log.Log("✅ ASKED USER (AskUserQuestion called)")
	}
}

func rule5CommandCheckpoints(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}

	// Track checkpoints — new format: 🔷 BP: skill [N/TOTAL], legacy: 🏁 [skill:step]
	if m := reCheckpointBP.FindStringSubmatch(p.Command); m != nil {
		activeCmd := m[1]
		checkpoint := activeCmd + ":" + m[2]
		state.AppendLine("checkpoints.txt", checkpoint)
		state.WriteText("active-command", activeCmd)
		log.Log("🔷 BP CHECKPOINT: " + checkpoint)
	} else if m := reCheckpoint.FindStringSubmatch(p.Command); m != nil {
		checkpoint := m[1]
		state.AppendLine("checkpoints.txt", checkpoint)
		activeCmd := strings.Split(checkpoint, ":")[0]
		state.WriteText("active-command", activeCmd)
		log.Log("🔷 BP CHECKPOINT: " + checkpoint)
	}

	// Prerequisite: /finish must AskUserQuestion before second gh pr merge
	// Exception: batch-flow, flow-auto, flow-auto-wt are autonomous pipelines
	// where merging PRs between rounds is a hard dependency, not discretionary.
	if reGhPrMerge.MatchString(p.Command) {
		activeCmd := state.ReadText("active-command")
		if isAutonomousPipeline(activeCmd) {
			log.Log("✅ gh pr merge in autonomous pipeline (" + activeCmd + ") — allowed")
		} else if activeCmd == "finish" {
			count := state.IncrInt("gh-pr-merge-count")
			if count >= 2 && !state.Exists("asked-user") {
				log.Warn("Second gh pr merge during /finish without AskUserQuestion. Step 6 requires asking the user before merging staging → main.")
			}
		}
	}

	// Prerequisite: gh pr create requires /pr, /finish, or pipeline context
	if strings.Contains(p.Command, "gh pr create") {
		activeCmd := state.ReadText("active-command")
		if isAutonomousPipeline(activeCmd) {
			log.Log("✅ gh pr create in autonomous pipeline (" + activeCmd + ") — allowed")
		} else if activeCmd != "pr" && activeCmd != "finish" && activeCmd != "hotfix-push" {
			log.Warn(fmt.Sprintf(
				"gh pr create called outside /pr, /finish, or /hotfix-push context. PRs should only be created via these skills. Active context: '%s'",
				activeCmd,
			))
		}
	}
}

func rule6PlanTemplateRead(p *Payload, state *SessionState, log *Logger, paths []string) {
	if !p.IsWriteLike() || len(paths) == 0 {
		return
	}
	for _, path := range paths {
		if rePlanningMd.MatchString(path) && !strings.Contains(path, "/backlog/") {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if !state.Exists("read-plan-template") {
					log.Block(
						"Creating new plan file without reading plan-template.md first.\n" +
							"Read ~/.claude/references/plan/plan-template.md before creating plans.",
					)
				}
			}
		}
	}
}

func rule7BlockWorkflowWithoutStagingBranch(p *Payload, state *SessionState, log *Logger, paths []string) {
	if !p.IsWriteLike() || len(paths) == 0 {
		return
	}

	cfg := LoadBlueprintConfig()

	for _, path := range paths {
		base := filepath.Base(path)
		if strings.Contains(path, ".github/workflows/") && (base == "claude-pr-reviewer.yml" || base == "tests.yml") {
			cmd := exec.Command("git", "branch", "-a", "--list", "*"+cfg.StagingBranch+"*")
			out, err := cmd.Output()
			if err != nil || strings.TrimSpace(string(out)) == "" {
				log.Block(fmt.Sprintf(
					"Creating '%s' but this project has no %s branch.\n\n"+
						"GitHub workflows (claude-pr-reviewer.yml, tests.yml) are only for projects using the\n"+
						"%s → main PR flow. Without %s, there's no PR pipeline to run CI on.\n\n"+
						"If this project should use the %s flow, create the branch first:\n"+
						"  git checkout -b %s && git push -u origin %s",
					base, cfg.StagingBranch, cfg.StagingBranch, cfg.StagingBranch,
					cfg.StagingBranch, cfg.StagingBranch, cfg.StagingBranch,
				))
			}
		}
	}
}

func rule8TestParallel(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}
	if !reTestCmd.MatchString(p.Command) {
		return
	}

	hasFilter := reFilter.MatchString(p.Command)
	hasTestFile := reTestFile.MatchString(p.Command)
	isFullSuite := !hasFilter && !hasTestFile

	if isFullSuite {
		hasParallel := reParallel.MatchString(p.Command)
		hasProcesses := reProcesses10.MatchString(p.Command)
		if !hasParallel || !hasProcesses {
			log.Block(
				"Wrong way to run the full test suite.\n\n" +
					"You MUST run it in parallel with 10 processes — otherwise it takes 40+ minutes:\n\n" +
					"  Full suite:     ./vendor/bin/pest --parallel --processes=10\n" +
					"  With coverage:  herd coverage ./vendor/bin/pest --coverage --parallel --processes=10\n\n" +
					"These are available as shell aliases:\n" +
					"  ptp   →  ./vendor/bin/pest --parallel --processes=10\n" +
					"  tcq   →  herd coverage ./vendor/bin/pest --coverage --parallel --processes=10\n\n" +
					"Targeted tests (always preferred for speed):\n" +
					"  vendor/bin/pest --filter=\"TestName\"\n" +
					"  vendor/bin/pest tests/Feature/Path/ToTest.php\n" +
					"  herd coverage vendor/bin/pest --coverage --filter=\"TestName\"",
			)
		}
	}
}

func rule9TaskDeletion(p *Payload, state *SessionState, log *Logger, paths []string) {
	if !p.IsWriteLike() || len(paths) == 0 {
		return
	}
	for _, path := range paths {
		if !rePlanningTodo.MatchString(path) {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		currentContent, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(currentContent)
		currentChecked := len(reTaskChecked.FindAllString(content, -1))
		currentUnchecked := len(reTaskUnchecked.FindAllString(content, -1))
		currentTotal := currentChecked + currentUnchecked

		oldString := firstInputString(p.Input, "old_string")
		newString := firstInputString(p.Input, "new_string")

		if oldString != "" && newString != "" {
			// Edit tool
			oldTasks := len(reTaskAny.FindAllString(oldString, -1))
			newTasks := len(reTaskAny.FindAllString(newString, -1))
			tasksRemoved := oldTasks - newTasks
			if tasksRemoved > 0 {
				oldUnchecked := len(reTaskUnchecked.FindAllString(oldString, -1))
				newUnchecked := len(reTaskUnchecked.FindAllString(newString, -1))
				uncheckedRemoved := oldUnchecked - newUnchecked
				if uncheckedRemoved > 0 {
					log.Warn(fmt.Sprintf(
						"🚨 PLAN TASK DELETION DETECTED: %d unchecked task(s) removed from '%s'.\n"+
							"   Before: %d tasks (%d unchecked) → After: %d tasks (%d unchecked)\n"+
							"   You MUST implement planned tasks, not delete them. If a task is no longer needed, mark it [x] SKIPPED with reason.\n"+
							"   Trust but CHECK — this edit looks like lying by omission.",
						uncheckedRemoved, filepath.Base(path), oldTasks, oldUnchecked, newTasks, newUnchecked,
					))
				}
			}
		} else if p.ToolName == "write" {
			// Write tool (full file replacement)
			newContent := firstInputString(p.Input, "content")
			if newContent != "" {
				newChecked := len(reTaskChecked.FindAllString(newContent, -1))
				newUnchecked := len(reTaskUnchecked.FindAllString(newContent, -1))
				newTotal := newChecked + newUnchecked
				tasksLost := currentTotal - newTotal
				if tasksLost > 0 {
					uncheckedLost := currentUnchecked - newUnchecked
					if uncheckedLost > 0 {
						log.Warn(fmt.Sprintf(
							"🚨 PLAN TASK DELETION DETECTED: %d task(s) disappeared from '%s'.\n"+
								"   Before: %d tasks (%d unchecked) → After: %d tasks (%d unchecked)\n"+
								"   %d unchecked task(s) were removed — these should be implemented, not deleted.\n"+
								"   If tasks are genuinely not needed, mark them [x] SKIPPED with reason.",
							tasksLost, filepath.Base(path), currentTotal, currentUnchecked, newTotal, newUnchecked, uncheckedLost,
						))
					}
				}
			}
		}

		// Save snapshot
		state.WriteText(
			fmt.Sprintf("plan-tasks-%s", strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))),
			fmt.Sprintf("%d:%d:%d", currentTotal, currentChecked, currentUnchecked),
		)
	}
}

func rule10BlockDangerous(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}

	cfg := LoadBlueprintConfig()

	// Block migrate:fresh
	if reMigrateFresh.MatchString(p.Command) {
		if !state.Exists("skill-read-squash-migrations") {
			log.Block(
				"migrate:fresh is BLOCKED — it drops all tables and destroys data.\n" +
					"Use 'php artisan migrate' for normal migrations.\n\n" +
					"If you need migrate:fresh for migration consolidation, invoke /squash-migrations first.\n" +
					"That skill unlocks migrate:fresh for the session.\n\n" +
					"If you truly need it outside that context, ask the user to run it manually.",
			)
		} else {
			log.Warn("migrate:fresh ALLOWED — /squash-migrations skill is active.\nThis is expected during migration consolidation. Proceed with caution.")
		}
	}

	// Block direct push to main
	isExplicitMainPush := reGitPushMain.MatchString(p.Command)
	isBarePush := reBarePush.MatchString(p.Command)
	if isExplicitMainPush || isBarePush {
		cwd, _ := os.Getwd()
		if m := reCdPath.FindStringSubmatch(p.Command); m != nil {
			cwd = os.ExpandEnv(m[1])
			if strings.HasPrefix(cwd, "~") {
				home, _ := os.UserHomeDir()
				cwd = filepath.Join(home, cwd[1:])
			}
		}

		branchCmd := exec.Command("git", "branch", "--show-current")
		if cwd != "" {
			branchCmd.Dir = cwd
		}
		branchOut, _ := branchCmd.Output()
		branch := strings.TrimSpace(string(branchOut))

		stagingCmd := exec.Command("git", "branch", "-a", "--list", "*"+cfg.StagingBranch+"*")
		if cwd != "" {
			stagingCmd.Dir = cwd
		}
		stagingOut, _ := stagingCmd.Output()
		hasStaging := strings.TrimSpace(string(stagingOut)) != ""

		isMainBranch := branch == "main"
		if hasStaging && (isExplicitMainPush || isMainBranch) {
			log.Block(
				"Direct push to main is BLOCKED.\n" +
					"Use the branch flow: feature/* → " + cfg.StagingBranch + " → main (via PR).\n" +
					"If this is an emergency, use /hotfix-push which creates a proper PR.",
			)
		}
	}

	// Block AI signatures in commits
	if reGitCommit.MatchString(p.Command) && reAISig.MatchString(p.Command) {
		log.Block(
			"AI signatures are BLOCKED in commits.\n" +
				"Remove any Co-Authored-By, 'Generated by AI', or similar attribution.\n" +
				"The commit-msg hook will also reject these, but catching it early.",
		)
	}

	// Block AI signatures in PR creation/editing
	if reGhPrCreateEdit.MatchString(p.Command) {
		for _, pattern := range aiSigPatternsPR {
			if pattern.MatchString(p.Command) {
				log.Block(fmt.Sprintf(
					"AI signatures are BLOCKED in pull requests.\n"+
						"Remove any 'Generated with Claude Code', 'Co-Authored-By', '🤖 Generated',\n"+
						"or any AI attribution from the PR title and body.\n"+
						"Matched pattern: %s",
					pattern.String(),
				))
			}
		}
	}

	// Block AI signatures in PR comments
	if reGhPrComment.MatchString(p.Command) {
		for _, pattern := range aiSigPatternsComment {
			if pattern.MatchString(p.Command) {
				log.Block(fmt.Sprintf(
					"AI signatures are BLOCKED in PR comments.\n"+
						"Remove any AI attribution from the comment body.\n"+
						"Matched pattern: %s",
					pattern.String(),
				))
			}
		}
	}
}

func rule11ClaudeReviewPrompt(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}
	if reGhPrComment.MatchString(p.Command) && reClaudeReview.MatchString(p.Command) {
		if !reCheckMerge.MatchString(p.Command) {
			log.Block(
				"@claude review comment is too short — agents must use the full prompt.\n\n" +
					"Required comment body:\n" +
					"  \"@claude review this PR and check if we are able to merge. " +
					"Analyze the code changes for any issues, security concerns, or improvements needed.\"\n\n" +
					"This ensures the CI reviewer performs a thorough analysis, not a superficial pass.",
			)
		}
	}
}

func rule12PlanCheckSkip(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}
	if !strings.Contains(p.Command, "gh pr create") && !reGhApiPulls.MatchString(p.Command) {
		return
	}

	checkpoints := state.ReadText("checkpoints.txt")
	hadPlanApproved := strings.Contains(checkpoints, "plan-approved:")
	hadPlanCheck := strings.Contains(checkpoints, "plan-check:")

	if hadPlanApproved && !hadPlanCheck {
		log.Warn(
			"Creating PR after /plan-approved but /plan-check was never run.\n" +
				"/plan-check is the quality gate — it audits code vs plan, catches orphaned tests,\n" +
				"and creates the audit commit. Run /plan-check before /pr.",
		)
	}
}

func rule13UncheckedAcceptance(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}
	if !strings.Contains(p.Command, "gh pr create") && !reGhPrEdit.MatchString(p.Command) {
		return
	}

	matches, _ := filepath.Glob("blueprint/*-todo.md")
	for _, pf := range matches {
		content, err := os.ReadFile(pf)
		if err != nil {
			continue
		}
		acMatch := reAcceptCriteria.FindSubmatch(content)
		if acMatch == nil {
			continue
		}
		acSection := string(acMatch[1])
		unchecked := reTaskUnchecked.FindAllString(acSection, -1)
		total := reTaskAny.FindAllString(acSection, -1)
		if len(unchecked) > 0 && len(total) > 0 {
			log.Warn(fmt.Sprintf(
				"PR has %d/%d unchecked acceptance criteria in '%s'.\n"+
					"/plan-check should verify and mark each criterion before creating the PR.\n"+
					"Run /plan-check or manually verify and check off the acceptance criteria.",
				len(unchecked), len(total), filepath.Base(pf),
			))
		}
	}
}

func rule14FlowAutoEnforcement(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}

	checkpoints := state.ReadText("checkpoints.txt")
	isFlowAuto := strings.Contains(checkpoints, "flow-auto:")

	isPRCreation := strings.Contains(p.Command, "gh pr create") || reGhApiPulls.MatchString(p.Command)

	if isFlowAuto && isPRCreation {
		hadStep5 := strings.Contains(checkpoints, "flow-auto:5")
		hadStep4 := strings.Contains(checkpoints, "flow-auto:4")

		if hadStep4 && !hadStep5 {
			log.Block(
				"flow-auto: Creating PR without running Step 5 (plan-check).\n\n" +
					"The flow-auto pipeline requires ALL steps in order:\n" +
					"  1. Initialize → 2. Plan → 3. Review → 4. Execute → 5. Plan Check → 6. PR → 7. Review Loop → 8. Report\n\n" +
					"You skipped Step 5 (auditing implementation). Run it now:\n" +
					"  echo \"🤖 [flow-auto:5] auditing implementation\"\n" +
					"  blueprint context --diffs\n" +
					"  # ... compare plan vs implementation, fix mismatches\n" +
					"  blueprint update",
			)
		}
	}

	// Enforce review loop
	isPRComment := reGhPrComment.MatchString(p.Command) && strings.Contains(p.Command, "Final Report")
	if isFlowAuto && isPRComment {
		hadStep7 := strings.Contains(checkpoints, "flow-auto:7")
		hadStep6 := strings.Contains(checkpoints, "flow-auto:6")
		if hadStep6 && !hadStep7 {
			log.Warn(
				"flow-auto: Posting final report without running Step 7 (review loop).\n\n" +
					"The review loop is mandatory — trigger @claude review on the PR:\n" +
					"  echo \"🤖 [flow-auto:7] starting review loop\"\n" +
					"  gh pr comment $PR_NUM --body \"@claude review this PR\"\n\n" +
					"If the GitHub Action is not set up, poll for 5 min then continue.\n" +
					"You MUST at least attempt the review step.",
			)
		}
	}
}

// isAutonomousPipeline returns true for contexts where the agent manages
// the full PR lifecycle (create, merge, push) autonomously.
func isAutonomousPipeline(activeCmd string) bool {
	switch activeCmd {
	case "batch-flow", "flow-auto", "flow-auto-wt":
		return true
	}
	return false
}

func rule15BacklogCLI(p *Payload, state *SessionState, log *Logger) {
	if p.ToolName != "bash" || p.Command == "" {
		return
	}

	if reBacklogGrep.MatchString(p.Command) || reBacklogLoop.MatchString(p.Command) || reBacklogCat.MatchString(p.Command) {
		log.Block(
			"Manual backlog file parsing is BLOCKED.\n\n" +
				"Use the CLI instead — it handles both YAML and legacy formats correctly:\n\n" +
				"  blueprint backlog                  # JSON output (default)\n" +
				"  blueprint backlog --format table   # Pretty table\n" +
				"  blueprint backlog --archive        # Include archived items\n" +
				"  blueprint backlog migrate           # Convert old format → YAML\n\n" +
				"Never parse backlog files with grep/sed/awk/cat — the CLI is faster and correct.",
		)
	}
}

func uniqueSkillNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, name := range skillMap {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}
