package cmd

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/tui"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Launch interactive dashboard",
	Long: `Launch an interactive TUI dashboard with:
- Recent daily notes
- Note preview
- Task summary
- Statistics

Navigation:
  tab/shift+tab  - Switch panels
  ↑/↓ or j/k     - Navigate items
  enter          - Open selected note
  r              - Refresh data
  q              - Quit`,
	Aliases: []string{"dash", "ui"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := config.LoadWithContext(ctx, "")
		if err != nil {
			return err
		}

		return LaunchDashboard(ctx, cfg)
	},
}

// LaunchDashboard starts the interactive TUI dashboard for note management.
// It displays recent notes, tasks, and provides fuzzy search capabilities.
func LaunchDashboard(ctx context.Context, cfg *config.LoadedConfig) error {
	// Create model
	m := tui.NewModel(ctx, cfg)

	// Create program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run
	if _, err := p.Run(); err != nil {
		utils.PrintError("running dashboard: %v", err)
		return err
	}

	return nil
}
