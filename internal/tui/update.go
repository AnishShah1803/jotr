package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
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
			m.err = nil
			m.errorRetryable = false
			m = setStatus(m, "Refreshing...", "info")

			return m, m.loadData()

		case "u":
			m = setStatus(m, "🔍 Checking for updates...", "info")
			return m, checkForUpdatesCmd()

		case "n":
			if m.err != nil && m.errorRetryable {
				err := createTodoFile(m.config.TodoPath)
				if err != nil {
					m = setStatus(m, "Failed to create file: "+err.Error(), "error")
				} else {
					m = setStatus(m, "Todo file created successfully", "success")
					m.err = nil
					m.errorRetryable = false
					return m, m.loadData()
				}
				return m, nil
			}
		}

		// If there's a retryable error, allow 'r' to retry
		if m.err != nil && m.errorRetryable && msg.String() == "r" {
			m.err = nil
			m.errorRetryable = false
			m = setStatus(m, "Retrying...", "info")
			return m, m.loadData()
		}

	case tickMsg:
		if m.statusMsg != "" && time.Since(m.statusMsgTime) > m.statusDuration {
			m = clearStatus(m)
		}
		return m, tickCmd()

	case dataLoadedMsg:
		m.notes = msg.notes
		m.tasks = msg.tasks
		m.streak = msg.streak
		m.totalNotes = msg.totalNotes
		m.totalTasks = msg.totalTasks
		m.completedTasks = msg.completedTasks
		m.editorConfigured = msg.editorConfigured
		m.editorFallback = msg.editorFallback
		m = setStatus(m, "Data loaded successfully", "success")

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
		m = clearStatus(m)
		if msg.err != nil {
			m.err = msg.err
			m.errorRetryable = true
		}

		return m, m.loadData()

	case editorOpenAttemptMsg:
		m = clearStatus(m)
		if msg.useShellFallback {
			return m.handleFileOpen()
		}
		return m, nil

	case editorFallbackMsg:
		m = clearStatus(m)
		if !m.editorConfigured {
			m = setStatus(m, "❌ No editor available - configure editor.default or set EDITOR env var", "error")
		}
		return m, nil

	case updateCheckMsg:
		if msg.err != nil {
			m = setStatus(m, fmt.Sprintf("❌ Update check failed: %v", msg.err), "error")
		} else if msg.hasUpdate {
			m.updateAvailable = true
			m.updateVersion = msg.version
			m = setStatus(m, fmt.Sprintf("🆕 Update available: %s (restart jotr and run 'jotr update')", msg.version), "info")
		} else {
			m = setStatus(m, "✅ You're running the latest version!", "success")
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
	if !m.editorConfigured {
		m = setStatus(m, "❌ No editor available - configure editor.default or set EDITOR env var", "error")
		return m, nil
	}

	if m.editorFallback {
		m = setStatus(m, "⚠️  editor not configured in config - using shell EDITOR", "warning")
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return editorOpenAttemptMsg{useShellFallback: true}
		})
	}

	return m.handleFileOpen()
}

func (m Model) handleFileOpen() (Model, tea.Cmd) {
	var filePath string
	var statusMsg string

	switch m.focusedPanel {
	case panelNotes:
		if len(m.notes) > 0 && m.selectedNote < len(m.notes) {
			filePath = m.notes[m.selectedNote]
			statusMsg = "Opening editor..."
		}
	case panelTasks:
		filePath = m.config.TodoPath
		statusMsg = "Opening todo list..."
	default:
		return m, nil
	}

	if filePath == "" {
		return m, nil
	}

	m = setStatus(m, statusMsg, "info")

	var c *exec.Cmd
	var err error

	if m.editorFallback {
		c, err = notes.GetEditorCmdWithShellFallback(m.ctx, filePath)
	} else {
		c, err = notes.GetEditorCmdWithContext(m.ctx, filePath)
	}

	if err != nil {
		m = setStatus(m, "Error: "+err.Error(), "error")
		return m, nil
	}

	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

type editorFinishedMsg struct{ err error }

type dataLoadedMsg struct {
	notes            []string
	tasks            []tasks.Task
	streak           int
	totalNotes       int
	totalTasks       int
	completedTasks   int
	editorConfigured bool
	editorFallback   bool
}

type previewLoadedMsg []byte

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
			// This is retryable - todo file might not exist yet
			return newErrorMsg(fmt.Errorf("failed to load tasks: %w", err), true)
		}

		// Check context cancellation
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

// updateViewportSizes updates all viewport dimensions based on current window size.
func (m *Model) updateViewportSizes() {
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
	streakColor := output.SuccessColor

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
			barColor = output.SuccessColor
		} else if completion >= 50 {
			barColor = output.WarningColor
		} else {
			barColor = output.SecondaryColor
		}

		barStyle := lipgloss.NewStyle().Foreground(barColor)
		content += fmt.Sprintf("  %s %.0f%%\n", barStyle.Render(bar), completion)
	}

	m.statsViewport.SetContent(content)
}

func createTodoFile(path string) error {
	return os.WriteFile(path, []byte("# Todo\n\n## Tasks\n\n\n\n"), 0644)
}

func isEditorAvailable(ctx context.Context) bool {
	editor := config.GetEditorWithContext(ctx)
	if editor == "" {
		return false
	}

	if err := utils.ValidateEditor(editor); err != nil {
		return false
	}

	return true
}

func isAnyEditorAvailable(ctx context.Context, cfg *config.LoadedConfig) bool {
	configEditor := cfg.GetDefaultEditor()

	if configEditor != "" {
		if err := utils.ValidateEditor(configEditor); err == nil {
			return true
		}
	}

	shellEditor := os.Getenv("EDITOR")
	if shellEditor != "" {
		if err := utils.ValidateEditor(shellEditor); err == nil {
			return true
		}
	}

	return false
}

func isConfigEditorAvailable(ctx context.Context, cfg *config.LoadedConfig) bool {
	configEditor := cfg.GetDefaultEditor()

	if configEditor == "" {
		return false
	}

	if err := utils.ValidateEditor(configEditor); err != nil {
		return false
	}

	return true
}

func isShellEditorAvailable(ctx context.Context) bool {
	shellEditor := os.Getenv("EDITOR")

	if shellEditor == "" {
		return false
	}

	if err := utils.ValidateEditor(shellEditor); err != nil {
		return false
	}

	return true
}
