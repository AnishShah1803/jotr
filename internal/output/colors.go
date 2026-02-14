package output

import "github.com/charmbracelet/lipgloss"

// Named color constants for CLI and TUI output to avoid magic numbers.
var (
	// UI palette
	PrimaryColor   = lipgloss.AdaptiveColor{Light: "86", Dark: "42"}
	SecondaryColor = lipgloss.AdaptiveColor{Light: "240", Dark: "244"}
	AccentColor    = lipgloss.AdaptiveColor{Light: "205", Dark: "219"}

	// Common semantic colors
	SuccessColor = lipgloss.AdaptiveColor{Light: "42", Dark: "48"}
	WarningColor = lipgloss.AdaptiveColor{Light: "214", Dark: "220"}
	ErrorColor   = lipgloss.AdaptiveColor{Light: "203", Dark: "224"}
	MutedColor   = lipgloss.AdaptiveColor{Light: "240", Dark: "244"}

	// Border / emphasis
	BorderColor = lipgloss.AdaptiveColor{Light: "196", Dark: "204"}
)

// Colorize returns the text wrapped in the given lipgloss color when enabled.
func Colorize(text string, color lipgloss.AdaptiveColor, enabled bool) string {
	if !enabled {
		return text
	}
	return lipgloss.NewStyle().Foreground(color).Render(text)
}
