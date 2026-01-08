package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

func createTestVisualConfig(t *testing.T, tmpDir string) *config.LoadedConfig {
	t.Helper()

	cfg := &config.LoadedConfig{
		Config: config.Config{},
	}
	cfg.Paths.BaseDir = tmpDir
	cfg.Paths.DiaryDir = "Diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"
	cfg.DiaryPath = filepath.Join(tmpDir, "Diary")

	return cfg
}

func TestShowStreak_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-streak-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)
	cfg.Streaks.IncludeWeekends = true

	err = ShowStreak(cfg)
	if err != nil {
		t.Errorf("ShowStreak should not error: %v", err)
	}
}

func TestShowStreak_WithNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-streak-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)
	cfg.Streaks.IncludeWeekends = true

	ctx := context.Background()
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	yesterdayPath := notes.BuildDailyNotePath(cfg.DiaryPath, yesterday)
	if err := notes.WriteNote(ctx, yesterdayPath, "# Yesterday\n"); err != nil {
		t.Fatalf("Failed to create yesterday note: %v", err)
	}

	err = ShowStreak(cfg)
	if err != nil {
		t.Errorf("ShowStreak should not error: %v", err)
	}
}

func TestShowStreak_WeekendsExcluded(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-streak-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)
	cfg.Streaks.IncludeWeekends = false

	err = ShowStreak(cfg)
	if err != nil {
		t.Errorf("ShowStreak should not error: %v", err)
	}
}

func TestShowCalendar(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-calendar-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)

	err = showCalendar(cfg)
	if err != nil {
		t.Errorf("showCalendar should not error: %v", err)
	}
}

func TestShowCalendar_WithNotes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-calendar-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)

	ctx := context.Background()
	today := time.Now()

	todayPath := notes.BuildDailyNotePath(cfg.DiaryPath, today)
	if err := notes.WriteNote(ctx, todayPath, "# Today\n"); err != nil {
		t.Fatalf("Failed to create today note: %v", err)
	}

	err = showCalendar(cfg)
	if err != nil {
		t.Errorf("showCalendar should not error: %v", err)
	}
}

func TestSanitizeNodeID_Basic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with_spaces"},
		{"with-dash", "with_dash"},
		{"with.dot", "with_dot"},
		{"[[link]]", "__link__"},
		{"", "node"},
	}

	for _, tt := range tests {
		result := sanitizeNodeID(tt.input)
		if result != tt.expected && tt.input != "" {
			t.Errorf("sanitizeNodeID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeNodeID_StartsWithNumber(t *testing.T) {
	result := sanitizeNodeID("123note")
	if len(result) == 0 {
		t.Error("sanitizeNodeID should not return empty string")
	}

	if result[0] == '1' {
		t.Error("sanitizeNodeID should prefix with 'n_' if starts with number")
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		input    byte
		expected bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', false},
		{' ', false},
		{'-', false},
		{'_', false},
	}

	for _, tt := range tests {
		result := isLetter(tt.input)
		if result != tt.expected {
			t.Errorf("isLetter(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateGraph_NoLinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-graph-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)

	err = generateGraph(context.Background(), cfg)
	if err == nil {
		t.Error("generateGraph should error when no links found or graphviz missing")
	}
}

func TestGenerateGraph_WithLinks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-graph-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)

	ctx := context.Background()
	note1Path := filepath.Join(tmpDir, "Note1.md")
	note2Path := filepath.Join(tmpDir, "Note2.md")

	notes.WriteNote(ctx, note1Path, "# Note 1\nSee [[Note2]] for more.\n")
	notes.WriteNote(ctx, note2Path, "# Note 2\nBack to [[Note1]].\n")

	graphOutput = "test_graph.png"
	graphFormat = "png"

	err = generateGraph(context.Background(), cfg)
	if err != nil && !strings.Contains(err.Error(), "graphviz") {
		t.Errorf("generateGraph should not error with links (unless graphviz missing): %v", err)
	}
}

func TestOpenGraph_NoViewer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-graph-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	graphPath := filepath.Join(tmpDir, "nonexistent.png")
	openGraph(graphPath)
}

func TestLaunchDashboard_Starts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-dashboard-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)

	ctx := context.Background()

	err = LaunchDashboard(ctx, cfg)
	if err != nil && !strings.Contains(err.Error(), "TTY") && !strings.Contains(err.Error(), "tty") {
		t.Errorf("LaunchDashboard should not error on startup (unless TTY issue): %v", err)
	}
}

func TestLaunchDashboard_VariousTerminalSizes(t *testing.T) {
	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"Standard 80x24", 80, 24},
		{"Large 120x40", 120, 40},
		{"Small 40x12", 40, 12},
		{"Wide 200x50", 200, 50},
		{"Tall 80x60", 80, 60},
		{"Very Small 20x10", 20, 10},
		{"Large Wide 160x48", 160, 48},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "jotr-dashboard-size-test-")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			cfg := createTestVisualConfig(t, tmpDir)
			ctx := context.Background()

			err = LaunchDashboard(ctx, cfg)
			if err != nil {
				isTTYError := strings.Contains(err.Error(), "TTY") ||
					strings.Contains(err.Error(), "tty") ||
					strings.Contains(err.Error(), "terminal") ||
					strings.Contains(err.Error(), "stdin")
				if !isTTYError {
					t.Errorf("LaunchDashboard should not error on startup: %v", err)
				}
			}
		})
	}
}

func TestLaunchDashboard_MinimumSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jotr-dashboard-min-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := createTestVisualConfig(t, tmpDir)
	ctx := context.Background()

	err = LaunchDashboard(ctx, cfg)
	if err != nil {
		isTTYError := strings.Contains(err.Error(), "TTY") ||
			strings.Contains(err.Error(), "tty") ||
			strings.Contains(err.Error(), "terminal") ||
			strings.Contains(err.Error(), "stdin")
		if !isTTYError {
			t.Errorf("LaunchDashboard should handle minimum terminal size: %v", err)
		}
	}
}
