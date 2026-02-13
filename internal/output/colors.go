package output

import "github.com/charmbracelet/lipgloss"

// Named color constants for CLI output to avoid magic numbers.
var (
	SuccessColor = lipgloss.Color("42")
	WarningColor = lipgloss.Color("214")
	ErrorColor   = lipgloss.Color("203")
	MutedColor   = lipgloss.Color("240")
)

// Colorize returns the text wrapped in the given lipgloss color when enabled.
func Colorize(text string, color lipgloss.Color, enabled bool) string {
	if !enabled {
		return text
	}
	return lipgloss.NewStyle().Foreground(color).Render(text)
}
