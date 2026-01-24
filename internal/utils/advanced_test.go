package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestAtomicWriteFile_TableDriven demonstrates table-driven test pattern.
func TestAtomicWriteFile_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		permissions os.FileMode
		wantErr     bool
	}{
		{
			name:        "basic write",
			data:        []byte("hello world"),
			permissions: 0644,
			wantErr:     false,
		},
		{
			name:        "empty file",
			data:        []byte(""),
			permissions: 0644,
			wantErr:     false,
		},
		{
			name:        "large file",
			data:        make([]byte, 1024*1024), // 1MB
			permissions: 0600,
			wantErr:     false,
		},
		{
			name:        "executable permissions",
			data:        []byte("#!/bin/bash\necho hello"),
			permissions: 0755,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "jotr-table-test-")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			testFile := filepath.Join(tmpDir, "test.txt")

			err = AtomicWriteFile(testFile, tt.data, tt.permissions)

			if (err != nil) != tt.wantErr {
				t.Errorf("AtomicWriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify content
				content, err := os.ReadFile(testFile)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}

				if len(content) != len(tt.data) {
					t.Errorf("Content length mismatch: got %d, want %d", len(content), len(tt.data))
				}

				// Verify permissions
				info, err := os.Stat(testFile)
				if err != nil {
					t.Errorf("Failed to stat file: %v", err)
					return
				}

				if info.Mode().Perm() != tt.permissions {
					t.Errorf("Permission mismatch: got %o, want %o", info.Mode().Perm(), tt.permissions)
				}
			}
		})
	}
}

// TestAtomicWriteFile_Subtests demonstrates subtest pattern.
func TestAtomicWriteFile_Subtests(t *testing.T) {
	// Setup shared resources
	tmpDir, err := os.MkdirTemp("", "jotr-subtest-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("concurrent writes", func(t *testing.T) {
		// Test concurrent atomic writes to same directory
		done := make(chan bool, 2)

		go func() {
			defer func() { done <- true }()

			testFile1 := filepath.Join(tmpDir, "concurrent1.txt")

			err := AtomicWriteFile(testFile1, []byte("data1"), 0644)
			if err != nil {
				t.Errorf("Concurrent write 1 failed: %v", err)
			}
		}()

		go func() {
			defer func() { done <- true }()

			testFile2 := filepath.Join(tmpDir, "concurrent2.txt")

			err := AtomicWriteFile(testFile2, []byte("data2"), 0644)
			if err != nil {
				t.Errorf("Concurrent write 2 failed: %v", err)
			}
		}()

		// Wait for both goroutines
		<-done
		<-done
	})

	t.Run("overwrite existing", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "overwrite.txt")

		// Create initial file
		err := AtomicWriteFile(testFile, []byte("initial"), 0644)
		if err != nil {
			t.Fatalf("Initial write failed: %v", err)
		}

		// Overwrite
		err = AtomicWriteFile(testFile, []byte("updated"), 0644)
		if err != nil {
			t.Fatalf("Overwrite failed: %v", err)
		}

		// Verify content
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(content) != "updated" {
			t.Errorf("Content not updated: got %q, want %q", string(content), "updated")
		}
	})
}

// BenchmarkAtomicWriteFile demonstrates benchmark testing.
func BenchmarkAtomicWriteFile(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "jotr-bench-")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	data := []byte("benchmark test data")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tmpDir, "bench.txt")

		err := AtomicWriteFile(testFile, data, 0644)
		if err != nil {
			b.Fatalf("AtomicWriteFile failed: %v", err)
		}
	}
}

// BenchmarkAtomicWriteFile_Sizes demonstrates benchmarking with different data sizes.
func BenchmarkAtomicWriteFile_Sizes(b *testing.B) {
	sizes := []int{
		1024,    // 1KB
		10240,   // 10KB
		102400,  // 100KB
		1048576, // 1MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			tmpDir, err := os.MkdirTemp("", "jotr-bench-")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			data := make([]byte, size)
			testFile := filepath.Join(tmpDir, "bench.txt")

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				err := AtomicWriteFile(testFile, data, 0644)
				if err != nil {
					b.Fatalf("AtomicWriteFile failed: %v", err)
				}
			}
		})
	}
}
