package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/skaisser/blueprint/internal/plan"
	"github.com/spf13/cobra"
)

var frontmatterCmd = &cobra.Command{
	Use:   "frontmatter",
	Short: "Validate and fix plan file frontmatter",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var frontmatterValidateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate frontmatter of a plan file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		result := plan.ValidateFrontmatter(filePath)
		fmt.Println(result.JSON())

		if !result.Valid {
			os.Exit(1)
		}
	},
}

var frontmatterFixDryRun bool

var frontmatterFixCmd = &cobra.Command{
	Use:   "fix <file>",
	Short: "Fix frontmatter of a plan file (add missing fields, correct status)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		changes, err := plan.FixFrontmatter(filePath, frontmatterFixDryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if changes == nil || len(changes) == 0 {
			fmt.Println("No changes needed.")
			return
		}

		if frontmatterFixDryRun {
			fmt.Println("Dry run — would make these changes:")
		} else {
			fmt.Println("Fixed:")
		}

		output, _ := json.Marshal(changes)
		fmt.Println(string(output))
	},
}

func init() {
	frontmatterFixCmd.Flags().BoolVar(&frontmatterFixDryRun, "dry-run", false, "Report changes without writing")
	frontmatterCmd.AddCommand(frontmatterValidateCmd)
	frontmatterCmd.AddCommand(frontmatterFixCmd)
}
