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

	m := Model{
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
			header, _ := m.renderHeader()
			return m, tea.Println("\n" + header)
		}
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
			cmdOutput := m.executeCommand(input)
			m.textInput.SetValue("")
			m.completions = []string{}
			m.selectedIdx = 0

			inset := 2
			insetStr := strings.Repeat(" ", inset)

			var block strings.Builder
			block.WriteString("\n")
			block.WriteString(promptStyle.Render("❯ ") + inputStyle.Render(input))
			if cmdOutput != "" {
				trimmed := strings.TrimSpace(cmdOutput)
				for _, line := range strings.Split(trimmed, "\n") {
					block.WriteString("\n")
					block.WriteString(insetStr + line)
				}
			}

			if m.quitting {
				return m, tea.Sequence(tea.Println(block.String()), tea.Quit)
			}
			return m, tea.Println(block.String())

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
				selected := m.completions[0]

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
			return m, nil
		}

	case outputMsg:
		chunk := string(msg)
		if chunk != "" {
			inset := 2
			insetStr := strings.Repeat(" ", inset)
			var block strings.Builder
			for _, line := range strings.Split(strings.TrimSpace(chunk), "\n") {
				block.WriteString(insetStr + line + "\n")
			}
			return m, tea.Println(strings.TrimRight(block.String(), "\n"))
		}
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.updateCompletions()
	return m, cmd
}

type outputMsg string

func (m *Model) updateCompletions() {
	value := m.textInput.Value()
	fields := strings.Fields(value)

	if value == "" {
		m.completions = m.autocomplete.GetAllCommands()
		m.selectedIdx = 0
		return
	}

	if len(fields) == 1 {
		cmdName := fields[0]

		if m.autocomplete.IsCommand(cmdName) {
			subCommands := m.autocomplete.GetSubCommands(cmdName)
			if subCommands != nil && len(subCommands) > 0 {
				m.completions = subCommands
				m.selectedIdx = 0
				return
			}

			if m.autocomplete.IsParamCommand(cmdName) {
				m.completions = m.autocomplete.GetParamCompletions(cmdName)
				m.selectedIdx = 0
				return
			}
		}

		m.completions = m.autocomplete.GetCompletions(cmdName)
		m.selectedIdx = 0
		return
	}

	m.completions = m.getCompletionsForInput(fields, strings.HasSuffix(value, " "))
	m.selectedIdx = 0
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
			// e.g. "note create path=" — user is typing the param key up to/including =
			partialIsParam = true
			key := strings.SplitN(last, "=", 2)[0]
			// Remove it from usedParams so it's still offered as a completion.
			usedParams[key+"="] = false
			partial = key + "="
		} else {
			// Last field has no `=`: it's either a partial subcommand name or partial param key.
			if len(cmdFields) > 0 {
				partial = cmdFields[len(cmdFields)-1]
				cmdFields = cmdFields[:len(cmdFields)-1]
			}
		}
	}

	cmdPath := strings.Join(cmdFields, " ")

	// No params typed yet — delegate to subcommand completion.
	if !partialIsParam && len(usedParams) == 0 {
		if hasTrailingSpace {
			subs := m.autocomplete.GetSubCommands(cmdPath)
			if subs != nil {
				return subs
			}
		} else {
			subs := m.autocomplete.GetSubCommandCompletions(cmdPath, partial)
			if subs != nil {
				return subs
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

	inset := 2

	separator := strings.Repeat(" ", inset) + strings.Repeat("─", 40)
	b.WriteString(separatorStyle.Render(separator))
	b.WriteString("\n")

	b.WriteString(strings.Repeat(" ", inset))
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(inputStyle.Render(m.textInput.View()))

	if len(m.completions) > 0 {
		b.WriteString("\n")
		b.WriteString(m.renderCompletions(inset))
	}

	return b.String()
}

func (m Model) renderCompletions(inset int) string {
	const maxLines = 10
	insetStr := strings.Repeat(" ", inset)
	lines := make([]string, maxLines)

	for i := 0; i < maxLines; i++ {
		if i < len(m.completions) {
			completion := m.completions[i]
			if i == m.selectedIdx {
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
