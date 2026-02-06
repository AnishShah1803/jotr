package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AnishShah1803/jotr/internal/tasks"
)

// TodoState represents the complete state of a todo list
type TodoState struct {
	LastSync    time.Time            `json:"lastSync"`
	LastArchive time.Time            `json:"lastArchive,omitempty"`
	Tasks       map[string]TaskState `json:"tasks"`
	Version     int                  `json:"version"`
}

// TaskState represents the state of a single task
type TaskState struct {
	Text         string    `json:"text"`
	Section      string    `json:"section"`
	Priority     string    `json:"priority,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	ID           string    `json:"id"`
	Completed    bool      `json:"completed"`
	CreatedAt    time.Time `json:"createdAt"`
	CompletedAt  time.Time `json:"completedAt,omitempty"`
	LastModified time.Time `json:"lastModified"`
	Source       string    `json:"source,omitempty"`
}

// NewTodoState creates a new empty TodoState
func NewTodoState() *TodoState {
	return &TodoState{
		Tasks:    make(map[string]TaskState),
		Version:  1,
		LastSync: time.Now(),
	}
}

// Read reads the state from a file
func Read(statePath string) (*TodoState, error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewTodoState(), nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state TodoState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Tasks == nil {
		state.Tasks = make(map[string]TaskState)
	}

	return &state, nil
}

// Write writes the state to a file
func (s *TodoState) Write(statePath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// AddTask adds or updates a task in the state
func (s *TodoState) AddTask(task tasks.Task, source string) {
	now := time.Now()

	ts := TaskState{
		Text:         task.Text,
		Section:      task.Section,
		Priority:     task.Priority,
		Tags:         task.Tags,
		ID:           task.ID,
		Completed:    task.Completed,
		LastModified: now,
		Source:       source,
	}

	if existing, ok := s.Tasks[task.ID]; ok {
		ts.CreatedAt = existing.CreatedAt
		ts.CompletedAt = existing.CompletedAt
	} else {
		ts.CreatedAt = now
	}

	if task.Completed && !ts.CompletedAt.IsZero() {
		ts.CompletedAt = now
	}

	s.Tasks[task.ID] = ts
	s.LastSync = now
}

// GetActiveTasks returns all non-completed tasks
func (s *TodoState) GetActiveTasks() []TaskState {
	var active []TaskState
	for _, task := range s.Tasks {
		if !task.Completed {
			active = append(active, task)
		}
	}
	return active
}

// GetCompletedTasks returns all completed tasks
func (s *TodoState) GetCompletedTasks() []TaskState {
	var completed []TaskState
	for _, task := range s.Tasks {
		if task.Completed {
			completed = append(completed, task)
		}
	}
	return completed
}

// HasTask checks if a task with the given ID exists
func (s *TodoState) HasTask(taskID string) bool {
	_, exists := s.Tasks[taskID]
	return exists
}

// MarkArchived records that tasks have been archived
func (s *TodoState) MarkArchived() {
	s.LastArchive = time.Now()
}

// ToTasks converts state tasks to tasks.Task slice
func (s *TodoState) ToTasks() []tasks.Task {
	var result []tasks.Task
	for _, ts := range s.Tasks {
		result = append(result, tasks.Task{
			Text:      ts.Text,
			Priority:  ts.Priority,
			Section:   ts.Section,
			ID:        ts.ID,
			Tags:      ts.Tags,
			Completed: ts.Completed,
		})
	}
	return result
}

// NeedsMigration checks if state needs to be migrated from markdown
func (s *TodoState) NeedsMigration() bool {
	return len(s.Tasks) == 0
}

// MigrateFromMarkdown populates state from existing markdown tasks
func (s *TodoState) MigrateFromMarkdown(tasksList []tasks.Task, source string) int {
	migrated := 0
	for _, task := range tasksList {
		tasks.EnsureTaskID(&task)
		s.AddTask(task, source)
		migrated++
	}
	return migrated
}
