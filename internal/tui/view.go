package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("86")
	secondaryColor = lipgloss.Color("240")
	accentColor    = lipgloss.Color("205")
	successColor   = lipgloss.Color("42")
	warningColor   = lipgloss.Color("214")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	focusedTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(accentColor)

	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTop(true).
			BorderRight(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	focusedPanelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderTop(true).
				BorderRight(true).
				BorderBottom(true).
				BorderLeft(true).
				BorderForeground(accentColor).
				Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Nerd Font Icons
	iconNote     = "" // nf-fa-file_text_o
	iconPreview  = "" // nf-fa-eye
	iconTask     = "" // nf-fa-check_square_o
	iconStats    = "" // nf-fa-bar_chart
	iconStreak   = "" // nf-fa-fire
	iconEmpty    = "" // nf-fa-circle_o

	// ASCII Art - Only used for large terminals (40+ lines, 50+ width)
	asciiArtLarge = `     ██╗ ██████╗ ████████╗██████╗
     ██║██╔═══██╗╚══██╔══╝██╔══██╗
     ██║██║   ██║   ██║   ██████╔╝
██   ██║██║   ██║   ██║   ██╔══██╗
╚█████╔╝╚██████╔╝   ██║   ██║  ██║
 ╚════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝`
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading jotr dashboard...\n\n"
	}

	if m.quitting {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press any key to continue...\n", m.err)
	}

	// Calculate panel dimensions
	// Note: lipgloss .Width() and .Height() include borders in the dimension
	// So if we set Width(50) with a border, the content area is 50 - 2 = 48

	// Header height calculation - dynamic based on terminal size
	// Only show ASCII art for large terminals (40+ lines, 50+ width)
	// Otherwise no header, just footer
	minWidthForAscii := 50
	minHeightForAscii := 40
	var headerFooterHeight int

	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		// Large terminal with full ASCII art
		// 1 (blank line above) + 6 (ASCII) + 1 (blank line below) + 2 (footer) + 2 (top/bottom margins) + 2 (extra) = 14
		headerFooterHeight = 14
	} else {
		// Smaller terminals: 1 blank line at top + footer
		// 1 (blank line at top) + 2 (footer with newline) = 3
		headerFooterHeight = 3
	}

	// Calculate dimensions
	// Account for margins: left=2, right=2 = 4 total
	// Account for gap between panels: 2 spaces
	// Account for borders on each panel: 2 per panel = 4 total for 2 panels
	// Account for padding inside borders: 2 per panel = 4 total for 2 panels
	// Total overhead: 4 (margins) + 2 (gap) + 4 (borders) + 4 (padding) = 14
	availableWidth := m.width - 14

	// Split width evenly between left and right panels
	// This is the CONTENT width (excluding borders and padding)
	leftPanelWidth := availableWidth / 2
	rightPanelWidth := availableWidth / 2
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

	// Render all panels
	notesPanel := m.renderNotesPanel(leftPanelWidth, panelHeight)
	previewPanel := m.renderPreviewPanel(rightPanelWidth, panelHeight)
	tasksPanel := m.renderTasksPanel(leftPanelWidth, panelHeight)
	statsPanel := m.renderStatsPanel(rightPanelWidth, panelHeight)

	// Combine panels with gap
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, notesPanel, "  ", previewPanel)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, tasksPanel, "  ", statsPanel)
	mainContent := lipgloss.JoinVertical(lipgloss.Left, topRow, "", bottomRow)

	// Add margin - always have top margin
	var mainWithMargin string
	if m.height < 40 {
		mainWithMargin = lipgloss.NewStyle().MarginTop(1).MarginBottom(0).MarginLeft(2).MarginRight(2).Render(mainContent)
	} else {
		mainWithMargin = lipgloss.NewStyle().Margin(1, 2).Render(mainContent)
	}

	// Add header and footer
	header := m.renderHeader()
	footer := m.renderFooter()

	// Calculate how much vertical space we have
	headerHeight := lipgloss.Height(header)
	mainHeight := lipgloss.Height(mainWithMargin)
	footerHeight := lipgloss.Height(footer)
	usedHeight := headerHeight + mainHeight + footerHeight

	// Add padding to push footer to bottom
	paddingNeeded := m.height - usedHeight
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}

	padding := strings.Repeat("\n", paddingNeeded)

	// Join all sections with padding before footer
	return lipgloss.JoinVertical(lipgloss.Left, header, mainWithMargin, padding, footer)
}

func (m Model) renderHeader() string {
	// Only show ASCII art for large terminals (40+ lines)
	minWidthForAscii := 50
	minHeightForAscii := 40

	var header string

	// Only show ASCII art if terminal is large enough
	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		// Large terminal: full ASCII art (6 lines)
		centeredArt := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(primaryColor).
			Render(asciiArtLarge)

		// Add blank lines above and below for spacing
		header = "\n\n" + centeredArt + "\n"
	} else {
		// For smaller terminals, add a single blank line at top
		header = "\n"
	}

	if m.statusMsg != "" {
		status := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Foreground(warningColor).
			Render(m.statusMsg)
		header += status + "\n"
	}

	return header
}

func (m Model) renderFooter() string {
	var helpText string

	switch m.focusedPanel {
	case panelNotes:
		helpText = "q: quit | tab: switch panel | ↑↓/jk: navigate | enter: open note | r: refresh"
	case panelPreview:
		helpText = "q: quit | tab: switch panel | ↑↓/jk: scroll | r: refresh"
	case panelTasks:
		helpText = "q: quit | tab: switch panel | ↑↓/jk: scroll | enter: open todo list | r: refresh"
	case panelStats:
		helpText = "q: quit | tab: switch panel | ↑↓/jk: scroll | r: refresh"
	default:
		helpText = "q: quit | tab: switch panel | r: refresh"
	}

	help := helpStyle.Render(helpText)
	return "\n" + help
}

func (m Model) renderNotesPanel(width, height int) string {
	var tStyle lipgloss.Style
	var style lipgloss.Style

	if m.focusedPanel == panelNotes {
		tStyle = focusedTitleStyle
		style = focusedPanelStyle
	} else {
		tStyle = titleStyle
		style = panelStyle
	}

	// Calculate content dimensions
	// Border takes 2 chars (1 on each side), padding takes 2 chars (1 on each side)
	// Total overhead: 4 chars for width
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	title := tStyle.Width(contentWidth).Render("Recent Notes")

	content := ""
	for i, notePath := range m.notes {
		basename := filepath.Base(notePath)
		basename = strings.TrimSuffix(basename, ".md")

		// Width for content: account for selection indicator (▶ = 2 chars) and small margin
		maxLen := contentWidth - 3
		if maxLen < 10 {
			maxLen = 10
		}
		if len(basename) > maxLen {
			basename = basename[:maxLen-3] + "..."
		}

		if i == m.selectedNote && m.focusedPanel == panelNotes {
			content += selectedItemStyle.Render(fmt.Sprintf("▶ %s", basename)) + "\n"
		} else {
			content += fmt.Sprintf(" %s\n", basename)
		}
	}

	if len(m.notes) == 0 {
		content = " No recent notes"
	}

	m.notesViewport.SetContent(content)

	panel := title + "\n" + m.notesViewport.View()

	return style.Width(width).Height(height).Render(panel)
}

func (m Model) renderPreviewPanel(width, height int) string {
	var tStyle lipgloss.Style
	var style lipgloss.Style

	if m.focusedPanel == panelPreview {
		tStyle = focusedTitleStyle
		style = focusedPanelStyle
	} else {
		tStyle = titleStyle
		style = panelStyle
	}

	// Calculate content width
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Render title with width constraint to prevent overflow
	title := tStyle.Width(contentWidth).Render("Preview")

	// Combine title and viewport
	panel := title + "\n" + m.previewViewport.View()

	// Render with border and exact width/height
	return style.Width(width).Height(height).Render(panel)
}

func (m Model) renderTasksPanel(width, height int) string {
	var tStyle lipgloss.Style
	var style lipgloss.Style

	if m.focusedPanel == panelTasks {
		tStyle = focusedTitleStyle
		style = focusedPanelStyle
	} else {
		tStyle = titleStyle
		style = panelStyle
	}

	// Calculate content dimensions
	// Border takes 2 chars, padding takes 2 chars = 4 total
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Render title with width constraint to prevent overflow
	title := tStyle.Width(contentWidth).Render("Tasks")

	// Build all content (not limited by height)
	content := ""
	count := 0

	for i, task := range m.tasks {
		if task.Completed {
			continue
		}

		taskText := task.Text
		if len(taskText) > contentWidth-3 {
			taskText = taskText[:contentWidth-6] + "..."
		}

		if i == m.selectedTask && m.focusedPanel == panelTasks {
			content += selectedItemStyle.Render(fmt.Sprintf("▶ %s", taskText)) + "\n"
		} else {
			content += fmt.Sprintf(" %s\n", taskText)
		}
		count++
	}

	if count == 0 {
		content = " No pending tasks"
	}

	// Set viewport content
	m.tasksViewport.SetContent(content)

	// Combine title and viewport
	panel := title + "\n" + m.tasksViewport.View()

	// Render with border and exact width/height
	return style.Width(width).Height(height).Render(panel)
}

func (m Model) renderStatsPanel(width, height int) string {
	var tStyle lipgloss.Style
	var style lipgloss.Style

	if m.focusedPanel == panelStats {
		tStyle = focusedTitleStyle
		style = focusedPanelStyle
	} else {
		tStyle = titleStyle
		style = panelStyle
	}

	// Calculate content width
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Render title with width constraint to prevent overflow
	title := tStyle.Width(contentWidth).Render("Quick Stats")

	// Combine title and viewport
	panel := title + "\n" + m.statsViewport.View()

	// Render with border and exact width/height
	return style.Width(width).Height(height).Render(panel)
}

