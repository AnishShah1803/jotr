package output

import "github.com/charmbracelet/lipgloss"

// Named color constants for CLI and TUI output to avoid magic numbers.
var (
	// UI palette
	PrimaryColor   = lipgloss.Color("86")
	SecondaryColor = lipgloss.Color("240")
	AccentColor    = lipgloss.Color("205")

	// Common semantic colors
	SuccessColor = lipgloss.Color("42")
	WarningColor = lipgloss.Color("214")
	ErrorColor   = lipgloss.Color("203")
	MutedColor   = lipgloss.Color("240")

	// Border / emphasis
	BorderColor = lipgloss.Color("196")
)

// Colorize returns the text wrapped in the given lipgloss color when enabled.
func Colorize(text string, color lipgloss.Color, enabled bool) string {
	if !enabled {
		return text
	}
	return lipgloss.NewStyle().Foreground(color).Render(text)
}
