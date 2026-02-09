package state

import (
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/tasks"
)

func TestCompareWithDailyNotes(t *testing.T) {
	tests := []struct {
		name        string
		state       *TodoState
		dailyTasks  []tasks.Task
		expected    int
		changeTypes []ChangeType
	}{
		{
			name: "no changes - all tasks match state",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {
						ID:     "abc123",
						Text:   "Test task",
						Source: "test.md",
					},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Test task", Section: "Tasks"},
			},
			expected: 0,
		},
		{
			name: "task without ID treated as new addition",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {
						ID:     "abc123",
						Text:   "Test task",
						Source: "test.md",
					},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "", Text: "Test task", Section: "Tasks"},
			},
			expected:    1,
			changeTypes: []ChangeType{Added},
		},
		{
			name: "new task added to daily note",
			state: &TodoState{
				Tasks: map[string]TaskState{},
			},
			dailyTasks: []tasks.Task{
				{ID: "new123", Text: "New task", Section: "Tasks"},
			},
			expected:    1,
			changeTypes: []ChangeType{Added},
		},
		{
			name: "task modified in daily note",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {
						ID:     "abc123",
						Text:   "Old text",
						Source: "test.md",
					},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Updated text", Section: "Tasks"},
			},
			expected:    1,
			changeTypes: []ChangeType{Modified},
		},
		{
			name: "multiple changes",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {
						ID:     "abc123",
						Text:   "Unchanged task",
						Source: "test.md",
					},
					"def456": {
						ID:     "def456",
						Text:   "Old text",
						Source: "test.md",
					},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Unchanged task", Section: "Tasks"},
				{ID: "def456", Text: "Modified text", Section: "Tasks"},
				{ID: "new789", Text: "New task", Section: "Tasks"},
			},
			expected:    2,
			changeTypes: []ChangeType{Modified, Added},
		},
		{
			name: "task completion status changed",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {
						ID:        "abc123",
						Text:      "Complete me",
						Completed: false,
						Source:    "test.md",
					},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Complete me", Completed: true, Section: "Tasks"},
			},
			expected:    1,
			changeTypes: []ChangeType{Modified},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := tt.state.CompareWithDailyNotes(tt.dailyTasks, "test.md")
			if len(changes) != tt.expected {
				t.Errorf("Expected %d changes, got %d", tt.expected, len(changes))
			}

			if len(tt.changeTypes) > 0 {
				for i, change := range changes {
					if i < len(tt.changeTypes) && change.ChangeType != tt.changeTypes[i] {
						t.Errorf("Expected change type %v, got %v", tt.changeTypes[i], change.ChangeType)
					}
				}
			}
		})
	}
}

func TestDetectConflicts(t *testing.T) {
	tests := []struct {
		name          string
		state         *TodoState
		dailyTasks    []tasks.Task
		todoTasks     []tasks.Task
		wantConflicts int
	}{
		{
			name: "no conflicts - different tasks modified",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Original", Source: "test.md"},
					"def456": {ID: "def456", Text: "Original 2", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Daily modified"},
			},
			todoTasks: []tasks.Task{
				{ID: "def456", Text: "Todo modified"},
			},
			wantConflicts: 0,
		},
		{
			name: "conflict - same task text modified differently",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Original", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Daily version"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Todo version"},
			},
			wantConflicts: 1,
		},
		{
			name: "no conflict - only todo list modified completion status",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Task", Completed: false, Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Task", Completed: false},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Task", Completed: true},
			},
			wantConflicts: 0,
		},
		{
			name: "no conflict - task only modified in one place",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Original", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Modified in daily"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Original"},
			},
			wantConflicts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dailyChanges := tt.state.CompareWithDailyNotes(tt.dailyTasks, "test.md")
			todoChanges := tt.state.CompareWithTodoList(tt.todoTasks)
			conflicts := tt.state.DetectConflicts(dailyChanges, todoChanges)

			if len(conflicts) != tt.wantConflicts {
				t.Errorf("Expected %d conflicts, got %d: %v", tt.wantConflicts, len(conflicts), conflicts)
			}
		})
	}
}

func TestBidirectionalSync(t *testing.T) {
	tests := []struct {
		name           string
		initialState   *TodoState
		dailyTasks     []tasks.Task
		todoTasks      []tasks.Task
		expectConflict bool
		stateUpdated   bool
		dailyChanged   bool
		todoChanged    bool
	}{
		{
			name: "sync from daily note only",
			initialState: &TodoState{
				Tasks: map[string]TaskState{},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "New task", Section: "Tasks"},
			},
			todoTasks:      []tasks.Task{},
			expectConflict: false,
			stateUpdated:   true,
			dailyChanged:   false,
			todoChanged:    true,
		},
		{
			name: "sync from todo list only",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Original", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Original"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Modified in todo"},
			},
			expectConflict: false,
			stateUpdated:   true,
			dailyChanged:   true,
			todoChanged:    false,
		},
		{
			name: "no changes - all in sync",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Task", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Task"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Task"},
			},
			expectConflict: false,
			stateUpdated:   false,
			dailyChanged:   false,
			todoChanged:    false,
		},
		{
			name: "conflict detected",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Original", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Daily version"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Todo version"},
			},
			expectConflict: true,
			stateUpdated:   false,
			dailyChanged:   false,
			todoChanged:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initialState.BidirectionalSync(tt.dailyTasks, tt.todoTasks, "test.md")

			if len(result.Conflicts) > 0 && !tt.expectConflict {
				t.Errorf("Expected no conflicts but got: %v", result.Conflicts)
			}

			if result.StateUpdated != tt.stateUpdated {
				t.Errorf("Expected StateUpdated=%v, got %v", tt.stateUpdated, result.StateUpdated)
			}

			if result.DailyChanged != tt.dailyChanged {
				t.Errorf("Expected DailyChanged=%v, got %v", tt.dailyChanged, result.DailyChanged)
			}

			if result.TodoChanged != tt.todoChanged {
				t.Errorf("Expected TodoChanged=%v, got %v", tt.todoChanged, result.TodoChanged)
			}
		})
	}
}

func TestSmartMerge(t *testing.T) {
	state := NewTodoState()
	now := time.Now()

	tests := []struct {
		name         string
		dailyChange  TaskChange
		todoChange   TaskChange
		expectMerge  bool
		expectedText string
	}{
		{
			name: "merge compatible changes - different tags",
			dailyChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task", Tags: []string{}, CreatedAt: now},
				NewTask:    &TaskState{ID: "abc123", Text: "Task", Tags: []string{"work"}},
			},
			todoChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task", Tags: []string{}, CreatedAt: now},
				NewTask:    &TaskState{ID: "abc123", Text: "Task", Tags: []string{"urgent"}},
			},
			expectMerge:  true,
			expectedText: "Task",
		},
		{
			name: "no merge - conflicting text",
			dailyChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task"},
				NewTask:    &TaskState{ID: "abc123", Text: "Daily version"},
			},
			todoChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task"},
				NewTask:    &TaskState{ID: "abc123", Text: "Todo version"},
			},
			expectMerge: false,
		},
		{
			name: "no merge - conflicting completion",
			dailyChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				NewTask:    &TaskState{ID: "abc123", Text: "Task", Completed: false},
			},
			todoChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				NewTask:    &TaskState{ID: "abc123", Text: "Task", Completed: true},
			},
			expectMerge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := state.smartMerge(tt.dailyChange, tt.todoChange)

			if tt.expectMerge && merged == nil {
				t.Error("Expected merge to succeed but got nil")
			}

			if !tt.expectMerge && merged != nil {
				t.Errorf("Expected merge to fail but got: %+v", merged)
			}

			if tt.expectMerge && merged != nil {
				if merged.Text != tt.expectedText {
					t.Errorf("Expected text '%s', got '%s'", tt.expectedText, merged.Text)
				}
			}
		})
	}
}

func TestIsTaskModified(t *testing.T) {
	s := NewTodoState()

	tests := []struct {
		name       string
		stateTask  TaskState
		sourceTask tasks.Task
		expectMod  bool
	}{
		{
			name:       "no modification - identical tasks",
			stateTask:  TaskState{Text: "Task", Priority: "P1", Completed: false, Tags: []string{"tag1"}},
			sourceTask: tasks.Task{Text: "Task", Priority: "P1", Completed: false, Tags: []string{"tag1"}},
			expectMod:  false,
		},
		{
			name:       "text modified",
			stateTask:  TaskState{Text: "Old text"},
			sourceTask: tasks.Task{Text: "New text"},
			expectMod:  true,
		},
		{
			name:       "priority modified",
			stateTask:  TaskState{Priority: "P1"},
			sourceTask: tasks.Task{Priority: "P2"},
			expectMod:  true,
		},
		{
			name:       "completion status modified",
			stateTask:  TaskState{Completed: false},
			sourceTask: tasks.Task{Completed: true},
			expectMod:  true,
		},
		{
			name:       "tags added",
			stateTask:  TaskState{Tags: []string{"tag1"}},
			sourceTask: tasks.Task{Tags: []string{"tag1", "tag2"}},
			expectMod:  true,
		},
		{
			name:       "tags removed",
			stateTask:  TaskState{Tags: []string{"tag1", "tag2"}},
			sourceTask: tasks.Task{Tags: []string{"tag1"}},
			expectMod:  true,
		},
		{
			name:       "same tags different order",
			stateTask:  TaskState{Tags: []string{"tag1", "tag2"}},
			sourceTask: tasks.Task{Tags: []string{"tag2", "tag1"}},
			expectMod:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modified := s.isTaskModified(tt.stateTask, tt.sourceTask)
			if modified != tt.expectMod {
				t.Errorf("Expected modified=%v, got %v", tt.expectMod, modified)
			}
		})
	}
}
