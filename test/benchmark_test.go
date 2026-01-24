package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	searchcmd "github.com/AnishShah1803/jotr/cmd/search"
	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/tasks"
)

func BenchmarkNoteCreation(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		err := notes.CreateNote(ctx, tempDir+"/benchmark-note", "Benchmark Test Note")

		b.StartTimer()

		if err != nil {
			b.Fatalf("Failed to create benchmark note: %v", err)
		}
	}
}

func BenchmarkNoteSearch(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	testNotes := []string{"search-test-1", "search-test-2", "search-test-3"}
	for _, note := range testNotes {
		err := notes.CreateNote(ctx, tempDir+"/"+note, "Search Test: "+note)
		if err != nil {
			b.Fatalf("Failed to create search test note: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := notes.SearchNotes(ctx, tempDir, "search")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkTaskSync(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	testNotes := []string{"sync-test-1", "sync-test-2"}
	for _, note := range testNotes {
		err := notes.CreateNote(ctx, tempDir+"/"+note, "Sync Test Note")
		if err != nil {
			b.Fatalf("Failed to create sync test note: %v", err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := notes.SearchNotes(ctx, tempDir, "sync")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkConfigLoading(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := config.Load()
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
	}
}

func BenchmarkMemoryUsage(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		runtime.GC()

		var m runtime.MemStats

		runtime.ReadMemStats(&m)

		if i == 0 {
			fmt.Printf("Memory Usage: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d\n",
				m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
		}
	}
}

func BenchmarkParseTasks(b *testing.B) {
	content := `## Tasks

### Todo
- [ ] Buy groceries <!-- id: abc12345 -->
- [ ] Call mom <!-- id: def67890 -->
- [ ] Finish quarterly report [P1] #work @important
- [ ] Schedule dentist appointment

### In Progress
- [x] Review project proposal [P2] <!-- id: fedcba09 -->
- [ ] Update documentation [P3] #documentation

### Done
- [x] Complete training module
- [x] Update team calendar`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tasks.ParseTasks(content)
	}
}

func BenchmarkParseTasksLarge(b *testing.B) {
	var sb strings.Builder
	sb.WriteString("## Tasks\n\n")

	for i := 0; i < 100; i++ {
		status := " "
		if i%3 == 0 {
			status = "x"
		}
		sb.WriteString(fmt.Sprintf("- [%s] Task number %d with some description text [P%d] #tag%d @context\n", status, i, i%4, i))
	}

	content := sb.String()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tasks.ParseTasks(content)
	}
}

func BenchmarkParseTasksEmpty(b *testing.B) {
	content := `# My Notes

This is just regular content without any tasks.

## Section One

Just some more text here.

## Section Two

More text without tasks.
`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tasks.ParseTasks(content)
	}
}

func BenchmarkParseTasksMixedFormats(b *testing.B) {
	content := `## Tasks

- [ ] Dash task
* [ ] Asterisk task
+ [ ] Plus task
- [x] Completed dash
* [x] Completed asterisk
+ [x] Completed plus

## With Priority and Tags
- [P1] #high @urgent High priority task
- [P2] Medium priority
- [P3] #low Low priority

## Complex Tasks
- [ ] Task with multiple tags #tag1 #tag2 #tag3
- [ ] Task with multiple mentions @person1 @person2
- [ ] Long task description that spans multiple lines in the source but should be parsed correctly
`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tasks.ParseTasks(content)
	}
}

func BenchmarkFindNotes(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 50; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			err := notes.CreateNote(ctx, notePath, fmt.Sprintf("Test Note %d-%d", i, j))
			if err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := notes.FindNotes(ctx, tempDir)
		if err != nil {
			b.Fatalf("Failed to find notes: %v", err)
		}
	}
}

func BenchmarkSearchNotes(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 50; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Test Note %d-%d\n\nThis note contains the search term for testing purposes.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := notes.SearchNotes(ctx, tempDir, "search term")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkSearchNotesLarge(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 100; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("subdir_%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 10; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Large Note %d-%d\n\nThis is a larger note content for benchmark testing with the keyword jotr and other searchable text.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matches, searchErr := notes.SearchNotes(ctx, tempDir, "jotr")
		if searchErr != nil {
			b.Fatalf("Failed to search notes: %v", searchErr)
		}
		_ = matches
	}
}

func BenchmarkSearchNotesNoMatches(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 50; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Test Note %d-%d\n\nThis note does not contain the searched term.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matches, searchErr := notes.SearchNotes(ctx, tempDir, "nonexistentterm12345")
		if searchErr != nil {
			b.Fatalf("Failed to search notes: %v", searchErr)
		}
		_ = matches
	}
}

func BenchmarkSearchNotesNested(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 20; i++ {
		for j := 0; j < 5; j++ {
			for k := 0; k < 3; k++ {
				subDir := filepath.Join(tempDir, fmt.Sprintf("level1_%d/level2_%d/level3_%d", i, j, k))
				err := os.MkdirAll(subDir, 0755)
				if err != nil {
					b.Fatalf("Failed to create nested directory: %v", err)
				}

				notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d_%d.md", i, j, k))
				content := fmt.Sprintf("# Deeply Nested Note %d-%d-%d\n\nThis is content for testing search performance with nested directories.", i, j, k)
				if err := notes.CreateNote(ctx, notePath, content); err != nil {
					b.Fatalf("Failed to create nested note: %v", err)
				}
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matches, searchErr := notes.SearchNotes(ctx, tempDir, "nested")
		if searchErr != nil {
			b.Fatalf("Failed to search notes: %v", searchErr)
		}
		_ = matches
	}
}

func BenchmarkSearchNotesCaseInsensitive(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 30; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Test Note %d-%d\n\nThis note has JOTR in different cases like Jotr, jotr, and JOTR for testing.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		matches, searchErr := notes.SearchNotes(ctx, tempDir, "jotr")
		if searchErr != nil {
			b.Fatalf("Failed to search notes: %v", searchErr)
		}
		_ = matches
	}
}

func BenchmarkSearchCmdWrapper(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 20; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/5))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Test Note %d-%d\n\nThis note contains the search term for benchmark testing purposes.\nMultiple lines of content to simulate real notes.\nLine with more content for context display.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	cfg := &config.LoadedConfig{
		Config: config.Config{
			Paths: config.PathsConfig{
				BaseDir: tempDir,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := searchcmd.SearchNotes(ctx, cfg, "search term")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkSearchCmdWrapperCount(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 30; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Note %d-%d with TODO item\n\n- [ ] Task to complete\n- [x] Completed task\n\nMore TODO comments here.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	cfg := &config.LoadedConfig{
		Config: config.Config{
			Paths: config.PathsConfig{
				BaseDir: tempDir,
			},
		},
	}

	originalCount := searchcmd.GetSearchCountForTest()
	searchcmd.SetSearchCountForTest(true)
	defer searchcmd.SetSearchCountForTest(originalCount)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := searchcmd.SearchNotes(ctx, cfg, "TODO")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkSearchCmdWrapperFiles(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 25; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/5))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 4; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf("# Project Update %d-%d\n\nUpdating project documentation.\nSee related files for more context.", i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	cfg := &config.LoadedConfig{
		Config: config.Config{
			Paths: config.PathsConfig{
				BaseDir: tempDir,
			},
		},
	}

	originalFiles := searchcmd.GetSearchFilesForTest()
	searchcmd.SetSearchFilesForTest(true)
	defer searchcmd.SetSearchFilesForTest(originalFiles)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := searchcmd.SearchNotes(ctx, cfg, "project")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkSearchCmdWrapperLarge(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 100; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("subdir_%d", i/10))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 10; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := fmt.Sprintf(`# Large Note %d-%d

## Overview
This is a comprehensive note with lots of content for benchmarking purposes.

## Details
The jotr tool is being tested for performance characteristics.
Search performance is critical for user experience.

## Implementation Notes
Additional context about the implementation.
Multiple paragraphs simulate real-world note content.

## References
See related documentation for more information.
jotr provides fast search across many notes.
`, i, j)
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	cfg := &config.LoadedConfig{
		Config: config.Config{
			Paths: config.PathsConfig{
				BaseDir: tempDir,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := searchcmd.SearchNotes(ctx, cfg, "jotr")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}

func BenchmarkSearchCmdWrapperNoMatches(b *testing.B) {
	ctx := context.Background()

	tempDir := b.TempDir()
	defer os.RemoveAll(tempDir)

	for i := 0; i < 25; i++ {
		subDir := filepath.Join(tempDir, fmt.Sprintf("dir%d", i/5))
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create subdirectory: %v", err)
		}

		for j := 0; j < 5; j++ {
			notePath := filepath.Join(subDir, fmt.Sprintf("note_%d_%d.md", i, j))
			content := "# Test Note\n\nThis note has some content but no special terms that would match."
			if err := notes.CreateNote(ctx, notePath, content); err != nil {
				b.Fatalf("Failed to create note: %v", err)
			}
		}
	}

	cfg := &config.LoadedConfig{
		Config: config.Config{
			Paths: config.PathsConfig{
				BaseDir: tempDir,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := searchcmd.SearchNotes(ctx, cfg, "nonexistentterm12345")
		if err != nil {
			b.Fatalf("Failed to search notes: %v", err)
		}
	}
}
