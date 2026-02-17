package state

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/constants"
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
		{
			name: "merge when both modified to same state - converged changes",
			dailyChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task", Completed: false, CreatedAt: now},
				NewTask:    &TaskState{ID: "abc123", Text: "Updated task", Completed: true},
			},
			todoChange: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask:    &TaskState{ID: "abc123", Text: "Task", Completed: false, CreatedAt: now},
				NewTask:    &TaskState{ID: "abc123", Text: "Updated task", Completed: true},
			},
			expectMerge:  true,
			expectedText: "Updated task",
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

func TestConcurrentBidirectionalSync(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"
	dailyNotePath := tempDir + "/daily.md"
	todoFilePath := tempDir + "/todo.md"

	initialState := NewTodoState()
	initialState.Tasks = map[string]TaskState{
		"task1": {ID: "task1", Text: "Task 1", Completed: false, Source: "old.md"},
		"task2": {ID: "task2", Text: "Task 2", Completed: false, Source: "old.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	dailyContent := `## Tasks
- [ ] Task 1 <!-- id: task1 -->
- [ ] Task 2 <!-- id: task2 -->
- [ ] New Task from Daily <!-- id: task3 -->
`
	if err := os.WriteFile(dailyNotePath, []byte(dailyContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to write daily note: %v", err)
	}

	todoContent := `## Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task from Todo
`
	if err := os.WriteFile(todoFilePath, []byte(todoContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var results []SyncResult

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start

			state, err := Read(statePath)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("goroutine %d: failed to read state: %w", idx, err))
				mu.Unlock()
				return
			}

			time.Sleep(10 * time.Millisecond)

			dailyTasks := []tasks.Task{
				{ID: "task1", Text: "Task 1", Completed: false, Section: "Tasks"},
				{ID: "task2", Text: "Task 2", Completed: false, Section: "Tasks"},
				{ID: "task3", Text: "New Task from Daily", Completed: false, Section: "Tasks"},
			}

			todoTasks := []tasks.Task{
				{ID: "task1", Text: "Task 1", Completed: false},
				{ID: "task2", Text: "Task 2", Completed: false},
				{ID: "task4", Text: "Task from Todo", Completed: false},
			}

			result := state.BidirectionalSync(dailyTasks, todoTasks, dailyNotePath)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			if result.StateUpdated {
				mu.Lock()
				if err := state.Write(statePath); err != nil {
					errors = append(errors, fmt.Errorf("goroutine %d: failed to write state: %w", idx, err))
					mu.Unlock()
					return
				}
				mu.Unlock()
			}
		}(i)
	}

	close(start)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock detected")
	}

	for _, err := range errors {
		if err != nil {
			t.Errorf("Goroutine failed: %v", err)
		}
	}

	for _, result := range results {
		if !result.StateUpdated {
			t.Error("Expected StateUpdated=true")
		}
		if len(result.Conflicts) > 0 {
			t.Logf("Had %d conflicts (acceptable for this test)", len(result.Conflicts))
		}
	}

	finalState, err := Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	baseTasks := []string{"task1", "task2"}
	for _, taskID := range baseTasks {
		if _, exists := finalState.Tasks[taskID]; !exists {
			t.Errorf("Expected base task %s to exist in final state", taskID)
		}
	}

	for id, task := range finalState.Tasks {
		if task.Text == "" {
			t.Errorf("Task %s has empty text (possible corruption)", id)
		}
		if task.ID == "" {
			t.Errorf("Task has empty ID (possible corruption): %v", task)
		}
	}

	_, hasTask3 := finalState.Tasks["task3"]
	_, hasTask4 := finalState.Tasks["task4"]
	if !hasTask3 && !hasTask4 {
		t.Error("Expected at least one new task (task3 or task4) to be present")
	}
}

func TestLockTimeoutBehavior(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"

	initialState := NewTodoState()
	initialState.Tasks = map[string]TaskState{
		"task1": {ID: "task1", Text: "Task 1", Completed: false, Source: "test.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	lockPath := statePath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, constants.FilePerm0600)
	if err != nil {
		t.Fatalf("Failed to open lock file: %v", err)
	}
	defer lockFile.Close()

	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after opening")
	}

	state, err := Read(statePath)
	if err != nil {
		t.Errorf("Should be able to read state while lock file exists: %v", err)
	}

	if len(state.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(state.Tasks))
	}
}

func TestTaskCompletionViaTodoSetsDates(t *testing.T) {
	state := NewTodoState()

	initialTask := TaskState{
		ID:          "abc123",
		Text:        "Review project proposal",
		Completed:   false,
		Source:      "Diary/2026-02-01.md",
		CreatedDate: "2026-02-01",
	}
	state.Tasks["abc123"] = initialTask

	dailyTasks := []tasks.Task{
		{ID: "abc123", Text: "Review project proposal", Completed: false, Section: "2026-02-01"},
	}

	todoTasks := []tasks.Task{
		{ID: "abc123", Text: "Review project proposal", Completed: true, Section: "Tasks"},
	}

	result := state.BidirectionalSync(dailyTasks, todoTasks, "Diary/2026-02-01.md")

	if !result.StateUpdated {
		t.Error("Expected StateUpdated=true, got false")
	}
	if !result.DailyChanged {
		t.Error("Expected DailyChanged=true (daily note should be updated with @completed tag)")
	}

	task, exists := state.Tasks["abc123"]
	if !exists {
		t.Fatal("Task abc123 should exist in state")
	}

	if task.CreatedDate != "2026-02-01" {
		t.Errorf("Expected CreatedDate='2026-02-01', got '%s'", task.CreatedDate)
	}

	if task.CompletedDate == "" {
		t.Error("Expected CompletedDate to be set, got empty string")
	}

	if len(task.CompletedDate) != 10 {
		t.Errorf("Expected CompletedDate in YYYY-MM-DD format, got '%s'", task.CompletedDate)
	}

	if !task.Completed {
		t.Error("Expected task to be marked complete")
	}
}

func TestTaskCompletionViaDailyNoteSetsCompletedDate(t *testing.T) {
	state := NewTodoState()

	initialTask := TaskState{
		ID:          "xyz789",
		Text:        "Complete from daily note",
		Completed:   false,
		Source:      "Diary/2026-02-03.md",
		CreatedDate: "2026-02-03",
	}
	state.Tasks["xyz789"] = initialTask

	// Daily note shows task as completed (user marked it done in the daily note)
	dailyTasks := []tasks.Task{
		{ID: "xyz789", Text: "Complete from daily note", Completed: true, Section: "2026-02-03"},
	}

	// Todo list still shows task as incomplete (hasn't been synced yet)
	todoTasks := []tasks.Task{
		{ID: "xyz789", Text: "Complete from daily note", Completed: false, Section: "Tasks"},
	}

	result := state.BidirectionalSync(dailyTasks, todoTasks, "Diary/2026-02-03.md")

	if !result.StateUpdated {
		t.Error("Expected StateUpdated=true, got false")
	}
	if !result.TodoChanged {
		t.Error("Expected TodoChanged=true (todo should be updated with completion)")
	}

	task, exists := state.Tasks["xyz789"]
	if !exists {
		t.Fatal("Task xyz789 should exist in state")
	}

	if task.CreatedDate != "2026-02-03" {
		t.Errorf("Expected CreatedDate='2026-02-03', got '%s'", task.CreatedDate)
	}

	if task.CompletedDate == "" {
		t.Error("Expected CompletedDate to be set, got empty string")
	}

	if len(task.CompletedDate) != 10 {
		t.Errorf("Expected CompletedDate in YYYY-MM-DD format, got '%s'", task.CompletedDate)
	}

	if !task.Completed {
		t.Error("Expected task to be marked complete")
	}
}

func TestDeadlockPrevention(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"
	dailyNotePath := tempDir + "/daily.md"
	todoFilePath := tempDir + "/todo.md"

	initialState := NewTodoState()
	initialState.Tasks = map[string]TaskState{
		"task1": {ID: "task1", Text: "Task 1", Completed: false, Source: "test.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	dailyContent := `## Tasks
- [ ] Task 1 <!-- id: task1 -->
`
	if err := os.WriteFile(dailyNotePath, []byte(dailyContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to write daily note: %v", err)
	}

	todoContent := `## Tasks
- [ ] Task 1
`
	if err := os.WriteFile(todoFilePath, []byte(todoContent), constants.FilePerm0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	numGoroutines := 5
	errors := make(chan error, numGoroutines)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			state, err := Read(statePath)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: failed to read state: %w", idx, err)
				return
			}

			dailyTasks := []tasks.Task{
				{ID: "task1", Text: "Task 1", Completed: false, Section: "Tasks"},
			}
			todoTasks := []tasks.Task{{ID: "task1", Text: "Task 1", Completed: false}}

			result := state.BidirectionalSync(dailyTasks, todoTasks, dailyNotePath)

			if result.StateUpdated {
				if err := state.Write(statePath); err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to write state: %w", idx, err)
					return
				}
			}

			errors <- nil
		}(i)
	}

	timeout := time.After(10 * time.Second)
	completed := 0

	for completed < numGoroutines {
		select {
		case <-done:
			completed++
		case err := <-errors:
			if err != nil {
				t.Errorf("Error in goroutine: %v", err)
			}
		case <-timeout:
			t.Fatalf("Test timed out after 10 seconds - possible deadlock detected (%d/%d completed)", completed, numGoroutines)
		}
	}

	finalState, err := Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	if len(finalState.Tasks) == 0 {
		t.Error("Final state has no tasks - possible data loss")
	}

	taskIDs := make(map[string]bool)
	for id := range finalState.Tasks {
		if taskIDs[id] {
			t.Errorf("Duplicate task found: %s", id)
		}
		taskIDs[id] = true
	}
}

func TestDetectDeletions(t *testing.T) {
	tests := []struct {
		name            string
		state           *TodoState
		dailyTasks      []tasks.Task
		todoTasks       []tasks.Task
		expectedDeletes int
		expectedTaskIDs []string
	}{
		{
			name: "no deletions - task exists in daily",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Task", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Task"},
			},
			todoTasks:       []tasks.Task{},
			expectedDeletes: 0,
		},
		{
			name: "no deletions - task exists in todo",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Task", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Task"},
			},
			expectedDeletes: 0,
		},
		{
			name: "deletion - task missing from both sources",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Deleted task", Source: "test.md"},
				},
			},
			dailyTasks:      []tasks.Task{},
			todoTasks:       []tasks.Task{},
			expectedDeletes: 1,
			expectedTaskIDs: []string{"abc123"},
		},
		{
			name: "multiple deletions",
			state: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Task 1", Source: "test.md"},
					"def456": {ID: "def456", Text: "Task 2", Source: "test.md"},
					"ghi789": {ID: "ghi789", Text: "Task 3", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "ghi789", Text: "Task 3"},
			},
			todoTasks:       []tasks.Task{},
			expectedDeletes: 2,
			expectedTaskIDs: []string{"abc123", "def456"},
		},
		{
			name: "empty state - no deletions",
			state: &TodoState{
				Tasks: map[string]TaskState{},
			},
			dailyTasks:      []tasks.Task{},
			todoTasks:       []tasks.Task{},
			expectedDeletes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deletions := tt.state.DetectDeletions(tt.dailyTasks, tt.todoTasks)

			if len(deletions) != tt.expectedDeletes {
				t.Errorf("Expected %d deletions, got %d", tt.expectedDeletes, len(deletions))
			}

			if len(tt.expectedTaskIDs) > 0 {
				foundIDs := make(map[string]bool)
				for _, del := range deletions {
					foundIDs[del.TaskID] = true
				}
				for _, expectedID := range tt.expectedTaskIDs {
					if !foundIDs[expectedID] {
						t.Errorf("Expected deletion of task %s not found", expectedID)
					}
				}
			}
		})
	}
}

func TestBidirectionalSyncWithDeletions(t *testing.T) {
	tests := []struct {
		name                 string
		initialState         *TodoState
		dailyTasks           []tasks.Task
		todoTasks            []tasks.Task
		expectDeleted        int
		stateUpdated         bool
		expectTaskInState    []string
		expectTaskNotInState []string
	}{
		{
			name: "deletion propagates - task removed from both sources",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Deleted task", Source: "test.md"},
				},
			},
			dailyTasks:           []tasks.Task{},
			todoTasks:            []tasks.Task{},
			expectDeleted:        1,
			stateUpdated:         true,
			expectTaskNotInState: []string{"abc123"},
		},
		{
			name: "completed tasks are NOT deleted",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Completed task", Completed: true, Source: "test.md"},
				},
			},
			dailyTasks:        []tasks.Task{},
			todoTasks:         []tasks.Task{},
			expectDeleted:     0,
			stateUpdated:      false,
			expectTaskInState: []string{"abc123"},
		},
		{
			name: "partial deletion - other tasks remain",
			initialState: &TodoState{
				Tasks: map[string]TaskState{
					"abc123": {ID: "abc123", Text: "Keep this", Source: "test.md"},
					"def456": {ID: "def456", Text: "Delete this", Source: "test.md"},
				},
			},
			dailyTasks: []tasks.Task{
				{ID: "abc123", Text: "Keep this"},
			},
			todoTasks: []tasks.Task{
				{ID: "abc123", Text: "Keep this"},
			},
			expectDeleted:        1,
			stateUpdated:         true,
			expectTaskInState:    []string{"abc123"},
			expectTaskNotInState: []string{"def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initialState.BidirectionalSync(tt.dailyTasks, tt.todoTasks, "test.md")

			if result.Deleted != tt.expectDeleted {
				t.Errorf("Expected %d deletions, got %d", tt.expectDeleted, result.Deleted)
			}

			if result.StateUpdated != tt.stateUpdated {
				t.Errorf("Expected StateUpdated=%v, got %v", tt.stateUpdated, result.StateUpdated)
			}

			for _, taskID := range tt.expectTaskInState {
				if _, exists := tt.initialState.Tasks[taskID]; !exists {
					t.Errorf("Expected task %s to exist in state", taskID)
				}
			}

			for _, taskID := range tt.expectTaskNotInState {
				if _, exists := tt.initialState.Tasks[taskID]; exists {
					t.Errorf("Expected task %s to NOT exist in state", taskID)
				}
			}
		})
	}
}

func TestBuildTaskChangeDetail(t *testing.T) {
	tests := []struct {
		name     string
		change   TaskChange
		expected TaskChangeDetail
	}{
		{
			name: "new task added",
			change: TaskChange{
				TaskID:     "abc123",
				ChangeType: Added,
				NewTask: &TaskState{
					ID:     "abc123",
					Text:   "New task text",
					Source: "test.md",
				},
			},
			expected: TaskChangeDetail{
				ID:      "abc123",
				Text:    "New task text",
				Change:  "added",
				Details: "new task added",
			},
		},
		{
			name: "task modified",
			change: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask: &TaskState{
					ID:     "abc123",
					Text:   "Old text",
					Source: "test.md",
				},
				NewTask: &TaskState{
					ID:     "abc123",
					Text:   "New text",
					Source: "test.md",
				},
			},
			expected: TaskChangeDetail{
				ID:      "abc123",
				Text:    "New text",
				Change:  "updated",
				From:    "Old text",
				To:      "New text",
				Details: "text changed",
			},
		},
		{
			name: "task deleted",
			change: TaskChange{
				TaskID:     "abc123",
				ChangeType: Deleted,
				OldTask: &TaskState{
					ID:     "abc123",
					Text:   "Deleted task",
					Source: "test.md",
				},
			},
			expected: TaskChangeDetail{
				ID:      "abc123",
				Text:    "",
				Change:  "deleted",
				From:    "Deleted task",
				Details: "task deleted",
			},
		},
		{
			name: "task completed",
			change: TaskChange{
				TaskID:     "abc123",
				ChangeType: Modified,
				OldTask: &TaskState{
					ID:        "abc123",
					Text:      "Task",
					Completed: false,
					Source:    "test.md",
				},
				NewTask: &TaskState{
					ID:        "abc123",
					Text:      "Task",
					Completed: true,
					Source:    "test.md",
				},
			},
			expected: TaskChangeDetail{
				ID:      "abc123",
				Text:    "Task",
				Change:  "updated",
				From:    "Task",
				To:      "Task",
				Details: "marked complete",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTaskChangeDetail(tt.change)

			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID=%q, got %q", tt.expected.ID, result.ID)
			}
			if result.Text != tt.expected.Text {
				t.Errorf("Expected Text=%q, got %q", tt.expected.Text, result.Text)
			}
			if result.Change != tt.expected.Change {
				t.Errorf("Expected Change=%q, got %q", tt.expected.Change, result.Change)
			}
			if result.From != tt.expected.From {
				t.Errorf("Expected From=%q, got %q", tt.expected.From, result.From)
			}
			if result.To != tt.expected.To {
				t.Errorf("Expected To=%q, got %q", tt.expected.To, result.To)
			}
			if result.Details != tt.expected.Details {
				t.Errorf("Expected Details=%q, got %q", tt.expected.Details, result.Details)
			}
		})
	}
}

func TestBuildChangeDetails(t *testing.T) {
	tests := []struct {
		name     string
		oldTask  *TaskState
		newTask  *TaskState
		expected string
	}{
		{
			name:     "text changed",
			oldTask:  &TaskState{ID: "abc123", Text: "Old text"},
			newTask:  &TaskState{ID: "abc123", Text: "New text"},
			expected: "text changed",
		},
		{
			name:     "nil old task",
			oldTask:  nil,
			newTask:  &TaskState{ID: "abc123", Text: "New text"},
			expected: "",
		},
		{
			name:     "nil new task",
			oldTask:  &TaskState{ID: "abc123", Text: "Old text"},
			newTask:  nil,
			expected: "",
		},
		{
			name:     "both nil",
			oldTask:  nil,
			newTask:  nil,
			expected: "",
		},
		{
			name:     "marked complete",
			oldTask:  &TaskState{ID: "abc123", Text: "Task", Completed: false},
			newTask:  &TaskState{ID: "abc123", Text: "Task", Completed: true},
			expected: "marked complete",
		},
		{
			name:     "marked incomplete",
			oldTask:  &TaskState{ID: "abc123", Text: "Task", Completed: true},
			newTask:  &TaskState{ID: "abc123", Text: "Task", Completed: false},
			expected: "marked incomplete",
		},
		{
			name:     "priority changed",
			oldTask:  &TaskState{ID: "abc123", Text: "Task", Priority: "P2"},
			newTask:  &TaskState{ID: "abc123", Text: "Task", Priority: "P1"},
			expected: "priority changed to P1",
		},
		{
			name:     "no changes",
			oldTask:  &TaskState{ID: "abc123", Text: "Task", Completed: false, Priority: "P2"},
			newTask:  &TaskState{ID: "abc123", Text: "Task", Completed: false, Priority: "P2"},
			expected: "modified",
		},
		{
			name:     "multiple changes",
			oldTask:  &TaskState{ID: "abc123", Text: "Old", Completed: false, Priority: "P2"},
			newTask:  &TaskState{ID: "abc123", Text: "New", Completed: true, Priority: "P1"},
			expected: "text changed, marked complete, priority changed to P1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildChangeDetails(tt.oldTask, tt.newTask)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
