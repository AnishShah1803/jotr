package tui

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := m.ctx
		if ctx == nil {
			ctx = context.Background()
		}

		select {
		case <-ctx.Done():
			return newErrorMsg(ctx.Err(), true)
		default:
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		var firstErr error

		var recentNotes []string
		var allTasks []tasks.Task
		var allNotes []string

		wg.Add(3)

		go func() {
			defer wg.Done()
			n, err := notes.GetRecentDailyNotes(ctx, m.config.DiaryPath, 10)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to load recent notes: %w", err)
			}
			recentNotes = n
			mu.Unlock()
		}()

		go func() {
			defer wg.Done()
			t, err := tasks.ReadTasks(ctx, m.config.TodoPath)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to load tasks: %w", err)
			}
			allTasks = t
			mu.Unlock()
		}()

		go func() {
			defer wg.Done()
			n, err := notes.FindNotes(ctx, m.config.Paths.BaseDir)
			mu.Lock()
			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to find notes: %w", err)
			}
			allNotes = n
			mu.Unlock()
		}()

		wg.Wait()

		if firstErr != nil {
			return newErrorMsg(firstErr, true)
		}

		total, completed, _ := tasks.CountTasks(allTasks)

		streak := calculateStreak(m.config)

		select {
		case <-ctx.Done():
			return newErrorMsg(ctx.Err(), true)
		default:
		}

		editorConfigured := isAnyEditorAvailable(m.config)
		editorFallback := !isConfigEditorAvailable(m.config) && isShellEditorAvailable()

		return dataLoadedMsg{
			notes:            recentNotes,
			tasks:            allTasks,
			streak:           streak,
			totalNotes:       len(allNotes),
			totalTasks:       total,
			completedTasks:   completed,
			editorConfigured: editorConfigured,
			editorFallback:   editorFallback,
		}
	}
}

func (m Model) loadPreview(notePath string) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(notePath)
		if err != nil {
			return previewLoadedMsg([]byte(fmt.Sprintf("Error loading preview: %v", err)))
		}

		return previewLoadedMsg(content)
	}
}

func handleDataLoaded(m Model, msg dataLoadedMsg) (Model, tea.Cmd) {
	m.isLoading = false
	m.notes = msg.notes
	m.tasks = msg.tasks
	m.streak = msg.streak
	m.totalNotes = msg.totalNotes
	m.totalTasks = msg.totalTasks
	m.completedTasks = msg.completedTasks
	m.editorConfigured = msg.editorConfigured
	m.editorFallback = msg.editorFallback
	m = setStatus(m, "Data loaded successfully", "success")

	m.updateStatsViewport()

	if len(m.notes) > 0 {
		return m, m.loadPreview(m.notes[m.selectedNote])
	}

	return m, nil
}

func handlePreviewLoaded(m Model, msg previewLoadedMsg) (Model, tea.Cmd) {
	m.previewContent = string(msg)
	m.updatePreviewViewport()

	return m, nil
}

func calculateStreak(cfg *config.LoadedConfig) int {
	today := time.Now()
	streak := 0
	firstValidDay := true

	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i)

		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if utils.FileExists(notePath) {
			streak++
		} else {
			if firstValidDay {
				break
			}
			if streak > 0 {
				break
			}
		}

		firstValidDay = false
	}

	return streak
}
