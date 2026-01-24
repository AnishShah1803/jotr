package updater

import (
	tea "github.com/charmbracelet/bubbletea"
)

// UpdateCheckResult contains the result of an update check.
type UpdateCheckResult struct {
	HasUpdate bool
	Version   string
	Err       error
}

// NewUpdateCheckCmd creates a bubble tea command that checks for updates.
// This integrates update checking into the TUI's command-based architecture.
func NewUpdateCheckCmd(checker UpdateChecker, currentVersion string) tea.Cmd {
	return func() tea.Msg {
		hasUpdate, version, err := checker.CheckForUpdate(currentVersion)
		if err != nil {
			return updateCheckMsg{
				err:       err,
				version:   "",
				hasUpdate: false,
			}
		}

		return updateCheckMsg{
			err:       nil,
			version:   version,
			hasUpdate: hasUpdate,
		}
	}
}

// updateCheckMsg is the message type for update check results.
// This is used internally by the TUI to receive update check results.
type updateCheckMsg struct {
	err       error
	version   string
	hasUpdate bool
}
