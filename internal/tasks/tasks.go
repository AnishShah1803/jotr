package tasks

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Task represents a task item.
type Task struct {
	Text          string
	Priority      string
	Section       string
	ID            string
	Tags          []string
	Line          int
	Completed     bool
	CompletedDate string // Extracted from @completed(YYYY-MM-DD) tag
}

// taskFormatRegex matches common markdown task list formats:
// - [ ] and - [x] (dash)
// * [ ] and * [x] (asterisk)
// + [ ] and + [x] (plus)
var taskFormatRegex = regexp.MustCompile(`^(\*|-|\+)\s*\[([ xX])\]\s*(.*)$`)

// ParseTasks parses tasks from markdown content.
func ParseTasks(content string) []Task {
	var tasks []Task

	lines := strings.Split(content, "\n")
	currentSection := ""

	for i, line := range lines {
		// Track sections
		if strings.HasPrefix(line, "## ") {
			currentSection = strings.TrimPrefix(line, "## ")
			continue
		}

		// Parse task lines - supports all common markdown formats:
		// - [ ] / - [x] (dash), * [ ] / * [x] (asterisk), + [ ] / + [x] (plus)
		trimmedLine := strings.TrimSpace(line)

		// Must match valid task format: bullet + space + [space/x/X] + optional text
		match := taskFormatRegex.FindStringSubmatch(trimmedLine)
		if len(match) == 0 {
			continue
		}

		task := Task{
			Line:    i + 1,
			Section: currentSection,
		}

		// Parse completed status and text from regex match
		checkbox := match[2]
		task.Completed = checkbox == "x" || checkbox == "X"
		task.Text = strings.TrimSpace(match[3])

		// Extract priority
		priorityRe := regexp.MustCompile(`\[P([0-3])\]`)
		if match := priorityRe.FindStringSubmatch(task.Text); len(match) > 1 {
			task.Priority = "P" + match[1]
		}

		// Extract tags
		tagRe := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)

		matches := tagRe.FindAllStringSubmatch(task.Text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				task.Tags = append(task.Tags, match[1])
			}
		}

		// Extract task ID
		task.ID = ExtractTaskID(task.Text)
		// Strip ID from text for clean display
		task.Text = StripTaskID(task.Text)

		// Extract completed date from @completed(date) tag
		task.CompletedDate = ExtractCompletedDate(task.Text)
		// Strip completed tag from text for clean display
		task.Text = StripCompletedTag(task.Text)

		tasks = append(tasks, task)
	}

	return tasks
}

// ReadTasks reads tasks from a file with context support.
func ReadTasks(ctx context.Context, path string) ([]Task, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseTasks(string(content)), nil
}

// FilterTasks filters tasks by various criteria.
func FilterTasks(tasks []Task, completed *bool, priority string, section string) []Task {
	var filtered []Task

	for _, task := range tasks {
		// Filter by completion status
		if completed != nil && task.Completed != *completed {
			continue
		}

		// Filter by priority
		if priority != "" && task.Priority != priority {
			continue
		}

		// Filter by section
		if section != "" && task.Section != section {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// CountTasks counts tasks by status.
func CountTasks(tasks []Task) (total, completed, pending int) {
	total = len(tasks)

	for _, task := range tasks {
		if task.Completed {
			completed++
		} else {
			pending++
		}
	}

	return
}

// GroupByPriority groups tasks by priority.
func GroupByPriority(tasks []Task) map[string][]Task {
	groups := make(map[string][]Task)

	for _, task := range tasks {
		priority := task.Priority
		if priority == "" {
			priority = "None"
		}

		groups[priority] = append(groups[priority], task)
	}

	return groups
}

// GroupBySection groups tasks by section.
func GroupBySection(tasks []Task) map[string][]Task {
	groups := make(map[string][]Task)

	for _, task := range tasks {
		section := task.Section
		if section == "" {
			section = "Uncategorized"
		}

		groups[section] = append(groups[section], task)
	}

	return groups
}

// FormatTask formats a task for display.
func FormatTask(task Task) string {
	checkbox := "○"
	if task.Completed {
		checkbox = "✓"
	}

	priority := ""
	if task.Priority != "" {
		priority = fmt.Sprintf("%s ", task.Priority)
	}

	return fmt.Sprintf("%s  %s%s", checkbox, priority, task.Text)
}

// IsOverdue checks if a task is overdue based on due date in text.
func IsOverdue(task Task) bool {
	// Simple check for "due:" pattern
	dueRe := regexp.MustCompile(`due:\s*(\d{4}-\d{2}-\d{2})`)
	if match := dueRe.FindStringSubmatch(task.Text); len(match) > 1 {
		dueDate, err := time.Parse("2006-01-02", match[1])
		if err == nil && dueDate.Before(time.Now()) && !task.Completed {
			return true
		}
	}

	return false
}

// GenerateTaskID generates a unique task ID based on content.
func GenerateTaskID(text string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return fmt.Sprintf("%x", hash)[:8]
}

// ExtractTaskID extracts task ID from task text.
func ExtractTaskID(text string) string {
	// Look for <!-- id: abc12345 --> pattern
	idRe := regexp.MustCompile(`<!-- id: ([a-f0-9]{8}) -->`)
	if match := idRe.FindStringSubmatch(text); len(match) > 1 {
		return match[1]
	}

	return ""
}

// EnsureTaskID ensures a task has an ID, generating one if needed.
func EnsureTaskID(task *Task) {
	if task.ID == "" {
		// Check if ID is embedded in text
		if id := ExtractTaskID(task.Text); id != "" {
			task.ID = id
		} else {
			// Generate new ID and embed in text
			task.ID = GenerateTaskID(task.Text)
			task.Text = task.Text + fmt.Sprintf(" <!-- id: %s -->", task.ID)
		}
	}
}

// StripTaskID removes task ID from task text for display.
func StripTaskID(text string) string {
	idRe := regexp.MustCompile(`\s*<!-- id: [a-f0-9]{8} -->`)
	return idRe.ReplaceAllString(text, "")
}

// StripCompletedTag removes @completed(YYYY-MM-DD) tag from task text for clean display.
func StripCompletedTag(text string) string {
	completedRe := regexp.MustCompile(`\s*@completed\(\d{4}-\d{2}-\d{2}\)`)
	return completedRe.ReplaceAllString(text, "")
}

func ExtractCompletedDate(text string) string {
	completedRe := regexp.MustCompile(`@completed\((\d{4}-\d{2}-\d{2})\)`)
	if match := completedRe.FindStringSubmatch(text); len(match) > 1 {
		return match[1]
	}
	return ""
}
