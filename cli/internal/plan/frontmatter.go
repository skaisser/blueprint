package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// PlanFrontmatter represents YAML frontmatter fields for plan files.
type PlanFrontmatter struct {
	ID         string   `yaml:"id,omitempty" json:"id,omitempty"`
	Title      string   `yaml:"title,omitempty" json:"title,omitempty"`
	Status     string   `yaml:"status,omitempty" json:"status,omitempty"`
	Complexity string   `yaml:"complexity,omitempty" json:"complexity,omitempty"`
	Branch     string   `yaml:"branch,omitempty" json:"branch,omitempty"`
	Tags       []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Issue      string   `yaml:"issue,omitempty" json:"issue,omitempty"`
	Created    string   `yaml:"created,omitempty" json:"created,omitempty"`
}

// ValidateResult holds the result of frontmatter validation.
type ValidateResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// JSON returns the validate result as JSON string.
func (v *ValidateResult) JSON() string {
	b, _ := json.Marshal(v)
	return string(b)
}

// detectFileType determines the file type from its directory location.
// Returns one of: "backlog", "live", "upstream", "expired", or "unknown".
func detectFileType(filePath string) string {
	dir := filepath.Dir(filePath)
	base := filepath.Base(dir)

	switch base {
	case "backlog":
		return "backlog"
	case "live":
		return "live"
	case "upstream":
		return "upstream"
	case "expired":
		return "expired"
	}

	// Check parent directory too (e.g., blueprint/backlog/archive/)
	parent := filepath.Base(filepath.Dir(dir))
	switch parent {
	case "backlog":
		return "backlog"
	case "live":
		return "live"
	case "upstream":
		return "upstream"
	case "expired":
		return "expired"
	}

	return "unknown"
}

// validStatuses defines which statuses are valid for each file type.
var validStatuses = map[string][]string{
	"backlog":  {"backlog"},
	"live":     {"todo", "in-progress", "approved", "awaiting-review"},
	"upstream": {"completed"},
	"expired":  {"expired", "cancelled"},
}

// ParseFrontmatterPublic is the exported version of parseFrontmatter for use by other packages.
func ParseFrontmatterPublic(content string) (*PlanFrontmatter, string, error) {
	return parseFrontmatter(content)
}

// parseFrontmatter extracts YAML frontmatter from a markdown file's content.
func parseFrontmatter(content string) (*PlanFrontmatter, string, error) {
	if !strings.HasPrefix(content, "---") {
		return nil, content, fmt.Errorf("no YAML frontmatter found (file does not start with ---)")
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 || strings.TrimSpace(parts[1]) == "" {
		return nil, content, fmt.Errorf("malformed YAML frontmatter")
	}

	fm := &PlanFrontmatter{}
	if err := yaml.Unmarshal([]byte(parts[1]), fm); err != nil {
		return nil, content, fmt.Errorf("invalid YAML: %w", err)
	}

	body := parts[2]
	return fm, body, nil
}

// ValidateFrontmatter validates a plan file's frontmatter against its expected type.
func ValidateFrontmatter(filePath string) *ValidateResult {
	result := &ValidateResult{Valid: true}

	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("cannot read file: %v", err))
		return result
	}

	fm, _, err := parseFrontmatter(string(content))
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	fileType := detectFileType(filePath)
	if fileType == "unknown" {
		result.Errors = append(result.Errors, "cannot determine file type from location")
		result.Valid = false
		return result
	}

	// Check required fields by type
	if fm.ID == "" {
		result.Errors = append(result.Errors, "missing field: id")
	}
	if fm.Title == "" {
		result.Errors = append(result.Errors, "missing field: title")
	}
	if fm.Status == "" {
		result.Errors = append(result.Errors, "missing field: status")
	}
	if fm.Tags == nil || len(fm.Tags) == 0 {
		result.Errors = append(result.Errors, "missing field: tags")
	}

	// Type-specific required fields
	switch fileType {
	case "live":
		if fm.Complexity == "" {
			result.Errors = append(result.Errors, "missing field: complexity")
		}
		if fm.Branch == "" {
			result.Errors = append(result.Errors, "missing field: branch")
		}
	case "upstream":
		if fm.Branch == "" {
			result.Errors = append(result.Errors, "missing field: branch")
		}
	}

	// Validate status matches location
	if fm.Status != "" && fileType != "unknown" {
		allowed, ok := validStatuses[fileType]
		if ok {
			found := false
			for _, s := range allowed {
				if fm.Status == s {
					found = true
					break
				}
			}
			if !found {
				result.Errors = append(result.Errors, fmt.Sprintf(
					"status '%s' but file is in %s/ (expected one of: %s)",
					fm.Status, fileType, strings.Join(allowed, ", "),
				))
			}
		}
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result
}

// FixFrontmatter fixes a plan file's frontmatter by adding missing fields and
// correcting status/location mismatches. Returns a description of changes made.
// If dryRun is true, no file is written.
func FixFrontmatter(filePath string, dryRun bool) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	text := string(content)
	fileType := detectFileType(filePath)
	if fileType == "unknown" {
		return nil, fmt.Errorf("cannot determine file type from location")
	}

	var changes []string
	var fm *PlanFrontmatter
	var body string

	fm, body, err = parseFrontmatter(text)
	if err != nil {
		// No frontmatter — create one from scratch
		fm = &PlanFrontmatter{}
		body = text
		changes = append(changes, "add YAML frontmatter block")
	}

	// Fix missing ID — derive from filename
	if fm.ID == "" {
		base := filepath.Base(filePath)
		m := numPrefix.FindStringSubmatch(base)
		if m != nil {
			fm.ID = m[1]
			changes = append(changes, fmt.Sprintf("set id: %s", fm.ID))
		}
	}

	// Fix missing title — derive from first heading in body
	if fm.Title == "" {
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				fm.Title = strings.TrimPrefix(line, "# ")
				changes = append(changes, fmt.Sprintf("set title: %s", fm.Title))
				break
			}
		}
		if fm.Title == "" {
			fm.Title = "Untitled"
			changes = append(changes, "set title: Untitled")
		}
	}

	// Fix missing or mismatched status
	if fm.Status == "" {
		defaultStatus := map[string]string{
			"backlog":  "backlog",
			"live":     "todo",
			"upstream": "completed",
			"expired":  "expired",
		}
		fm.Status = defaultStatus[fileType]
		changes = append(changes, fmt.Sprintf("set status: %s", fm.Status))
	} else {
		// Check if status matches location
		allowed, ok := validStatuses[fileType]
		if ok {
			found := false
			for _, s := range allowed {
				if fm.Status == s {
					found = true
					break
				}
			}
			if !found {
				oldStatus := fm.Status
				defaultStatus := map[string]string{
					"backlog":  "backlog",
					"live":     "todo",
					"upstream": "completed",
					"expired":  "expired",
				}
				fm.Status = defaultStatus[fileType]
				changes = append(changes, fmt.Sprintf("fix status: %s -> %s (matches %s/ location)", oldStatus, fm.Status, fileType))
			}
		}
	}

	// Fix missing tags
	if fm.Tags == nil || len(fm.Tags) == 0 {
		fm.Tags = []string{"untagged"}
		changes = append(changes, "set tags: [untagged]")
	}

	// Fix type-specific missing fields
	switch fileType {
	case "live":
		if fm.Complexity == "" {
			fm.Complexity = "S"
			changes = append(changes, "set complexity: S")
		}
		if fm.Branch == "" {
			fm.Branch = "unknown"
			changes = append(changes, "set branch: unknown")
		}
	case "upstream":
		if fm.Branch == "" {
			fm.Branch = "unknown"
			changes = append(changes, "set branch: unknown")
		}
	}

	// Fix missing created date
	if fm.Created == "" {
		fm.Created = time.Now().Format("2006-01-02")
		changes = append(changes, fmt.Sprintf("set created: %s", fm.Created))
	}

	if len(changes) == 0 {
		return nil, nil
	}

	if dryRun {
		return changes, nil
	}

	// Write the fixed file
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal frontmatter: %w", err)
	}

	newContent := "---\n" + string(yamlBytes) + "---\n" + body

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("cannot write file: %w", err)
	}

	return changes, nil
}
