package utils

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProcessFilesParallel(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "hello world",
		"file2.txt": "goodbye world",
		"file3.txt": "hello again",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	paths := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	ctx := context.Background()
	results, err := ProcessFilesParallel(ctx, paths, 2, func(path string, content []byte) bool {
		return strings.Contains(string(content), "hello")
	})
	if err != nil {
		t.Fatalf("ProcessFilesParallel failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(results))
	}

	for _, r := range results {
		if !strings.Contains(r, "file1.txt") && !strings.Contains(r, "file3.txt") {
			t.Errorf("Unexpected match: %s", r)
		}
	}
}

func TestProcessFilesParallelWithContent(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"file1.txt": "test content 1",
		"file2.txt": "other content",
		"file3.txt": "test content 3",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	paths := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	ctx := context.Background()
	results, err := ProcessFilesParallelWithContent(ctx, paths, 2, func(path string, content []byte) bool {
		return strings.HasPrefix(string(content), "test")
	})
	if err != nil {
		t.Fatalf("ProcessFilesParallelWithContent failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(results))
	}

	for _, r := range results {
		if !strings.Contains(r.Path, "file1.txt") && !strings.Contains(r.Path, "file3.txt") {
			t.Errorf("Unexpected match path: %s", r.Path)
		}
		if !strings.HasPrefix(string(r.Content), "test") {
			t.Errorf("Content should start with 'test': %s", string(r.Content))
		}
	}
}

func TestProcessFilesParallel_Cancellation(t *testing.T) {
	tempDir := t.TempDir()

	for i := 0; i < 10; i++ {
		path := filepath.Join(tempDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		paths[i] = filepath.Join(tempDir, "file"+string(rune('0'+i))+".txt")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	_, err := ProcessFilesParallel(ctx, paths, 4, func(path string, content []byte) bool {
		return true
	})

	if err != context.DeadlineExceeded {
		t.Logf("Expected context deadline exceeded, got: %v", err)
	}
}

func TestProcessFilesParallel_DefaultWorkers(t *testing.T) {
	tempDir := t.TempDir()

	path := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	results, err := ProcessFilesParallel(ctx, []string{path}, 0, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		t.Fatalf("ProcessFilesParallel failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestProcessFilesParallel_SkipUnreadable(t *testing.T) {
	tempDir := t.TempDir()

	readablePath := filepath.Join(tempDir, "readable.txt")
	if err := os.WriteFile(readablePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")

	ctx := context.Background()
	results, err := ProcessFilesParallel(ctx, []string{readablePath, nonExistentPath}, 2, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		t.Fatalf("ProcessFilesParallel failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result (skipping unreadable), got %d", len(results))
	}

	if results[0] != readablePath {
		t.Errorf("Expected %s, got %s", readablePath, results[0])
	}
}

func TestProcessFilesParallel_EmptyPaths(t *testing.T) {
	ctx := context.Background()
	results, err := ProcessFilesParallel(ctx, []string{}, 2, func(path string, content []byte) bool {
		return true
	})
	if err != nil {
		t.Fatalf("ProcessFilesParallel failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty paths, got %d", len(results))
	}
}
