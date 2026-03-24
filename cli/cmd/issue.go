package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/skaisser/blueprint/internal/github"
	"github.com/skaisser/blueprint/internal/plan"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "GitHub issue operations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var issueFetchCmd = &cobra.Command{
	Use:   "fetch <number>",
	Short: "Fetch a GitHub issue as JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		number := args[0]
		result, err := fetchIssue(number)
		if err != nil {
			errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintln(os.Stderr, string(errJSON))
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

var issueCloseCmd = &cobra.Command{
	Use:   "close <number>",
	Short: "Close a GitHub issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		number := args[0]
		result, err := closeIssue(number)
		if err != nil {
			errJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintln(os.Stderr, string(errJSON))
			os.Exit(1)
		}
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	},
}

var prBodyIssueFlag string
var prBodyPlanFlag string

var prBodyCmd = &cobra.Command{
	Use:   "pr-body",
	Short: "Generate a PR body template",
	Run: func(cmd *cobra.Command, args []string) {
		body := generatePRBody(prBodyIssueFlag, prBodyPlanFlag)
		fmt.Print(body)
	},
}

// fetchIssue calls gh to fetch issue details. Returns raw JSON string.
func fetchIssue(number string) (string, error) {
	if err := checkGHInstalled(); err != nil {
		return "", err
	}

	out, stderr, err := github.Run("gh", "issue", "view", number,
		"--json", "title,body,labels,assignees,state")
	if err != nil {
		if strings.Contains(stderr, "Could not resolve") || strings.Contains(stderr, "not found") ||
			strings.Contains(out, "not found") {
			return "", fmt.Errorf("issue #%s not found", number)
		}
		if strings.Contains(stderr, "auth") || strings.Contains(stderr, "login") {
			return "", fmt.Errorf("gh auth required: run 'gh auth login'")
		}
		return "", fmt.Errorf("gh error: %s", stderr)
	}

	return out, nil
}

// closeIssueResult holds the result of closing an issue.
type closeIssueResult struct {
	Number          string `json:"number"`
	State           string `json:"state"`
	WasAlreadyClosed bool  `json:"was_already_closed"`
}

// closeIssue checks state and closes if open.
func closeIssue(number string) (*closeIssueResult, error) {
	if err := checkGHInstalled(); err != nil {
		return nil, err
	}

	// Check current state
	out, stderr, err := github.Run("gh", "issue", "view", number, "--json", "state", "-q", ".state")
	if err != nil {
		if strings.Contains(stderr, "not found") || strings.Contains(out, "not found") {
			return nil, fmt.Errorf("issue #%s not found", number)
		}
		return nil, fmt.Errorf("gh error: %s", stderr)
	}

	state := strings.TrimSpace(strings.ToUpper(out))

	if state == "CLOSED" {
		return &closeIssueResult{
			Number:          number,
			State:           "closed",
			WasAlreadyClosed: true,
		}, nil
	}

	// Close the issue
	_, stderr, err = github.Run("gh", "issue", "close", number)
	if err != nil {
		return nil, fmt.Errorf("failed to close issue: %s", stderr)
	}

	return &closeIssueResult{
		Number:          number,
		State:           "closed",
		WasAlreadyClosed: false,
	}, nil
}

// checkGHInstalled verifies that gh CLI is available.
func checkGHInstalled() error {
	_, _, err := github.Run("gh", "--version")
	if err != nil {
		return fmt.Errorf("gh CLI not installed: install from https://cli.github.com")
	}
	return nil
}

// generatePRBody creates a PR body markdown template.
func generatePRBody(issueNumber, planFile string) string {
	var summary string
	var changes []string
	var related []string
	var planIssue string

	if planFile != "" {
		summary, changes, planIssue = extractPlanInfo(planFile)
	}

	// If no explicit issue but plan has one, use that
	if issueNumber == "" && planIssue != "" {
		issueNumber = planIssue
	}

	var sb strings.Builder

	sb.WriteString("## Summary\n")
	if summary != "" {
		sb.WriteString(summary + "\n")
	} else {
		sb.WriteString("<!-- Describe the changes -->\n")
	}

	sb.WriteString("\n## Changes\n")
	if len(changes) > 0 {
		for _, c := range changes {
			sb.WriteString("- " + c + "\n")
		}
	} else {
		sb.WriteString("- <!-- List key changes -->\n")
	}

	if issueNumber != "" {
		related = append(related, fmt.Sprintf("Closes #%s", strings.TrimPrefix(issueNumber, "#")))
	}

	if len(related) > 0 {
		sb.WriteString("\n## Related\n")
		for _, r := range related {
			sb.WriteString(r + "\n")
		}
	}

	sb.WriteString("\n## Test Plan\n")
	sb.WriteString("- <!-- How to verify the changes -->\n")

	return sb.String()
}

// extractPlanInfo reads a plan file and extracts summary, phase names, and issue ref.
func extractPlanInfo(planFile string) (summary string, changes []string, issueRef string) {
	content, err := os.ReadFile(planFile)
	if err != nil {
		return "", nil, ""
	}

	text := string(content)

	// Parse frontmatter for title and issue
	fm, body, err := parsePlanFrontmatterForPR(text)
	if err == nil {
		if fm.Title != "" {
			summary = fm.Title
		}
		if fm.Issue != "" {
			issueRef = fm.Issue
		}
	} else {
		body = text
	}

	// Extract phase names from ## Phase N: <title> headers
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Phase") {
			// Extract title after "## Phase N: "
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				title := strings.TrimSpace(parts[1])
				// Remove trailing complexity markers like [S], [H], [O]
				title = strings.TrimRight(title, " ")
				for _, suffix := range []string{" [S]", " [H]", " [O]"} {
					title = strings.TrimSuffix(title, suffix)
				}
				if title != "" {
					changes = append(changes, title)
				}
			}
		}
	}

	return summary, changes, issueRef
}

// parsePlanFrontmatterForPR extracts YAML frontmatter for PR body generation.
// This is a lightweight wrapper to avoid exporting internal plan parsing.
func parsePlanFrontmatterForPR(content string) (*plan.PlanFrontmatter, string, error) {
	return plan.ParseFrontmatterPublic(content)
}

func init() {
	issueCmd.AddCommand(issueFetchCmd)
	issueCmd.AddCommand(issueCloseCmd)

	prBodyCmd.Flags().StringVar(&prBodyIssueFlag, "issue", "", "GitHub issue number to reference")
	prBodyCmd.Flags().StringVar(&prBodyPlanFlag, "plan", "", "Plan file to extract info from")
}
