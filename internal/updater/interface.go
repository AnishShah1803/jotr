package updater

// UpdateChecker defines the interface for checking updates.
// This allows the TUI to depend on an abstraction rather than concrete implementation.
type UpdateChecker interface {
	// CheckForUpdate checks if a newer version is available.
	// Returns: hasUpdate bool, latestVersion string, err error
	CheckForUpdate(currentVersion string) (bool, string, error)
}

// DefaultUpdateChecker is the default implementation of UpdateChecker.
type DefaultUpdateChecker struct{}

// NewDefaultUpdateChecker creates a new DefaultUpdateChecker.
func NewDefaultUpdateChecker() *DefaultUpdateChecker {
	return &DefaultUpdateChecker{}
}

// CheckForUpdate implements the UpdateChecker interface.
func (c *DefaultUpdateChecker) CheckForUpdate(currentVersion string) (bool, string, error) {
	hasUpdate, latestVersion, _, err := CheckForUpdates(currentVersion)
	return hasUpdate, latestVersion, err
}
