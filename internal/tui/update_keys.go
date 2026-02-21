package tui

import (
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AnishShah1803/jotr/internal/notes"
)

func handleKeyEvent(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Sequence(tea.ExitAltScreen, tea.Quit)

	case key.Matches(msg, m.keys.Tab):
		m.cycleFocus(true)
		return m, nil

	case key.Matches(msg, m.keys.TabReverse):
		m.cycleFocus(false)
		return m, nil

	case key.Matches(msg, m.keys.Up):
		return m.handleUp()

	case key.Matches(msg, m.keys.Down):
		return m.handleDown()

	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()

	case key.Matches(msg, m.keys.Refresh):
		m.err = nil
		m.errorRetryable = false
		m = setStatus(m, "Refreshing...", "info")
		m.updateCachedKeyMap()

		return m, m.loadData()

	case key.Matches(msg, m.keys.Update):
		m = setStatus(m, "Checking for updates...", "info")
		return m, checkForUpdatesCmd()

	case key.Matches(msg, m.keys.NewTaskFile):
		if m.err != nil && m.errorRetryable {
			err := createTodoFile(m.config.TodoPath)
			if err != nil {
				m = setStatus(m, "Failed to create file: "+err.Error(), "error")
			} else {
				m = setStatus(m, "Todo file created successfully", "success")
				m.err = nil
				m.errorRetryable = false
				m.updateCachedKeyMap()
				return m, m.loadData()
			}
			return m, nil
		}
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

func handleEditorFinished(m Model, msg editorFinishedMsg) (Model, tea.Cmd) {
	m = clearStatus(m)
	if msg.err != nil {
		m.err = msg.err
		m.errorRetryable = true
		m.updateCachedKeyMap()
	}

	return m, m.loadData()
}

func handleEditorOpenAttempt(m Model, msg editorOpenAttemptMsg) (Model, tea.Cmd) {
	m = clearStatus(m)
	if msg.useShellFallback {
		return m.handleFileOpen()
	}
	return m, nil
}

func handleEditorFallback(m Model) (Model, tea.Cmd) {
	m = clearStatus(m)
	if !m.editorConfigured {
		m = setStatus(m, "❌ No editor available - configure editor.default or set EDITOR env var", "error")
	}
	return m, nil
}
