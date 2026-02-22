package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AnishShah1803/jotr/internal/updater"
	"github.com/AnishShah1803/jotr/internal/version"
)

type updateChecker interface {
	CheckForUpdate(currentVersion string) (bool, string, error)
}

type defaultUpdateChecker struct{}

func (c *defaultUpdateChecker) CheckForUpdate(currentVersion string) (bool, string, error) {
	hasUpdate, latestVersion, _, err := updater.CheckForUpdates(currentVersion)
	return hasUpdate, latestVersion, err
}

func checkForUpdatesCmd() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, version, err := checkForUpdatesFromTUI()

		return updateCheckMsg{
			hasUpdate: hasUpdate,
			version:   version,
			err:       err,
		}
	}
}

func checkForUpdatesFromTUI() (bool, string, error) {
	var checker updateChecker = &defaultUpdateChecker{}
	currentVersion := version.GetVersion()

	return checker.CheckForUpdate(currentVersion)
}

func handleUpdateCheck(m Model, msg updateCheckMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m = setStatus(m, fmt.Sprintf("❌ Update check failed: %v", msg.err), "error")
	} else if msg.hasUpdate {
		m.updateAvailable = true
		m.updateVersion = msg.version
		m = setStatus(m, fmt.Sprintf("🆕 Update available: %s (restart jotr and run 'jotr update')", msg.version), "info")
	} else {
		m = setStatus(m, "✅ You're running the latest version!", "success")
	}

	return m, nil
}

func handleError(m Model, msg errorMsg) (Model, tea.Cmd) {
	m.err = msg.err
	m.errorRetryable = msg.retryable
	m.updateCachedKeyMap()

	return m, nil
}
