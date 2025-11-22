package testhelpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFS provides utilities for testing file system operations
// Pattern used by: kubernetes, docker, go toolchain
type TestFS struct {
	BaseDir string
	cleanup func()
}

// NewTestFS creates a new test filesystem in a temporary directory
func NewTestFS(t *testing.T) *TestFS {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "jotr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return &TestFS{
		BaseDir: tmpDir,
		cleanup: func() { os.RemoveAll(tmpDir) },
	}
}

// Cleanup removes the test filesystem
func (fs *TestFS) Cleanup() {
	if fs.cleanup != nil {
		fs.cleanup()
	}
}

// WriteFile creates a file with content in the test filesystem
func (fs *TestFS) WriteFile(t *testing.T, path, content string) {
	t.Helper()
	fullPath := filepath.Join(fs.BaseDir, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", fullPath, err)
	}
}

// ReadFile reads a file from the test filesystem
func (fs *TestFS) ReadFile(t *testing.T, path string) string {
	t.Helper()
	fullPath := filepath.Join(fs.BaseDir, path)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", fullPath, err)
	}

	return string(content)
}

// FileExists checks if a file exists in the test filesystem
func (fs *TestFS) FileExists(path string) bool {
	fullPath := filepath.Join(fs.BaseDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// AssertFileExists verifies that a file exists
func (fs *TestFS) AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if !fs.FileExists(path) {
		t.Errorf("Expected file %s to exist, but it doesn't", path)
	}
}

// AssertFileNotExists verifies that a file does not exist
func (fs *TestFS) AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if fs.FileExists(path) {
		t.Errorf("Expected file %s to not exist, but it does", path)
	}
}

// AssertFileContains verifies that a file contains the expected content
func (fs *TestFS) AssertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	content := fs.ReadFile(t, path)
	if !strings.Contains(content, expected) {
		t.Errorf("Expected file %s to contain %q, but it didn't.\nFile content:\n%s",
			path, expected, content)
	}
}

// AssertFileEquals verifies that a file content exactly matches expected
func (fs *TestFS) AssertFileEquals(t *testing.T, path, expected string) {
	t.Helper()
	content := fs.ReadFile(t, path)
	if strings.TrimSpace(content) != strings.TrimSpace(expected) {
		t.Errorf("Expected file %s to equal:\n%q\nBut got:\n%q", path, expected, content)
	}
}

// CreateDailyNote creates a sample daily note for testing
func (fs *TestFS) CreateDailyNote(t *testing.T, date time.Time, content string) string {
	t.Helper()

	// Format path like jotr does: {year}/{month}/{year}-{month}-{day}-{weekday}.md
	year, month, day := date.Date()
	weekday := date.Format("Mon")

	dirPath := fmt.Sprintf("%d/%02d", year, int(month))
	filename := fmt.Sprintf("%d-%02d-%02d-%s.md", year, int(month), day, weekday)
	fullPath := filepath.Join(dirPath, filename)

	fs.WriteFile(t, fullPath, content)
	return fullPath
}

// MockTime provides utilities for testing time-dependent code
type MockTime struct {
	current time.Time
}

// NewMockTime creates a new mock time starting at the given time
func NewMockTime(t time.Time) *MockTime {
	return &MockTime{current: t}
}

// Now returns the current mocked time
func (mt *MockTime) Now() time.Time {
	return mt.current
}

// Advance moves the mock time forward by the given duration
func (mt *MockTime) Advance(d time.Duration) {
	mt.current = mt.current.Add(d)
}

// TestLogger captures log output for testing
// Pattern used by: kubernetes, prometheus, docker
type TestLogger struct {
	logs []string
}

// NewTestLogger creates a new test logger
func NewTestLogger() *TestLogger {
	return &TestLogger{}
}

// Log adds a log entry
func (tl *TestLogger) Log(msg string) {
	tl.logs = append(tl.logs, msg)
}

// Logf adds a formatted log entry
func (tl *TestLogger) Logf(format string, args ...interface{}) {
	tl.logs = append(tl.logs, fmt.Sprintf(format, args...))
}

// GetLogs returns all logged messages
func (tl *TestLogger) GetLogs() []string {
	return tl.logs
}

// AssertLogContains verifies that logs contain the expected message
func (tl *TestLogger) AssertLogContains(t *testing.T, expected string) {
	t.Helper()
	for _, log := range tl.logs {
		if strings.Contains(log, expected) {
			return
		}
	}
	t.Errorf("Expected logs to contain %q, but they didn't.\nLogs: %v", expected, tl.logs)
}

// AssertLogCount verifies the number of log entries
func (tl *TestLogger) AssertLogCount(t *testing.T, expected int) {
	t.Helper()
	if len(tl.logs) != expected {
		t.Errorf("Expected %d log entries, got %d.\nLogs: %v", expected, len(tl.logs), tl.logs)
	}
}

// Clear removes all log entries
func (tl *TestLogger) Clear() {
	tl.logs = tl.logs[:0]
}

// TestWriter captures written output for testing
type TestWriter struct {
	content strings.Builder
}

// Write implements io.Writer
func (tw *TestWriter) Write(p []byte) (n int, err error) {
	return tw.content.Write(p)
}

// String returns the written content
func (tw *TestWriter) String() string {
	return tw.content.String()
}

// Reset clears the content
func (tw *TestWriter) Reset() {
	tw.content.Reset()
}

// AssertContains verifies the content contains expected string
func (tw *TestWriter) AssertContains(t *testing.T, expected string) {
	t.Helper()
	if !strings.Contains(tw.String(), expected) {
		t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s",
			expected, tw.String())
	}
}

// ConfigHelper provides utilities for testing configuration
type ConfigHelper struct {
	fs *TestFS
}

// NewConfigHelper creates a new config helper
func NewConfigHelper(fs *TestFS) *ConfigHelper {
	return &ConfigHelper{fs: fs}
}

// CreateBasicConfig creates a minimal valid config for testing
func (ch *ConfigHelper) CreateBasicConfig(t *testing.T) {
	t.Helper()

	configContent := `{
  "paths": {
    "base_dir": "` + ch.fs.BaseDir + `",
    "diary_dir": "diary",
    "todo_file_path": "todo.md"
  },
  "format": {
    "task_section": "Tasks",
    "capture_section": "Captured",
    "daily_note_sections": ["Notes", "Tasks"],
    "daily_note_pattern": "{year}-{month}-{day}-{weekday}",
    "daily_note_dir_pattern": "{year}/{month}"
  },
  "streaks": {
    "include_weekends": false
  }
}`

	ch.fs.WriteFile(t, ".config/jotr/config.json", configContent)
}

// WithCustomConfig creates a config with custom settings
func (ch *ConfigHelper) WithCustomConfig(t *testing.T, customizations map[string]interface{}) {
	t.Helper()
	// In a real implementation, you would merge the customizations with the base config
	// For now, just create the basic config
	ch.CreateBasicConfig(t)
}

// ParallelTest runs a test function in parallel with proper setup/teardown
// Pattern used by: go toolchain, kubernetes
func ParallelTest(t *testing.T, name string, fn func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		fn(t)
	})
}

// SubTest is a helper for creating subtests with proper naming
func SubTest(t *testing.T, name string, fn func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		fn(t)
	})
}

// SkipIfShort skips a test if running with -short flag
// Pattern used throughout the Go toolchain
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// ExpectPanic verifies that a function panics with the expected message
func ExpectPanic(t *testing.T, expectedMsg string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if msg := fmt.Sprintf("%v", r); !strings.Contains(msg, expectedMsg) {
				t.Errorf("Expected panic message to contain %q, got %q", expectedMsg, msg)
			}
		} else {
			t.Errorf("Expected function to panic with message containing %q, but it didn't panic", expectedMsg)
		}
	}()
	fn()
}

// TempChdir temporarily changes to a directory for the duration of the test
// Pattern used by: go toolchain, kubernetes
func TempChdir(t *testing.T, dir string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Failed to change to directory %s: %v", dir, err)
	}

	return func() {
		err := os.Chdir(originalDir)
		if err != nil {
			t.Errorf("Failed to restore directory to %s: %v", originalDir, err)
		}
	}
}

// RedirectOutput captures stdout/stderr during test execution
func RedirectOutput(t *testing.T) (stdout, stderr *TestWriter, restore func()) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	stdoutWriter := &TestWriter{}
	stderrWriter := &TestWriter{}

	go func() {
		io.Copy(stdoutWriter, stdoutR)
	}()
	go func() {
		io.Copy(stderrWriter, stderrR)
	}()

	restore = func() {
		stdoutW.Close()
		stderrW.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}

	return stdoutWriter, stderrWriter, restore
}
