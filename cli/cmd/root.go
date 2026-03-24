package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "blueprint",
	Short:   "BLUEPRINT SDLC — development lifecycle for Claude Code",
	Long:    "blueprint — development lifecycle CLI for Claude Code projects.",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Core commands (were under sdlcCmd, now directly on root)
	rootCmd.AddCommand(metaCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(fullCmd)
	rootCmd.AddCommand(commitFmtCmd)
	rootCmd.AddCommand(backlogCmd)

	// Audit & PR
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(prReviewCmd)
	rootCmd.AddCommand(updateCmd)

	// Config & frontmatter
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(frontmatterCmd)

	// Branch operations
	rootCmd.AddCommand(baseBranchCmd)
	rootCmd.AddCommand(branchCmd)

	// Plan operations
	rootCmd.AddCommand(planTasksCmd)
	rootCmd.AddCommand(planProfileCmd)
	rootCmd.AddCommand(planStatusCmd)
	rootCmd.AddCommand(validateCmd)

	// GitHub integration
	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(prBodyCmd)

	// Complex operations
	rootCmd.AddCommand(worktreeCmd)
	rootCmd.AddCommand(mergeChainCmd)
	rootCmd.AddCommand(reviewPollCmd)
	rootCmd.AddCommand(detectStackCmd)
	rootCmd.AddCommand(timeCmd)
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
