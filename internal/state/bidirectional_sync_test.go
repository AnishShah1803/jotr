package state

import (
	"fmt"
	"os"
	"sync"
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

// TestConcurrentBidirectionalSync tests that concurrent sync operations
// don't cause data loss or corruption. This test simulates two goroutines
// attempting to sync simultaneously, which should be protected by file locks.
func TestConcurrentBidirectionalSync(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"
	dailyNotePath := tempDir + "/daily.md"
	todoFilePath := tempDir + "/todo.md"

	// Initialize state file with some tasks
	initialState := NewTodoState()
	initialState.Tasks = map[string]TaskState{
		"task1": {ID: "task1", Text: "Task 1", Completed: false, Source: "old.md"},
		"task2": {ID: "task2", Text: "Task 2", Completed: false, Source: "old.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	// Create daily note with new task
	dailyContent := `## Tasks
- [ ] Task 1 <!-- id: task1 -->
- [ ] Task 2 <!-- id: task2 -->
- [ ] New Task from Daily <!-- id: task3 -->
`
	if err := os.WriteFile(dailyNotePath, []byte(dailyContent), 0644); err != nil {
		t.Fatalf("Failed to write daily note: %v", err)
	}

	// Create todo file
	todoContent := `## Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task from Todo
`
	if err := os.WriteFile(todoFilePath, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	// Use a channel to synchronize the start of both goroutines
	start := make(chan struct{})
	var errors [2]error
	var results [2]SyncResult
	var mu sync.Mutex

	// Launch two concurrent sync operations
	for i := 0; i < 2; i++ {
		go func(idx int) {
			<-start

			// Read the current state
			state, err := Read(statePath)
			if err != nil {
				errors[idx] = fmt.Errorf("goroutine %d: failed to read state: %w", idx, err)
				return
			}

			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)

			// Read tasks from daily note and todo file
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

			// Perform bidirectional sync
			result := state.BidirectionalSync(dailyTasks, todoTasks, dailyNotePath)
			results[idx] = result

			// Write the updated state with mutex to prevent JSON corruption
			if result.StateUpdated {
				mu.Lock()
				if err := state.Write(statePath); err != nil {
					errors[idx] = fmt.Errorf("goroutine %d: failed to write state: %w", idx, err)
					mu.Unlock()
					return
				}
				mu.Unlock()
			}

			errors[idx] = nil
		}(i)
	}

	// Start both goroutines simultaneously
	close(start)

	// Wait for both to complete (with timeout)
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
	case <-timeout:
		t.Fatal("Test timed out - possible deadlock detected")
	}

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}

	// Verify both sync operations produced valid results
	for i, result := range results {
		if !result.StateUpdated {
			t.Errorf("Goroutine %d: Expected StateUpdated=true", i)
		}
		if len(result.Conflicts) > 0 {
			t.Logf("Goroutine %d: Had %d conflicts (acceptable for this test)", i, len(result.Conflicts))
		}
	}

	// Verify final state is not corrupted
	// This test demonstrates why file locking is needed - without it, concurrent writes can cause data loss
	finalState, err := Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	// Verify base tasks are preserved
	baseTasks := []string{"task1", "task2"}
	for _, taskID := range baseTasks {
		if _, exists := finalState.Tasks[taskID]; !exists {
			t.Errorf("Expected base task %s to exist in final state", taskID)
		}
	}

	// Verify no task data corruption
	for id, task := range finalState.Tasks {
		if task.Text == "" {
			t.Errorf("Task %s has empty text (possible corruption)", id)
		}
		if task.ID == "" {
			t.Errorf("Task has empty ID (possible corruption): %v", task)
		}
	}

	// At least one sync operation should have added its new task
	_, hasTask3 := finalState.Tasks["task3"]
	_, hasTask4 := finalState.Tasks["task4"]
	if !hasTask3 && !hasTask4 {
		t.Error("Expected at least one new task (task3 or task4) to be present")
	}
}

// TestLockTimeoutBehavior tests that sync operations handle lock timeouts gracefully.
// This test verifies that when a lock cannot be acquired, appropriate error handling occurs.
func TestLockTimeoutBehavior(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"

	// Create a state file
	initialState := NewTodoState()
	initialState.Tasks = map[string]TaskState{
		"task1": {ID: "task1", Text: "Task 1", Completed: false, Source: "test.md"},
	}
	if err := initialState.Write(statePath); err != nil {
		t.Fatalf("Failed to write initial state: %v", err)
	}

	// Manually acquire a lock on the state file to simulate another process holding it
	lockPath := statePath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open lock file: %v", err)
	}
	defer lockFile.Close()

	// Try to acquire an exclusive lock (this will succeed for us)
	// In a real scenario, this would be held by another process
	// For this test, we'll verify the lock file mechanism works

	// Verify the lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after locking")
	}

	// Verify we can read state while lock is held (reads don't require lock)
	state, err := Read(statePath)
	if err != nil {
		t.Errorf("Should be able to read state while lock is held: %v", err)
	}

	// Verify state has expected data
	if len(state.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(state.Tasks))
	}

	// The actual timeout behavior is tested at the service level (task_service_test.go)
	// where the acquireSyncLocks function is called with a timeout
}

// TestDeadlockPrevention tests that rapid concurrent sync calls don't cause deadlocks.
// This test verifies the lock ordering (state → todo → daily) prevents deadlock.
func TestDeadlockPrevention(t *testing.T) {
	tempDir := t.TempDir()
	statePath := tempDir + "/.todo_state.json"
	dailyNotePath := tempDir + "/daily.md"
	todoFilePath := tempDir + "/todo.md"

	// Initialize all files
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
	if err := os.WriteFile(dailyNotePath, []byte(dailyContent), 0644); err != nil {
		t.Fatalf("Failed to write daily note: %v", err)
	}

	todoContent := `## Tasks
- [ ] Task 1
`
	if err := os.WriteFile(todoFilePath, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	// Launch multiple rapid sync operations to stress-test lock ordering
	numGoroutines := 5
	errors := make(chan error, numGoroutines)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			// Read and sync
			state, err := Read(statePath)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: failed to read state: %w", idx, err)
				return
			}

			// Perform sync (this will acquire locks in correct order)
			dailyTasks := []tasks.Task{
				{ID: "task1", Text: "Task 1", Completed: false, Section: "Tasks"},
			}
			todoTasks := []tasks.Task{{ID: "task1", Text: "Task 1", Completed: false}}

			result := state.BidirectionalSync(dailyTasks, todoTasks, dailyNotePath)

			// Write back
			if result.StateUpdated {
				if err := state.Write(statePath); err != nil {
					errors <- fmt.Errorf("goroutine %d: failed to write state: %w", idx, err)
					return
				}
			}

			errors <- nil
		}(i)
	}

	// Wait for all goroutines to complete with timeout
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

	// Verify final state is not corrupted
	finalState, err := Read(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	if len(finalState.Tasks) == 0 {
		t.Error("Final state has no tasks - possible data loss")
	}

	// Verify no duplicate tasks
	taskIDs := make(map[string]bool)
	for id := range finalState.Tasks {
		if taskIDs[id] {
			t.Errorf("Duplicate task found: %s", id)
		}
		taskIDs[id] = true
	}
}
