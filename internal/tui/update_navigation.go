package tui

import tea "github.com/charmbracelet/bubbletea"

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
	case panelSearch:
		m.searchViewport.LineUp(1)
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
	case panelSearch:
		m.searchViewport.LineDown(1)
	}

	return m, nil
}

func (m *Model) cycleFocus(forward bool) {
	if forward {
		m.focusedPanel = (m.focusedPanel + 1) % numPanels
	} else {
		m.focusedPanel = (m.focusedPanel + numPanels - 1) % numPanels
	}
	m.updateCachedKeyMap()
}
