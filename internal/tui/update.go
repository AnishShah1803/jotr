package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowSizeMsg(m, msg)

	case tea.KeyMsg:
		return handleKeyEvent(m, msg)

	case tickMsg:
		return handleTickMsg(m)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case dataLoadedMsg:
		return handleDataLoaded(m, msg)

	case previewLoadedMsg:
		return handlePreviewLoaded(m, msg)

	case editorFinishedMsg:
		return handleEditorFinished(m, msg)

	case editorOpenAttemptMsg:
		return handleEditorOpenAttempt(m, msg)

	case editorFallbackMsg:
		return handleEditorFallback(m)

	case updateCheckMsg:
		return handleUpdateCheck(m, msg)

	case errorMsg:
		return handleError(m, msg)

	case searchResultsMsg:
		return handleSearchResults(m, msg)
	}

	return m, nil
}

func handleWindowSizeMsg(m Model, msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	m.updateViewportSizes()

	if !m.ready {
		m.ready = true
		return m, m.loadData()
	}
	return m, nil
}
