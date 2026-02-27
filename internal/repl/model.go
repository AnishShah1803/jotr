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
	}
}

// Init initializes the REPL.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the REPL.
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
			return m, nil

		case tea.KeyDown:
			next := m.history.Next()
			if next != "" {
				m.textInput.SetValue(next)
				m.textInput.CursorEnd()
			} else {
				m.textInput.SetValue("")
			}
			return m, nil

		case tea.KeyTab:
			current := m.textInput.Value()
			completed := m.autocomplete.Complete(current)
			if completed != current {
				m.textInput.SetValue(completed)
				m.textInput.CursorEnd()
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
	return m, cmd
}

type outputMsg string

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

	return b.String()
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
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPadStr+logo,
		strings.Repeat(" ", m.width-lipgloss.Width(logo)-lipgloss.Width(ver)-leftPad-2),
		ver,
	)
	contentWidth := leftPad + lipgloss.Width(hint)
	return header + "\n" + leftPadStr + hint, contentWidth
}
