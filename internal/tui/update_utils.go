package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/utils"
)

func setStatus(m Model, msg string, level string) Model {
	m.statusMsg = msg
	m.statusMsgTime = time.Now()
	m.statusLevel = level
	if level == "error" {
		m.statusDuration = 5 * time.Second
	} else {
		m.statusDuration = 1 * time.Second
	}
	return m
}

func clearStatus(m Model) Model {
	m.statusMsg = ""
	m.statusLevel = ""
	m.statusDuration = 0
	return m
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func handleTickMsg(m Model) (Model, tea.Cmd) {
	if m.statusMsg != "" && time.Since(m.statusMsgTime) > m.statusDuration {
		m = clearStatus(m)
	}
	return m, tickCmd()
}

func createTodoFile(path string) error {
	return os.WriteFile(path, []byte("# Todo\n\n## Tasks\n\n\n\n"), constants.FilePerm0644)
}

func isAnyEditorAvailable(cfg *config.LoadedConfig) bool {
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

func isConfigEditorAvailable(cfg *config.LoadedConfig) bool {
	configEditor := cfg.GetDefaultEditor()

	if configEditor == "" {
		return false
	}

	if err := utils.ValidateEditor(configEditor); err != nil {
		return false
	}

	return true
}

func isShellEditorAvailable() bool {
	shellEditor := os.Getenv("EDITOR")

	if shellEditor == "" {
		return false
	}

	if err := utils.ValidateEditor(shellEditor); err != nil {
		return false
	}

	return true
}

func (m *Model) updateViewportSizes() {
	var headerFooterHeight int

	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		headerFooterHeight = 13
	} else {
		headerFooterHeight = 2
	}

	availableWidth := m.width - 8
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

	leftContentWidth := leftPanelWidth - 4
	rightContentWidth := rightPanelWidth - 4
	contentHeight := panelHeight - 3

	if leftContentWidth < 10 {
		leftContentWidth = 10
	}

	if rightContentWidth < 10 {
		rightContentWidth = 10
	}

	if contentHeight < 3 {
		contentHeight = 3
	}

	m.notesViewport.Width = leftContentWidth
	m.notesViewport.Height = contentHeight
	m.previewViewport.Width = rightContentWidth
	m.previewViewport.Height = contentHeight
	m.tasksViewport.Width = leftContentWidth
	m.tasksViewport.Height = contentHeight
	m.statsViewport.Width = rightContentWidth
	m.statsViewport.Height = contentHeight

	m.updatePreviewViewport()
	m.updateStatsViewport()
}

func (m *Model) updatePreviewViewport() {
	content := m.previewContent
	if content == "" {
		content = "Select a note to preview"
	}

	contentWidth := m.previewViewport.Width
	if contentWidth < 10 {
		contentWidth = 10
	}

	lines := strings.Split(content, "\n")

	maxWidth := contentWidth - 1
	if maxWidth < 10 {
		maxWidth = 10
	}

	for i, line := range lines {
		if len(line) > maxWidth {
			lines[i] = line[:maxWidth-3] + "..."
		}
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

	content := " " + streakStyleBase.Foreground(streakColor).Render(fmt.Sprintf("%s %d day streak", streakIcon, m.streak)) + "\n\n"

	content += " " + labelStyle.Render("Notes") + "\n"
	content += fmt.Sprintf("  Total: %d\n", m.totalNotes)
	content += fmt.Sprintf("  Recent: %d\n\n", len(m.notes))

	content += " " + labelStyle.Render("Tasks") + "\n"
	pendingTasks := m.totalTasks - m.completedTasks
	content += fmt.Sprintf("  Pending: %d\n", pendingTasks)
	content += fmt.Sprintf("  Done: %d\n", m.completedTasks)

	if m.totalTasks > 0 {
		completion := float64(m.completedTasks) / float64(m.totalTasks) * 100

		barWidth := contentWidth - 3
		if barWidth > 20 {
			barWidth = 20
		}

		if barWidth < 5 {
			barWidth = 5
		}

		filled := int(float64(barWidth) * completion / 100)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

		var barColor lipgloss.AdaptiveColor
		if completion >= 80 {
			barColor = output.SuccessColor
		} else if completion >= 50 {
			barColor = output.WarningColor
		} else {
			barColor = output.SecondaryColor
		}

		barStyle := barStyleBase.Foreground(barColor)
		content += fmt.Sprintf("  %s %.0f%%\n", barStyle.Render(bar), completion)
	}

	m.statsViewport.SetContent(content)
}
