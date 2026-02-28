package repl

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/output"
	"github.com/AnishShah1803/jotr/internal/version"
)

type Model struct {
	ctx          context.Context
	config       *config.LoadedConfig
	rootCmd      *cobra.Command
	textInput    textinput.Model
	history      *History
	autocomplete *Autocomplete
	width        int
	height       int
	ready        bool
	quitting     bool
	lastOutput   string
	showOutput   bool
	completions  []string
	selectedIdx  int
}

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
)

func NewModel(ctx context.Context, cfg *config.LoadedConfig, rootCmd *cobra.Command) Model {
	ti := textinput.New()
	ti.Placeholder = "type a command..."
	ti.Prompt = ""
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	return Model{
		ctx:          ctx,
		config:       cfg,
		rootCmd:      rootCmd,
		textInput:    ti,
		history:      NewHistory(),
		autocomplete: NewAutocomplete(rootCmd),
		width:        80,
		height:       24,
		ready:        false,
		quitting:     false,
		completions:  []string{},
		selectedIdx:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = m.width - 10
		if m.textInput.Width < 20 {
			m.textInput.Width = 20
		}
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.textInput.Value() == "" {
				m.quitting = true
				return m, tea.Quit
			}
			m.textInput.SetValue("")
			m.completions = []string{}
			m.selectedIdx = 0
			return m, nil

		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}

			m.history.Add(input)
			output := m.executeCommand(input)
			m.textInput.SetValue("")
			m.lastOutput = output
			m.showOutput = true
			m.completions = []string{}
			m.selectedIdx = 0
			if m.quitting {
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyUp:
			prev := m.history.Previous()
			if prev != "" {
				m.textInput.SetValue(prev)
				m.textInput.CursorEnd()
			}
			m.updateCompletions()
			return m, nil

		case tea.KeyDown:
			next := m.history.Next()
			if next != "" {
				m.textInput.SetValue(next)
				m.textInput.CursorEnd()
			} else {
				m.textInput.SetValue("")
			}
			m.updateCompletions()
			return m, nil

		case tea.KeyTab:
			current := m.textInput.Value()
			if len(m.completions) > 0 {
				m.textInput.SetValue(m.completions[0] + " ")
				m.textInput.CursorEnd()
				m.completions = []string{}
				m.selectedIdx = 0
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
			return m, nil
		}

	case outputMsg:
		m.lastOutput = string(msg)
		m.showOutput = true
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.updateCompletions()
	return m, cmd
}

type outputMsg string

func (m *Model) updateCompletions() {
	input := strings.TrimSpace(m.textInput.Value())
	if !strings.Contains(input, " ") && len(input) > 0 {
		m.completions = m.autocomplete.GetCompletions(input)
		m.selectedIdx = 0
	} else if input == "" {
		m.completions = m.autocomplete.GetAllCommands()
		m.selectedIdx = 0
	} else {
		m.completions = []string{}
		m.selectedIdx = 0
	}
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading jotr...\n\n"
	}

	if m.quitting {
		return "\n  Goodbye!\n\n"
	}

	var b strings.Builder

	topPad := m.height * 5 / 100
	if topPad < 1 {
		topPad = 1
	}
	b.WriteString(strings.Repeat("\n", topPad))

	header, contentWidth := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	inset := 2
	separatorLen := contentWidth - inset
	if separatorLen < 10 {
		separatorLen = 10
	}
	separator := strings.Repeat(" ", inset) + strings.Repeat("─", separatorLen)
	b.WriteString(separatorStyle.Render(separator))
	b.WriteString("\n")

	gapPrompt := m.height * 2 / 100
	if gapPrompt < 1 {
		gapPrompt = 1
	}
	b.WriteString(strings.Repeat("\n", gapPrompt))

	if m.showOutput && m.lastOutput != "" {
		b.WriteString(m.lastOutput)
		b.WriteString("\n\n")
	}

	b.WriteString(strings.Repeat(" ", inset))
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(inputStyle.Render(m.textInput.View()))
	input := strings.TrimSpace(m.textInput.Value())
	if !strings.Contains(input, " ") && len(m.completions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderCompletions(inset))
	}

	return b.String()
}

func (m Model) renderCompletions(inset int) string {
	if len(m.completions) == 0 {
		return ""
	}

	const maxLines = 10
	insetStr := strings.Repeat(" ", inset)
	count := len(m.completions)
	if count > maxLines {
		count = maxLines
	}
	lines := make([]string, count)

	for i := 0; i < count; i++ {
		completion := m.completions[i]
		if i == m.selectedIdx {
			lines[i] = insetStr + selectedCompletionStyle.Render("  "+completion)
		} else {
			lines[i] = insetStr + completionStyle.Render("  "+completion)
		}
	}

	return strings.Join(lines, "\n")
}

func renderLogo() string {
	return replLogoStyle.Render(replAsciiArt)
}

func (m Model) renderHeader() (string, int) {
	const helpHint = "tab to autocomplete · ↑/↓ for history · ctrl+c to quit"
	const minWidthForArt = 50
	const minHeightForArt = 15

	leftPad := 2
	leftPadStr := strings.Repeat(" ", leftPad)

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
		verPadded := strings.Repeat("\n", replAsciiArtHeight-1) + verText
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
		header := leftPadStr + logo + "\n" + leftPadStr + ver

		maxContent := lipgloss.Width(logo)
		if w := lipgloss.Width(ver); w > maxContent {
			maxContent = w
		}
		if w := lipgloss.Width(hint); w > maxContent {
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
	return header + "\n" + leftPadStr + hint, contentWidth
}
