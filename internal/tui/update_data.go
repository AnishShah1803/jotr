package tui

import (
	"context"
	"fmt"
	"os"
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

		recentNotes, err := notes.GetRecentDailyNotes(ctx, m.config.DiaryPath, 10)
		if err != nil {
			return newErrorMsg(fmt.Errorf("failed to load recent notes: %w", err), true)
		}

		select {
		case <-ctx.Done():
			return newErrorMsg(ctx.Err(), true)
		default:
		}
		allTasks, err := tasks.ReadTasks(ctx, m.config.TodoPath)
		if err != nil {
			return newErrorMsg(fmt.Errorf("failed to load tasks: %w", err), true)
		}

		select {
		case <-ctx.Done():
			return newErrorMsg(ctx.Err(), true)
		default:
		}

		total, completed, _ := tasks.CountTasks(allTasks)

		streak := calculateStreak(m.config)

		select {
		case <-ctx.Done():
			return newErrorMsg(ctx.Err(), true)
		default:
		}
		allNotes, err := notes.FindNotes(ctx, m.config.Paths.BaseDir)
		if err != nil {
			return newErrorMsg(fmt.Errorf("failed to find notes: %w", err), true)
		}

		editorConfigured := isAnyEditorAvailable(ctx, m.config)
		editorFallback := !isConfigEditorAvailable(ctx, m.config) && isShellEditorAvailable(ctx)

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
