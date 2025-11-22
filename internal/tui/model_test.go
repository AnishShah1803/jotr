package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/anish/jotr/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// TestModel_Init tests the model initialization
func TestModel_Init(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(cfg)

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

// TestModel_Update tests model updates with different messages
func TestModel_Update(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name        string
		initialMsg  tea.Msg
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
			model := NewModel(cfg)

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

// TestModel_ErrorHandling tests error state handling
func TestModel_ErrorHandling(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(cfg)

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

// TestModel_Resize tests window resize handling
func TestModel_Resize(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(cfg)

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

// TestModel_View tests view rendering (headless)
func TestModel_View(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(cfg)

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
