package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	gitpkg "github.com/skaisser/blueprint/internal/git"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// branchCmd is the parent for branch subcommands and also hosts base-branch as the default.
var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Branch utilities (safety, naming)",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// ─── base-branch ─────────────────────────────────────────────────────────────

var baseBranchCmd = &cobra.Command{
	Use:   "base-branch",
	Short: "Resolve the PR target branch (staging if exists, else main)",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository: %v\n", err)
			os.Exit(1)
		}

		stagingBranch := readStagingBranch()
		base := repo.DetectBaseBranch(stagingBranch)
		current := repo.CurrentBranch()
		hasStaging := base == stagingBranch

		result := map[string]interface{}{
			"base":        base,
			"has_staging": hasStaging,
			"current":     current,
		}

		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	},
}

// ─── branch safety ───────────────────────────────────────────────────────────

var branchSafetyCmd = &cobra.Command{
	Use:   "safety",
	Short: "Check if current branch allows push (blocks main/master)",
	Run: func(cmd *cobra.Command, args []string) {
		repo, err := gitpkg.Open("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository: %v\n", err)
			os.Exit(1)
		}

		branch := repo.CurrentBranch()
		if branch == "" {
			branch = "HEAD"
		}

		blocked := branch == "main" || branch == "master"

		var result map[string]interface{}
		if blocked {
			result = map[string]interface{}{
				"safe":   false,
				"branch": branch,
				"reason": "direct push to " + branch + " is blocked",
			}
		} else {
			result = map[string]interface{}{
				"safe":   true,
				"branch": branch,
			}
		}

		out, _ := json.Marshal(result)
		fmt.Println(string(out))

		if blocked {
			os.Exit(1)
		}
	},
}

// ─── branch-name ─────────────────────────────────────────────────────────────

var branchNameCmd = &cobra.Command{
	Use:   "branch-name <description>",
	Short: "Generate a sanitized branch name from a description",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := GenerateBranchName(args[0])
		fmt.Println(name)
	},
}

// GenerateBranchName creates a sanitized branch name from a description.
func GenerateBranchName(description string) string {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return "feat/unnamed"
	}

	lower := strings.ToLower(desc)

	// Detect prefix from description
	prefix := "feat"
	knownPrefixes := []string{"fix", "refactor", "hotfix", "chore", "docs", "test", "perf", "style", "build", "ci"}
	for _, p := range knownPrefixes {
		if strings.HasPrefix(lower, p+"/") || strings.HasPrefix(lower, p+" ") || strings.HasPrefix(lower, p+":") {
			prefix = p
			// Remove the prefix from the description
			lower = strings.TrimSpace(lower[len(p):])
			// Remove leading separator
			lower = strings.TrimLeft(lower, " /:")
			break
		}
	}

	// Replace spaces and underscores with hyphens
	slug := strings.ReplaceAll(lower, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special chars (keep only alphanumeric and hyphens)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "")

	// Collapse multiple hyphens
	reMulti := regexp.MustCompile(`-{2,}`)
	slug = reMulti.ReplaceAllString(slug, "-")

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "unnamed"
	}

	result := prefix + "/" + slug

	// Truncate to 50 chars
	if len(result) > 50 {
		result = result[:50]
		// Don't end with a hyphen after truncation
		result = strings.TrimRight(result, "-")
	}

	return result
}

// readStagingBranch reads the staging_branch from blueprint/.config.yml.
// Returns "staging" as default if config cannot be read.
func readStagingBranch() string {
	data, err := os.ReadFile("blueprint/.config.yml")
	if err != nil {
		return "staging"
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "staging"
	}

	if val, ok := raw["staging_branch"]; ok {
		if s, ok := val.(string); ok && s != "" {
			return s
		}
	}

	return "staging"
}

func init() {
	branchCmd.AddCommand(branchSafetyCmd)
	branchCmd.AddCommand(branchNameCmd)
}
