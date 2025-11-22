package cmd

import (
	"fmt"
	"strings"

	"github.com/anish/jotr/internal/updater"
	"github.com/anish/jotr/internal/version"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	force     bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update jotr to the latest version",
	Long: `Check for and install the latest version of jotr.
	
Examples:
  jotr update               # Check and update if newer version available
  jotr update --check       # Only check for updates, don't install
  jotr update --force       # Force update even if same version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
	updateCmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if same version")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate() error {
	fmt.Println("🔍 Checking for updates...")
	
	currentVersion := "v" + version.Version
	hasUpdate, latestVersion, release, err := updater.CheckForUpdates(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	currentVer := strings.TrimPrefix(currentVersion, "v")
	latestVer := strings.TrimPrefix(latestVersion, "v")

	if latestVer == currentVer && !force {
		fmt.Println("✅ You're already running the latest version!")
		return nil
	}

	if hasUpdate || force {
		fmt.Printf("\n🆕 New version available: %s\n", latestVersion)
		fmt.Printf("📅 Released: %s\n", release.PublishedAt.Format("2006-01-02"))
		
		if release.Body != "" {
			fmt.Println("\n📝 What's New:")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println(updater.FormatChangelog(release.Body))
		}

		if checkOnly {
			fmt.Println("\n💡 Run 'jotr update' to install the latest version")
			return nil
		}

		fmt.Print("\n❓ Do you want to update? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Update cancelled")
			return nil
		}

		fmt.Println("\n⬇️  Downloading update...")
		
		if err := updater.PerformUpdate(release); err != nil {
			return err
		}

		fmt.Printf("✅ Successfully updated to version %s!\n", latestVersion)
		fmt.Println("🔄 Please restart jotr to use the new version.")
	} else {
		fmt.Printf("ℹ️  Version %s is newer than the latest release %s\n", currentVersion, latestVersion)
	}

	return nil
}

// CheckForUpdates is exported for use by TUI
func CheckForUpdates() (bool, string, error) {
	currentVersion := "v" + version.Version
	hasUpdate, latestVersion, _, err := updater.CheckForUpdates(currentVersion)
	return hasUpdate, latestVersion, err
}