package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "blueprint",
	Short:   "BLUEPRINT SDLC — development lifecycle for Claude Code",
	Long:    "blueprint is the CLI for the BLUEPRINT SDLC pipeline. All commands are top-level.",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(prReviewCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(metaCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(backlogCmd)
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
