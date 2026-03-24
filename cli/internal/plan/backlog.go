package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// BacklogItem represents a single backlog idea.
type BacklogItem struct {
	ID       string   `json:"id" yaml:"id"`
	Title    string   `json:"title" yaml:"title"`
	Type     string   `json:"type" yaml:"type"`
	Status   string   `json:"status" yaml:"status"`
	Priority string   `json:"priority" yaml:"priority"`
	Size     string   `json:"size" yaml:"size"`
	Project  string   `json:"project,omitempty" yaml:"project"`
	Tags     []string `json:"tags,omitempty" yaml:"tags"`
	Linear   *string  `json:"linear,omitempty" yaml:"linear"`
	Created  string   `json:"created,omitempty" yaml:"created"`
	Plan     *string  `json:"plan,omitempty" yaml:"plan"`
	Depends  []string `json:"depends,omitempty" yaml:"depends"`
	File     string   `json:"file" yaml:"-"`
}

// Regex for parsing old blockquote format.
var (
	bqStatus   = regexp.MustCompile(`(?i)>\s*\*\*Status:\*\*\s*(.+)`)
	bqPriority = regexp.MustCompile(`(?i)>\s*\*\*Priority:\*\*\s*(.+)`)
	bqSize     = regexp.MustCompile(`(?i)>\s*\*\*Size:\*\*\s*(.+)`)
	bqCreated  = regexp.MustCompile(`(?i)>\s*\*\*Created:\*\*\s*(.+)`)
	bqPlan     = regexp.MustCompile(`(?i)>\s*\*\*Plan:\*\*\s*(.+)`)
	bqTitle    = regexp.MustCompile(`^#\s+(.+)`)
	bqIDFromFn = regexp.MustCompile(`^(\d{4})-`)
	bqType     = regexp.MustCompile(`^(\d{4})-([a-z]+)-`)
)

// ParseBacklogFile parses a single backlog markdown file (both YAML and blockquote formats).
func ParseBacklogFile(path string) (*BacklogItem, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(content)
	filename := filepath.Base(path)

	// Try YAML frontmatter first
	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 3 && strings.TrimSpace(parts[1]) != "" {
			item := &BacklogItem{}
			if err := yaml.Unmarshal([]byte(parts[1]), item); err == nil && item.ID != "" {
				item.File = filename
				// Ensure type from filename if missing
				if item.Type == "" {
					if m := bqType.FindStringSubmatch(filename); m != nil {
						item.Type = m[2]
					}
				}
				return item, nil
			}
		}
	}

	// Fallback: blockquote format
	item := &BacklogItem{File: filename}

	// ID from filename
	if m := bqIDFromFn.FindStringSubmatch(filename); m != nil {
		item.ID = m[1]
	}
	// Type from filename
	if m := bqType.FindStringSubmatch(filename); m != nil {
		item.Type = m[2]
	}

	// Title from first H1
	if m := bqTitle.FindStringSubmatch(text); m != nil {
		item.Title = strings.TrimSpace(m[1])
	} else {
		// Derive from filename
		slug := strings.TrimSuffix(filename, ".md")
		slug = regexp.MustCompile(`^\d{4}-[a-z]+-`).ReplaceAllString(slug, "")
		item.Title = strings.ReplaceAll(slug, "-", " ")
	}

	if m := bqStatus.FindStringSubmatch(text); m != nil {
		item.Status = strings.TrimSpace(m[1])
	} else {
		item.Status = "new"
	}
	if m := bqPriority.FindStringSubmatch(text); m != nil {
		item.Priority = strings.TrimSpace(m[1])
	}
	if m := bqSize.FindStringSubmatch(text); m != nil {
		item.Size = strings.TrimSpace(m[1])
	}
	if m := bqCreated.FindStringSubmatch(text); m != nil {
		item.Created = strings.TrimSpace(m[1])
	}
	if m := bqPlan.FindStringSubmatch(text); m != nil {
		val := strings.TrimSpace(m[1])
		if !strings.Contains(val, "not yet") && val != "" {
			item.Plan = &val
		}
	}

	return item, nil
}

// ScanBacklogDir scans a backlog directory and returns all items.
func ScanBacklogDir(dir string) ([]*BacklogItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var items []*BacklogItem
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		item, err := ParseBacklogFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	// Sort by ID
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	return items, nil
}

// BacklogResult holds the combined active + optional archive results.
type BacklogResult struct {
	Active   []*BacklogItem `json:"active"`
	Archived []*BacklogItem `json:"archived,omitempty"`
	Summary  BacklogSummary `json:"summary"`
}

// BacklogSummary has counts.
type BacklogSummary struct {
	ActiveCount   int `json:"active_count"`
	ArchivedCount int `json:"archived_count"`
}

// ScanBacklog scans blueprint/backlog/ and optionally archive/.
func ScanBacklog(planningDir string, includeArchive bool) (*BacklogResult, error) {
	backlogDir := filepath.Join(planningDir, "backlog")
	archiveDir := filepath.Join(backlogDir, "archive")

	active, err := ScanBacklogDir(backlogDir)
	if err != nil {
		return nil, fmt.Errorf("scanning backlog: %w", err)
	}

	result := &BacklogResult{
		Active: active,
	}

	if includeArchive {
		archived, err := ScanBacklogDir(archiveDir)
		if err == nil {
			result.Archived = archived
		}
	}

	result.Summary.ActiveCount = len(result.Active)
	if result.Archived != nil {
		result.Summary.ArchivedCount = len(result.Archived)
	}

	return result, nil
}

// JSON returns the result as indented JSON.
func (r *BacklogResult) JSON() string {
	b, _ := json.MarshalIndent(r, "", "  ")
	return string(b)
}

// Table returns a formatted table of backlog items.
func (r *BacklogResult) Table() string {
	var sb strings.Builder

	if len(r.Active) > 0 {
		sb.WriteString(formatBacklogTable("Active Backlog", r.Active))
	} else {
		sb.WriteString("No active backlog items.\n")
	}

	if len(r.Archived) > 0 {
		sb.WriteString("\n")
		sb.WriteString(formatBacklogTable("Archived", r.Archived))
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d active", r.Summary.ActiveCount))
	if r.Summary.ArchivedCount > 0 {
		sb.WriteString(fmt.Sprintf(", %d archived", r.Summary.ArchivedCount))
	}
	sb.WriteString("\n")

	return sb.String()
}

// MigrateResult holds result of a migration run.
type MigrateResult struct {
	Migrated []string
	Skipped  []string
}

// IsOldFormat checks if a file uses blockquote format (no YAML frontmatter).
func IsOldFormat(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(content)
	return !strings.HasPrefix(text, "---") && bqStatus.MatchString(text)
}

// MigrateBacklogFile converts a single old-format file to YAML frontmatter.
func MigrateBacklogFile(path, project string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	text := string(content)

	// Parse using existing blockquote parser
	item, err := ParseBacklogFile(path)
	if err != nil {
		return err
	}

	// Build YAML frontmatter
	fm := map[string]interface{}{
		"id":       item.ID,
		"title":    item.Title,
		"type":     item.Type,
		"status":   strings.ToLower(item.Status),
		"priority": strings.ToLower(item.Priority),
		"size":     strings.ToLower(item.Size),
		"project":  project,
		"tags":     []string{},
		"linear":   nil,
		"created":  item.Created,
		"plan":     nil,
		"depends":  nil,
	}

	if item.Plan != nil {
		fm["plan"] = *item.Plan
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("YAML marshal error: %w", err)
	}

	// Find the body: everything from the first H1 heading onwards
	// Strip blockquote metadata lines
	lines := strings.Split(text, "\n")
	var bodyLines []string
	inBlockquote := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inBlockquote {
			// Skip blockquote metadata lines
			if strings.HasPrefix(trimmed, ">") || trimmed == "" {
				continue
			}
			inBlockquote = false
		}
		bodyLines = append(bodyLines, line)
	}

	body := strings.Join(bodyLines, "\n")
	// Ensure body starts with newline
	if !strings.HasPrefix(body, "\n") {
		body = "\n" + body
	}

	output := "---\n" + string(yamlBytes) + "---\n" + body

	return os.WriteFile(path, []byte(output), 0644)
}

// MigrateBacklogDir migrates all old-format files in a directory.
func MigrateBacklogDir(dir, project string) (*MigrateResult, error) {
	result := &MigrateResult{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if IsOldFormat(path) {
			if err := MigrateBacklogFile(path, project); err != nil {
				result.Skipped = append(result.Skipped, fmt.Sprintf("%s: %v", e.Name(), err))
			} else {
				result.Migrated = append(result.Migrated, e.Name())
			}
		} else {
			result.Skipped = append(result.Skipped, e.Name()+" (already YAML)")
		}
	}

	return result, nil
}

func formatBacklogTable(header string, items []*BacklogItem) string {
	var sb strings.Builder

	// Calculate column widths
	idW, typeW, titleW, prioW, sizeW, statusW := 4, 4, 5, 8, 4, 6
	for _, item := range items {
		if len(item.ID) > idW {
			idW = len(item.ID)
		}
		if len(item.Type) > typeW {
			typeW = len(item.Type)
		}
		t := item.Title
		if len(t) > 50 {
			t = t[:47] + "..."
		}
		if len(t) > titleW {
			titleW = len(t)
		}
		if len(item.Priority) > prioW {
			prioW = len(item.Priority)
		}
		if len(item.Size) > sizeW {
			sizeW = len(item.Size)
		}
		if len(item.Status) > statusW {
			statusW = len(item.Status)
		}
	}

	// Header
	sb.WriteString(fmt.Sprintf("── %s (%d items) ", header, len(items)))
	sb.WriteString(strings.Repeat("─", 40))
	sb.WriteString("\n")

	// Column headers
	sb.WriteString(fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s",
		idW, "ID", typeW, "Type", titleW, "Title", prioW, "Priority", sizeW, "Size", statusW, "Status"))

	// Plan column for items that have it
	sb.WriteString("  Plan")
	sb.WriteString("\n")

	// Separator
	sb.WriteString(fmt.Sprintf("  %s  %s  %s  %s  %s  %s  %s\n",
		strings.Repeat("─", idW),
		strings.Repeat("─", typeW),
		strings.Repeat("─", titleW),
		strings.Repeat("─", prioW),
		strings.Repeat("─", sizeW),
		strings.Repeat("─", statusW),
		strings.Repeat("─", 6),
	))

	// Rows
	for _, item := range items {
		title := item.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		planStr := "—"
		if item.Plan != nil && *item.Plan != "" {
			planStr = *item.Plan
		}
		sb.WriteString(fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s  %-*s  %s\n",
			idW, item.ID,
			typeW, item.Type,
			titleW, title,
			prioW, item.Priority,
			sizeW, item.Size,
			statusW, item.Status,
			planStr,
		))
	}

	return sb.String()
}
