package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/updater"
	"github.com/AnishShah1803/jotr/internal/version"
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
	return []key.Binding{k.Quit, k.Refresh, k.NewTaskFile, k.Update}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.TabReverse},
		{k.Enter, k.NewTaskFile, k.Refresh, k.Update, k.Quit},
	}
}

func newKeyMap() keyMap {
	return keyMap{
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
			key.WithHelp("n", "create file"),
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
}

func (m Model) getKeyMap() keyMap {
	keys := m.keys

	enterEnabled := m.focusedPanel == panelNotes || m.focusedPanel == panelTasks
	keys.Enter.SetEnabled(enterEnabled)

	return keys
}

// updateChecker is the interface for update checking.
// This allows for easier testing and decoupling from concrete implementation.
type updateChecker interface {
	CheckForUpdate(currentVersion string) (bool, string, error)
}

// defaultUpdateChecker wraps the updater package's CheckForUpdates function.
type defaultUpdateChecker struct{}

func (c *defaultUpdateChecker) CheckForUpdate(currentVersion string) (bool, string, error) {
	hasUpdate, latestVersion, _, err := updater.CheckForUpdates(currentVersion)
	return hasUpdate, latestVersion, err
}

type panel int

const (
	panelNotes panel = iota
	panelPreview
	panelTasks
	panelStats
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
	keys             keyMap
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
}

type tickMsg time.Time

type updateCheckMsg struct {
	err       error
	version   string
	hasUpdate bool
}

type errorMsg struct {
	err       error
	retryable bool
}

type editorFallbackMsg struct{}

type editorOpenAttemptMsg struct {
	useShellFallback bool
}

func newErrorMsg(err error, retryable bool) errorMsg {
	return errorMsg{err: err, retryable: retryable}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

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

func checkForUpdatesCmd() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, version, err := checkForUpdatesFromTUI()

		return updateCheckMsg{
			hasUpdate: hasUpdate,
			version:   version,
			err:       err,
		}
	}
}

func checkForUpdatesFromTUI() (bool, string, error) {
	var checker updateChecker = &defaultUpdateChecker{}
	currentVersion := version.GetVersion()

	return checker.CheckForUpdate(currentVersion)
}

func NewModel(ctx context.Context, cfg *config.LoadedConfig) Model {
	helpModel := help.New()
	helpModel.Styles.ShortKey = helpModel.Styles.ShortKey.Foreground(output.SecondaryColor)
	helpModel.Styles.ShortDesc = helpModel.Styles.ShortDesc.Foreground(output.SecondaryColor)

	return Model{
		ctx:             ctx,
		config:          cfg,
		focusedPanel:    panelNotes,
		notes:           []string{},
		tasks:           []tasks.Task{},
		notesViewport:   viewport.New(0, 0),
		previewViewport: viewport.New(0, 0),
		tasksViewport:   viewport.New(0, 0),
		statsViewport:   viewport.New(0, 0),
		helpModel:       helpModel,
		keys:            newKeyMap(),
		width:           80, // Default width
		height:          24, // Default height (will be updated by WindowSizeMsg)
		statusLevel:     "",
		statusDuration:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}
