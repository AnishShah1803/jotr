package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/version"
)

var (
	primaryColor   = output.PrimaryColor
	secondaryColor = output.SecondaryColor
	accentColor    = output.AccentColor
	successColor   = output.SuccessColor
	warningColor   = output.WarningColor
	errorColor     = output.ErrorColor

	iconStreak = "🔥"
	iconEmpty  = "○"

	// Styles.
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

	// ASCII Art - Only used for large terminals (40+ lines, 50+ width).
	asciiArtLarge = `     ██╗ ██████╗ ████████╗██████╗
     ██║██╔═══██╗╚══██╔══╝██╔══██╗
     ██║██║   ██║   ██║   ██████╔╝
██   ██║██║   ██║   ██║   ██╔══██╗
╚█████╔╝╚██████╔╝   ██║   ██║  ██║
 ╚════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝`

	// Error style - base style without width (applied dynamically)
	errorStyleBase = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(1, 2).
			Foreground(errorColor)

	// Margin styles for small and large terminals
	marginStyleSmall = lipgloss.NewStyle().
				MarginTop(1).MarginBottom(0).MarginLeft(2).MarginRight(2)

	marginStyleLarge = lipgloss.NewStyle().
				Margin(1, 2)

	// Header styles
	asciiArtStyleBase = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Foreground(primaryColor)

	statusStyleBase = lipgloss.NewStyle().
			Align(lipgloss.Center)

	versionStyleBase = lipgloss.NewStyle().
				Align(lipgloss.Center).
				Foreground(secondaryColor)
)

func (m Model) errorStyle() lipgloss.Style {
	width := m.width - 8
	if width < 1 {
		width = 1
	}
	return errorStyleBase.Width(width)
}

func (m Model) marginStyle() lipgloss.Style {
	if m.height < 40 {
		return marginStyleSmall
	}
	return marginStyleLarge
}

func (m Model) asciiArtStyle() lipgloss.Style {
	width := m.width
	if width < 1 {
		width = 1
	}
	return asciiArtStyleBase.Width(width)
}

func (m Model) statusStyle(color lipgloss.AdaptiveColor) lipgloss.Style {
	width := m.width
	if width < 1 {
		width = 1
	}
	return statusStyleBase.Width(width).Foreground(color)
}

func (m Model) versionStyle() lipgloss.Style {
	width := m.width
	if width < 1 {
		width = 1
	}
	return versionStyleBase.Width(width)
}

func calculateLayoutHeight(header, main, footer string) int {
	return lipgloss.Height(header) + lipgloss.Height(main) + lipgloss.Height(footer)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading jotr dashboard...\n\n"
	}

	if m.quitting {
		return ""
	}

	if m.err != nil {
		var helpText string
		if m.errorRetryable {
			helpText = fmt.Sprintf("Press '%s' to create the file, '%s' to retry, or '%s' to quit",
				m.keys.NewTaskFile.Help().Key, m.keys.Refresh.Help().Key, m.keys.Quit.Help().Key)
		} else {
			helpText = fmt.Sprintf("Press '%s' to quit", m.keys.Quit.Help().Key)
		}

		errorTitle := "❌ Error"
		errorContent := fmt.Sprintf("%v\n\n%s", m.err, helpText)

		return "\n" + m.errorStyle().Render(errorTitle+"\n\n"+errorContent) + "\n"
	}

	// Calculate panel dimensions
	// Note: lipgloss .Width() and .Height() include borders in the dimension
	// So if we set Width(50) with a border, the content area is 50 - 2 = 48

	var headerFooterHeight int

	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		headerFooterHeight = 14 // Large terminal with ASCII art and footer
	} else {
		headerFooterHeight = 3 // Small terminal: minimal header + footer
	}

	// Calculate available width: account for margins, gaps, borders, and padding
	availableWidth := m.width - 14

	// Split width evenly between left and right panels
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
	mainWithMargin := m.marginStyle().Render(mainContent)

	// Add header and footer
	header := m.renderHeader()
	footer := m.renderFooter()

	usedHeight := calculateLayoutHeight(header, mainWithMargin, footer)

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
	var header string

	if m.height >= minHeightForAscii && m.width >= minWidthForAscii {
		centeredArt := m.asciiArtStyle().Render(asciiArtLarge)

		header = "\n\n" + centeredArt + "\n"
	} else {
		header = "\n"
	}

	if m.statusMsg != "" {
		var statusColor lipgloss.AdaptiveColor
		switch m.statusLevel {
		case "error":
			statusColor = errorColor
		case "success":
			statusColor = successColor
		case "warning":
			statusColor = warningColor
		default:
			statusColor = warningColor
		}
		status := m.statusStyle(statusColor).Render(m.statusMsg)
		header += status + "\n"
	} else if !m.updateAvailable {
		versionInfo := version.GetVersion()
		version := m.versionStyle().Render(versionInfo)
		header += version + "\n"
	}

	return header
}

func (m Model) renderFooter() string {
	help := m.helpModel.View(m.cachedKeyMap)

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
	contentWidth := width - 4 // Account for border and padding
	if contentWidth < 10 {
		contentWidth = 10
	}

	title := tStyle.Width(contentWidth).Render("Recent Notes")

	content := ""

	for i, notePath := range m.notes {
		basename := filepath.Base(notePath)
		basename = strings.TrimSuffix(basename, ".md")

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
	contentWidth := width - 4 // Account for border and padding
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
