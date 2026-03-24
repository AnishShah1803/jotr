package repl

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	taskcmd "github.com/AnishShah1803/jotr/cmd/task"
	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/services"
	"github.com/AnishShah1803/jotr/internal/version"
)

type Model struct {
	ctx                  context.Context
	config               *config.LoadedConfig
	rootCmd              *cobra.Command
	textInput            textinput.Model
	history              *History
	autocomplete         *Autocomplete
	width                int
	height               int
	ready                bool
	quitting             bool
	completions          []string
	selectedIdx          int
	completionsOffset    int
	browsingCompletions  bool
	inHistoryNav         bool
	taskAddPromptActive  bool
	taskAddPromptStep    int
	taskAddPromptValues  []string
	taskAddPromptError   string
	taskPromptMode       string
	taskPromptActive     bool
	taskPromptStep       int
	taskPromptValues     []string
	taskPromptError      string
	taskEditActive       bool
	taskEditStep         int
	taskEditValues       []string
	taskEditError        string
	taskEditTaskNumber   int
	taskEditTaskID       string
	taskEditExistingText string
	taskEditPriority     string
	taskEditTags         string
	streakResult         services.StreakResult
	transcript           []transcriptEntry
}

type transcriptEntry struct {
	command string
	output  string
}

const (
	transcriptTaskListPending   = "task list"
	transcriptTaskListForEditor = "task list (for edit)"
)

const replAsciiArt = `      ░░
      ░░▒░
      ░░░▒░
     ░░░░▒▒
    ░░░░░░▒█░
    ░░░░░░▒██░
   ░░░░░░▒▒██▒
   ░ ░░░░▒░████
   ░░ ░░▒▒▒████
    ░  ▒▒░▒███▒
    ░░ ░░▒████▌
     ▒░░▒███▛
       ▀▀▀▀    `

const replAsciiArtHeight = 13

var (
	primaryColor   = output.PrimaryColor
	secondaryColor = output.SecondaryColor
	accentColor    = output.AccentColor

	dropColor = lipgloss.Color("#c97a98")

	replLogoStyle = lipgloss.NewStyle().Foreground(dropColor).Bold(true)

	logoStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	separatorStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	promptStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	completionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	selectedCompletionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("51")).
				Bold(true)
	defaultREPLPrompt      = ""
	defaultREPLPlaceholder = "type a command..."
)

const completionsMaxLines = 10
const transcriptMaxEntries = 10

func NewModel(ctx context.Context, cfg *config.LoadedConfig, rootCmd *cobra.Command) Model {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.Prompt = ""
	ti.PromptStyle = promptStyle
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(accentColor)
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60
	ti.TextStyle = inputStyle
	ti.PlaceholderStyle = completionStyle

	m := Model{
		ctx:                 ctx,
		config:              cfg,
		rootCmd:             rootCmd,
		textInput:           ti,
		history:             NewHistory(),
		autocomplete:        NewAutocomplete(rootCmd),
		width:               80,
		height:              24,
		ready:               false,
		quitting:            false,
		completions:         []string{},
		selectedIdx:         0,
		streakResult:        services.CalculateStreak(cfg),
		taskAddPromptValues: make([]string, 0, 3),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.ClearScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		wasReady := m.ready
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = m.width - 10
		if m.textInput.Width < 20 {
			m.textInput.Width = 20
		}
		m.ready = true
		if !wasReady {
			return m, nil
		}
		return m, nil

	case tea.KeyMsg:
		if m.taskAddPromptActive {
			switch msg.Type {
			case tea.KeyEsc, tea.KeyCtrlC:
				m.clearTaskAddPrompt()
				m.textInput.SetValue("")
				m.clearCompletionState()
				return m, nil

			case tea.KeyEnter:
				value := strings.TrimSpace(m.textInput.Value())
				if m.taskAddPromptStep == 0 && value == "" {
					m.taskAddPromptError = "Task text is required"
					return m, nil
				}

				m.taskAddPromptError = ""
				m.taskAddPromptValues = append(m.taskAddPromptValues, value)
				m.textInput.SetValue("")
				m.textInput.CursorStart()

				if m.taskAddPromptStep >= 2 {
					m.history.Add("task add")
					out, err := m.runTaskAddFromPrompt()
					m.appendTranscript("task add", out)
					m.clearTaskAddPrompt()
					m.clearCompletionState()
					if err != nil {
						m.appendTranscript("task add", fmt.Sprintf("Error: %v", err))
					}
					return m, nil
				}

				m.taskAddPromptStep++
				m.textInput.Prompt = m.taskAddPromptLabel()
				return m, nil
			}

			prevValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.browsingCompletions && m.textInput.Value() != prevValue {
				m.browsingCompletions = false
			}
			return m, cmd
		}
		if m.taskPromptActive {
			switch msg.Type {
			case tea.KeyEsc, tea.KeyCtrlC:
				m.clearTaskPrompt()
				m.textInput.SetValue("")
				m.clearCompletionState()
				return m, nil

			case tea.KeyEnter:
				value := strings.TrimSpace(m.textInput.Value())
				if value == "" {
					m.taskPromptError = "Task number(s) are required"
					return m, nil
				}

				m.taskPromptValues = append(m.taskPromptValues, value)
				m.textInput.SetValue("")
				m.textInput.CursorStart()

				out, err := m.runTaskPrompt()
				m.appendTranscript("task "+m.taskPromptMode, out)
				m.clearTaskPrompt()
				m.clearCompletionState()
				if err != nil {
					m.appendTranscript("task "+m.taskPromptMode, fmt.Sprintf("Error: %v", err))
				}
				return m, nil
			}

			prevValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.browsingCompletions && m.textInput.Value() != prevValue {
				m.browsingCompletions = false
			}
			return m, cmd
		}
		if m.taskEditActive {
			switch msg.Type {
			case tea.KeyEsc, tea.KeyCtrlC:
				m.clearTaskEdit()
				m.textInput.SetValue("")
				m.clearCompletionState()
				return m, nil

			case tea.KeyEnter:
				value := strings.TrimSpace(m.textInput.Value())
				if m.taskEditStep == 0 && value == "" {
					m.taskEditError = "Task number is required"
					return m, nil
				}
				if m.taskEditStep == 1 && value == "" {
					m.taskEditError = "Task text is required"
					return m, nil
				}

				m.taskEditValues = append(m.taskEditValues, value)
				m.taskEditError = ""
				if m.taskEditStep == 0 {
					if err := m.beginTaskEditTextPrompt(); err != nil {
						m.taskEditError = err.Error()
					}
					return m, nil
				}

				if m.taskEditStep >= 3 {
					out, err := m.runTaskEditFromPrompt()
					m.appendTranscript("task edit", out)
					m.clearTaskEdit()
					m.textInput.SetValue("")
					m.clearCompletionState()
					if err != nil {
						m.appendTranscript("task edit", fmt.Sprintf("Error: %v", err))
					}
					return m, nil
				}

				m.taskEditStep++
				m.textInput.Prompt = m.taskEditPromptLabel()
				m.textInput.SetValue(m.taskEditPromptDefaultValue())
				m.textInput.CursorEnd()
				return m, nil
			}

			prevValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.browsingCompletions && m.textInput.Value() != prevValue {
				m.browsingCompletions = false
			}
			return m, cmd
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.textInput.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
			m.textInput.SetValue("")
			m.clearCompletionState()
			m.history.Reset()
			return m, nil

		case tea.KeyEnter:
			if m.browsingCompletions && len(m.completions) > 0 {
				// Accept the highlighted completion into the input; don't execute.
				selected := m.completions[m.selectedIdx]
				m.textInput.SetValue(selected)
				m.textInput.CursorEnd()
				m.browsingCompletions = false
				m.selectedIdx = 0
				m.completionsOffset = 0
				m.updateCompletions()
				return m, nil
			}
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}

			m.history.Add(input)
			if m.startTaskAddPrompt(input) {
				return m, nil
			}
			if m.startTaskEditPrompt(input) {
				return m, nil
			}
			if m.startTaskPrompt(input) {
				return m, nil
			}
			cmdOutput := m.executeCommand(input)
			m.appendTranscript(input, cmdOutput)

			m.textInput.SetValue("")
			m.textInput.CursorStart()
			m.clearCompletionState()

			if m.quitting {
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyUp:
			if m.browsingCompletions {
				m.selectedIdx--
				if m.selectedIdx < 0 {
					m.browsingCompletions = false
					m.selectedIdx = 0
					m.completionsOffset = 0
				} else {
					if m.selectedIdx < m.completionsOffset {
						m.completionsOffset = m.selectedIdx
					}
				}
				return m, nil
			}
			// Only navigate history when input is empty.
			if m.textInput.Value() == "" {
				prev := m.history.Previous()
				if prev != "" {
					m.inHistoryNav = true
					m.textInput.SetValue(prev)
					m.textInput.CursorEnd()
				}
				m.updateCompletions()
			}
			return m, nil

		case tea.KeyDown:
			if m.browsingCompletions {
				if m.selectedIdx < len(m.completions)-1 {
					m.selectedIdx++
					if m.selectedIdx >= m.completionsOffset+completionsMaxLines {
						m.completionsOffset++
					}
				}
				return m, nil
			}

			if m.inHistoryNav {
				// Navigate forward through history.
				next := m.history.Next()
				if next != "" {
					m.textInput.SetValue(next)
					m.textInput.CursorEnd()
					m.updateCompletions()
				} else {
					// Reached blank state — exit history nav, restore empty input.
					// The next ↓ press will enter completion browsing.
					m.inHistoryNav = false
					m.textInput.SetValue("")
					m.updateCompletions()
				}
			} else if len(m.completions) > 0 {
				// Not in history nav — enter completion browsing directly.
				m.browsingCompletions = true
				m.selectedIdx = 0
				m.completionsOffset = 0
			}
			return m, nil

		case tea.KeyTab:
			current := m.textInput.Value()
			if len(m.completions) > 0 {
				selected := m.completions[m.selectedIdx]

				subCommands := m.autocomplete.GetSubCommands(selected)

				if subCommands != nil {
					m.textInput.SetValue(selected + " ")
				} else {
					m.textInput.SetValue(selected)
				}
				m.textInput.CursorEnd()
				m.selectedIdx = 0
				m.updateCompletions()
			} else {
				completed := m.autocomplete.Complete(current)
				if completed != current {
					m.textInput.SetValue(completed)
					m.textInput.CursorEnd()
				}
				m.updateCompletions()
			}
			return m, nil

		case tea.KeyShiftTab:
			if len(m.completions) > 0 {
				if !m.browsingCompletions {
					// Enter completion browsing from the last item.
					m.browsingCompletions = true
					m.selectedIdx = len(m.completions) - 1
					m.completionsOffset = 0
				} else {
					m.selectedIdx--
					if m.selectedIdx < 0 {
						m.selectedIdx = len(m.completions) - 1
					}
					if m.selectedIdx < m.completionsOffset {
						m.completionsOffset = m.selectedIdx
					}
				}
			}
			return m, nil
		}

	}

	prevValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	if m.browsingCompletions && m.textInput.Value() != prevValue {
		m.browsingCompletions = false
	}
	m.updateCompletions()
	return m, cmd
}

func (m *Model) startTaskAddPrompt(input string) bool {
	parts := strings.Fields(input)
	if len(parts) != 2 || parts[0] != "task" || parts[1] != "add" {
		return false
	}

	m.taskAddPromptActive = true
	m.taskAddPromptStep = 0
	m.taskAddPromptValues = m.taskAddPromptValues[:0]
	m.taskAddPromptError = ""
	m.textInput.SetValue("")
	m.textInput.Prompt = m.taskAddPromptLabel()
	m.textInput.Placeholder = ""
	m.textInput.CursorStart()
	m.clearCompletionState()
	return true
}

func (m *Model) startTaskPrompt(input string) bool {
	parts := strings.Fields(input)
	if len(parts) != 2 || parts[0] != "task" {
		return false
	}

	switch parts[1] {
	case "complete":
		if err := m.ensureTaskListContext(false); err != nil {
			m.appendTranscript("task complete", fmt.Sprintf("Error: %v", err))
			return true
		}
		m.taskPromptMode = "complete"
		m.taskPromptActive = true
		m.taskPromptStep = 0
		m.taskPromptValues = m.taskPromptValues[:0]
		m.taskPromptError = ""
		m.appendTaskListBeforePrompt("complete")
		m.textInput.SetValue("")
		m.textInput.Prompt = "Task numbers to complete: "
		m.clearCompletionState()
		return true
	default:
		return false
	}
}

func (m *Model) clearTaskPrompt() {
	m.taskPromptMode = ""
	m.taskPromptActive = false
	m.taskPromptStep = 0
	m.taskPromptValues = m.taskPromptValues[:0]
	m.taskPromptError = ""
	if !m.taskAddPromptActive {
		m.textInput.Prompt = defaultREPLPrompt
		m.textInput.Placeholder = defaultREPLPlaceholder
	}
}

func (m *Model) runTaskPrompt() (string, error) {
	switch m.taskPromptMode {
	case "complete":
		err := taskcmd.RunTaskComplete(m.ctx, m.config, []string{strings.Join(m.taskPromptValues, " ")})
		return "", err
	default:
		return "", fmt.Errorf("unknown task prompt mode")
	}
}

func (m *Model) startTaskEditPrompt(input string) bool {
	parts := strings.Fields(input)
	if len(parts) != 2 || parts[0] != "task" || parts[1] != "edit" {
		return false
	}

	if err := m.ensureTaskListContext(true); err != nil {
		m.appendTranscript("task edit", fmt.Sprintf("Error: %v", err))
		return true
	}

	m.taskEditActive = true
	m.taskEditStep = 0
	m.taskEditValues = m.taskEditValues[:0]
	m.taskEditError = ""
	m.appendTaskListBeforePrompt("edit")
	m.textInput.SetValue("")
	m.textInput.Prompt = "Task number to edit: "
	m.clearCompletionState()
	return true
}

func (m *Model) appendTaskListBeforePrompt(mode string) {
	if m.lastTranscriptIsTaskList() {
		return
	}

	listOutput, err := m.renderTaskListForPrompt(mode)
	if strings.TrimSpace(listOutput) != "" {
		m.appendTranscript("task "+mode, listOutput)
		return
	}
	if err != nil {
		m.appendTranscript("task "+mode, fmt.Sprintf("Error: %v", err))
	}
}

func (m *Model) renderTaskListForPrompt(mode string) (string, error) {
	switch mode {
	case "complete":
		listOutput, err := captureProcessOutput(func() error {
			return taskcmd.RunTaskComplete(m.ctx, m.config, []string{})
		})
		if strings.TrimSpace(listOutput) != "" {
			return listOutput, nil
		}
		return "", err
	case "edit":
		listOutput, err := captureProcessOutput(func() error {
			return taskcmd.RunTaskEdit(m.ctx, m.config, []string{})
		})
		if strings.TrimSpace(listOutput) != "" {
			return listOutput, nil
		}
		return "", err
	default:
		return "", fmt.Errorf("unknown task prompt mode")
	}
}

func (m *Model) lastTranscriptIsTaskList() bool {
	if len(m.transcript) == 0 {
		return false
	}

	last := m.transcript[len(m.transcript)-1]
	if last.command == "task list" || last.command == "task ls" {
		return true
	}

	return looksLikeTaskListOutput(last.output)
}

func looksLikeTaskListOutput(output string) bool {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return false
	}

	lines := strings.Split(trimmed, "\n")
	hasHeader := false
	hasItem := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			hasHeader = true
			continue
		}

		i := 0
		for i < len(line) && line[i] >= '0' && line[i] <= '9' {
			i++
		}
		if i > 0 && i+1 < len(line) && line[i] == '.' && line[i+1] == ' ' {
			hasItem = true
		}
	}

	return hasHeader && hasItem
}

func (m *Model) beginTaskEditTextPrompt() error {
	if len(m.taskEditValues) == 0 {
		return fmt.Errorf("task number is required")
	}

	if len(m.taskEditValues) > 0 {
		if _, err := strconv.Atoi(strings.TrimSpace(m.taskEditValues[0])); err != nil {
			return fmt.Errorf("invalid task number: %w", err)
		}
	}

	ordered, err := taskcmd.LoadTasksForEdit(m.ctx, m.config)
	if err != nil {
		return err
	}
	if len(ordered) == 0 {
		return fmt.Errorf("no tasks to edit")
	}

	n, err := strconv.Atoi(strings.TrimSpace(m.taskEditValues[0]))
	if err != nil {
		return fmt.Errorf("invalid task number: %w", err)
	}
	if n < 1 || n > len(ordered) {
		return fmt.Errorf("task number %d is out of range", n)
	}

	selected := ordered[n-1]
	editableText := selected.Text
	if selected.Priority != "" {
		editableText = strings.ReplaceAll(editableText, "["+selected.Priority+"]", "")
	}
	if len(selected.Tags) > 0 {
		for _, tag := range selected.Tags {
			editableText = strings.ReplaceAll(editableText, "#"+tag, "")
		}
	}
	editableText = strings.Join(strings.Fields(strings.TrimSpace(editableText)), " ")

	m.taskEditTaskNumber = n
	m.taskEditTaskID = selected.ID
	m.taskEditExistingText = editableText
	m.taskEditPriority = selected.Priority
	m.taskEditTags = strings.Join(selected.Tags, ",")
	m.taskEditStep = 1
	m.taskEditValues = m.taskEditValues[:0]
	m.textInput.SetValue(editableText)
	m.textInput.CursorEnd()
	m.textInput.Prompt = m.taskEditPromptLabel()
	return nil
}

func (m *Model) taskEditPromptLabel() string {
	switch m.taskEditStep {
	case 1:
		return "Task text: "
	case 2:
		return "Priority (P0-P3, press enter to skip): "
	case 3:
		return "Tags (comma-separated, press enter to skip): "
	default:
		return "Task number to edit: "
	}
}

func (m *Model) taskEditPromptDefaultValue() string {
	switch m.taskEditStep {
	case 2:
		return strings.TrimPrefix(strings.TrimSpace(strings.ToUpper(m.taskEditPriority)), "P")
	case 3:
		return m.taskEditTags
	default:
		return ""
	}
}

func (m *Model) clearTaskEdit() {
	m.taskEditActive = false
	m.taskEditStep = 0
	m.taskEditValues = m.taskEditValues[:0]
	m.taskEditError = ""
	m.taskEditTaskNumber = 0
	m.taskEditTaskID = ""
	m.taskEditExistingText = ""
	m.taskEditPriority = ""
	m.taskEditTags = ""
	if !m.taskAddPromptActive && !m.taskPromptActive {
		m.textInput.Prompt = defaultREPLPrompt
		m.textInput.Placeholder = defaultREPLPlaceholder
	}
}

func (m *Model) runTaskEditFromPrompt() (string, error) {
	if len(m.taskEditValues) == 0 {
		return "", fmt.Errorf("task edit prompt is incomplete")
	}

	newText := strings.TrimSpace(m.taskEditValues[0])
	priority := ""
	if len(m.taskEditValues) > 1 {
		priority = strings.TrimSpace(m.taskEditValues[1])
	}
	var tags []string
	if len(m.taskEditValues) > 2 && strings.TrimSpace(m.taskEditValues[2]) != "" {
		for _, part := range strings.Split(m.taskEditValues[2], ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	taskService := services.NewTaskService()
	_, err := taskService.UpdateTask(m.ctx, services.UpdateTaskOptions{
		DiaryPath: m.config.DiaryPath,
		TodoPath:  m.config.TodoPath,
		StatePath: m.config.StatePath,
		TaskID:    m.taskEditTaskID,
		Text:      newText,
		Priority:  priority,
		Tags:      tags,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated task %d\n", m.taskEditTaskNumber), nil
}

func (m *Model) clearTaskAddPrompt() {
	m.taskAddPromptActive = false
	m.taskAddPromptStep = 0
	m.taskAddPromptValues = m.taskAddPromptValues[:0]
	m.taskAddPromptError = ""
	if !m.taskPromptActive {
		m.textInput.Prompt = defaultREPLPrompt
		m.textInput.Placeholder = defaultREPLPlaceholder
	}
}

func (m *Model) taskAddPromptLabel() string {
	switch m.taskAddPromptStep {
	case 0:
		return "Task text: "
	case 1:
		return "Priority (P0-P3, press enter to skip): "
	case 2:
		return "Tags (comma-separated, press enter to skip): "
	default:
		return "Task text: "
	}
}

func (m *Model) runTaskAddFromPrompt() (string, error) {
	if len(m.taskAddPromptValues) < 3 {
		return "", fmt.Errorf("task add prompt is incomplete")
	}

	text := m.taskAddPromptValues[0]
	priority := m.taskAddPromptValues[1]
	tagsRaw := m.taskAddPromptValues[2]

	var tags []string
	if tagsRaw != "" {
		for _, part := range strings.Split(tagsRaw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	return taskcmd.RunTaskAdd(m.ctx, m.config, text, priority, tags)
}

func (m *Model) clearCompletionState() {
	m.completions = nil
	m.selectedIdx = 0
	m.completionsOffset = 0
	m.browsingCompletions = false
	m.inHistoryNav = false
}

func (m *Model) appendTranscript(command, output string) {
	m.transcript = append(m.transcript, transcriptEntry{
		command: command,
		output:  strings.TrimSpace(output),
	})
	if len(m.transcript) > transcriptMaxEntries {
		m.transcript = m.transcript[len(m.transcript)-transcriptMaxEntries:]
	}
}

func (m *Model) hasTaskListContext(includeCompleted bool) bool {
	if len(m.transcript) == 0 {
		return false
	}
	last := m.transcript[len(m.transcript)-1]
	if includeCompleted {
		return last.command == transcriptTaskListForEditor || last.command == transcriptTaskListPending
	}
	return last.command == transcriptTaskListPending
}

func (m *Model) ensureTaskListContext(includeCompleted bool) error {
	if m.hasTaskListContext(includeCompleted) {
		return nil
	}
	out, err := captureProcessOutput(func() error {
		return taskcmd.RunTaskList(m.ctx, m.config, includeCompleted)
	})
	if err != nil {
		return err
	}
	cmd := transcriptTaskListPending
	if includeCompleted {
		cmd = transcriptTaskListForEditor
	}
	m.appendTranscript(cmd, out)
	return nil
}

func (m *Model) updateCompletions() {
	value := m.textInput.Value()
	fields := strings.Fields(value)

	var newCompletions []string
	if value == "" {
		newCompletions = m.autocomplete.GetAllCommands()
	} else if len(fields) == 1 {
		cmdName := fields[0]
		prefixMatches := m.autocomplete.GetCompletions(cmdName)
		if m.autocomplete.IsCommand(cmdName) && len(prefixMatches) == 1 {
			// Exact unambiguous match — show subcommands or params.
			subCommands := m.autocomplete.GetSubCommands(cmdName)
			if len(subCommands) > 0 {
				newCompletions = subCommands
			} else if m.autocomplete.IsParamCommand(cmdName) {
				newCompletions = m.autocomplete.GetParamCompletions(cmdName)
			} else {
				newCompletions = prefixMatches
			}
		} else {
			// Partial or ambiguous — show prefix matches.
			newCompletions = prefixMatches
		}
	} else {
		newCompletions = m.getCompletionsForInput(fields, strings.HasSuffix(value, " "))
	}

	m.completions = newCompletions
	if !m.browsingCompletions {
		m.selectedIdx = 0
	} else {
		// Clamp in case the list shrank.
		if m.selectedIdx >= len(m.completions) {
			m.selectedIdx = len(m.completions) - 1
		}
		if m.selectedIdx < 0 {
			m.selectedIdx = 0
		}
		// Clamp offset so the selected item is always visible.
		if m.completionsOffset > m.selectedIdx {
			m.completionsOffset = m.selectedIdx
		}
		if m.completionsOffset < 0 {
			m.completionsOffset = 0
		}
	}
}

func (m *Model) getCompletionsForInput(fields []string, hasTrailingSpace bool) []string {
	if len(fields) == 0 {
		return []string{}
	}

	// Split fields into the command prefix (no `=`) and already-used param tokens.
	cmdFields := []string{}
	usedParams := map[string]bool{}
	for _, f := range fields {
		if strings.Contains(f, "=") {
			key := strings.SplitN(f, "=", 2)[0]
			usedParams[key+"="] = true
		} else {
			cmdFields = append(cmdFields, f)
		}
	}

	// If the last field is a partial param token, treat it as the partial being typed.
	var partial string
	var partialIsParam bool
	if !hasTrailingSpace && len(fields) > 0 {
		last := fields[len(fields)-1]
		if strings.Contains(last, "=") {
			parts := strings.SplitN(last, "=", 2)
			key := parts[0]
			value := parts[1]
			if value == "" {
				// e.g. "configure diaryDir=" — user is typing the param key up to/including =
				partialIsParam = true
				// Remove it from usedParams so it's still offered as a completion.
				usedParams[key+"="] = false
				partial = key + "="
			}
			// If value != "", the param is complete (e.g., "diaryDir=/some/path")
			// and should remain in usedParams
		} else {
			// Last field has no `=`: it's either a partial subcommand name or partial param key.
			if len(cmdFields) > 0 {
				partial = cmdFields[len(cmdFields)-1]
				cmdFields = cmdFields[:len(cmdFields)-1]
			}
		}
	}

	cmdPath := strings.Join(cmdFields, " ")

	if !hasTrailingSpace && len(fields) > 0 {
		last := fields[len(fields)-1]
		if strings.Contains(last, "=") {
			parts := strings.SplitN(last, "=", 2)
			key := parts[0]
			value := parts[1]
			valueCompletions := m.autocomplete.GetParamValueCompletionsFromRegistry(cmdPath, key, value)
			if len(valueCompletions) > 0 {
				committedBefore := strings.Join(fields[:len(fields)-1], " ")
				var result []string
				for _, pc := range valueCompletions {
					entry := key + "=" + pc
					if committedBefore != "" {
						entry = committedBefore + " " + entry
					}
					result = append(result, entry)
				}
				return result
			}
		}
	}
	if !partialIsParam && len(usedParams) == 0 {
		if hasTrailingSpace {
			subs := m.autocomplete.GetSubCommands(cmdPath)
			if subs != nil {
				return subs
			}
			// cmdPath may be a leaf subcommand — fall through to param lookup.
			// Only echo the leaf if there are no params registered.
			if parentPath, _, ok := strings.Cut(cmdPath, " "); ok {
				if m.autocomplete.GetSubCommands(parentPath) != nil {
					if m.autocomplete.GetParamsForCommand(cmdPath) == nil {
						return []string{cmdPath}
					}
					// Has params — fall through to param section below.
				}
			}
		} else {
			subs := m.autocomplete.GetSubCommandCompletions(cmdPath, partial)
			if subs != nil {
				return subs
			}
			// Only echo the leaf if there are no params; otherwise fall through.
			if parentPath, _, ok := strings.Cut(cmdPath, " "); ok {
				if m.autocomplete.GetSubCommands(parentPath) != nil {
					if m.autocomplete.GetParamsForCommand(cmdPath) == nil {
						return []string{cmdPath}
					}
					// Has params — fall through to param section below.
				}
			}
		}
	}

	// Resolve the param set for the deepest matching command path.
	params := m.autocomplete.GetParamsForCommand(cmdPath)
	if params == nil {
		for i := len(cmdFields); i > 0; i-- {
			p := strings.Join(cmdFields[:i], " ")
			if pp := m.autocomplete.GetParamsForCommand(p); pp != nil {
				params = pp
				break
			}
		}
	}
	if params == nil {
		return []string{}
	}

	// Build the already-committed portion of the input for prefixing suggestions.
	var committedPrefix string
	if hasTrailingSpace {
		committedPrefix = strings.Join(fields, " ")
	} else {
		committedPrefix = strings.Join(fields[:len(fields)-1], " ")
	}

	// Remove already-used params from the available set.
	var remaining []string
	for _, p := range params {
		if !usedParams[p] {
			remaining = append(remaining, p)
		}
	}

	if hasTrailingSpace {
		var result []string
		for _, p := range remaining {
			result = append(result, committedPrefix+" "+p)
		}
		return result
	}

	// Filter remaining by prefix match on the partial token.
	var result []string
	for _, p := range remaining {
		if strings.HasPrefix(p, partial) {
			result = append(result, committedPrefix+" "+p)
		}
	}
	return result
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading jotr...\n\n"
	}

	if m.quitting {
		return ""
	}

	var b strings.Builder
	header, _ := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n\n")

	inset := 2
	if len(m.transcript) > 0 {
		transcriptSeparator := strings.Repeat(" ", inset) + strings.Repeat("─", 40)
		b.WriteString(separatorStyle.Render(transcriptSeparator))
		b.WriteString("\n\n")
	}

	for _, entry := range m.transcript {
		b.WriteString(strings.Repeat(" ", inset))
		b.WriteString(promptStyle.Render("❯ "))
		b.WriteString(inputStyle.Render(entry.command))
		b.WriteString("\n")
		if entry.output != "" {
			for _, line := range strings.Split(strings.TrimRight(entry.output, "\n"), "\n") {
				b.WriteString(strings.Repeat(" ", inset*2))
				b.WriteString(line)
				b.WriteString("\n")
			}
		} else {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	separator := strings.Repeat(" ", inset) + strings.Repeat("─", 40)
	b.WriteString(separatorStyle.Render(separator))
	b.WriteString("\n")

	b.WriteString(strings.Repeat(" ", inset))
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(m.textInput.View())
	if (m.taskAddPromptActive && m.taskAddPromptError != "") || (m.taskPromptActive && m.taskPromptError != "") || (m.taskEditActive && m.taskEditError != "") {
		b.WriteString("\n")
		b.WriteString(strings.Repeat(" ", inset))
		if m.taskAddPromptActive {
			b.WriteString(helpStyle.Render(m.taskAddPromptError))
		} else if m.taskEditActive {
			b.WriteString(helpStyle.Render(m.taskEditError))
		} else {
			b.WriteString(helpStyle.Render(m.taskPromptError))
		}
	}

	if !m.taskAddPromptActive && !m.taskPromptActive && !m.taskEditActive && len(m.completions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderCompletions(inset))
	}

	return b.String()
}

func (m Model) renderCompletions(inset int) string {
	insetStr := strings.Repeat(" ", inset)
	lines := make([]string, completionsMaxLines)

	for i := 0; i < completionsMaxLines; i++ {
		ci := m.completionsOffset + i
		if ci < len(m.completions) {
			completion := m.completions[ci]
			if ci == m.selectedIdx && m.browsingCompletions {
				lines[i] = insetStr + selectedCompletionStyle.Render("> "+completion)
			} else if ci == m.selectedIdx {
				lines[i] = insetStr + selectedCompletionStyle.Render("  "+completion)
			} else {
				lines[i] = insetStr + completionStyle.Render("  "+completion)
			}
		} else {
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n")
}

func renderLogo() string {
	return replLogoStyle.Render(replAsciiArt)
}

func (m Model) renderHeader() (string, int) {
	const helpHint = "tab to autocomplete · ↑ history · ↓ scroll options · ctrl+c to quit"
	const minWidthForArt = 50
	const minHeightForArt = 15

	leftPad := 2
	leftPadStr := strings.Repeat(" ", leftPad)

	streak := m.streakResult
	var streakText string
	if streak.CurrentStreak > 0 {
		streakText = versionStyle.Render(fmt.Sprintf("%d day streak 🔥", streak.CurrentStreak))
	} else {
		streakText = helpStyle.Render("no streak yet")
	}

	if m.width >= minWidthForArt && m.height >= minHeightForArt {
		artRendered := renderLogo()
		artWidth := lipgloss.Width(artRendered)

		gapRight := m.width * 2 / 100
		if gapRight < 1 {
			gapRight = 1
		}
		gapBelow := m.height * 3 / 100
		if gapBelow < 2 {
			gapBelow = 2
		}

		verText := versionStyle.Render("jotr " + version.GetVersion())
		verPadded := strings.Repeat("\n", replAsciiArtHeight-1) + verText + "\n" + streakText
		rightColWidth := m.width - artWidth - leftPad - gapRight
		if rightColWidth < 20 {
			rightColWidth = 20
		}
		rightCol := lipgloss.NewStyle().Width(rightColWidth).Render(verPadded)

		gapStr := strings.Repeat(" ", gapRight)
		row := lipgloss.JoinHorizontal(lipgloss.Top, leftPadStr, artRendered, gapStr, rightCol)

		controlsRendered := helpStyle.Render(helpHint)
		controls := leftPadStr + controlsRendered
		contentWidth := leftPad + lipgloss.Width(controlsRendered)

		return row + strings.Repeat("\n", gapBelow) + controls, contentWidth
	}

	logo := logoStyle.Render("jotr")
	ver := versionStyle.Render(version.GetVersion())
	hint := helpStyle.Render(helpHint)

	requiredWidth := leftPad + lipgloss.Width(logo) + lipgloss.Width(ver) + 2

	if m.width < requiredWidth {
		header := leftPadStr + logo + "\n" + leftPadStr + ver + "\n" + leftPadStr + streakText

		maxContent := lipgloss.Width(logo)
		if w := lipgloss.Width(ver); w > maxContent {
			maxContent = w
		}
		if w := lipgloss.Width(hint); w > maxContent {
			maxContent = w
		}
		if w := lipgloss.Width(streakText); w > maxContent {
			maxContent = w
		}

		contentWidth := leftPad + maxContent
		return header + "\n" + leftPadStr + hint, contentWidth
	}

	padding := m.width - lipgloss.Width(logo) - lipgloss.Width(ver) - leftPad - 2
	if padding < 0 {
		padding = 0
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPadStr+logo,
		strings.Repeat(" ", padding),
		ver,
	)
	contentWidth := leftPad + lipgloss.Width(hint)
	return header + "\n" + leftPadStr + streakText + "\n" + leftPadStr + hint, contentWidth
}
