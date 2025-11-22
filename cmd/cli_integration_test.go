package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/testhelpers"
	"github.com/spf13/cobra"
)

// Example comprehensive CLI tests using the new testing helpers
// This demonstrates the patterns from major Go projects

func TestCaptureCommand_Integration(t *testing.T) {
	// Set up test configuration
	configDir, cleanup := testhelpers.SetupTestConfig(t)
	defer cleanup()

	// Create test diary structure
	diaryPath := filepath.Join(configDir, "diary")
	err := os.MkdirAll(diaryPath, 0755)
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
		args           []string
		expectSuccess  bool
		expectedStdout string
		expectedStderr string
	}]{
		{
			Name: "valid capture",
			Data: struct {
				args           []string
				expectSuccess  bool
				expectedStdout string
				expectedStderr string
			}{
				args:          []string{"capture", "Test capture text"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture as task",
			Data: struct {
				args           []string
				expectSuccess  bool
				expectedStdout string
				expectedStderr string
			}{
				args:          []string{"capture", "--task", "Test task"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture with alias",
			Data: struct {
				args           []string
				expectSuccess  bool
				expectedStdout string
				expectedStderr string
			}{
				args:          []string{"cap", "Using alias"},
				expectSuccess: true,
			},
		},
		{
			Name: "capture without text should fail",
			Data: struct {
				args           []string
				expectSuccess  bool
				expectedStdout string
				expectedStderr string
			}{
				args:           []string{"capture"},
				expectSuccess:  false,
				expectedStderr: "text to capture is required",
			},
		},
	}

	// Execute table-driven tests
	testhelpers.RunNamedTableTests(t, tests, func(t *testing.T, tt struct {
		args           []string
		expectSuccess  bool
		expectedStdout string
		expectedStderr string
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
	result.AssertStdoutContains(t, "A powerful journaling and note-taking tool")
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

// createTestRootCommand creates a fresh command tree for testing
// This mirrors the pattern used by kubectl and other major CLI tools
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
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("jotr version 1.0.0-test")
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "capture [text]",
		Short: "Quick capture to daily note",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simplified capture logic for testing
			cmd.Printf("Captured: %s\n", args[0])
			return nil
		},
	})

	return rootCmd
}

// Benchmark test example - pattern used by major Go projects
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

// Example of testing with environment variables
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

// Example test using golden files pattern (if you need it)
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
