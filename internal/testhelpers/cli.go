package testhelpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// CLIResult represents the result of executing a CLI command
type CLIResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// ExecuteCommand executes a jotr command with the given arguments
// This is the primary helper function for testing CLI commands
// Pattern used by: kubectl, gh, docker, hugo
func ExecuteCommand(rootCmd *cobra.Command, args ...string) *CLIResult {
	// Create buffers to capture output
	var stdout, stderr bytes.Buffer

	// Configure the command for testing
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)

	// Execute the command
	err := rootCmd.Execute()

	// Determine exit code
	exitCode := 0
	if err != nil {
		exitCode = 1
		// Check if it's an exec.ExitError for proper exit codes
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Error:    err,
	}
}

// ExecuteCommandWithInput executes a command with stdin input
// Useful for testing interactive commands
func ExecuteCommandWithInput(rootCmd *cobra.Command, input string, args ...string) *CLIResult {
	var stdout, stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetIn(strings.NewReader(input))
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Error:    err,
	}
}

// createTestRootCommand creates a fresh root command for testing
// This ensures test isolation and prevents global state issues
func createTestRootCommand() *cobra.Command {
	// Note: This is a placeholder implementation
	// In practice, you would create a fresh command tree based on your actual root command

	return &cobra.Command{
		Use: "jotr",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

// AssertSuccess verifies that a command executed successfully
func (r *CLIResult) AssertSuccess(t *testing.T) {
	t.Helper()
	if r.ExitCode != 0 {
		t.Fatalf("Command failed with exit code %d\nStdout: %s\nStderr: %s\nError: %v",
			r.ExitCode, r.Stdout, r.Stderr, r.Error)
	}
}

// AssertFailure verifies that a command failed with the expected exit code
func (r *CLIResult) AssertFailure(t *testing.T, expectedExitCode int) {
	t.Helper()
	if r.ExitCode != expectedExitCode {
		t.Fatalf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s",
			expectedExitCode, r.ExitCode, r.Stdout, r.Stderr)
	}
}

// AssertStdoutContains verifies stdout contains the expected string
func (r *CLIResult) AssertStdoutContains(t *testing.T, expected string) {
	t.Helper()
	if !strings.Contains(r.Stdout, expected) {
		t.Fatalf("Expected stdout to contain %q, but it didn't.\nStdout: %s",
			expected, r.Stdout)
	}
}

// AssertStderrContains verifies stderr contains the expected string
func (r *CLIResult) AssertStderrContains(t *testing.T, expected string) {
	t.Helper()
	if !strings.Contains(r.Stderr, expected) {
		t.Fatalf("Expected stderr to contain %q, but it didn't.\nStderr: %s",
			expected, r.Stderr)
	}
}

// AssertStdoutEquals verifies stdout exactly matches the expected string
func (r *CLIResult) AssertStdoutEquals(t *testing.T, expected string) {
	t.Helper()
	if strings.TrimSpace(r.Stdout) != strings.TrimSpace(expected) {
		t.Fatalf("Expected stdout to equal:\n%q\nBut got:\n%q",
			expected, r.Stdout)
	}
}

// AssertStdoutEmpty verifies stdout is empty
func (r *CLIResult) AssertStdoutEmpty(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(r.Stdout) != "" {
		t.Fatalf("Expected empty stdout, but got: %q", r.Stdout)
	}
}

// AssertStderrEmpty verifies stderr is empty
func (r *CLIResult) AssertStderrEmpty(t *testing.T) {
	t.Helper()
	if strings.TrimSpace(r.Stderr) != "" {
		t.Fatalf("Expected empty stderr, but got: %q", r.Stderr)
	}
}

// CaptureStdout captures stdout during function execution
// Useful for testing functions that write directly to os.Stdout
func CaptureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the function
	fn()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// CaptureStderr captures stderr during function execution
// Useful for testing functions that write directly to os.Stderr
func CaptureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Execute the function
	fn()

	// Restore stderr
	w.Close()
	os.Stderr = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// SetupTestConfig creates a temporary config for testing
// This ensures tests don't interfere with user's actual config
func SetupTestConfig(t *testing.T) (configDir string, cleanup func()) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set environment variables to use temp config
	originalHome := os.Getenv("HOME")
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")

	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", "")

	cleanup = func() {
		os.RemoveAll(tmpDir)
		os.Setenv("HOME", originalHome)
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
	}

	return tmpDir, cleanup
}

// CheckStringContains is a helper function for verifying string content
// Pattern used by: kubernetes, docker, helm
func CheckStringContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected string to contain %q, but it didn't.\nFull string: %s",
			needle, haystack)
	}
}

// CheckStringNotContains verifies a string does NOT contain the needle
func CheckStringNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("Expected string to NOT contain %q, but it did.\nFull string: %s",
			needle, haystack)
	}
}

// CheckStringEquals verifies exact string match with helpful error message
func CheckStringEquals(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("Strings don't match.\nExpected: %q\nActual:   %q", expected, actual)
	}
}

// RunTableTests executes a slice of test cases
// This is a common pattern for table-driven tests in major Go projects
func RunTableTests[T any](t *testing.T, tests []T, runner func(t *testing.T, tt T)) {
	t.Helper()
	for i, tt := range tests {
		testName := fmt.Sprintf("test_%d", i)
		t.Run(testName, func(t *testing.T) {
			runner(t, tt)
		})
	}
}

// RunNamedTableTests executes named test cases
type NamedTest[T any] struct {
	Name string
	Data T
}

func RunNamedTableTests[T any](t *testing.T, tests []NamedTest[T], runner func(t *testing.T, tt T)) {
	t.Helper()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner(t, test.Data)
		})
	}
}
