package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	gitpkg "github.com/skaisser/blueprint/internal/git"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Git worktree management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var worktreeCreateBase string

var worktreeCreateCmd = &cobra.Command{
	Use:   "create <branch>",
	Short: "Create a git worktree for the given branch",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]

		repo, err := gitpkg.Open("")
		if err != nil {
			printJSON(map[string]interface{}{"error": "not in a git repository"})
			os.Exit(1)
		}

		// Determine base ref
		base := worktreeCreateBase
		if base == "" {
			stagingBranch := readStagingBranch()
			base = repo.DetectBaseBranch(stagingBranch)
		}

		// Compute worktree path
		wtPath := computeWorktreePath(branch)

		// Fetch origin
		fetchCmd := exec.Command("git", "fetch", "origin")
		fetchCmd.Stderr = os.Stderr
		_ = fetchCmd.Run() // best effort

		// Create worktree
		gitArgs := []string{"worktree", "add", "-b", branch, wtPath, base}
		// If branch already exists, try without -b
		wtCmd := exec.Command("git", gitArgs...)
		wtCmd.Stderr = os.Stderr
		if err := wtCmd.Run(); err != nil {
			// Retry without -b (branch already exists)
			gitArgs = []string{"worktree", "add", wtPath, branch}
			wtCmd = exec.Command("git", gitArgs...)
			wtCmd.Stderr = os.Stderr
			if err := wtCmd.Run(); err != nil {
				printJSON(map[string]interface{}{"error": fmt.Sprintf("failed to create worktree: %v", err)})
				os.Exit(1)
			}
		}

		// Verify it exists
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			printJSON(map[string]interface{}{"error": "worktree path does not exist after creation"})
			os.Exit(1)
		}

		absPath, _ := filepath.Abs(wtPath)
		printJSON(map[string]interface{}{
			"path":   absPath,
			"branch": branch,
			"base":   base,
		})
	},
}

var worktreeRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove a git worktree",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wtPath := args[0]

		// Safety: verify path exists
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			printJSON(map[string]interface{}{"error": "path does not exist", "path": wtPath})
			os.Exit(1)
		}

		// Safety: verify it's a worktree
		if !isWorktree(wtPath) {
			printJSON(map[string]interface{}{"error": "path is not a git worktree", "path": wtPath})
			os.Exit(1)
		}

		// Remove worktree
		rmCmd := exec.Command("git", "worktree", "remove", "--force", wtPath)
		rmCmd.Stderr = os.Stderr
		if err := rmCmd.Run(); err != nil {
			printJSON(map[string]interface{}{"error": fmt.Sprintf("failed to remove worktree: %v", err)})
			os.Exit(1)
		}

		// Prune
		pruneCmd := exec.Command("git", "worktree", "prune")
		_ = pruneCmd.Run()

		absPath, _ := filepath.Abs(wtPath)
		printJSON(map[string]interface{}{
			"removed": true,
			"path":    absPath,
		})
	},
}

// computeWorktreePath determines the worktree directory path.
// If branch contains a plan number (e.g., feat/0002-something), uses ../projectN.
// Otherwise uses ../project-branch (sanitized).
func computeWorktreePath(branch string) string {
	cwd, _ := os.Getwd()
	parent := filepath.Dir(cwd)
	project := filepath.Base(cwd)

	// Try to extract plan number from branch name
	rePlanNum := regexp.MustCompile(`(\d{4})`)
	if m := rePlanNum.FindString(branch); m != "" {
		return filepath.Join(parent, project+m)
	}

	// Sanitize branch name for directory
	slug := strings.ReplaceAll(branch, "/", "-")
	slug = strings.ReplaceAll(slug, " ", "-")
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	slug = re.ReplaceAllString(slug, "")

	return filepath.Join(parent, project+"-"+slug)
}

// isWorktree checks if a path is a git worktree (has .git file, not directory).
func isWorktree(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	// Worktrees have a .git file (not directory) pointing to the main repo
	return !info.IsDir()
}

// printJSON marshals and prints a map as JSON.
func printJSON(v interface{}) {
	out, _ := json.Marshal(v)
	fmt.Println(string(out))
}

func init() {
	worktreeCreateCmd.Flags().StringVar(&worktreeCreateBase, "base", "", "Base ref for the worktree (default: auto-detect)")
	worktreeCmd.AddCommand(worktreeCreateCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)
}
