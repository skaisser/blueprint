package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	gitpkg "github.com/skaisser/blueprint/internal/git"
	"github.com/skaisser/blueprint/internal/plan"
	"github.com/spf13/cobra"
)

// ─── meta ───────────────────────────────────────────────────────────────────

var metaCmd = &cobra.Command{
	Use:   "meta [field]",
	Short: "Plan metadata as JSON, or a single field value",
	Long: `Output plan metadata. With no arguments, prints full JSON.
With a field name, prints just that field's raw value.

Supported fields:
  next_num, base_branch, branch, plan_file, plan_num,
  status, progress, project, git_remote, today`,
	Example: `  blueprint meta              # full JSON
  blueprint meta plan_file    # just the plan file path
  blueprint meta branch       # just the current branch`,
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Println(`{"error": "Not in a git repository"}`)
			os.Exit(1)
		}

		branch := repo.CurrentBranch()
		if branch == "" {
			branch = "main"
		}

		baseBranch := repo.DetectBaseBranchSimple()
		planFile := plan.FindPlanFile(".planning", branch)
		if planFile == "" {
			planFile = plan.FindPlanFile("blueprint", branch)
		}
		header := plan.ParsePlanHeader(planFile)

		nextNumPlanning := plan.GetNextNum(".planning")
		nextNumBlueprint := plan.GetNextNum("blueprint")
		nextNum := nextNumPlanning
		if nextNumBlueprint > nextNum {
			nextNum = nextNumBlueprint
		}

		project := gitpkg.ProjectName()
		gitRemote := repo.RemoteURL("origin")

		result := &plan.MetaResult{
			NextNum:    nextNum,
			BaseBranch: baseBranch,
			Branch:     branch,
			PlanFile:   planFile,
			PlanNum:    header.PlanNum,
			Status:     header.Status,
			Progress:   header.Progress,
			Project:    project,
			GitRemote:  gitRemote,
			Today:      time.Now().Format("02/01/2006 15:04"),
		}

		// Single field mode
		if len(args) > 0 {
			field := args[0]
			switch field {
			case "next_num":
				fmt.Println(result.NextNum)
			case "base_branch":
				fmt.Println(result.BaseBranch)
			case "branch":
				fmt.Println(result.Branch)
			case "plan_file":
				fmt.Println(result.PlanFile)
			case "plan_num":
				fmt.Println(result.PlanNum)
			case "status":
				fmt.Println(result.Status)
			case "progress":
				fmt.Println(result.Progress)
			case "project":
				fmt.Println(result.Project)
			case "git_remote":
				fmt.Println(result.GitRemote)
			case "today":
				fmt.Println(result.Today)
			default:
				fmt.Fprintf(os.Stderr, "Unknown field: %s\n", field)
				os.Exit(1)
			}
			return
		}

		fmt.Println(result.JSON())
	},
}

// ─── context ────────────────────────────────────────────────────────────────

var (
	contextDiffs  bool
	contextSkipPR bool
)

var contextCmd = &cobra.Command{
	Use:   "context [BASE_BRANCH]",
	Short: "Git context for PR/plan-check (commits, files, diffs)",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Println("❌ Not in a git repository")
			os.Exit(1)
		}

		var base string
		if len(args) > 0 {
			base = args[0]
		} else {
			base = repo.DetectBaseBranch()
		}

		current := repo.CurrentBranch()
		if current == "" {
			fmt.Println("❌ Could not determine current branch")
			os.Exit(1)
		}

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("BASE_BRANCH:    %s\n", base)
		fmt.Printf("CURRENT_BRANCH: %s\n", current)
		fmt.Printf("%s\n", strings.Repeat("=", 60))

		// Commits since base
		fmt.Println(gitpkg.Section("Commits Since " + base))
		commits, err := gitpkg.CommitsSince(base)
		if err != nil || commits == "" {
			fmt.Printf("(no commits since %s)\n", base)
		} else {
			fmt.Println(commits)
		}

		// Changed files
		fmt.Println(gitpkg.Section("Changed Files"))
		changedFiles, err := gitpkg.ChangedFiles(base)
		if err != nil || len(changedFiles) == 0 {
			fmt.Println("(no changed files)")
		} else {
			fmt.Println(strings.Join(changedFiles, "\n"))
		}

		// Diff stat
		fmt.Println(gitpkg.Section("Diff Stat"))
		stat, err := gitpkg.DiffStat(base)
		if err != nil || stat == "" {
			fmt.Println("(no diff)")
		} else {
			lines := strings.Split(stat, "\n")
			if len(lines) <= 20 {
				fmt.Println(stat)
			} else {
				fmt.Println(strings.Join(lines[:15], "\n"))
				fmt.Printf("  ... (%d more files)\n", len(lines)-16)
				fmt.Println(lines[len(lines)-1])
			}
		}

		// Per-file diffs
		if contextDiffs && len(changedFiles) > 0 {
			fmt.Println(gitpkg.Section("File Diffs"))
			for _, fname := range changedFiles {
				padding := 54 - len(fname)
				if padding < 0 {
					padding = 0
				}
				fmt.Printf("\n──── %s %s\n", fname, strings.Repeat("─", padding))
				diff, err := gitpkg.FileDiff(base, fname)
				if err != nil || diff == "" {
					fmt.Println("  (no diff)")
				} else {
					diffLines := strings.Split(diff, "\n")
					if len(diffLines) > 120 {
						fmt.Println(strings.Join(diffLines[:120], "\n"))
						fmt.Printf("  ... (%d more lines)\n", len(diffLines)-120)
					} else {
						fmt.Println(diff)
					}
				}
			}
		}

		// Associated PR
		if !contextSkipPR {
			fmt.Println(gitpkg.Section("Associated PR"))
			pr, err := gitpkg.PRInfo()
			if err != nil {
				fmt.Println("(no open PR for this branch)")
			} else {
				num := pr["number"]
				state := pr["state"]
				title := pr["title"]
				url := pr["url"]
				fmt.Printf("#%v — %v — %v\n", num, state, title)
				if url != nil {
					fmt.Println(url)
				}
			}
		}

		fmt.Printf("\n%s\n\n", strings.Repeat("=", 60))
	},
}

// ─── sync ───────────────────────────────────────────────────────────────────

var (
	syncFinish   bool
	syncPRNumber string
)

var syncCmd = &cobra.Command{
	Use:   "sync [plan-file]",
	Short: "Sync plan frontmatter counts from checkboxes",
	Run: func(cmd *cobra.Command, args []string) {
		planFile := ""
		if len(args) > 0 {
			planFile = args[0]
		}

		result, err := plan.SyncPlanFile(planFile, syncFinish, syncPRNumber)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if result.Finished {
			fmt.Printf("Finished: status=completed pr=%s tasks=%d/%d\n",
				result.PR, result.TasksDone, result.TasksTotal)
		} else {
			fmt.Printf("Synced: %d/%d tasks, %d/%d phases, %d sessions\n",
				result.TasksDone, result.TasksTotal,
				result.PhasesDone, result.PhasesTotal,
				result.Sessions)
		}
	},
}

// ─── status (was "full") ────────────────────────────────────────────────────

var (
	statusDiffs  bool
	statusSkipPR bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Full SDLC status (meta JSON + git context)",
	Run: func(cmd *cobra.Command, args []string) {
		// Run meta
		metaCmd.Run(cmd, nil)
		fmt.Print("\n---\n\n")
		// Run context with flags
		contextDiffs = statusDiffs
		contextSkipPR = statusSkipPR
		contextCmd.Run(cmd, nil)
	},
}

// ─── commit ─────────────────────────────────────────────────────────────────

var commitCmd = &cobra.Command{
	Use:   "commit <message> [files...]",
	Short: "Commit plan + code changes with formatting",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		message := args[0]
		files := args[1:]
		if err := plan.Commit(message, files); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// ─── backlog ────────────────────────────────────────────────────────────────

var (
	backlogArchive bool
	backlogFormat  string
)

var backlogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "List backlog items (JSON default, --format table for pretty output)",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := plan.ScanBacklog("blueprint", backlogArchive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		switch strings.ToLower(backlogFormat) {
		case "table":
			fmt.Print(result.Table())
		default:
			fmt.Println(result.JSON())
		}
	},
}

var backlogMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Convert old blockquote backlog files to YAML frontmatter format",
	Run: func(cmd *cobra.Command, args []string) {
		project := gitpkg.ProjectName()

		// Migrate active
		activeResult, err := plan.MigrateBacklogDir("blueprint/backlog", project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Migrate archive
		archiveResult, err := plan.MigrateBacklogDir("blueprint/backlog/archive", project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: archive migration error: %v\n", err)
		}

		total := len(activeResult.Migrated)
		if archiveResult != nil {
			total += len(archiveResult.Migrated)
		}

		if total == 0 {
			fmt.Println("No old-format backlog files found. All files already use YAML frontmatter.")
			return
		}

		fmt.Printf("Migrated %d files to YAML frontmatter:\n", total)
		for _, f := range activeResult.Migrated {
			fmt.Printf("  ✅ backlog/%s\n", f)
		}
		if archiveResult != nil {
			for _, f := range archiveResult.Migrated {
				fmt.Printf("  ✅ archive/%s\n", f)
			}
		}
	},
}

func init() {
	// context flags
	contextCmd.Flags().BoolVar(&contextDiffs, "diffs", false, "Include full per-file diffs")
	contextCmd.Flags().BoolVar(&contextSkipPR, "skip-pr", false, "Skip gh pr view call (saves ~500ms)")

	// sync flags
	syncCmd.Flags().BoolVar(&syncFinish, "finish", false, "Mark plan as completed")
	syncCmd.Flags().StringVar(&syncPRNumber, "pr", "", "PR number (use with --finish)")

	// status flags
	statusCmd.Flags().BoolVar(&statusDiffs, "diffs", false, "Include full per-file diffs")
	statusCmd.Flags().BoolVar(&statusSkipPR, "skip-pr", false, "Skip gh pr view call")

	// backlog flags
	backlogCmd.Flags().BoolVar(&backlogArchive, "archive", false, "Include archived items")
	backlogCmd.Flags().StringVar(&backlogFormat, "format", "json", "Output format: json or table")
	backlogCmd.AddCommand(backlogMigrateCmd)
}
