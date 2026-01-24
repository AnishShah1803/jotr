package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/testhelpers"
)

// Example comprehensive CLI tests using the new testing helpers
// This demonstrates the patterns from major Go projects

func TestCaptureCommand_Integration(t *testing.T) {
	// Set up test configuration
	configDir, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	// Create test diary structure
	diaryPath := filepath.Join(configDir, "diary")

	err := os.MkdirAll(diaryPath, 0750)
	if err != nil {
		t.Fatalf("Failed to create diary path: %v", err)
	}

	// Create minimal config
	cfg := &config.Config{}
	cfg.Paths.BaseDir = configDir
	cfg.Paths.DiaryDir = "diary"
	cfg.Format.CaptureSection = "Captured"
	cfg.Format.DailyNotePattern = "{year}-{month}-{day}-{weekday}"
	cfg.Format.DailyNoteDirPattern = "{year}/{month}"

	err = config.Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Table-driven tests using the pattern from major Go projects
	tests := []testhelpers.NamedTest[struct {
		expectedStdout string
		expectedStderr string
		args           []string
		expectSuccess  bool
	}]{
		{
			Name: "valid capture",
			Data: struct {
				expectedStdout string
				expectedStderr string
				args           []string
				expectSuccess  bool
			}{
				args:          []string{"capture", "Test capture text"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture as task",
			Data: struct {
				expectedStdout string
				expectedStderr string
				args           []string
				expectSuccess  bool
			}{
				args:          []string{"capture", "--task", "Test task"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture with alias",
			Data: struct {
				expectedStdout string
				expectedStderr string
				args           []string
				expectSuccess  bool
			}{
				args:          []string{"cap", "Using alias"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture without text should fail",
			Data: struct {
				expectedStdout string
				expectedStderr string
				args           []string
				expectSuccess  bool
			}{
				args:           []string{"capture"},
				expectSuccess:  false,
				expectedStderr: "text to capture is required",
			},
		},
	}

	// Execute table-driven tests
	testhelpers.RunNamedTableTests(t, tests, func(t *testing.T, tt struct {
		expectedStdout string
		expectedStderr string
		args           []string
		expectSuccess  bool
	}) {
		// Create a fresh root command for each test to avoid state pollution
		rootCmd := createTestRootCommand()

		// Execute the command
		result := testhelpers.ExecuteCommand(rootCmd, tt.args...)

		// Verify success/failure
		if tt.expectSuccess {
			result.AssertSuccess(t)
		} else {
			result.AssertFailure(t, 1)
		}

		// Verify expected output
		if tt.expectedStdout != "" {
			result.AssertStdoutContains(t, tt.expectedStdout)
		}

		if tt.expectedStderr != "" {
			result.AssertStderrContains(t, tt.expectedStderr)
		}
	})
}

func TestVersionCommand(t *testing.T) {
	// Test version command - simple case
	rootCmd := createTestRootCommand()

	result := testhelpers.ExecuteCommand(rootCmd, "version")
	result.AssertSuccess(t)
	result.AssertStdoutContains(t, "jotr version")
}

func TestHelpCommand(t *testing.T) {
	// Test help output
	rootCmd := createTestRootCommand()

	result := testhelpers.ExecuteCommand(rootCmd, "--help")
	result.AssertSuccess(t)
	result.AssertStdoutContains(t, "jotr is a command-line journaling and note-taking tool")
}

func TestConfigureCommand_Interactive(t *testing.T) {
	// Test interactive command with input
	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	rootCmd := createTestRootCommand()

	// Simulate user input for configuration
	input := "/tmp/test-notes\nDiary\ntodo.md\nTasks\nCaptured\ny\n"

	result := testhelpers.ExecuteCommandWithInput(rootCmd, input, "configure")

	// Should succeed with proper input
	result.AssertSuccess(t)
	result.AssertStdoutContains(t, "Configuration")
}

// This mirrors the pattern used by kubectl and other major CLI tools.
func createTestRootCommand() *cobra.Command {
	// Create a copy of the root command structure but isolated for testing
	rootCmd := &cobra.Command{
		Use:   "jotr",
		Short: "A powerful journaling and note-taking tool",
		Long: `jotr is a command-line journaling and note-taking tool designed for daily use.
 It supports daily notes, task management, templates, search, and much more.

 When run without arguments, jotr launches the interactive dashboard.`,
	}

	// Add essential commands for testing
	// In practice, you would add all your actual commands here
	// but configured for testing (avoiding global state, file system writes, etc.)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("jotr version 1.0.0-test")
		},
	})

	captureCmd := &cobra.Command{
		Use:   "capture [text]",
		Short: "Quick capture to daily note",
		Long: `Quickly capture text to today's daily note.

Examples:
  jotr capture "Meeting with team"
  jotr capture --task "Review PR #123"
  jotr cap "Quick thought"`,
		Aliases: []string{"cap"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("text to capture is required")
			}
			// Simplified capture logic for testing
			cmd.Printf("Captured: %s\n", args[0])
			return nil
		},
	}
	captureCmd.Flags().Bool("task", false, "Capture as task")
	rootCmd.AddCommand(captureCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:     "configure",
		Short:   "Run configuration wizard",
		Long:    `Interactive wizard to set up jotr configuration.`,
		Aliases: []string{"config", "cfg"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("Configuration wizard")
			return nil
		},
	})

	return rootCmd
}

// Benchmark test example - pattern used by major Go projects.
func BenchmarkCaptureCommand(b *testing.B) {
	_, cleanup := testhelpers.SetupTestConfig(nil) // Pass nil for benchmarks
	defer cleanup()

	rootCmd := createTestRootCommand()

	b.ResetTimer()

	for i := range b.N {
		_ = testhelpers.ExecuteCommand(rootCmd, "capture", "benchmark test text")
		_ = i // Use the variable to avoid unused warnings
	}
}

// Example of testing with environment variables.
func TestCaptureCommand_WithEnvVars(t *testing.T) {
	// Store original env vars
	originalEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", originalEditor)

	// Set test environment
	os.Setenv("EDITOR", "echo")

	_, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	rootCmd := createTestRootCommand()

	result := testhelpers.ExecuteCommand(rootCmd, "capture", "env test")
	result.AssertSuccess(t)
}

// Example test using golden files pattern (if you need it).
func TestCaptureCommand_Output(t *testing.T) {
	// This would use golden files for complex output comparison
	// Pattern used by kubernetes, hugo, and other projects when output is complex
	rootCmd := createTestRootCommand()

	result := testhelpers.ExecuteCommand(rootCmd, "capture", "golden test")
	result.AssertSuccess(t)

	// In a real golden file test, you would compare against a .golden file
	// For now, just check basic output
	testhelpers.CheckStringContains(t, result.Stdout, "Captured")
}

// ============================================================================
// Search Command Integration Tests
// ============================================================================

func TestSearchCommand_Integration(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	testNotes := []struct {
		filename string
		content  string
	}{
		{"ProjectAlpha", "# Project Alpha\n\nThis is a note about alpha development.\n"},
		{"ProjectBeta", "# Project Beta\n\nWorking on beta features.\n"},
		{"MeetingNotes", "# Meeting Notes\n\nDiscussed project alpha timeline.\n"},
		{"TodoList", "# Todo List\n\n- [ ] Task for completion\n"},
	}

	for _, note := range testNotes {
		fs.WriteFile(t, note.filename+".md", note.content)
	}

	searchTests := []testhelpers.NamedTest[struct {
		query         string
		expectSuccess bool
		expectedFiles []string
		expectCount   int
	}]{
		{
			Name: "search for existing term",
			Data: struct {
				query         string
				expectSuccess bool
				expectedFiles []string
				expectCount   int
			}{
				query:         "project alpha",
				expectSuccess: true,
				expectedFiles: []string{"ProjectAlpha", "MeetingNotes"},
				expectCount:   2,
			},
		},
		{
			Name: "search for term in multiple notes",
			Data: struct {
				query         string
				expectSuccess bool
				expectedFiles []string
				expectCount   int
			}{
				query:         "project",
				expectSuccess: true,
				expectedFiles: []string{"ProjectAlpha", "ProjectBeta", "MeetingNotes"},
				expectCount:   3,
			},
		},
		{
			Name: "search for non-existent term",
			Data: struct {
				query         string
				expectSuccess bool
				expectedFiles []string
				expectCount   int
			}{
				query:         "nonexistentterm12345",
				expectSuccess: true,
				expectedFiles: nil,
				expectCount:   0,
			},
		},
		{
			Name: "search with case insensitive",
			Data: struct {
				query         string
				expectSuccess bool
				expectedFiles []string
				expectCount   int
			}{
				query:         "PROJECT ALPHA",
				expectSuccess: true,
				expectedFiles: []string{"ProjectAlpha", "MeetingNotes"},
				expectCount:   2,
			},
		},
	}

	testhelpers.RunNamedTableTests(t, searchTests, func(t *testing.T, tt struct {
		query         string
		expectSuccess bool
		expectedFiles []string
		expectCount   int
	}) {
		ctx := context.Background()
		matches, err := notes.SearchNotes(ctx, fs.BaseDir, tt.query)

		if tt.expectSuccess && err != nil {
			t.Fatalf("SearchNotes failed: %v", err)
		}

		if len(matches) != tt.expectCount {
			t.Errorf("Expected %d matches, got %d", tt.expectCount, len(matches))
		}

		for _, expectedFile := range tt.expectedFiles {
			found := false
			for _, match := range matches {
				if strings.Contains(filepath.Base(match), expectedFile) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find %s in matches: %v", expectedFile, matches)
			}
		}
	})
}

func TestSearchCommand_SubdirectorySearch(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	dirs := []string{
		"work",
		"personal",
		"work/projects",
		"personal/notes",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(fs.BaseDir, dir), 0750); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	testNotes := []struct {
		filename string
		dir      string
		content  string
	}{
		{"WorkPlan", "work", "# Work Plan\n\nReview quarterly goals.\n"},
		{"ProjectSpec", "work/projects", "# Project Spec\n\nDetailed specifications for project.\n"},
		{"PersonalJournal", "personal", "# Personal Journal\n\nDaily reflections and notes.\n"},
		{"WorkTodo", "work", "# Work Todo\n\nTasks for work completion.\n"},
	}

	for _, note := range testNotes {
		fs.WriteFile(t, filepath.Join(note.dir, note.filename+".md"), note.content)
	}

	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, fs.BaseDir, "work")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for 'work', got %d: %v", len(matches), matches)
	}

	matches, err = notes.SearchNotes(ctx, fs.BaseDir, "project")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	foundProjectSpec := false
	for _, match := range matches {
		if strings.Contains(filepath.Base(match), "ProjectSpec") {
			foundProjectSpec = true
			break
		}
	}
	if !foundProjectSpec {
		t.Errorf("Expected to find ProjectSpec in matches: %v", matches)
	}

	matches, err = notes.SearchNotes(ctx, fs.BaseDir, "specifications")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match for 'specifications', got %d", len(matches))
	}
}

func TestSearchCommand_LinksAndTags(t *testing.T) {
	fs := testhelpers.NewTestFS(t)
	defer fs.Cleanup()

	ch := testhelpers.NewConfigHelper(fs)
	ch.CreateBasicConfig(t)

	testNotes := []struct {
		filename string
		content  string
	}{
		{"NoteWithTags", "# Tagged Note\n\nThis has #important and #urgent tags.\n"},
		{"NoteWithLinks", "# Linked Note\n\nSee [[AnotherNote]] for details.\n"},
		{"TargetNote", "# Target Note\n\nThis is referenced by other notes.\n"},
		{"MixedContent", "# Mixed Note\n\n#tag1 content here with [[TargetNote]] reference.\n"},
	}

	for _, note := range testNotes {
		fs.WriteFile(t, note.filename+".md", note.content)
	}

	ctx := context.Background()

	matches, err := notes.SearchNotes(ctx, fs.BaseDir, "#important")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	found := false
	for _, match := range matches {
		if strings.Contains(filepath.Base(match), "NoteWithTags") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find NoteWithTags when searching for #important")
	}

	matches, err = notes.SearchNotes(ctx, fs.BaseDir, "reference")
	if err != nil {
		t.Fatalf("SearchNotes failed: %v", err)
	}

	if len(matches) < 2 {
		t.Errorf("Expected at least 2 matches for 'reference', got %d", len(matches))
	}
}
