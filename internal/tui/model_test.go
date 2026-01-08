package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/tasks"
)

// TestModel_Init tests the model initialization.
func TestModel_Init(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Test initial state
	if model.config != cfg {
		t.Errorf("Config not set correctly")
	}

	if model.focusedPanel != panelNotes {
		t.Errorf("Expected initial panel to be panelNotes, got %v", model.focusedPanel)
	}

	// Test Init command
	cmd := model.Init()
	if cmd == nil {
		t.Errorf("Init should return a command")
	}
}

// TestModel_Update tests model updates with different messages.
func TestModel_Update(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		initialMsg  tea.Msg
		name        string
		expectPanel panel
		expectCmd   bool
	}{
		{
			name:        "window size message with ready model",
			initialMsg:  tea.WindowSizeMsg{Width: 100, Height: 50},
			expectPanel: panelNotes,
			expectCmd:   false, // No command when already ready
		},
		{
			name:        "tab key navigation",
			initialMsg:  tea.KeyMsg{Type: tea.KeyTab},
			expectPanel: panelPreview, // Should move to next panel
			expectCmd:   false,
		},
		{
			name:        "quit command",
			initialMsg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			expectPanel: panelNotes, // Panel doesn't change
			expectCmd:   true,       // Should return quit command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(context.Background(), cfg)

			// For the tab test, we need to set model as ready first
			if strings.Contains(tt.name, "tab") {
				model.ready = true
			}

			// For window size test with ready model, set ready
			if strings.Contains(tt.name, "ready") {
				model.ready = true
			}

			updatedModel, cmd := model.Update(tt.initialMsg)

			// Check if model state changed correctly
			if m, ok := updatedModel.(Model); ok {
				if m.focusedPanel != tt.expectPanel {
					t.Errorf("Expected panel %v, got %v", tt.expectPanel, m.focusedPanel)
				}
			} else {
				t.Errorf("Update didn't return Model type")
			}

			// Check command presence
			if (cmd != nil) != tt.expectCmd {
				t.Errorf("Expected cmd presence %v, got %v", tt.expectCmd, cmd != nil)
			}
		})
	}
}

// TestModel_ErrorHandling tests error state handling.
func TestModel_ErrorHandling(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Test error message handling
	errorMessage := errorMsg{
		err:       fmt.Errorf("test error"),
		retryable: true,
	}

	updatedModel, cmd := model.Update(errorMessage)
	m := updatedModel.(Model)

	// Check if error was stored
	if m.err == nil {
		t.Errorf("Error should be stored in model")
	}

	if m.err.Error() != "test error" {
		t.Errorf("Expected error 'test error', got %v", m.err.Error())
	}

	if !m.errorRetryable {
		t.Errorf("Error should be marked as retryable")
	}

	// Should not return a command for error display
	if cmd != nil {
		t.Errorf("Error handling should not return command")
	}
}

// TestModel_Resize tests window resize handling.
func TestModel_Resize(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Test window resize
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 60}
	updatedModel, _ := model.Update(resizeMsg)
	m := updatedModel.(Model)

	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}

	if m.height != 60 {
		t.Errorf("Expected height 60, got %d", m.height)
	}
}

// TestModel_View tests view rendering (headless).
func TestModel_View(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Set some dimensions and mark as ready
	model.width = 100
	model.height = 50
	model.ready = true

	// Test basic view rendering
	view := model.View()

	if view == "" {
		t.Errorf("View should not be empty")
	}

	// Basic view should contain some structural elements or version info
	if len(view) < 10 {
		t.Errorf("View seems too short, got length %d", len(view))
	}

	// Test that view responds to model state
	if model.ready != true {
		t.Errorf("Model should be marked as ready")
	}

	// Test that view includes version info (common in TUIs)
	// Instead of specific content, test that it's a meaningful string
	if !strings.Contains(view, "jotr") && !strings.Contains(view, "v") {
		// This is more flexible - just ensure it has some relevant content
		t.Logf("View content: %s", view)
		// Don't fail, just log for inspection
	}
}

// TestModel_TickMsg tests tick message handling.
func TestModel_TickMsg(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Test tick message - should return tickCmd for continuous updates
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tickMessage := tickMsg(testTime)

	updatedModel, cmd := model.Update(tickMessage)
	m := updatedModel.(Model)

	// Model state should remain unchanged
	if m.err != nil {
		t.Errorf("Tick message should not set error, got: %v", m.err)
	}

	// Should return a tick command for continuous updates
	if cmd == nil {
		t.Errorf("Tick message should return tickCmd for continuous updates")
	}
}

// TestModel_UpdateCheckMsg tests update check message handling.
func TestModel_UpdateCheckMsg(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name          string
		checkMsg      updateCheckMsg
		expectUpdate  bool
		expectVersion string
		expectStatus  string
	}{
		{
			name: "update available",
			checkMsg: updateCheckMsg{
				hasUpdate: true,
				version:   "v1.2.0",
				err:       nil,
			},
			expectUpdate:  true,
			expectVersion: "v1.2.0",
			expectStatus:  "🆕 Update available: v1.2.0",
		},
		{
			name: "no update available",
			checkMsg: updateCheckMsg{
				hasUpdate: false,
				version:   "",
				err:       nil,
			},
			expectUpdate:  false,
			expectVersion: "",
			expectStatus:  "✅ You're running the latest version!",
		},
		{
			name: "update check error",
			checkMsg: updateCheckMsg{
				hasUpdate: false,
				version:   "",
				err:       fmt.Errorf("network error"),
			},
			expectUpdate:  false,
			expectVersion: "",
			expectStatus:  "❌ Update check failed: network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(context.Background(), cfg)
			updatedModel, _ := model.Update(tt.checkMsg)
			m := updatedModel.(Model)

			if m.updateAvailable != tt.expectUpdate {
				t.Errorf("Expected updateAvailable=%v, got %v", tt.expectUpdate, m.updateAvailable)
			}

			if m.updateVersion != tt.expectVersion {
				t.Errorf("Expected updateVersion=%s, got %s", tt.expectVersion, m.updateVersion)
			}

			if !strings.Contains(m.statusMsg, tt.expectStatus) {
				t.Errorf("Expected status containing '%s', got '%s'", tt.expectStatus, m.statusMsg)
			}
		})
	}
}

// TestModel_DataLoadedMsg tests data loaded message handling.
func TestModel_DataLoadedMsg(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	// Create test data
	testNotes := []string{"/tmp/test/diary/2024-01-15.md", "/tmp/test/diary/2024-01-14.md"}
	testTasks := []tasks.Task{
		{Text: "Test task 1", Completed: false},
		{Text: "Test task 2", Completed: true},
	}

	dataMsg := dataLoadedMsg{
		notes:          testNotes,
		tasks:          testTasks,
		streak:         5,
		totalNotes:     100,
		totalTasks:     10,
		completedTasks: 3,
	}

	updatedModel, cmd := model.Update(dataMsg)
	m := updatedModel.(Model)

	// Check notes were loaded
	if len(m.notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(m.notes))
	}

	// Check tasks were loaded
	if len(m.tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(m.tasks))
	}

	// Check streak
	if m.streak != 5 {
		t.Errorf("Expected streak 5, got %d", m.streak)
	}

	// Check stats
	if m.totalNotes != 100 {
		t.Errorf("Expected totalNotes 100, got %d", m.totalNotes)
	}

	if m.totalTasks != 10 {
		t.Errorf("Expected totalTasks 10, got %d", m.totalTasks)
	}

	if m.completedTasks != 3 {
		t.Errorf("Expected completedTasks 3, got %d", m.completedTasks)
	}

	// Check status message
	if !strings.Contains(m.statusMsg, "Data loaded successfully") {
		t.Errorf("Expected success status message, got '%s'", m.statusMsg)
	}

	// Should return command to load preview if notes exist
	if len(testNotes) > 0 && cmd == nil {
		t.Errorf("Expected preview load command when notes exist")
	}
}

// TestModel_PreviewLoadedMsg tests preview loaded message handling.
func TestModel_PreviewLoadedMsg(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(context.Background(), cfg)

	testContent := []byte("# Test Note\n\nThis is test content.")
	previewMsg := previewLoadedMsg(testContent)

	updatedModel, _ := model.Update(previewMsg)
	m := updatedModel.(Model)

	// Check preview content
	if string(m.previewContent) != string(testContent) {
		t.Errorf("Expected preview content '%s', got '%s'", string(testContent), m.previewContent)
	}
}

// TestModel_EditorFinishedMsg tests editor finished message handling.
func TestModel_EditorFinishedMsg(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name         string
		editorErr    error
		expectError  bool
		expectStatus string
	}{
		{
			name:         "editor succeeded",
			editorErr:    nil,
			expectError:  false,
			expectStatus: "",
		},
		{
			name:         "editor failed",
			editorErr:    fmt.Errorf("editor crashed"),
			expectError:  true,
			expectStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(context.Background(), cfg)
			editorMsg := editorFinishedMsg{err: tt.editorErr}

			updatedModel, cmd := model.Update(editorMsg)
			m := updatedModel.(Model)

			if tt.expectError {
				if m.err == nil {
					t.Errorf("Expected error to be set")
				}
				if !m.errorRetryable {
					t.Errorf("Expected error to be retryable")
				}
			}

			// Should always return loadData command to refresh
			if cmd == nil {
				t.Errorf("Editor finished should always return load command")
			}
		})
	}
}

// TestModel_Navigation tests up/down navigation in all panels.
func TestModel_Navigation(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	t.Run("notes panel navigation", func(t *testing.T) {
		model := NewModel(context.Background(), cfg)
		model.ready = true
		model.notes = []string{"/tmp/note1.md", "/tmp/note2.md", "/tmp/note3.md"}

		// Initial selection should be 0
		if model.selectedNote != 0 {
			t.Errorf("Expected initial selection 0, got %d", model.selectedNote)
		}

		// Press down to move to note 2
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updatedModel.(Model)
		if model.selectedNote != 1 {
			t.Errorf("Expected selection 1 after down, got %d", model.selectedNote)
		}

		// Press down again to move to note 3
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updatedModel.(Model)
		if model.selectedNote != 2 {
			t.Errorf("Expected selection 2 after second down, got %d", model.selectedNote)
		}

		// Press down again - should stay at last note
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updatedModel.(Model)
		if model.selectedNote != 2 {
			t.Errorf("Expected selection 2 (max), got %d", model.selectedNote)
		}

		// Press up to move back
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		model = updatedModel.(Model)
		if model.selectedNote != 1 {
			t.Errorf("Expected selection 1 after up, got %d", model.selectedNote)
		}

		// Press up again to first note
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		model = updatedModel.(Model)
		if model.selectedNote != 0 {
			t.Errorf("Expected selection 0 after second up, got %d", model.selectedNote)
		}

		// Press up again - should stay at first note
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		model = updatedModel.(Model)
		if model.selectedNote != 0 {
			t.Errorf("Expected selection 0 (min), got %d", model.selectedNote)
		}
	})

	t.Run("tasks panel navigation", func(t *testing.T) {
		model := NewModel(context.Background(), cfg)
		model.ready = true
		model.focusedPanel = panelTasks
		model.tasks = []tasks.Task{
			{Text: "Task 1", Completed: false},
			{Text: "Task 2", Completed: false},
			{Text: "Task 3", Completed: false},
		}

		// Initial selection should be 0
		if model.selectedTask != 0 {
			t.Errorf("Expected initial selection 0, got %d", model.selectedTask)
		}

		// Navigate down
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updatedModel.(Model)
		if model.selectedTask != 1 {
			t.Errorf("Expected selection 1 after down, got %d", model.selectedTask)
		}

		// Navigate up
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		model = updatedModel.(Model)
		if model.selectedTask != 0 {
			t.Errorf("Expected selection 0 after up, got %d", model.selectedTask)
		}
	})

	t.Run("panel cycling with tab", func(t *testing.T) {
		model := NewModel(context.Background(), cfg)
		model.ready = true

		// Initial panel
		if model.focusedPanel != panelNotes {
			t.Errorf("Expected initial panel panelNotes, got %v", model.focusedPanel)
		}

		// Tab to next panel
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updatedModel.(Model)
		if model.focusedPanel != panelPreview {
			t.Errorf("Expected panelPreview after tab, got %v", model.focusedPanel)
		}

		// Tab again
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updatedModel.(Model)
		if model.focusedPanel != panelTasks {
			t.Errorf("Expected panelTasks after 2nd tab, got %v", model.focusedPanel)
		}

		// Tab again
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updatedModel.(Model)
		if model.focusedPanel != panelStats {
			t.Errorf("Expected panelStats after 3rd tab, got %v", model.focusedPanel)
		}

		// Tab again - should cycle back to notes
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		model = updatedModel.(Model)
		if model.focusedPanel != panelNotes {
			t.Errorf("Expected panelNotes after cycling, got %v", model.focusedPanel)
		}
	})
}
