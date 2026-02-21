package tui

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/tasks"
)

type keyMap struct {
	Quit        key.Binding
	Tab         key.Binding
	TabReverse  key.Binding
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	NewTaskFile key.Binding
	Refresh     key.Binding
	Update      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Refresh, k.Enter, k.Update}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.TabReverse},
		{k.Enter, k.NewTaskFile, k.Refresh, k.Update, k.Quit},
	}
}

var defaultKeyMap = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	TabReverse: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "switch panel"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "navigate"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "navigate"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	NewTaskFile: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "create todo file"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Update: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "check updates"),
	),
}

func (m *Model) updateCachedKeyMap() {
	keys := m.keys

	enterEnabled := m.focusedPanel == panelNotes || m.focusedPanel == panelTasks
	keys.Enter.SetEnabled(enterEnabled)

	keys.NewTaskFile.SetEnabled(m.err != nil && m.errorRetryable)

	m.cachedKeyMap = keys
}

type panel int

const (
	panelNotes panel = iota
	panelPreview
	panelTasks
	panelStats
	numPanels = 4
)

const (
	minWidthForAscii  = 50
	minHeightForAscii = 40
)

// Model represents the TUI state for the dashboard.
// It manages notes, tasks, and statistics display with keyboard navigation.
type Model struct {
	ctx              context.Context
	err              error
	config           *config.LoadedConfig
	previewContent   string
	updateVersion    string
	statusMsg        string
	statusMsgTime    time.Time
	statusLevel      string
	statusDuration   time.Duration
	notes            []string
	tasks            []tasks.Task
	statsViewport    viewport.Model
	tasksViewport    viewport.Model
	notesViewport    viewport.Model
	previewViewport  viewport.Model
	helpModel        help.Model
	spinner          spinner.Model
	keys             keyMap
	cachedKeyMap     keyMap
	completedTasks   int
	selectedNote     int
	streak           int
	totalNotes       int
	totalTasks       int
	selectedTask     int
	focusedPanel     panel
	height           int
	width            int
	ready            bool
	quitting         bool
	errorRetryable   bool
	updateAvailable  bool
	editorConfigured bool
	editorFallback   bool
	isLoading        bool
	isNonInteractive bool
}

func NewModel(ctx context.Context, cfg *config.LoadedConfig) Model {
	helpModel := help.New()
	helpModel.Styles.ShortKey = helpModel.Styles.ShortKey.Foreground(output.SecondaryColor)
	helpModel.Styles.ShortDesc = helpModel.Styles.ShortDesc.Foreground(output.SecondaryColor)
	helpModel.Styles.ShortSeparator = helpModel.Styles.ShortSeparator.Foreground(output.SecondaryColor)

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = lipgloss.NewStyle().Foreground(secondaryColor)

	m := Model{
		ctx:              ctx,
		config:           cfg,
		focusedPanel:     panelNotes,
		notes:            []string{},
		tasks:            []tasks.Task{},
		notesViewport:    viewport.New(0, 0),
		previewViewport:  viewport.New(0, 0),
		tasksViewport:    viewport.New(0, 0),
		statsViewport:    viewport.New(0, 0),
		helpModel:        helpModel,
		spinner:          s,
		keys:             defaultKeyMap,
		width:            80, // Default width
		height:           24, // Default height (will be updated by WindowSizeMsg)
		statusLevel:      "",
		statusDuration:   0,
		isLoading:        true,
		isNonInteractive: !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()),
	}
	m.updateCachedKeyMap()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.spinner.Tick)
}
