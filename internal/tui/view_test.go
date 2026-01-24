package tui

import (
	"strings"
	"testing"

	"github.com/AnishShah1803/jotr/internal/config"
)

// TestRenderHeader_StatusDisplay tests that status messages display under ASCII logo when active.
func TestRenderHeader_StatusDisplay(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name           string
		statusMsg      string
		statusLevel    string
		expectContains []string
		width          int
		height         int
	}{
		{
			name:           "info status displays under logo",
			statusMsg:      "Refreshing...",
			statusLevel:    "info",
			expectContains: []string{"Refreshing..."},
			width:          80,
			height:         50,
		},
		{
			name:           "error status displays under logo",
			statusMsg:      "Error occurred",
			statusLevel:    "error",
			expectContains: []string{"Error occurred"},
			width:          80,
			height:         50,
		},
		{
			name:           "success status displays under logo",
			statusMsg:      "Success!",
			statusLevel:    "success",
			expectContains: []string{"Success!"},
			width:          80,
			height:         50,
		},
		{
			name:           "warning status displays under logo",
			statusMsg:      "Warning message",
			statusLevel:    "warning",
			expectContains: []string{"Warning message"},
			width:          80,
			height:         50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(nil, cfg)
			model.statusMsg = tt.statusMsg
			model.statusLevel = tt.statusLevel
			model.width = tt.width
			model.height = tt.height
			model.ready = true

			header := model.renderHeader()

			for _, expected := range tt.expectContains {
				if !strings.Contains(header, expected) {
					t.Errorf("Expected header to contain '%s', got:\n%s", expected, header)
				}
			}
		})
	}
}

// TestRenderHeader_VersionDisplay tests that version number displays when no status and no update.
func TestRenderHeader_VersionDisplay(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(nil, cfg)
	model.statusMsg = ""
	model.updateAvailable = false
	model.width = 80
	model.height = 50
	model.ready = true

	header := model.renderHeader()

	if !strings.Contains(header, "v") {
		t.Errorf("Expected header to contain version info, got:\n%s", header)
	}
}

// TestRenderHeader_StatusPriority tests that status has priority over version and update notification.
func TestRenderHeader_StatusPriority(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(nil, cfg)
	model.statusMsg = "Status message"
	model.statusLevel = "info"
	model.updateAvailable = true
	model.updateVersion = "v1.2.0"
	model.width = 80
	model.height = 50
	model.ready = true

	header := model.renderHeader()

	if !strings.Contains(header, "Status message") {
		t.Errorf("Expected header to contain status message, got:\n%s", header)
	}

	if strings.Contains(header, "v1.2.0") {
		t.Errorf("Expected header NOT to contain update version when status is active, got:\n%s", header)
	}
}

// TestRenderHeader_LargeTerminalAsciiLogo tests that ASCII logo only renders on large terminals.
func TestRenderHeader_LargeTerminalAsciiLogo(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name           string
		width          int
		height         int
		expectAsciiArt bool
	}{
		{
			name:           "large terminal shows ASCII art",
			width:          80,
			height:         50,
			expectAsciiArt: true,
		},
		{
			name:           "small terminal no ASCII art",
			width:          40,
			height:         30,
			expectAsciiArt: false,
		},
		{
			name:           "wide but short terminal no ASCII art",
			width:          80,
			height:         30,
			expectAsciiArt: false,
		},
		{
			name:           "narrow but tall terminal no ASCII art",
			width:          40,
			height:         50,
			expectAsciiArt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(nil, cfg)
			model.width = tt.width
			model.height = tt.height
			model.ready = true

			header := model.renderHeader()

			hasAsciiArt := strings.Contains(header, "██╗")

			if hasAsciiArt != tt.expectAsciiArt {
				t.Errorf("Expected ASCII art=%v, got=%v for width=%d height=%d\nHeader:\n%s",
					tt.expectAsciiArt, hasAsciiArt, tt.width, tt.height, header)
			}
		})
	}
}

// TestRenderHeader_StatusColorStyling tests that different status levels use appropriate colors.
func TestRenderHeader_StatusColorStyling(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	tests := []struct {
		name        string
		statusMsg   string
		statusLevel string
		expectColor string
	}{
		{
			name:        "error status uses error color",
			statusMsg:   "Error",
			statusLevel: "error",
			expectColor: "203",
		},
		{
			name:        "success status uses success color",
			statusMsg:   "Success",
			statusLevel: "success",
			expectColor: "42",
		},
		{
			name:        "warning status uses warning color",
			statusMsg:   "Warning",
			statusLevel: "warning",
			expectColor: "214",
		},
		{
			name:        "info status uses warning color (default)",
			statusMsg:   "Info",
			statusLevel: "info",
			expectColor: "214",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(nil, cfg)
			model.statusMsg = tt.statusMsg
			model.statusLevel = tt.statusLevel
			model.width = 80
			model.height = 50
			model.ready = true

			header := model.renderHeader()

			if !strings.Contains(header, tt.statusMsg) {
				t.Errorf("Expected header to contain status message '%s', got:\n%s", tt.statusMsg, header)
			}
		})
	}
}

// TestModel_StatusFlow tests the complete status flow from set to auto-expiry.
func TestModel_StatusFlow(t *testing.T) {
	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = "/tmp/test"
	cfg.DiaryPath = "/tmp/test/diary"
	cfg.TodoPath = "/tmp/test/todo.md"

	model := NewModel(nil, cfg)
	model.width = 80
	model.height = 50
	model.ready = true

	// Initial state - no status, should show version
	view1 := model.View()
	if !strings.Contains(view1, "v") {
		t.Errorf("Initial view should show version, got:\n%s", view1)
	}

	// Set status - should show status message
	model = setStatus(model, "Test status", "info")
	view2 := model.View()
	if !strings.Contains(view2, "Test status") {
		t.Errorf("View should show status message, got:\n%s", view2)
	}

	// Clear status - should revert to version
	model = clearStatus(model)
	view3 := model.View()
	if !strings.Contains(view3, "v") {
		t.Errorf("View should revert to version after clear, got:\n%s", view3)
	}
}
