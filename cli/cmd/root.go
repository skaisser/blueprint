package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "blueprint",
	Short:   "BLUEPRINT SDLC — development lifecycle for Claude Code",
	Long:    "blueprint consolidates SDLC scripts (audit, sdlc, pr-review) into a single high-performance Go binary.",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(sdlcCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(prReviewCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(frontmatterCmd)
	rootCmd.AddCommand(baseBranchCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(prBodyCmd)
	rootCmd.AddCommand(planTasksCmd)
	rootCmd.AddCommand(planProfileCmd)
	rootCmd.AddCommand(planStatusCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(timeCmd)
	rootCmd.AddCommand(worktreeCmd)
	rootCmd.AddCommand(mergeChainCmd)
	rootCmd.AddCommand(commitFmtCmd)
	rootCmd.AddCommand(reviewPollCmd)
	rootCmd.AddCommand(detectStackCmd)
}

func Execute() error {
	rootCmd.Version = Version
	// If called as just "blueprint" with no args, show help
	if len(os.Args) == 1 {
		rootCmd.Help()
		return nil
	}
	return rootCmd.Execute()
}
