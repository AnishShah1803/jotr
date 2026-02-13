package state

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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
	Text          string    `json:"text"`
	Section       string    `json:"section"`
	Priority      string    `json:"priority,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	ID            string    `json:"id"`
	Completed     bool      `json:"completed"`
	CreatedAt     time.Time `json:"createdAt"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
	LastModified  time.Time `json:"lastModified"`
	Source        string    `json:"source,omitempty"`
	CreatedDate   string    `json:"createdDate,omitempty"`
	CompletedDate string    `json:"completedDate,omitempty"`
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

// isDateSection checks if a string matches YYYY-MM-DD format
func isDateSection(s string) bool {
	if len(s) != 10 {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// AddTask adds or updates a task in the state
func (s *TodoState) AddTask(task tasks.Task, source string) {
	now := time.Now()
	today := now.Format("2006-01-02")

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
		ts.CreatedDate = existing.CreatedDate
		ts.CompletedDate = existing.CompletedDate
	} else {
		ts.CreatedAt = now
		if isDateSection(task.Section) {
			ts.CreatedDate = task.Section
		} else {
			ts.CreatedDate = today
		}
	}

	if task.Completed && ts.CompletedAt.IsZero() {
		ts.CompletedAt = now
		ts.CompletedDate = today
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

// RemoveTask removes a task from state by ID
func (s *TodoState) RemoveTask(taskID string) {
	delete(s.Tasks, taskID)
	s.LastSync = time.Now()
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

// ChangeType represents the type of change detected
type ChangeType int

const (
	NoChange ChangeType = iota
	Added
	Modified
	Deleted
)

// TaskChange represents a detected change for a task
type TaskChange struct {
	TaskID     string
	ChangeType ChangeType
	OldTask    *TaskState // nil for Added changes
	NewTask    *TaskState // nil for Deleted changes
	Source     string     // Where the change was detected (e.g., "daily-note", "todo-list")
}

// TaskChangeDetail represents detailed information about a task change for reporting
type TaskChangeDetail struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Change  string `json:"change"`            // "added", "updated", "deleted"
	From    string `json:"from,omitempty"`    // previous value (for updates)
	To      string `json:"to,omitempty"`      // new value (for updates)
	Details string `json:"details,omitempty"` // "marked complete", "priority changed to P1", etc.
}

// ConflictDetail represents detailed information about a conflict for reporting
type ConflictDetail struct {
	ID        string `json:"id"`
	TextDaily string `json:"text_daily"`
	TextTodo  string `json:"text_todo"`
	Reason    string `json:"reason"`
}

// CompareWithDailyNotes compares the state with tasks from daily notes
// Returns a list of changes detected
func (s *TodoState) CompareWithDailyNotes(dailyTasks []tasks.Task, source string) []TaskChange {
	changes := []TaskChange{}
	dailyTaskMap := make(map[string]tasks.Task)

	for _, task := range dailyTasks {
		if task.ID == "" {
			continue
		}
		dailyTaskMap[task.ID] = task
	}

	for id, stateTask := range s.Tasks {
		dailyTask, exists := dailyTaskMap[id]
		if !exists {
			continue
		}

		if s.isTaskModified(stateTask, dailyTask) {
			changes = append(changes, TaskChange{
				TaskID:     id,
				ChangeType: Modified,
				OldTask:    &stateTask,
				NewTask: &TaskState{
					Text:      dailyTask.Text,
					Section:   dailyTask.Section,
					Priority:  dailyTask.Priority,
					Tags:      dailyTask.Tags,
					ID:        dailyTask.ID,
					Completed: dailyTask.Completed,
					Source:    source,
				},
				Source: source,
			})
		}
	}

	for _, task := range dailyTasks {
		if task.ID == "" || !s.HasTask(task.ID) {
			changes = append(changes, TaskChange{
				TaskID:     task.ID,
				ChangeType: Added,
				NewTask: &TaskState{
					Text:      task.Text,
					Section:   task.Section,
					Priority:  task.Priority,
					Tags:      task.Tags,
					ID:        task.ID,
					Completed: task.Completed,
					Source:    source,
				},
				Source: source,
			})
		}
	}

	return changes
}

// CompareWithTodoList compares the state with tasks from the todo list
// Returns a list of changes detected
func (s *TodoState) CompareWithTodoList(todoTasks []tasks.Task) []TaskChange {
	changes := []TaskChange{}
	todoTaskMap := make(map[string]tasks.Task)

	for _, task := range todoTasks {
		if task.ID == "" {
			continue
		}
		todoTaskMap[task.ID] = task
	}

	for id, stateTask := range s.Tasks {
		todoTask, exists := todoTaskMap[id]
		if !exists {
			continue
		}

		if s.isTaskModified(stateTask, todoTask) {
			changes = append(changes, TaskChange{
				TaskID:     id,
				ChangeType: Modified,
				OldTask:    &stateTask,
				NewTask: &TaskState{
					Text:      todoTask.Text,
					Section:   todoTask.Section,
					Priority:  todoTask.Priority,
					Tags:      todoTask.Tags,
					ID:        todoTask.ID,
					Completed: todoTask.Completed,
				},
				Source: "todo-list",
			})
		}
	}

	return changes
}

// DetectDeletions finds tasks that exist in state but are missing from sources.
// Returns TaskChange entries for tasks that should be deleted.
// A task is considered deleted if it exists in state but is missing from BOTH sources.
func (s *TodoState) DetectDeletions(dailyTasks, todoTasks []tasks.Task) []TaskChange {
	deletions := []TaskChange{}

	dailyTaskMap := make(map[string]tasks.Task)
	for _, task := range dailyTasks {
		if task.ID != "" {
			dailyTaskMap[task.ID] = task
		}
	}

	todoTaskMap := make(map[string]tasks.Task)
	for _, task := range todoTasks {
		if task.ID != "" {
			todoTaskMap[task.ID] = task
		}
	}

	for id, stateTask := range s.Tasks {
		_, inDaily := dailyTaskMap[id]
		_, inTodo := todoTaskMap[id]

		if !inDaily && !inTodo {
			deletions = append(deletions, TaskChange{
				TaskID:     id,
				ChangeType: Deleted,
				OldTask:    &stateTask,
				Source:     "deletion-detected",
			})
		}
	}

	return deletions
}

func (s *TodoState) isTaskModified(stateTask TaskState, sourceTask tasks.Task) bool {
	if stateTask.Text != sourceTask.Text {
		return true
	}
	if stateTask.Priority != sourceTask.Priority {
		return true
	}
	if stateTask.Completed != sourceTask.Completed {
		return true
	}

	stateTags := make(map[string]bool)
	for _, tag := range stateTask.Tags {
		stateTags[tag] = true
	}
	sourceTags := make(map[string]bool)
	for _, tag := range sourceTask.Tags {
		sourceTags[tag] = true
	}

	if len(stateTags) != len(sourceTags) {
		return true
	}
	for tag := range stateTags {
		if !sourceTags[tag] {
			return true
		}
	}

	return false
}

// DetectConflicts checks if there are conflicting changes between daily notes and todo list
// Returns a map of task IDs to conflict descriptions
func (s *TodoState) DetectConflicts(dailyChanges, todoChanges []TaskChange) map[string]string {
	conflicts := make(map[string]string)

	dailyChangeMap := make(map[string]TaskChange)
	for _, change := range dailyChanges {
		dailyChangeMap[change.TaskID] = change
	}

	todoChangeMap := make(map[string]TaskChange)
	for _, change := range todoChanges {
		todoChangeMap[change.TaskID] = change
	}

	for id, dailyChange := range dailyChangeMap {
		if todoChange, exists := todoChangeMap[id]; exists {
			if dailyChange.ChangeType == Modified && todoChange.ChangeType == Modified {
				if dailyChange.NewTask != nil && todoChange.NewTask != nil {
					var conflictParts []string
					if dailyChange.NewTask.Text != todoChange.NewTask.Text {
						conflictParts = append(conflictParts, fmt.Sprintf("text differs (daily: '%s', todo: '%s')",
							dailyChange.NewTask.Text, todoChange.NewTask.Text))
					}
					if dailyChange.NewTask.Completed != todoChange.NewTask.Completed {
						conflictParts = append(conflictParts, fmt.Sprintf("completion differs (daily: %v, todo: %v)",
							dailyChange.NewTask.Completed, todoChange.NewTask.Completed))
					}
					if len(conflictParts) > 0 {
						conflicts[id] = strings.Join(conflictParts, "; ")
					}
				}
			}
		}
	}

	return conflicts
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	StateUpdated   bool
	DailyChanged   bool
	TodoChanged    bool
	Conflicts      map[string]string
	AppliedDaily   int
	AppliedTodo    int
	Deleted        int // Number of tasks deleted during sync
	Skipped        int
	ChangedTaskIDs []string // Task IDs that changed and may need their source files updated
	DeletedTaskIDs []string // Task IDs that were deleted
}

// BidirectionalSync performs bidirectional sync between daily notes and todo list
// Compares both sources with state and propagates changes appropriately
func (s *TodoState) BidirectionalSync(dailyTasks, todoTasks []tasks.Task, dailySourcePath string) SyncResult {
	result := SyncResult{
		Conflicts: make(map[string]string),
	}

	dailyChanges := s.CompareWithDailyNotes(dailyTasks, dailySourcePath)
	todoChanges := s.CompareWithTodoList(todoTasks)

	conflicts := s.DetectConflicts(dailyChanges, todoChanges)
	if len(conflicts) > 0 {
		result.Conflicts = conflicts
		return result
	}

	dailyChangeMap := make(map[string]TaskChange)
	for _, change := range dailyChanges {
		dailyChangeMap[change.TaskID] = change
	}

	todoChangeMap := make(map[string]TaskChange)
	for _, change := range todoChanges {
		todoChangeMap[change.TaskID] = change
	}

	for taskID, dailyChange := range dailyChangeMap {
		todoChange, todoHasChange := todoChangeMap[taskID]

		if !todoHasChange {
			s.applyChange(dailyChange)
			result.AppliedDaily++
			result.StateUpdated = true
			result.TodoChanged = true
			result.ChangedTaskIDs = append(result.ChangedTaskIDs, taskID)
		} else if dailyChange.ChangeType == Modified && todoChange.ChangeType == Modified {
			merged := s.smartMerge(dailyChange, todoChange)
			if merged != nil {
				s.applyChange(TaskChange{
					TaskID:     taskID,
					ChangeType: Modified,
					NewTask:    merged,
					Source:     "merged",
				})
				result.AppliedDaily++
				result.AppliedTodo++
				result.StateUpdated = true
				result.DailyChanged = true
				result.TodoChanged = true
				result.ChangedTaskIDs = append(result.ChangedTaskIDs, taskID)
			} else {
				result.Skipped++
			}
		}
	}

	for taskID, todoChange := range todoChangeMap {
		if _, dailyHasChange := dailyChangeMap[taskID]; !dailyHasChange {
			s.applyChange(todoChange)
			result.AppliedTodo++
			result.StateUpdated = true
			result.DailyChanged = true
			result.ChangedTaskIDs = append(result.ChangedTaskIDs, taskID)
		}
	}

	deletions := s.DetectDeletions(dailyTasks, todoTasks)
	for _, deletion := range deletions {
		if deletion.OldTask != nil && deletion.OldTask.Completed {
			continue
		}
		s.RemoveTask(deletion.TaskID)
		result.Deleted++
		result.StateUpdated = true
		result.DeletedTaskIDs = append(result.DeletedTaskIDs, deletion.TaskID)
	}

	return result
}

func (s *TodoState) applyChange(change TaskChange) {
	if change.NewTask == nil {
		return
	}

	now := time.Now()
	task := *change.NewTask

	// Preserve fields from existing task
	if existing, exists := s.Tasks[change.TaskID]; exists {
		task.CreatedAt = existing.CreatedAt
		task.CompletedAt = existing.CompletedAt
		task.CreatedDate = existing.CreatedDate
		task.CompletedDate = existing.CompletedDate
	}

	// Set CompletedDate if task transitioned to complete
	if task.Completed && (change.OldTask == nil || !change.OldTask.Completed) {
		if task.CompletedDate == "" {
			task.CompletedDate = now.Format("2006-01-02")
		}
	}

	task.LastModified = now
	s.Tasks[change.TaskID] = task
	s.LastSync = now
}

// smartMerge attempts to merge non-conflicting changes from both sources
// Returns nil if merge is not possible
func (s *TodoState) smartMerge(dailyChange, todoChange TaskChange) *TaskState {
	if dailyChange.NewTask == nil || todoChange.NewTask == nil {
		return nil
	}

	merged := TaskState{
		ID:        dailyChange.NewTask.ID,
		Text:      dailyChange.NewTask.Text,
		Section:   dailyChange.NewTask.Section,
		Priority:  dailyChange.NewTask.Priority,
		Tags:      dailyChange.NewTask.Tags,
		Completed: dailyChange.NewTask.Completed,
		Source:    "merged",
	}

	var wasCompleted bool
	var createdDate string
	if dailyChange.OldTask != nil {
		merged.CreatedAt = dailyChange.OldTask.CreatedAt
		merged.CompletedAt = dailyChange.OldTask.CompletedAt
		wasCompleted = dailyChange.OldTask.Completed
		createdDate = dailyChange.OldTask.CreatedDate
	} else if todoChange.OldTask != nil {
		merged.CreatedAt = todoChange.OldTask.CreatedAt
		merged.CompletedAt = todoChange.OldTask.CompletedAt
		wasCompleted = todoChange.OldTask.Completed
		createdDate = todoChange.OldTask.CreatedDate
	}
	merged.CreatedDate = createdDate
	if merged.Completed && !wasCompleted {
		merged.CompletedDate = time.Now().Format("2006-01-02")
	} else if wasCompleted {
		// Preserve existing CompletedDate if task was already completed
		if dailyChange.OldTask != nil {
			merged.CompletedDate = dailyChange.OldTask.CompletedDate
		} else if todoChange.OldTask != nil {
			merged.CompletedDate = todoChange.OldTask.CompletedDate
		}
	}

	merged.LastModified = time.Now()

	// Merge tags from both sources
	mergedTags := make(map[string]bool)
	for _, tag := range dailyChange.NewTask.Tags {
		mergedTags[tag] = true
	}
	for _, tag := range todoChange.NewTask.Tags {
		mergedTags[tag] = true
	}
	merged.Tags = make([]string, 0, len(mergedTags))
	for tag := range mergedTags {
		merged.Tags = append(merged.Tags, tag)
	}
	sort.Strings(merged.Tags)

	// Priority: daily note takes precedence, fallback to todo if daily has none
	merged.Priority = dailyChange.NewTask.Priority
	if merged.Priority == "" {
		merged.Priority = todoChange.NewTask.Priority
	}

	// If both modified to the same state, just use that state (no real conflict)
	if dailyChange.NewTask.Text == todoChange.NewTask.Text &&
		dailyChange.NewTask.Completed == todoChange.NewTask.Completed {
		return &merged
	}

	// Real conflict - return nil to indicate merge not possible
	return nil
}
