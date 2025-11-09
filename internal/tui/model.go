package tui

import (
	"time"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/tasks"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type panel int

const (
	panelNotes panel = iota
	panelPreview
	panelTasks
	panelStats
)

type Model struct {
	config         *config.LoadedConfig
	width          int
	height         int
	focusedPanel   panel

	// Notes panel
	notes          []string
	selectedNote   int
	notesViewport  viewport.Model

	// Preview panel
	previewContent string
	previewViewport viewport.Model

	// Tasks panel
	tasks          []tasks.Task
	selectedTask   int
	tasksViewport  viewport.Model

	// Stats panel
	statsViewport  viewport.Model

	// Stats
	streak         int
	totalNotes     int
	totalTasks     int
	completedTasks int

	// State
	ready          bool
	quitting       bool
	err            error
	statusMsg      string
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewModel(cfg *config.LoadedConfig) Model {
	return Model{
		config:          cfg,
		focusedPanel:    panelNotes,
		notes:           []string{},
		tasks:           []tasks.Task{},
		notesViewport:   viewport.New(0, 0),
		previewViewport: viewport.New(0, 0),
		tasksViewport:   viewport.New(0, 0),
		statsViewport:   viewport.New(0, 0),
		width:           80,  // Default width
		height:          24,  // Default height (will be updated by WindowSizeMsg)
	}
}

func (m Model) Init() tea.Cmd {
	// Note: tea.EnterAltScreen is already handled by tea.WithAltScreen() in program initialization
	// Calling it here was redundant and may have been clearing the top margin
	return tickCmd()
}

