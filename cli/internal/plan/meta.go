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

// GetNextNum scans blueprint/live/ and blueprint/ for the highest numbered plan file and returns next.
func GetNextNum(planningDir string) string {
	maxNum := 0

	// Scan both live/ subdirectory and root planning dir
	dirs := []string{
		filepath.Join(planningDir, "live"),
		planningDir,
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			m := numPrefix.FindStringSubmatch(e.Name())
			if m != nil {
				n, _ := strconv.Atoi(m[1])
				if n > maxNum {
					maxNum = n
				}
			}
		}
	}

	if maxNum == 0 {
		return "0001"
	}
	return fmt.Sprintf("%04d", maxNum+1)
}

// collectPlanFiles scans a directory for plan files and returns todo files and all plan files
// with their paths relative to the base planningDir.
func collectPlanFiles(dir, planningDir string) (todoFiles []string, allPlanFiles []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}
		// Build path relative to planningDir so results always use planningDir as base
		relPath := name
		if dir != planningDir {
			rel, err := filepath.Rel(planningDir, filepath.Join(dir, name))
			if err == nil {
				relPath = rel
			}
		}
		if strings.HasSuffix(name, "-todo.md") {
			todoFiles = append(todoFiles, relPath)
		}
		if numPrefix.MatchString(name) {
			allPlanFiles = append(allPlanFiles, relPath)
		}
	}
	return
}

// FindPlanFile finds the active plan file matching the branch, or falls back to most recent.
// Searches blueprint/live/ first, then falls back to blueprint/ root.
func FindPlanFile(planningDir, branch string) string {
	var todoFiles []string
	var allPlanFiles []string

	// Search live/ subdirectory first, then root
	dirs := []string{
		filepath.Join(planningDir, "live"),
		planningDir,
	}

	for _, dir := range dirs {
		td, ap := collectPlanFiles(dir, planningDir)
		todoFiles = append(todoFiles, td...)
		allPlanFiles = append(allPlanFiles, ap...)
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
		base := filepath.Base(f)
		fSlug := regexp.MustCompile(`^\d{4}-[a-z]+-`).ReplaceAllString(base, "")
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
