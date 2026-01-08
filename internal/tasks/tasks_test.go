package tasks

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/testhelpers"
)

func TestParseTasks(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	content := `# Tasks

- [x] Completed task one <!-- id: abc12345 -->
- [ ] In progress task two <!-- id: def45678 -->
- [ ] Pending task three <!-- id: a1b2c3d4 -->

## Next Section

More task content here`

	// Create a temporary file with task content
	taskFile := "tasks.md"
	fs.WriteFile(t, taskFile, content)

	// Test parsing tasks (use full path from fs)
	ctx := context.Background()

	tasks, err := ReadTasks(ctx, filepath.Join(fs.BaseDir, taskFile))
	if err != nil {
		t.Fatalf("ReadTasks() error = %v", err)
	}

	// Verify parsed tasks
	if len(tasks) != 3 {
		t.Errorf("ParseTasks() returned %d tasks, want 3", len(tasks))
	}

	// Check first task
	if tasks[0].Text != "Completed task one" {
		t.Errorf("Task 1 text = %q, want %q", tasks[0].Text, "Completed task one")
	}

	// Check second task
	if tasks[1].Text != "In progress task two" {
		t.Errorf("Task 2 text = %q, want %q", tasks[1].Text, "In progress task two")
	}

	// Check third task
	if tasks[2].Text != "Pending task three" {
		t.Errorf("Task 3 text = %q, want %q", tasks[2].Text, "Pending task three")
	}
}

func TestFormatTask(t *testing.T) {
	task := &Task{
		Text:      "Test task",
		Completed: false,
		Priority:  "",
		Tags:      []string{"work", "feature"},
		Line:      1,
		Section:   "Backlog",
		ID:        "test123",
	}

	formatted := FormatTask(*task)
	expected := "○  Test task"

	if formatted != expected {
		t.Errorf("FormatTask() = %q, want %q", formatted, expected)
	}
}

func TestReadTasks(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	// Create a temporary file with tasks
	taskFile := "tasks.md"
	content := `# Tasks

- [x] Task one <!-- id: abc12345 -->
- [ ] Task two <!-- id: def45678 -->

Another line here
- [ ] Task three <!-- id: ghi78901 -->
`

	fs.WriteFile(t, taskFile, content)

	// Test reading tasks (use full path from fs)
	ctx := context.Background()

	tasks, err := ReadTasks(ctx, filepath.Join(fs.BaseDir, taskFile))
	if err != nil {
		t.Fatalf("ReadTasks() error = %v", err)
	}

	// Verify parsed tasks
	if len(tasks) != 3 {
		t.Errorf("ReadTasks() returned %d tasks, want 3", len(tasks))
	}

	// Check task content
	if tasks[0].Text != "Task one" {
		t.Errorf("Task 1 text = %q, want %q", tasks[0].Text, "Task one")
	}
}

// TestGenerateTaskID tests that GenerateTaskID produces deterministic and unique IDs.
func TestGenerateTaskID(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantLen     int
		wantPattern string // regex pattern the ID should match
	}{
		{
			name:        "simple text",
			text:        "Review proposal",
			wantLen:     8,
			wantPattern: "^[a-f0-9]{8}$",
		},
		{
			name:        "empty text",
			text:        "",
			wantLen:     8,
			wantPattern: "^[a-f0-9]{8}$",
		},
		{
			name:        "long text",
			text:        "This is a very long task description that should still generate a consistent 8-character hex ID",
			wantLen:     8,
			wantPattern: "^[a-f0-9]{8}$",
		},
		{
			name:        "text with special chars",
			text:        "Task with @#$%^&*() special characters!",
			wantLen:     8,
			wantPattern: "^[a-f0-9]{8}$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateTaskID(tt.text)
			if len(id) != tt.wantLen {
				t.Errorf("GenerateTaskID(%q) length = %d, want %d", tt.text, len(id), tt.wantLen)
			}
			// Check pattern
			matched, _ := regexp.MatchString(tt.wantPattern, id)
			if !matched {
				t.Errorf("GenerateTaskID(%q) = %q, does not match pattern %q", tt.text, id, tt.wantPattern)
			}
		})
	}
}

// TestGenerateTaskID_Deterministic tests that GenerateTaskID produces the same ID for the same input.
func TestGenerateTaskID_Deterministic(t *testing.T) {
	text := "Review proposal document"

	id1 := GenerateTaskID(text)
	id2 := GenerateTaskID(text)
	id3 := GenerateTaskID(text)

	if id1 != id2 || id2 != id3 {
		t.Errorf("GenerateTaskID is not deterministic: got %q, %q, %q", id1, id2, id3)
	}
}

// TestGenerateTaskID_Uniqueness tests that GenerateTaskID produces different IDs for different inputs.
func TestGenerateTaskID_Uniqueness(t *testing.T) {
	tasks := []string{
		"Review proposal",
		"Review proposal document",
		"Review",
		"review proposal", // lowercase
	}

	ids := make(map[string]int)
	for i, text := range tasks {
		id := GenerateTaskID(text)
		if prevIdx, exists := ids[id]; exists {
			t.Errorf("GenerateTaskID produced duplicate IDs: %q for task %d (%q) and task %d (%q)",
				id, prevIdx, tasks[prevIdx], i, text)
		}
		ids[id] = i
	}

	if len(ids) != len(tasks) {
		t.Errorf("GenerateTaskID did not produce unique IDs for different inputs: got %d unique IDs for %d inputs",
			len(ids), len(tasks))
	}
}

// TestExtractTaskID tests that ExtractTaskID correctly extracts task IDs from text.
func TestExtractTaskID(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantID   string
		wantBool bool
	}{
		{
			name:     "valid ID",
			text:     "Review proposal <!-- id: abc12345 -->",
			wantID:   "abc12345",
			wantBool: true,
		},
		{
			name:     "valid ID with leading space",
			text:     "Task with ID <!-- id: def45678 -->",
			wantID:   "def45678",
			wantBool: true,
		},
		{
			name:     "no ID",
			text:     "Task without ID",
			wantID:   "",
			wantBool: false,
		},
		{
			name:     "empty text",
			text:     "",
			wantID:   "",
			wantBool: false,
		},
		{
			name:     "ID with wrong format",
			text:     "Task <!-- id: wrong -->",
			wantID:   "",
			wantBool: false,
		},
		{
			name:     "ID with short hex",
			text:     "Task <!-- id: abc12 -->",
			wantID:   "",
			wantBool: false,
		},
		{
			name:     "ID with uppercase hex",
			text:     "Task <!-- id: ABC12345 -->",
			wantID:   "",
			wantBool: false, // only lowercase a-f
		},
		{
			name:     "ID in middle of text",
			text:     "<!-- id: 12345678 --> Task with ID in middle",
			wantID:   "12345678",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := ExtractTaskID(tt.text)
			if id != tt.wantID {
				t.Errorf("ExtractTaskID(%q) = %q, want %q", tt.text, id, tt.wantID)
			}
			// Verify we can distinguish between found and not found
			found := id != ""
			if found != tt.wantBool {
				t.Errorf("ExtractTaskID(%q) found = %v, want %v", tt.text, found, tt.wantBool)
			}
		})
	}
}

// TestStripTaskID tests that StripTaskID correctly removes task ID comments from text.
func TestStripTaskID(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantText string
	}{
		{
			name:     "strips ID at end",
			text:     "Review proposal <!-- id: abc12345 -->",
			wantText: "Review proposal",
		},
		{
			name:     "strips ID with leading space",
			text:     "Task text <!-- id: def45678 -->",
			wantText: "Task text",
		},
		{
			name:     "no ID returns original",
			text:     "Task without ID",
			wantText: "Task without ID",
		},
		{
			name:     "empty text",
			text:     "",
			wantText: "",
		},
		{
			name:     "only ID returns empty",
			text:     "<!-- id: abc12345 -->",
			wantText: "",
		},
		{
			name:     "ID at end preserves leading text",
			text:     "Task with ID at end <!-- id: abc12345 -->",
			wantText: "Task with ID at end",
		},
		{
			name:     "strips second ID",
			text:     "First task <!-- id: abc12345 -->",
			wantText: "First task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText := StripTaskID(tt.text)
			if gotText != tt.wantText {
				t.Errorf("StripTaskID(%q) = %q, want %q", tt.text, gotText, tt.wantText)
			}
		})
	}
}

// TestEnsureTaskID tests that EnsureTaskID correctly handles task ID assignment.
func TestEnsureTaskID(t *testing.T) {
	tests := []struct {
		name       string
		task       Task
		wantID     bool // should have ID after
		wantTextID bool // text should contain ID comment
	}{
		{
			name:       "task without ID generates new ID",
			task:       Task{Text: "New task without ID", ID: ""},
			wantID:     true,
			wantTextID: true,
		},
		{
			name:       "task with existing ID keeps it",
			task:       Task{Text: "Task with existing ID", ID: "existing"},
			wantID:     true,
			wantTextID: false, // text doesn't have embedded ID, only struct has it
		},
		{
			name:       "task without struct ID but with embedded ID in text extracts it",
			task:       Task{Text: "Task with embedded ID <!-- id: embedded1 -->", ID: ""},
			wantID:     true,
			wantTextID: false, // ID is extracted from text, not added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalText := tt.task.Text
			EnsureTaskID(&tt.task)

			if tt.wantID && tt.task.ID == "" {
				t.Errorf("EnsureTaskID() did not set ID, want non-empty ID")
			}
			if !tt.wantID && tt.task.ID != "" {
				t.Errorf("EnsureTaskID() set ID = %q, want empty ID", tt.task.ID)
			}

			// Check if text contains embedded ID
			textHasID := strings.Contains(tt.task.Text, "<!-- id: "+tt.task.ID+" -->")
			if tt.wantTextID && !textHasID {
				t.Errorf("EnsureTaskID() text = %q, does not contain embedded ID %q",
					tt.task.Text, tt.task.ID)
			}
			if !tt.wantTextID && originalText == tt.task.Text && textHasID {
				t.Errorf("EnsureTaskID() modified text when it shouldn't have")
			}
		})
	}
}

// TestEnsureTaskID_GeneratesConsistentID tests that EnsureTaskID generates the same ID for same text.
func TestEnsureTaskID_ConsistentID(t *testing.T) {
	task1 := Task{Text: "Review proposal", ID: ""}
	task2 := Task{Text: "Review proposal", ID: ""}

	EnsureTaskID(&task1)
	EnsureTaskID(&task2)

	if task1.ID != task2.ID {
		t.Errorf("EnsureTaskID generated different IDs for same text: %q vs %q",
			task1.ID, task2.ID)
	}
}

// TestEnsureTaskID_EmbeddedIDOverride tests that embedded IDs in text take precedence.
func TestEnsureTaskID_EmbeddedIDOverride(t *testing.T) {
	task := Task{Text: "Task with embedded ID <!-- id: a1b2c3d4 -->", ID: ""}

	EnsureTaskID(&task)

	if task.ID != "a1b2c3d4" {
		t.Errorf("EnsureTaskID() ID = %q, want embedded ID %q", task.ID, "a1b2c3d4")
	}
}

// TestEnsureTaskID_NoDuplicateEmbedding tests that EnsureTaskID doesn't add duplicate ID comments.
func TestEnsureTaskID_NoDuplicateEmbedding(t *testing.T) {
	task := Task{Text: "New task", ID: ""}

	EnsureTaskID(&task)
	id1 := task.ID

	// Call EnsureTaskID again
	EnsureTaskID(&task)
	id2 := task.ID

	if id1 != id2 {
		t.Errorf("EnsureTaskID() changed ID on second call: %q -> %q", id1, id2)
	}

	// Count ID comment occurrences
	count := strings.Count(task.Text, "<!-- id: "+id1+" -->")
	if count != 1 {
		t.Errorf("EnsureTaskID() embedded ID %d times, want 1", count)
	}
}

// TestParseTasks_AsteriskFormat tests that ParseTasks correctly parses asterisk-format tasks.
func TestParseTasks_AsteriskFormat(t *testing.T) {
	content := `# Tasks

* [ ] Pending asterisk task
* [x] Completed asterisk task
* [X] Another completed (uppercase X)

## Next Section

More content`

	tasks := ParseTasks(content)

	if len(tasks) != 3 {
		t.Errorf("ParseTasks() returned %d tasks, want 3", len(tasks))
	}

	// Check first task (pending)
	if tasks[0].Completed {
		t.Errorf("Task 1 should be pending, but is completed")
	}
	if tasks[0].Text != "Pending asterisk task" {
		t.Errorf("Task 1 text = %q, want %q", tasks[0].Text, "Pending asterisk task")
	}

	// Check second task (completed with lowercase x)
	if !tasks[1].Completed {
		t.Errorf("Task 2 should be completed")
	}
	if tasks[1].Text != "Completed asterisk task" {
		t.Errorf("Task 2 text = %q, want %q", tasks[1].Text, "Completed asterisk task")
	}

	// Check third task (completed with uppercase X)
	if !tasks[2].Completed {
		t.Errorf("Task 3 should be completed")
	}
	if tasks[2].Text != "Another completed (uppercase X)" {
		t.Errorf("Task 3 text = %q, want %q", tasks[2].Text, "Another completed (uppercase X)")
	}
}

// TestParseTasks_PlusFormat tests that ParseTasks correctly parses plus-format tasks.
func TestParseTasks_PlusFormat(t *testing.T) {
	content := `# Tasks

+ [ ] Pending plus task
+ [x] Completed plus task
+ [X] Another completed (uppercase X)

## Next Section

More content`

	tasks := ParseTasks(content)

	if len(tasks) != 3 {
		t.Errorf("ParseTasks() returned %d tasks, want 3", len(tasks))
	}

	// Check first task (pending)
	if tasks[0].Completed {
		t.Errorf("Task 1 should be pending, but is completed")
	}
	if tasks[0].Text != "Pending plus task" {
		t.Errorf("Task 1 text = %q, want %q", tasks[0].Text, "Pending plus task")
	}

	// Check second task (completed with lowercase x)
	if !tasks[1].Completed {
		t.Errorf("Task 2 should be completed")
	}
	if tasks[1].Text != "Completed plus task" {
		t.Errorf("Task 2 text = %q, want %q", tasks[1].Text, "Completed plus task")
	}

	// Check third task (completed with uppercase X)
	if !tasks[2].Completed {
		t.Errorf("Task 3 should be completed")
	}
	if tasks[2].Text != "Another completed (uppercase X)" {
		t.Errorf("Task 3 text = %q, want %q", tasks[2].Text, "Another completed (uppercase X)")
	}
}

// TestParseTasks_MixedFormats tests that ParseTasks correctly handles mixed task formats.
func TestParseTasks_MixedFormats(t *testing.T) {
	content := `# Tasks

- [ ] Dash pending task
* [ ] Asterisk pending task
+ [ ] Plus pending task
- [x] Dash completed
* [x] Asterisk completed
+ [x] Plus completed`

	tasks := ParseTasks(content)

	if len(tasks) != 6 {
		t.Errorf("ParseTasks() returned %d tasks, want 6", len(tasks))
	}

	// Verify order is preserved
	expectedTexts := []string{
		"Dash pending task",
		"Asterisk pending task",
		"Plus pending task",
		"Dash completed",
		"Asterisk completed",
		"Plus completed",
	}

	for i, expected := range expectedTexts {
		if tasks[i].Text != expected {
			t.Errorf("Task %d text = %q, want %q", i+1, tasks[i].Text, expected)
		}
	}

	// Verify completion status
	pendingCount := 0
	completedCount := 0
	for _, task := range tasks {
		if task.Completed {
			completedCount++
		} else {
			pendingCount++
		}
	}

	if pendingCount != 3 {
		t.Errorf("Expected 3 pending tasks, got %d", pendingCount)
	}
	if completedCount != 3 {
		t.Errorf("Expected 3 completed tasks, got %d", completedCount)
	}
}

// TestParseTasks_AllFormatsWithFeatures tests that all formats work with priority, tags, and IDs.
func TestParseTasks_AllFormatsWithFeatures(t *testing.T) {
	content := `# Tasks

- [ ] Dash with priority [P1] and #tag1
* [ ] Asterisk with priority [P2] and #tag2
+ [ ] Plus with priority [P3] and #tag3
- [x] Completed dash #tag4
* [x] Completed asterisk #tag5
+ [x] Completed plus #tag6`

	tasks := ParseTasks(content)

	if len(tasks) != 6 {
		t.Errorf("ParseTasks() returned %d tasks, want 6", len(tasks))
	}

	// Check priorities
	priorities := []string{"P1", "P2", "P3"}
	for i, expected := range priorities {
		if tasks[i].Priority != expected {
			t.Errorf("Task %d priority = %q, want %q", i+1, tasks[i].Priority, expected)
		}
	}

	// Check tags
	tags := [][]string{
		{"tag1"},
		{"tag2"},
		{"tag3"},
		{"tag4"},
		{"tag5"},
		{"tag6"},
	}
	for i, expectedTags := range tags {
		if len(tasks[i].Tags) != len(expectedTags) {
			t.Errorf("Task %d has %d tags, want %d", i+1, len(tasks[i].Tags), len(expectedTags))
			continue
		}
		for j, tag := range expectedTags {
			if tasks[i].Tags[j] != tag {
				t.Errorf("Task %d tag %d = %q, want %q", i+1, j, tasks[i].Tags[j], tag)
			}
		}
	}
}

// TestParseTasks_NonTaskLines tests that non-task lines are correctly ignored.
func TestParseTasks_NonTaskLines(t *testing.T) {
	content := `# Tasks

This is a regular line
- Not a task (missing brackets)
* [ incomplete asterisk
+ [ incomplete plus
[ ] No bullet character

- [ ] Actual task 1
* [ ] Actual task 2
+ [ ] Actual task 3`

	tasks := ParseTasks(content)

	if len(tasks) != 3 {
		t.Errorf("ParseTasks() returned %d tasks, want 3", len(tasks))
	}

	expected := []string{"Actual task 1", "Actual task 2", "Actual task 3"}
	for i, expectedText := range expected {
		if tasks[i].Text != expectedText {
			t.Errorf("Task %d text = %q, want %q", i+1, tasks[i].Text, expectedText)
		}
	}
}

// TestIsOverdue tests that IsOverdue correctly identifies overdue tasks.
func TestIsOverdue(t *testing.T) {
	tests := []struct {
		name        string
		task        Task
		wantOverdue bool
	}{
		{
			name: "past due date",
			task: Task{
				Text:      "Review proposal due: 2020-01-01",
				Completed: false,
			},
			wantOverdue: true,
		},
		{
			name: "future due date",
			task: Task{
				Text:      "Review proposal due: 2099-12-31",
				Completed: false,
			},
			wantOverdue: false,
		},
		{
			name: "no due date",
			task: Task{
				Text:      "Review proposal",
				Completed: false,
			},
			wantOverdue: false,
		},
		{
			name: "completed task with past due date",
			task: Task{
				Text:      "Review proposal due: 2020-01-01",
				Completed: true,
			},
			wantOverdue: false,
		},
		{
			name: "due date with different format",
			task: Task{
				Text:      "Review proposal due:2023-05-15",
				Completed: false,
			},
			wantOverdue: true,
		},
		{
			name: "due date with extra spaces",
			task: Task{
				Text:      "Review proposal due:   2020-01-01",
				Completed: false,
			},
			wantOverdue: true,
		},
		{
			name: "due date with other content after",
			task: Task{
				Text:      "Review proposal due: 2020-01-01 #important",
				Completed: false,
			},
			wantOverdue: true,
		},
		{
			name: "due date with other content before",
			task: Task{
				Text:      "Important task due: 2020-01-01",
				Completed: false,
			},
			wantOverdue: true,
		},
		{
			name: "invalid date format",
			task: Task{
				Text:      "Review proposal due: 2020/01/01",
				Completed: false,
			},
			wantOverdue: false,
		},
		{
			name: "incomplete date format",
			task: Task{
				Text:      "Review proposal due: 2020-01",
				Completed: false,
			},
			wantOverdue: false,
		},
		{
			name: "text contains due but not a date pattern",
			task: Task{
				Text:      "Review proposal due next week",
				Completed: false,
			},
			wantOverdue: false,
		},
		{
			name: "empty task text",
			task: Task{
				Text:      "",
				Completed: false,
			},
			wantOverdue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOverdue(tt.task)
			if got != tt.wantOverdue {
				t.Errorf("IsOverdue(%q, completed=%v) = %v, want %v",
					tt.task.Text, tt.task.Completed, got, tt.wantOverdue)
			}
		})
	}
}

// TestIsOverdue_PastDateVariants tests IsOverdue with various past date formats.
func TestIsOverdue_PastDateVariants(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantOverdue bool
	}{
		{
			name:        "standard format YYYY-MM-DD",
			text:        "Task due: 2020-01-01",
			wantOverdue: true,
		},
		{
			name:        "mid-year date",
			text:        "Task due: 2023-06-15",
			wantOverdue: true,
		},
		{
			name:        "yesterday date",
			text:        "Task due: " + time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			wantOverdue: true,
		},
		{
			name:        "tomorrow date",
			text:        "Task due: " + time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			wantOverdue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Text: tt.text, Completed: false}
			got := IsOverdue(task)
			if got != tt.wantOverdue {
				t.Errorf("IsOverdue(%q) = %v, want %v", tt.text, got, tt.wantOverdue)
			}
		})
	}
}

// TestIsOverdue_IntegrationWithParseTasks tests IsOverdue with tasks parsed from content.
func TestIsOverdue_IntegrationWithParseTasks(t *testing.T) {
	content := `# Tasks

- [ ] Overdue task due: 2020-01-01
- [x] Completed overdue task due: 2020-01-01
- [ ] Future task due: 2099-01-01
- [ ] No due date task
`

	tasks := ParseTasks(content)

	if len(tasks) != 4 {
		t.Fatalf("ParseTasks() returned %d tasks, want 4", len(tasks))
	}

	// Task 1: Overdue and pending
	if !IsOverdue(tasks[0]) {
		t.Errorf("Task 1 should be overdue")
	}

	// Task 2: Overdue but completed
	if IsOverdue(tasks[1]) {
		t.Errorf("Task 2 should not be overdue (completed)")
	}

	// Task 3: Future due date
	if IsOverdue(tasks[2]) {
		t.Errorf("Task 3 should not be overdue (future)")
	}

	// Task 4: No due date
	if IsOverdue(tasks[3]) {
		t.Errorf("Task 4 should not be overdue (no date)")
	}
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// TestFilterTasks tests that FilterTasks correctly filters tasks by various criteria.
func TestFilterTasks(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Completed: false, Priority: "P1", Section: "Section A"},
		{Text: "Task 2", Completed: true, Priority: "P1", Section: "Section A"},
		{Text: "Task 3", Completed: false, Priority: "P2", Section: "Section B"},
		{Text: "Task 4", Completed: true, Priority: "P2", Section: "Section B"},
		{Text: "Task 5", Completed: false, Priority: "P3", Section: "Section C"},
	}

	tests := []struct {
		name      string
		tasks     []Task
		completed *bool
		priority  string
		section   string
		wantCount int
		wantTexts []string
	}{
		{
			name:      "filter by completed",
			tasks:     tasks,
			completed: boolPtr(true),
			wantCount: 2,
			wantTexts: []string{"Task 2", "Task 4"},
		},
		{
			name:      "filter by pending",
			tasks:     tasks,
			completed: boolPtr(false),
			wantCount: 3,
			wantTexts: []string{"Task 1", "Task 3", "Task 5"},
		},
		{
			name:      "filter by priority P1",
			tasks:     tasks,
			priority:  "P1",
			wantCount: 2,
			wantTexts: []string{"Task 1", "Task 2"},
		},
		{
			name:      "filter by priority P2",
			tasks:     tasks,
			priority:  "P2",
			wantCount: 2,
			wantTexts: []string{"Task 3", "Task 4"},
		},
		{
			name:      "filter by section Section A",
			tasks:     tasks,
			section:   "Section A",
			wantCount: 2,
			wantTexts: []string{"Task 1", "Task 2"},
		},
		{
			name:      "filter by section Section B",
			tasks:     tasks,
			section:   "Section B",
			wantCount: 2,
			wantTexts: []string{"Task 3", "Task 4"},
		},
		{
			name:      "filter by completed and priority",
			tasks:     tasks,
			completed: boolPtr(false),
			priority:  "P2",
			wantCount: 1,
			wantTexts: []string{"Task 3"},
		},
		{
			name:      "filter by completed and section",
			tasks:     tasks,
			completed: boolPtr(true),
			section:   "Section B",
			wantCount: 1,
			wantTexts: []string{"Task 4"},
		},
		{
			name:      "filter by priority and section",
			tasks:     tasks,
			priority:  "P1",
			section:   "Section A",
			wantCount: 2,
			wantTexts: []string{"Task 1", "Task 2"},
		},
		{
			name:      "filter by all criteria",
			tasks:     tasks,
			completed: boolPtr(false),
			priority:  "P3",
			section:   "Section C",
			wantCount: 1,
			wantTexts: []string{"Task 5"},
		},
		{
			name:      "no matching tasks",
			tasks:     tasks,
			completed: boolPtr(true),
			priority:  "P1",
			section:   "Section B",
			wantCount: 0,
			wantTexts: nil,
		},
		{
			name:      "no filter returns all",
			tasks:     tasks,
			completed: nil,
			priority:  "",
			section:   "",
			wantCount: 5,
			wantTexts: []string{"Task 1", "Task 2", "Task 3", "Task 4", "Task 5"},
		},
		{
			name:      "empty tasks slice",
			tasks:     []Task{},
			wantCount: 0,
			wantTexts: nil,
		},
		{
			name:      "filter non-existent priority",
			tasks:     tasks,
			priority:  "P99",
			wantCount: 0,
			wantTexts: nil,
		},
		{
			name:      "filter non-existent section",
			tasks:     tasks,
			section:   "NonExistent",
			wantCount: 0,
			wantTexts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterTasks(tt.tasks, tt.completed, tt.priority, tt.section)

			if len(filtered) != tt.wantCount {
				t.Errorf("FilterTasks() returned %d tasks, want %d", len(filtered), tt.wantCount)
			}

			for i, task := range filtered {
				if i < len(tt.wantTexts) && task.Text != tt.wantTexts[i] {
					t.Errorf("Filtered task %d text = %q, want %q", i+1, task.Text, tt.wantTexts[i])
				}
			}
		})
	}
}

// TestFilterTasks_PriorityWithNoMatch tests FilterTasks when priority doesn't match.
func TestFilterTasks_PriorityWithNoMatch(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Priority: "P1"},
		{Text: "Task 2", Priority: "P2"},
	}

	filtered := FilterTasks(tasks, nil, "P3", "")

	if len(filtered) != 0 {
		t.Errorf("FilterTasks() returned %d tasks, want 0", len(filtered))
	}
}

// TestFilterTasks_SectionWithNoMatch tests FilterTasks when section doesn't match.
func TestFilterTasks_SectionWithNoMatch(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Section: "Section A"},
		{Text: "Task 2", Section: "Section B"},
	}

	filtered := FilterTasks(tasks, nil, "", "Section C")

	if len(filtered) != 0 {
		t.Errorf("FilterTasks() returned %d tasks, want 0", len(filtered))
	}
}

// TestCountTasks tests that CountTasks correctly counts tasks by status.
func TestCountTasks(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []Task
		wantTotal     int
		wantCompleted int
		wantPending   int
	}{
		{
			name:          "all pending",
			tasks:         []Task{{Text: "Task 1"}, {Text: "Task 2"}, {Text: "Task 3"}},
			wantTotal:     3,
			wantCompleted: 0,
			wantPending:   3,
		},
		{
			name: "all completed",
			tasks: []Task{
				{Text: "Task 1", Completed: true},
				{Text: "Task 2", Completed: true},
				{Text: "Task 3", Completed: true},
			},
			wantTotal:     3,
			wantCompleted: 3,
			wantPending:   0,
		},
		{
			name: "mixed completed and pending",
			tasks: []Task{
				{Text: "Task 1", Completed: false},
				{Text: "Task 2", Completed: true},
				{Text: "Task 3", Completed: false},
				{Text: "Task 4", Completed: true},
			},
			wantTotal:     4,
			wantCompleted: 2,
			wantPending:   2,
		},
		{
			name:          "empty slice",
			tasks:         []Task{},
			wantTotal:     0,
			wantCompleted: 0,
			wantPending:   0,
		},
		{
			name:          "single pending task",
			tasks:         []Task{{Text: "Task 1", Completed: false}},
			wantTotal:     1,
			wantCompleted: 0,
			wantPending:   1,
		},
		{
			name:          "single completed task",
			tasks:         []Task{{Text: "Task 1", Completed: true}},
			wantTotal:     1,
			wantCompleted: 1,
			wantPending:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, completed, pending := CountTasks(tt.tasks)

			if total != tt.wantTotal {
				t.Errorf("CountTasks() total = %d, want %d", total, tt.wantTotal)
			}
			if completed != tt.wantCompleted {
				t.Errorf("CountTasks() completed = %d, want %d", completed, tt.wantCompleted)
			}
			if pending != tt.wantPending {
				t.Errorf("CountTasks() pending = %d, want %d", pending, tt.wantPending)
			}
		})
	}
}

// TestCountTasks_IntegrationWithParseTasks tests CountTasks with parsed tasks.
func TestCountTasks_IntegrationWithParseTasks(t *testing.T) {
	content := `# Tasks

- [x] Completed task 1
- [ ] Pending task 2
- [x] Completed task 3
- [ ] Pending task 4
- [ ] Pending task 5
`

	tasks := ParseTasks(content)

	if len(tasks) != 5 {
		t.Fatalf("ParseTasks() returned %d tasks, want 5", len(tasks))
	}

	total, completed, pending := CountTasks(tasks)

	if total != 5 {
		t.Errorf("CountTasks() total = %d, want 5", total)
	}
	if completed != 2 {
		t.Errorf("CountTasks() completed = %d, want 2", completed)
	}
	if pending != 3 {
		t.Errorf("CountTasks() pending = %d, want 3", pending)
	}
}

// TestGroupByPriority tests that GroupByPriority correctly groups tasks by priority.
func TestGroupByPriority(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Priority: "P1"},
		{Text: "Task 2", Priority: "P2"},
		{Text: "Task 3", Priority: "P1"},
		{Text: "Task 4", Priority: "P3"},
		{Text: "Task 5", Priority: ""},
		{Text: "Task 6", Priority: "P2"},
	}

	groups := GroupByPriority(tasks)

	// Check P1 group
	p1Group, exists := groups["P1"]
	if !exists {
		t.Error("GroupByPriority() missing P1 group")
	}
	if len(p1Group) != 2 {
		t.Errorf("P1 group has %d tasks, want 2", len(p1Group))
	}

	// Check P2 group
	p2Group, exists := groups["P2"]
	if !exists {
		t.Error("GroupByPriority() missing P2 group")
	}
	if len(p2Group) != 2 {
		t.Errorf("P2 group has %d tasks, want 2", len(p2Group))
	}

	// Check P3 group
	p3Group, exists := groups["P3"]
	if !exists {
		t.Error("GroupByPriority() missing P3 group")
	}
	if len(p3Group) != 1 {
		t.Errorf("P3 group has %d tasks, want 1", len(p3Group))
	}

	// Check "None" group for empty priority
	noneGroup, exists := groups["None"]
	if !exists {
		t.Error("GroupByPriority() missing None group for empty priority")
	}
	if len(noneGroup) != 1 {
		t.Errorf("None group has %d tasks, want 1", len(noneGroup))
	}
	if noneGroup[0].Text != "Task 5" {
		t.Errorf("None group task text = %q, want %q", noneGroup[0].Text, "Task 5")
	}

	// Check total count
	totalCount := 0
	for _, group := range groups {
		totalCount += len(group)
	}
	if totalCount != 6 {
		t.Errorf("Total tasks in all groups = %d, want 6", totalCount)
	}
}

// TestGroupByPriority_EmptySlice tests GroupByPriority with empty task slice.
func TestGroupByPriority_EmptySlice(t *testing.T) {
	groups := GroupByPriority([]Task{})

	if len(groups) != 0 {
		t.Errorf("GroupByPriority() returned %d groups, want 0", len(groups))
	}
}

// TestGroupByPriority_AllEmptyPriority tests GroupByPriority when all tasks have empty priority.
func TestGroupByPriority_AllEmptyPriority(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Priority: ""},
		{Text: "Task 2", Priority: ""},
	}

	groups := GroupByPriority(tasks)

	noneGroup, exists := groups["None"]
	if !exists {
		t.Error("GroupByPriority() missing None group")
	}
	if len(noneGroup) != 2 {
		t.Errorf("None group has %d tasks, want 2", len(noneGroup))
	}
}

// TestGroupBySection tests that GroupBySection correctly groups tasks by section.
func TestGroupBySection(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Section: "Section A"},
		{Text: "Task 2", Section: "Section B"},
		{Text: "Task 3", Section: "Section A"},
		{Text: "Task 4", Section: "Section C"},
		{Text: "Task 5", Section: ""},
		{Text: "Task 6", Section: "Section B"},
	}

	groups := GroupBySection(tasks)

	// Check Section A group
	sectionAGroup, exists := groups["Section A"]
	if !exists {
		t.Error("GroupBySection() missing Section A group")
	}
	if len(sectionAGroup) != 2 {
		t.Errorf("Section A group has %d tasks, want 2", len(sectionAGroup))
	}

	// Check Section B group
	sectionBGroup, exists := groups["Section B"]
	if !exists {
		t.Error("GroupBySection() missing Section B group")
	}
	if len(sectionBGroup) != 2 {
		t.Errorf("Section B group has %d tasks, want 2", len(sectionBGroup))
	}

	// Check Section C group
	sectionCGroup, exists := groups["Section C"]
	if !exists {
		t.Error("GroupBySection() missing Section C group")
	}
	if len(sectionCGroup) != 1 {
		t.Errorf("Section C group has %d tasks, want 1", len(sectionCGroup))
	}

	// Check "Uncategorized" group for empty section
	uncategorizedGroup, exists := groups["Uncategorized"]
	if !exists {
		t.Error("GroupBySection() missing Uncategorized group for empty section")
	}
	if len(uncategorizedGroup) != 1 {
		t.Errorf("Uncategorized group has %d tasks, want 1", len(uncategorizedGroup))
	}
	if uncategorizedGroup[0].Text != "Task 5" {
		t.Errorf("Uncategorized group task text = %q, want %q", uncategorizedGroup[0].Text, "Task 5")
	}

	// Check total count
	totalCount := 0
	for _, group := range groups {
		totalCount += len(group)
	}
	if totalCount != 6 {
		t.Errorf("Total tasks in all groups = %d, want 6", totalCount)
	}
}

// TestGroupBySection_EmptySlice tests GroupBySection with empty task slice.
func TestGroupBySection_EmptySlice(t *testing.T) {
	groups := GroupBySection([]Task{})

	if len(groups) != 0 {
		t.Errorf("GroupBySection() returned %d groups, want 0", len(groups))
	}
}

// TestGroupBySection_AllEmptySection tests GroupBySection when all tasks have empty section.
func TestGroupBySection_AllEmptySection(t *testing.T) {
	tasks := []Task{
		{Text: "Task 1", Section: ""},
		{Text: "Task 2", Section: ""},
	}

	groups := GroupBySection(tasks)

	uncategorizedGroup, exists := groups["Uncategorized"]
	if !exists {
		t.Error("GroupBySection() missing Uncategorized group")
	}
	if len(uncategorizedGroup) != 2 {
		t.Errorf("Uncategorized group has %d tasks, want 2", len(uncategorizedGroup))
	}
}

// TestGroupBySection_IntegrationWithParseTasks tests GroupBySection with parsed tasks.
func TestGroupBySection_IntegrationWithParseTasks(t *testing.T) {
	content := `# Tasks

- [ ] Task in Tasks section
- [ ] Another Task in Tasks section

## Section A

- [ ] Task in Section A

## Section B

- [ ] Task in Section B
- [ ] Another Task in Section B
`

	tasks := ParseTasks(content)

	groups := GroupBySection(tasks)

	// Check Uncategorized section (empty section for tasks before first ## heading)
	uncategorizedGroup, exists := groups["Uncategorized"]
	if !exists {
		t.Error("GroupBySection() missing Uncategorized group for tasks before first ## heading")
	}
	if len(uncategorizedGroup) != 2 {
		t.Errorf("Uncategorized group has %d tasks, want 2", len(uncategorizedGroup))
	}

	// Check Section A group
	sectionAGroup, exists := groups["Section A"]
	if !exists {
		t.Error("GroupBySection() missing Section A group")
	}
	if len(sectionAGroup) != 1 {
		t.Errorf("Section A group has %d tasks, want 1", len(sectionAGroup))
	}

	// Check Section B group
	sectionBGroup, exists := groups["Section B"]
	if !exists {
		t.Error("GroupBySection() missing Section B group")
	}
	if len(sectionBGroup) != 2 {
		t.Errorf("Section B group has %d tasks, want 2", len(sectionBGroup))
	}

	// Check total count
	totalCount := 0
	for _, group := range groups {
		totalCount += len(group)
	}
	if totalCount != 5 {
		t.Errorf("Total tasks in all groups = %d, want 5", totalCount)
	}
}
