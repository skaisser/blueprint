package plan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var numPrefix = regexp.MustCompile(`^(\d{4})`)

// MetaResult is the JSON output of the meta command.
type MetaResult struct {
	NextNum    string `json:"next_num"`
	BaseBranch string `json:"base_branch"`
	Branch     string `json:"branch"`
	PlanFile   string `json:"plan_file"`
	PlanNum    string `json:"plan_num"`
	Status     string `json:"status"`
	Progress   string `json:"progress"`
	Project    string `json:"project"`
	GitRemote  string `json:"git_remote"`
	Today      string `json:"today"`
}

// GetNextNum scans blueprint/ for the highest numbered plan file and returns next.
func GetNextNum(planningDir string) string {
	entries, err := os.ReadDir(planningDir)
	if err != nil {
		return "0001"
	}

	maxNum := 0
	for _, e := range entries {
		m := numPrefix.FindStringSubmatch(e.Name())
		if m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > maxNum {
				maxNum = n
			}
		}
	}

	if maxNum == 0 {
		return "0001"
	}
	return fmt.Sprintf("%04d", maxNum+1)
}

// FindPlanFile finds the active plan file matching the branch, or falls back to most recent.
func FindPlanFile(planningDir, branch string) string {
	entries, err := os.ReadDir(planningDir)
	if err != nil {
		return ""
	}

	var todoFiles []string
	var allPlanFiles []string

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}
		if strings.HasSuffix(name, "-todo.md") {
			todoFiles = append(todoFiles, name)
		}
		if numPrefix.MatchString(name) {
			allPlanFiles = append(allPlanFiles, name)
		}
	}

	candidates := todoFiles
	if len(candidates) == 0 {
		candidates = allPlanFiles
	}
	if len(candidates) == 0 {
		return ""
	}

	// Extract suffix after type prefix (feat/fix-name → fix-name)
	branchSuffix := regexp.MustCompile(`^[^/]+/`).ReplaceAllString(branch, "")

	// Sort descending (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(candidates)))

	// Try to match branch suffix against filename
	for _, f := range candidates {
		fSlug := regexp.MustCompile(`^\d{4}-[a-z]+-`).ReplaceAllString(f, "")
		fSlug = strings.TrimSuffix(fSlug, "-todo.md")
		fSlug = strings.TrimSuffix(fSlug, "-completed.md")

		if fSlug != "" && branchSuffix != "" &&
			(strings.Contains(fSlug, branchSuffix) || strings.Contains(branchSuffix, fSlug)) {
			return filepath.Join(planningDir, f)
		}
	}

	// Fallback: most recently modified
	type fileWithTime struct {
		name    string
		modTime time.Time
	}
	var fwt []fileWithTime
	for _, f := range candidates {
		info, err := os.Stat(filepath.Join(planningDir, f))
		if err == nil {
			fwt = append(fwt, fileWithTime{f, info.ModTime()})
		}
	}
	sort.Slice(fwt, func(i, j int) bool {
		return fwt[i].modTime.After(fwt[j].modTime)
	})
	if len(fwt) > 0 {
		return filepath.Join(planningDir, fwt[0].name)
	}
	return filepath.Join(planningDir, candidates[0])
}

// PlanHeader holds parsed header fields from a plan file.
type PlanHeader struct {
	PlanNum  string
	Status   string
	Progress string
}

// ParsePlanHeader reads the first 30 lines of a plan file for status/progress.
func ParsePlanHeader(planFile string) PlanHeader {
	result := PlanHeader{}
	if planFile == "" {
		return result
	}

	// Extract plan number from filename
	base := filepath.Base(planFile)
	m := numPrefix.FindStringSubmatch(base)
	if m != nil {
		result.PlanNum = m[1]
	}

	f, err := os.Open(planFile)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() && lineNum < 30 {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "> **Status:**") {
			result.Status = strings.TrimSpace(strings.TrimPrefix(line, "> **Status:**"))
		} else if strings.HasPrefix(line, "> **Progress:**") {
			result.Progress = strings.TrimSpace(strings.TrimPrefix(line, "> **Progress:**"))
		}
		lineNum++
	}

	return result
}

// MetaJSON returns the meta result as indented JSON string.
func (m *MetaResult) JSON() string {
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}
