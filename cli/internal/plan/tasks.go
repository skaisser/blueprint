package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Task represents a single checkbox task from a plan file.
type Task struct {
	Text       string `json:"text"`
	Done       bool   `json:"done"`
	Complexity string `json:"complexity,omitempty"`
}

// TasksResult is the JSON output of plan-tasks.
type TasksResult struct {
	PlanFile string `json:"plan_file"`
	Tasks    []Task `json:"tasks"`
	Total    int    `json:"total"`
	Done     int    `json:"done"`
	Added    []Task `json:"added,omitempty"`
	Removed  []Task `json:"removed,omitempty"`
}

// JSON returns the result as indented JSON.
func (r *TasksResult) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

// ProfileResult is the JSON output of plan-profile.
type ProfileResult struct {
	OpenTasks     int    `json:"open_tasks"`
	DoneTasks     int    `json:"done_tasks"`
	TotalTasks    int    `json:"total_tasks"`
	Phases        int    `json:"phases"`
	MaxComplexity string `json:"max_complexity"`
	H             int    `json:"h"`
	S             int    `json:"s"`
	O             int    `json:"o"`
}

// JSON returns the result as indented JSON.
func (r *ProfileResult) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

// StatusResult is the JSON output of plan-status.
type StatusResult struct {
	PlanFile   string `json:"plan_file"`
	Total      int    `json:"total"`
	Done       int    `json:"done"`
	Phases     int    `json:"phases"`
	PhasesDone int    `json:"phases_done"`
	Status     string `json:"status"`
}

// JSON returns the result as indented JSON.
func (r *StatusResult) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

var (
	taskLineRe     = regexp.MustCompile(`^(\s*)-\s*\[([ xX])\]\s*(.+)$`)
	complexityRe   = regexp.MustCompile(`\[([HSOhso])\]`)
	phaseLineRe    = regexp.MustCompile(`(?im)^##+ Phase \d+`)
)

// ParseTasks extracts checkbox tasks from plan content.
func ParseTasks(content string) []Task {
	var tasks []Task
	for _, line := range strings.Split(content, "\n") {
		m := taskLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		done := strings.ToLower(m[2]) == "x"
		text := strings.TrimSpace(m[3])

		complexity := ""
		cm := complexityRe.FindStringSubmatch(text)
		if cm != nil {
			complexity = strings.ToUpper(cm[1])
		}

		tasks = append(tasks, Task{
			Text:       text,
			Done:       done,
			Complexity: complexity,
		})
	}
	return tasks
}

// ExtractTasks reads a plan file and extracts all tasks.
func ExtractTasks(planFile string) (*TasksResult, error) {
	content, err := os.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read plan file: %w", err)
	}

	tasks := ParseTasks(string(content))
	done := 0
	for _, t := range tasks {
		if t.Done {
			done++
		}
	}

	return &TasksResult{
		PlanFile: planFile,
		Tasks:    tasks,
		Total:    len(tasks),
		Done:     done,
	}, nil
}

// ExtractTasksWithBaseline extracts tasks and computes diff against a baseline commit.
func ExtractTasksWithBaseline(planFile, baseline string) (*TasksResult, error) {
	result, err := ExtractTasks(planFile)
	if err != nil {
		return nil, err
	}

	// Get old content from git
	oldContent, err := gitShowFile(baseline, planFile)
	if err != nil {
		return nil, fmt.Errorf("cannot get baseline content: %w", err)
	}

	oldTasks := ParseTasks(oldContent)

	// Build maps for diff
	oldMap := make(map[string]Task)
	for _, t := range oldTasks {
		oldMap[t.Text] = t
	}
	newMap := make(map[string]Task)
	for _, t := range result.Tasks {
		newMap[t.Text] = t
	}

	// Find added tasks (in new but not in old)
	var added []Task
	for _, t := range result.Tasks {
		if _, exists := oldMap[t.Text]; !exists {
			added = append(added, t)
		}
	}

	// Find removed tasks (in old but not in new)
	var removed []Task
	for _, t := range oldTasks {
		if _, exists := newMap[t.Text]; !exists {
			removed = append(removed, t)
		}
	}

	result.Added = added
	result.Removed = removed

	return result, nil
}

// gitShowFile retrieves file content at a specific commit using git show.
func gitShowFile(commit, filePath string) (string, error) {
	cmd := exec.Command("git", "show", commit+":"+filePath)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git show failed: %w", err)
	}
	return string(out), nil
}

// PlanProfile computes a profile (counts) from a plan file.
func PlanProfile(planFile string) (*ProfileResult, error) {
	content, err := os.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read plan file: %w", err)
	}

	return PlanProfileFromContent(string(content)), nil
}

// PlanProfileFromContent computes a profile from plan content string.
func PlanProfileFromContent(content string) *ProfileResult {
	tasks := ParseTasks(content)

	open := 0
	done := 0
	h, s, o := 0, 0, 0

	for _, t := range tasks {
		if t.Done {
			done++
		} else {
			open++
		}
		switch t.Complexity {
		case "H":
			h++
		case "S":
			s++
		case "O":
			o++
		}
	}

	// Count phases
	phases := len(phaseLineRe.FindAllString(content, -1))

	// Determine max complexity
	maxComplexity := ""
	if o > 0 {
		maxComplexity = "O"
	} else if s > 0 {
		maxComplexity = "S"
	} else if h > 0 {
		maxComplexity = "H"
	}

	return &ProfileResult{
		OpenTasks:     open,
		DoneTasks:     done,
		TotalTasks:    len(tasks),
		Phases:        phases,
		MaxComplexity: maxComplexity,
		H:             h,
		S:             s,
		O:             o,
	}
}

// PlanStatus computes plan completion status from the active plan.
func PlanStatus(planningDir, branch string) (*StatusResult, error) {
	planFile := FindPlanFile(planningDir, branch)
	if planFile == "" {
		return nil, fmt.Errorf("no active plan found")
	}

	content, err := os.ReadFile(planFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read plan file: %w", err)
	}

	return PlanStatusFromContent(string(content), planFile), nil
}

// PlanStatusFromContent computes plan status from content and plan file path.
func PlanStatusFromContent(content, planFile string) *StatusResult {
	tasks := ParseTasks(content)

	total := len(tasks)
	done := 0
	for _, t := range tasks {
		if t.Done {
			done++
		}
	}

	// Count phases and done phases
	lines := strings.Split(content, "\n")
	phases := 0
	phasesDone := 0

	// Find phase boundaries
	type phaseRange struct {
		start int
		end   int
	}
	var phaseRanges []phaseRange

	for i, line := range lines {
		if phaseLineRe.MatchString(line) {
			phases++
			if len(phaseRanges) > 0 {
				phaseRanges[len(phaseRanges)-1].end = i
			}
			phaseRanges = append(phaseRanges, phaseRange{start: i, end: len(lines)})
		}
	}

	// Check if each phase is fully done
	for _, pr := range phaseRanges {
		phaseDone := 0
		phaseOpen := 0
		for i := pr.start; i < pr.end; i++ {
			m := taskLineRe.FindStringSubmatch(lines[i])
			if m == nil {
				continue
			}
			if strings.ToLower(m[2]) == "x" {
				phaseDone++
			} else {
				phaseOpen++
			}
		}
		if phaseDone > 0 && phaseOpen == 0 {
			phasesDone++
		}
	}

	// Determine status
	status := "not-started"
	if done > 0 && done < total {
		status = "in-progress"
	} else if done == total && total > 0 {
		status = "completed"
	}

	return &StatusResult{
		PlanFile:   planFile,
		Total:      total,
		Done:       done,
		Phases:     phases,
		PhasesDone: phasesDone,
		Status:     status,
	}
}
