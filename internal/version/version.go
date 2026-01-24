package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Version is set during build using ldflags
// Default to dev if not set at build time
var Version = "dev"

// BuildTime is set during build using ldflags
var BuildTime = "unknown"

func getCurrentVersion() string {
	if Version != "dev" {
		return "v" + Version
	}

	now := time.Now()
	yearInDev := now.Year() - 2025

	calVer := fmt.Sprintf("v%d.%d.0-dev", yearInDev, now.Month())

	if latestVersion, err := getLatestReleaseVersion(); err == nil {
		if isNewerVersion(latestVersion, calVer) {
			return latestVersion
		}
	}

	return calVer
}

func getLatestReleaseVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/repos/AnishShah1803/jotr/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return "v" + result.TagName, nil
}

func isNewerVersion(current, latest string) bool {
	currentParts := strings.TrimPrefix(current, "v")
	latestParts := strings.TrimPrefix(latest, "v")

	return currentParts != latestParts && latestParts > currentParts
}

func GetVersion() string {
	return getCurrentVersion()
}
