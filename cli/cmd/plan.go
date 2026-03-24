package cmd

import (
	"fmt"
	"os"

	gitpkg "github.com/skaisser/blueprint/internal/git"
	"github.com/skaisser/blueprint/internal/plan"
	"github.com/spf13/cobra"
)

// ─── plan-tasks ──────────────────────────────────────────────────────────────

var planTasksBaseline string

var planTasksCmd = &cobra.Command{
	Use:   "plan-tasks",
	Short: "Extract tasks from active plan, optionally diff vs baseline commit",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
			os.Exit(1)
		}

		branch := repo.CurrentBranch()
		if branch == "" {
			branch = "main"
		}

		planFile := plan.FindPlanFile("blueprint", branch)
		if planFile == "" {
			fmt.Fprintf(os.Stderr, "Error: no active plan found\n")
			os.Exit(1)
		}

		var result *plan.TasksResult
		if planTasksBaseline != "" {
			result, err = plan.ExtractTasksWithBaseline(planFile, planTasksBaseline)
		} else {
			result, err = plan.ExtractTasks(planFile)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result.JSON())
	},
}

// ─── plan-profile ────────────────────────────────────────────────────────────

var planProfileCmd = &cobra.Command{
	Use:   "plan-profile <plan-file>",
	Short: "Quick plan profile as JSON (tasks, phases, complexity counts)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := plan.PlanProfile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result.JSON())
	},
}

// ─── plan-status ─────────────────────────────────────────────────────────────

var planStatusCmd = &cobra.Command{
	Use:   "plan-status",
	Short: "Plan completion status from active plan",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
			os.Exit(1)
		}

		branch := repo.CurrentBranch()
		if branch == "" {
			branch = "main"
		}

		result, err := plan.PlanStatus("blueprint", branch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result.JSON())
	},
}

// ─── validate ────────────────────────────────────────────────────────────────

var validateCmd = &cobra.Command{
	Use:   "validate <plan-file>",
	Short: "Full plan completeness check (markers, strategy, acceptance criteria, file refs)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := plan.ValidatePlan(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result.JSON())

		if !result.Valid {
			os.Exit(1)
		}
	},
}

func init() {
	planTasksCmd.Flags().StringVar(&planTasksBaseline, "baseline", "", "Compare tasks vs this commit")
}
