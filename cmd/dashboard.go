package cmd

import (
	"fmt"
	"os"

	"github.com/anish/jotr/internal/config"
	"github.com/anish/jotr/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
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
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return launchDashboard(cfg)
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func launchDashboard(cfg *config.LoadedConfig) error {
	// Create model
	m := tui.NewModel(cfg)

	// Create program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running dashboard: %v\n", err)
		return err
	}

	return nil
}

