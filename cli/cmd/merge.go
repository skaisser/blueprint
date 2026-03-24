package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	gh "github.com/skaisser/blueprint/internal/github"
	"github.com/spf13/cobra"
)

var mergeChainStaging string

var mergeChainCmd = &cobra.Command{
	Use:   "merge-chain <branch>",
	Short: "Full PR create + merge chain: feat→staging→main",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]

		staging := mergeChainStaging
		if staging == "" {
			staging = readStagingBranch()
		}

		result := runMergeChain(branch, staging)
		out, _ := json.Marshal(result)
		fmt.Println(string(out))

		if _, ok := result["error"]; ok {
			os.Exit(1)
		}
	},
}

// runMergeChain executes the full merge chain: branch→staging→main.
func runMergeChain(branch, staging string) map[string]interface{} {
	// Step 1: Create PR from branch → staging
	stagingPR, err := createPR(branch, staging)
	if err != nil {
		return map[string]interface{}{"error": err.Error(), "step": "staging_pr"}
	}

	// Step 2: Merge staging PR
	if err := mergePR(stagingPR); err != nil {
		return map[string]interface{}{"error": err.Error(), "step": "staging_merge", "staging_pr": stagingPR}
	}

	// Step 3: Create PR from staging → main
	mainPR, err := createPR(staging, "main")
	if err != nil {
		return map[string]interface{}{"error": err.Error(), "step": "main_pr", "staging_pr": stagingPR}
	}

	// Step 4: Merge main PR
	if err := mergePR(mainPR); err != nil {
		return map[string]interface{}{"error": err.Error(), "step": "main_merge", "staging_pr": stagingPR, "main_pr": mainPR}
	}

	return map[string]interface{}{
		"staging_pr": stagingPR,
		"main_pr":    mainPR,
		"merged":     true,
	}
}

// createPR creates a PR from head to base and returns the PR number.
func createPR(head, base string) (int, error) {
	title := fmt.Sprintf("🔀 merge: %s → %s", head, base)
	out, stderr, err := gh.Run("gh", "pr", "create",
		"--head", head,
		"--base", base,
		"--title", title,
		"--body", fmt.Sprintf("Automated merge: %s → %s", head, base),
	)
	if err != nil {
		return 0, fmt.Errorf("gh pr create failed: %s %s", stderr, err)
	}

	// Extract PR number from URL (last segment)
	parts := strings.Split(strings.TrimSpace(out), "/")
	if len(parts) == 0 {
		return 0, fmt.Errorf("could not parse PR URL: %s", out)
	}
	numStr := parts[len(parts)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("could not parse PR number from: %s", out)
	}
	return num, nil
}

// mergePR merges a PR by number.
func mergePR(prNum int) error {
	_, stderr, err := gh.Run("gh", "pr", "merge", strconv.Itoa(prNum), "--merge", "--delete-branch=false")
	if err != nil {
		return fmt.Errorf("gh pr merge failed: %s %s", stderr, err)
	}
	return nil
}

// ─── commit format ──────────────────────────────────────────────────────────

var commitFmtCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit utilities",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var commitFormatCmd = &cobra.Command{
	Use:   "format <type> <message>",
	Short: "Format a commit message with the correct emoji",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		formatted, err := FormatCommitMessage(args[0], args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(formatted)
	},
}

// FormatCommitMessage formats a commit message with the correct emoji for the type.
func FormatCommitMessage(commitType, message string) (string, error) {
	// Import from audit package
	types := map[string]string{
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

	emoji, ok := types[commitType]
	if !ok {
		valid := make([]string, 0, len(types))
		for k := range types {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("invalid commit type %q, valid types: %s", commitType, strings.Join(valid, ", "))
	}

	return fmt.Sprintf("%s %s: %s", emoji, commitType, message), nil
}

// ─── review-poll ────────────────────────────────────────────────────────────

var reviewPollTimeout string

var reviewPollCmd = &cobra.Command{
	Use:   "review-poll <pr-number>",
	Short: "Poll for review comments on a PR",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prNum := args[0]

		timeout, err := time.ParseDuration(reviewPollTimeout)
		if err != nil {
			timeout = 20 * time.Minute
		}

		result := pollForReview(prNum, timeout)
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	},
}

// pollForReview polls a PR for reviews until timeout.
func pollForReview(prNum string, timeout time.Duration) map[string]interface{} {
	deadline := time.Now().Add(timeout)
	interval := 30 * time.Second

	for time.Now().Before(deadline) {
		comments, approved, err := checkPRReviewStatus(prNum)
		if err == nil && comments > 0 {
			return map[string]interface{}{
				"reviewed": true,
				"comments": comments,
				"approved": approved,
			}
		}

		// Check if we'd go past deadline on next sleep
		if time.Now().Add(interval).After(deadline) {
			break
		}
		time.Sleep(interval)
	}

	return map[string]interface{}{"timeout": true}
}

// checkPRReviewStatus checks how many review comments a PR has and if it's approved.
func checkPRReviewStatus(prNum string) (int, bool, error) {
	// Get reviews
	out, _, err := gh.Run("gh", "pr", "view", prNum,
		"--json", "reviews",
		"--jq", ".reviews | length")
	if err != nil {
		return 0, false, err
	}

	count, _ := strconv.Atoi(strings.TrimSpace(out))

	// Check for approval
	approvedOut, _, _ := gh.Run("gh", "pr", "view", prNum,
		"--json", "reviews",
		"--jq", `[.reviews[] | select(.state == "APPROVED")] | length`)
	approvedCount, _ := strconv.Atoi(strings.TrimSpace(approvedOut))

	// Also count issue comments (bot responses)
	repo, repoErr := gh.GetRepo()
	if repoErr == nil {
		commentOut, _, _ := gh.Run("gh", "api",
			fmt.Sprintf("repos/%s/issues/%s/comments", repo, prNum),
			"--jq", ". | length")
		commentCount, _ := strconv.Atoi(strings.TrimSpace(commentOut))
		count += commentCount
	}

	return count, approvedCount > 0, nil
}

// ─── detect-stack ───────────────────────────────────────────────────────────

var detectStackCmd = &cobra.Command{
	Use:   "detect-stack",
	Short: "Auto-detect project stack from manifest files",
	Run: func(cmd *cobra.Command, args []string) {
		result := DetectStack()
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	},
}

// DetectStack detects the project's tech stack from manifest files.
func DetectStack() map[string]string {
	result := map[string]string{}

	// Check Go
	if _, err := os.Stat("go.mod"); err == nil {
		result["language"] = "go"
		result["framework"] = detectGoFramework()
		result["version"] = detectGoVersion()
		result["test_runner"] = "go test"
		return result
	}

	// Check PHP / Laravel
	if _, err := os.Stat("composer.json"); err == nil {
		result["language"] = "php"
		result["framework"] = detectPHPFramework()
		result["version"] = detectLaravelVersion()
		result["test_runner"] = detectPHPTestRunner()
		return result
	}

	// Check Node.js
	if _, err := os.Stat("package.json"); err == nil {
		result["language"] = "javascript"
		result["framework"] = detectNodeFramework()
		result["version"] = ""
		result["test_runner"] = detectNodeTestRunner()
		return result
	}

	// Check Ruby
	if _, err := os.Stat("Gemfile"); err == nil {
		result["language"] = "ruby"
		result["framework"] = "unknown"
		result["test_runner"] = "rspec"
		return result
	}

	// Check Python
	if _, err := os.Stat("requirements.txt"); err == nil {
		result["language"] = "python"
		result["framework"] = "unknown"
		result["test_runner"] = "pytest"
		return result
	}

	result["language"] = "unknown"
	return result
}

func detectGoFramework() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "none"
	}
	content := string(data)
	if strings.Contains(content, "github.com/gin-gonic/gin") {
		return "gin"
	}
	if strings.Contains(content, "github.com/gofiber/fiber") {
		return "fiber"
	}
	if strings.Contains(content, "github.com/labstack/echo") {
		return "echo"
	}
	if strings.Contains(content, "github.com/spf13/cobra") {
		return "cobra"
	}
	return "none"
}

func detectGoVersion() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "go ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "go"))
		}
	}
	return ""
}

func detectPHPFramework() string {
	data, err := os.ReadFile("composer.json")
	if err != nil {
		return "unknown"
	}
	if strings.Contains(string(data), "laravel/framework") {
		return "laravel"
	}
	if strings.Contains(string(data), "symfony/") {
		return "symfony"
	}
	return "unknown"
}

func detectLaravelVersion() string {
	data, err := os.ReadFile("composer.lock")
	if err != nil {
		return ""
	}
	content := string(data)
	// Simple extraction: find laravel/framework version
	idx := strings.Index(content, `"name": "laravel/framework"`)
	if idx < 0 {
		return ""
	}
	// Look for version after this
	sub := content[idx:]
	vIdx := strings.Index(sub, `"version": "`)
	if vIdx < 0 {
		return ""
	}
	start := vIdx + len(`"version": "`)
	end := strings.Index(sub[start:], `"`)
	if end < 0 {
		return ""
	}
	ver := sub[start : start+end]
	// Convert v11.x.y to 11.x
	ver = strings.TrimPrefix(ver, "v")
	parts := strings.Split(ver, ".")
	if len(parts) >= 2 {
		return parts[0] + ".x"
	}
	return ver
}

func detectPHPTestRunner() string {
	if _, err := os.Stat("vendor/bin/pest"); err == nil {
		return "pest"
	}
	data, err := os.ReadFile("composer.json")
	if err != nil {
		return "phpunit"
	}
	if strings.Contains(string(data), "pestphp/pest") {
		return "pest"
	}
	return "phpunit"
}

func detectNodeFramework() string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return "unknown"
	}
	content := string(data)
	if strings.Contains(content, "\"next\"") {
		return "next"
	}
	if strings.Contains(content, "\"react-native\"") || strings.Contains(content, "\"expo\"") {
		return "react-native"
	}
	if strings.Contains(content, "\"react\"") {
		return "react"
	}
	if strings.Contains(content, "\"vue\"") {
		return "vue"
	}
	if strings.Contains(content, "\"svelte\"") {
		return "svelte"
	}
	if strings.Contains(content, "\"express\"") {
		return "express"
	}
	return "unknown"
}

func detectNodeTestRunner() string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return "unknown"
	}
	content := string(data)
	if strings.Contains(content, "\"vitest\"") {
		return "vitest"
	}
	if strings.Contains(content, "\"jest\"") {
		return "jest"
	}
	if strings.Contains(content, "\"mocha\"") {
		return "mocha"
	}
	return "unknown"
}

// ─── exec helper (for testing) ──────────────────────────────────────────────

// execCommand is a variable to allow test injection.
var execCommand = exec.Command

func init() {
	mergeChainCmd.Flags().StringVar(&mergeChainStaging, "staging", "", "Override staging branch name")
	reviewPollCmd.Flags().StringVar(&reviewPollTimeout, "timeout", "20m", "Timeout duration (e.g., 20m, 1h)")

	commitFmtCmd.AddCommand(commitFormatCmd)
	commitFmtCmd.AddCommand(commitPlanCmd)
}
