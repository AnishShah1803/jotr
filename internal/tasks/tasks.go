package tasks

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Task represents a task item
type Task struct {
	Text      string
	Completed bool
	Priority  string
	Tags      []string
	Line      int
	Section   string
}

// ParseTasks parses tasks from markdown content
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

		// Parse task lines
		if strings.HasPrefix(strings.TrimSpace(line), "- [") {
			task := Task{
				Line:    i + 1,
				Section: currentSection,
			}

			// Check if completed
			if strings.Contains(line, "- [x]") || strings.Contains(line, "- [X]") {
				task.Completed = true
				task.Text = strings.TrimSpace(strings.TrimPrefix(line, "- [x]"))
				task.Text = strings.TrimSpace(strings.TrimPrefix(task.Text, "- [X]"))
			} else {
				task.Completed = false
				task.Text = strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
			}

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

			tasks = append(tasks, task)
		}
	}

	return tasks
}

// ReadTasks reads tasks from a file
func ReadTasks(path string) ([]Task, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseTasks(string(content)), nil
}

// FilterTasks filters tasks by various criteria
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

// CountTasks counts tasks by status
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

// GroupByPriority groups tasks by priority
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

// GroupBySection groups tasks by section
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

// FormatTask formats a task for display
func FormatTask(task Task) string {
	checkbox := "○"
	if task.Completed {
		checkbox = "✓"
	}

	priority := ""
	if task.Priority != "" {
		priority = fmt.Sprintf("[%s] ", task.Priority)
	}

	return fmt.Sprintf("%s %s%s", checkbox, priority, task.Text)
}

// IsOverdue checks if a task is overdue based on due date in text
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

