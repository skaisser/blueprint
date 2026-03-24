package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// PlanValidateResult is the JSON output of plan validate.
type PlanValidateResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

// JSON returns the result as indented JSON.
func (r *PlanValidateResult) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

var (
	backtickPathRe    = regexp.MustCompile("`([a-zA-Z0-9_./-]+\\.[a-zA-Z0-9]+)`")
	executionStratRe  = regexp.MustCompile(`(?im)^##+ .*Execution Strategy`)
	acceptanceCritRe  = regexp.MustCompile(`(?im)^##+ .*Acceptance Criteria`)
)

// ValidatePlan performs a full plan completeness check.
func ValidatePlan(planFile string) (*PlanValidateResult, error) {
	content, err := os.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read plan file: %w", err)
	}

	return ValidatePlanContent(string(content)), nil
}

// ValidatePlanContent validates plan content without reading from disk.
func ValidatePlanContent(content string) *PlanValidateResult {
	result := &PlanValidateResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
	}

	// Check for Execution Strategy section
	if !executionStratRe.MatchString(content) {
		result.Errors = append(result.Errors, "missing Execution Strategy section")
	}

	// Check for Acceptance Criteria section
	if !acceptanceCritRe.MatchString(content) {
		result.Errors = append(result.Errors, "missing Acceptance Criteria section")
	}

	// Check tasks for complexity markers
	tasks := ParseTasks(content)
	missingMarkers := 0
	for _, t := range tasks {
		if t.Complexity == "" {
			missingMarkers++
		}
	}
	if missingMarkers > 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("%d tasks missing complexity markers", missingMarkers))
	}

	// Check referenced files exist (backtick-quoted paths that look like file paths)
	paths := backtickPathRe.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	for _, m := range paths {
		path := m[1]
		if seen[path] {
			continue
		}
		seen[path] = true

		// Skip paths that are clearly not local files
		if strings.Contains(path, "://") || strings.HasPrefix(path, "http") {
			continue
		}
		// Skip common non-file patterns
		if strings.HasPrefix(path, "v") && strings.Contains(path, ".") && !strings.Contains(path, "/") {
			// Version strings like v1.2.3
			continue
		}
		// Only check paths that contain a directory separator (more likely to be real file paths)
		if !strings.Contains(path, "/") {
			continue
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("file %s referenced but not found", path))
		}
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}
