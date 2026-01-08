package version

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const Version = "1.2.0"

func getCurrentVersion() string {
	if Version != "dev" {
		return "v" + Version
	}

	now := time.Now()
	calVer := fmt.Sprintf("v%d.%d.0", now.Year()%100, int(now.Month()))

	if latestVersion, err := getLatestReleaseVersion(); err == nil {
		if isNewerVersion(latestVersion, calVer) {
			return latestVersion
		}
	}

	return calVer
}

func getLatestReleaseVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/AnishShah1803/jotr/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	return "v1.2.0", nil
}

func isNewerVersion(current, latest string) bool {
	currentParts := strings.TrimPrefix(current, "v")
	latestParts := strings.TrimPrefix(latest, "v")

	return currentParts != latestParts && latestParts > currentParts
}

func GetVersion() string {
	return getCurrentVersion()
}
