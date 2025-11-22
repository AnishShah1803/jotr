package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/notes"
	"github.com/anish/jotr/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport dimensions
		m.updateViewportSizes()

		if !m.ready {
			m.ready = true
			return m, m.loadData()
		}
		// Window was resized, just re-render with new dimensions
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)

		case "tab":
			m.focusedPanel = (m.focusedPanel + 1) % 4
			return m, nil

		case "shift+tab":
			m.focusedPanel = (m.focusedPanel + 3) % 4
			return m, nil

		case "up", "k":
			return m.handleUp()

		case "down", "j":
			return m.handleDown()

		case "enter":
			return m.handleEnter()

		case "r":
			// Clear any previous errors when refreshing
			m.err = nil
			m.errorRetryable = false
			m.statusMsg = "Refreshing..."
			return m, m.loadData()

		case "u":
			m.statusMsg = "🔍 Checking for updates..."
			return m, checkForUpdatesCmd()

		case "escape":
			// Clear errors on escape
			if m.err != nil {
				m.err = nil
				m.errorRetryable = false
				m.statusMsg = ""
				return m, nil
			}
		}

		// If there's a retryable error, allow 'r' to retry
		if m.err != nil && m.errorRetryable && msg.String() == "r" {
			m.err = nil
			m.errorRetryable = false
			m.statusMsg = "Retrying..."
			return m, m.loadData()
		}

	case tickMsg:
		return m, tickCmd()

	case dataLoadedMsg:
		m.notes = msg.notes
		m.tasks = msg.tasks
		m.streak = msg.streak
		m.totalNotes = msg.totalNotes
		m.totalTasks = msg.totalTasks
		m.completedTasks = msg.completedTasks
		m.statusMsg = "Data loaded successfully"

		// Update stats viewport with new data
		m.updateStatsViewport()

		if len(m.notes) > 0 {
			return m, m.loadPreview(m.notes[m.selectedNote])
		}
		return m, nil

	case previewLoadedMsg:
		m.previewContent = string(msg)
		m.updatePreviewViewport()
		return m, nil

	case editorFinishedMsg:
		// Editor closed, refresh data
		m.statusMsg = ""
		if msg.err != nil {
			m.err = msg.err
			m.errorRetryable = true
		}
		return m, m.loadData()

	case updateCheckMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("❌ Update check failed: %v", msg.err)
		} else if msg.hasUpdate {
			m.updateAvailable = true
			m.updateVersion = msg.version
			m.statusMsg = fmt.Sprintf("🆕 Update available: %s (restart jotr and run 'jotr update')", msg.version)
		} else {
			m.statusMsg = "✅ You're running the latest version!"
		}
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.errorRetryable = msg.retryable
		return m, nil
	}

	return m, nil
}

func (m Model) handleUp() (Model, tea.Cmd) {
	switch m.focusedPanel {
	case panelNotes:
		if m.selectedNote > 0 {
			m.selectedNote--
			if len(m.notes) > 0 {
				return m, m.loadPreview(m.notes[m.selectedNote])
			}
		}
	case panelPreview:
		m.previewViewport.LineUp(1)
	case panelTasks:
		if m.selectedTask > 0 {
			m.selectedTask--
		}
	case panelStats:
		m.statsViewport.LineUp(1)
	}
	return m, nil
}

func (m Model) handleDown() (Model, tea.Cmd) {
	switch m.focusedPanel {
	case panelNotes:
		if m.selectedNote < len(m.notes)-1 {
			m.selectedNote++
			if len(m.notes) > 0 {
				return m, m.loadPreview(m.notes[m.selectedNote])
			}
		}
	case panelPreview:
		m.previewViewport.LineDown(1)
	case panelTasks:
		if m.selectedTask < len(m.tasks)-1 {
			m.selectedTask++
		}
	case panelStats:
		m.statsViewport.LineDown(1)
	}
	return m, nil
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.focusedPanel {
	case panelNotes:
		if len(m.notes) > 0 && m.selectedNote < len(m.notes) {
			// Open note in editor
			notePath := m.notes[m.selectedNote]
			m.statusMsg = "Opening editor..."
			c := notes.GetEditorCmd(notePath)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				return editorFinishedMsg{err}
			})
		}
	case panelTasks:
		// Open todo list in editor
		todoPath := m.config.TodoPath
		m.statusMsg = "Opening todo list..."
		c := notes.GetEditorCmd(todoPath)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		})
	}
	return m, nil
}

type editorFinishedMsg struct{ err error }

type dataLoadedMsg struct {
	notes          []string
	tasks          []tasks.Task
	streak         int
	totalNotes     int
	totalTasks     int
	completedTasks int
}

type previewLoadedMsg []byte

func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		// Load recent notes with error handling
		recentNotes, err := notes.GetRecentDailyNotes(m.config.DiaryPath, 10)
		if err != nil {
			return newErrorMsg(fmt.Errorf("failed to load recent notes: %w", err), true)
		}

		// Load tasks with error handling
		allTasks, err := tasks.ReadTasks(m.config.TodoPath)
		if err != nil {
			// This is retryable - todo file might not exist yet
			return newErrorMsg(fmt.Errorf("failed to load tasks: %w", err), true)
		}
		total, completed, _ := tasks.CountTasks(allTasks)

		// Calculate streak
		streak := calculateStreak(m.config)

		// Count total notes with error handling
		allNotes, err := notes.FindNotes(m.config.Paths.BaseDir)
		if err != nil {
			return newErrorMsg(fmt.Errorf("failed to find notes: %w", err), true)
		}

		return dataLoadedMsg{
			notes:          recentNotes,
			tasks:          allTasks,
			streak:         streak,
			totalNotes:     len(allNotes),
			totalTasks:     total,
			completedTasks: completed,
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



func calculateStreak(cfg *config.LoadedConfig) int {
	today := time.Now()
	streak := 0
	firstValidDay := true

	for i := 0; i < 365; i++ {
		date := today.AddDate(0, 0, -i)

		// Skip weekends if configured
		if !cfg.Streaks.IncludeWeekends {
			weekday := date.Weekday()
			if weekday == time.Saturday || weekday == time.Sunday {
				continue
			}
		}

		notePath := notes.BuildDailyNotePath(cfg.DiaryPath, date)

		if notes.FileExists(notePath) {
			streak++
		} else {
			// If this is the first valid day we're checking and there's no note, streak is 0
			if firstValidDay {
				break
			}
			// If we already have a streak and hit a gap, stop counting
			if streak > 0 {
				break
			}
		}

		firstValidDay = false
	}

	return streak
}

// updateViewportSizes updates all viewport dimensions based on current window size
func (m *Model) updateViewportSizes() {
	// Must match the calculation in View() exactly
	minWidthForAscii := 50
	minHeightForAscii := 40
	var headerFooterHeight int

	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		headerFooterHeight = 13 // Large terminal with ASCII art
	} else {
		headerFooterHeight = 2 // Small terminal: minimal header
	}

	// Calculate dimensions - must match View()
	availableWidth := m.width - 8 // Account for margins
	leftPanelWidth := (availableWidth - 2) / 2
	rightPanelWidth := availableWidth - leftPanelWidth - 2
	panelHeight := (m.height - headerFooterHeight - 4) / 2

	// Ensure minimum size
	if leftPanelWidth < 30 {
		leftPanelWidth = 30
	}
	if rightPanelWidth < 30 {
		rightPanelWidth = 30
	}
	if panelHeight < 8 {
		panelHeight = 8
	}

	// Calculate content dimensions for each panel
	leftContentWidth := leftPanelWidth - 4   // Account for border and padding
	rightContentWidth := rightPanelWidth - 4 // Account for border and padding
	contentHeight := panelHeight - 3         // Account for border and title

	if leftContentWidth < 10 {
		leftContentWidth = 10
	}
	if rightContentWidth < 10 {
		rightContentWidth = 10
	}
	if contentHeight < 3 {
		contentHeight = 3
	}

	// Update viewports with correct widths
	m.notesViewport.Width = leftContentWidth
	m.notesViewport.Height = contentHeight
	m.previewViewport.Width = rightContentWidth
	m.previewViewport.Height = contentHeight
	m.tasksViewport.Width = leftContentWidth
	m.tasksViewport.Height = contentHeight
	m.statsViewport.Width = rightContentWidth
	m.statsViewport.Height = contentHeight

	// Update viewport contents with new dimensions
	m.updatePreviewViewport()
	m.updateStatsViewport()
}

func (m *Model) updatePreviewViewport() {
	content := m.previewContent
	if content == "" {
		content = "Select a note to preview"
	}

	// Calculate content width based on viewport
	contentWidth := m.previewViewport.Width
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Process lines: truncate if needed and add margin
	lines := strings.Split(content, "\n")
	maxWidth := contentWidth - 1
	if maxWidth < 10 {
		maxWidth = 10
	}
	for i, line := range lines {
		if len(line) > maxWidth {
			lines[i] = line[:maxWidth-3] + "..."
		}
		// Add small left margin
		lines[i] = " " + lines[i]
	}

	displayContent := strings.Join(lines, "\n")
	m.previewViewport.SetContent(displayContent)
}

func (m *Model) updateStatsViewport() {
	contentWidth := m.statsViewport.Width
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Streak with visual indicator
	streakIcon := iconStreak
	streakColor := successColor
	if m.streak == 0 {
		streakIcon = iconEmpty
		streakColor = secondaryColor
	} else if m.streak >= 30 {
		streakIcon = iconStreak + iconStreak + iconStreak
	} else if m.streak >= 7 {
		streakIcon = iconStreak + iconStreak
	}

	streakStyle := lipgloss.NewStyle().Foreground(streakColor).Bold(true)
	content := " " + streakStyle.Render(fmt.Sprintf("%s %d day streak", streakIcon, m.streak)) + "\n\n"

	// Notes stats
	content += " " + lipgloss.NewStyle().Foreground(primaryColor).Render("Notes") + "\n"
	content += fmt.Sprintf("  Total: %d\n", m.totalNotes)
	content += fmt.Sprintf("  Recent: %d\n\n", len(m.notes))

	// Task stats with progress bar
	content += " " + lipgloss.NewStyle().Foreground(primaryColor).Render("Tasks") + "\n"
	pendingTasks := m.totalTasks - m.completedTasks
	content += fmt.Sprintf("  Pending: %d\n", pendingTasks)
	content += fmt.Sprintf("  Done: %d\n", m.completedTasks)

	if m.totalTasks > 0 {
		completion := float64(m.completedTasks) / float64(m.totalTasks) * 100

		// Progress bar
		barWidth := contentWidth - 3
		if barWidth > 20 {
			barWidth = 20
		}
		if barWidth < 5 {
			barWidth = 5
		}
		filled := int(float64(barWidth) * completion / 100)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

		var barColor lipgloss.Color
		if completion >= 80 {
			barColor = successColor
		} else if completion >= 50 {
			barColor = warningColor
		} else {
			barColor = secondaryColor
		}

		barStyle := lipgloss.NewStyle().Foreground(barColor)
		content += fmt.Sprintf("  %s %.0f%%\n", barStyle.Render(bar), completion)
	}

	m.statsViewport.SetContent(content)
}

