//go:build dev

package config

// getDefaultConfigPath returns the default config path for dev builds
func getDefaultConfigPath() string {
	return "dev-config.json"
}
