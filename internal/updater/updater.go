package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/version"
)

// Asset represents a downloadable release asset from GitHub.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	PublishedAt time.Time `json:"published_at"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	Draft      bool `json:"draft"`
	Prerelease bool `json:"prerelease"`
}

const (
	githubRepo = "AnishShah1803/jotr"
)

// CheckForUpdates checks if a newer version is available.
func CheckForUpdates(currentVersion string) (bool, string, *GitHubRelease, error) {
	latest, err := getLatestRelease()
	if err != nil {
		return false, "", nil, err
	}

	currentVer := strings.TrimPrefix(currentVersion, "v")
	latestVer := strings.TrimPrefix(latest.TagName, "v")

	hasUpdate := isNewerVersion(latestVer, currentVer)

	return hasUpdate, latest.TagName, latest, nil
}

// CheckAndUpdate checks for updates and optionally performs the update.
func CheckAndUpdate(performUpdate bool) error {
	currentVersion := version.GetVersion()

	hasUpdate, _, release, err := CheckForUpdates(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdate {
		return nil // No update needed
	}

	if performUpdate {
		return PerformUpdate(release)
	}

	return nil
}

func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)

	// Use longer timeout for API calls, with retry logic
	client := &http.Client{Timeout: 30 * time.Second}

	var resp *http.Response

	var err error

	// Retry logic for network issues
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Get(url)
		if err == nil {
			break
		}

		// If this isn't the last attempt, wait before retrying
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info after %d attempts: %w", maxRetries, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func validateBeforeUpdate() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)

	testFile := filepath.Join(exeDir, ".jotr-update-test")
	defer os.Remove(testFile)

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("insufficient permissions to update binary: %w", err)
	}

	return nil
}

func isNewerVersion(latest, current string) bool {
	// Simple version comparison
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	maxLen := len(latestParts)
	if len(currentParts) > maxLen {
		maxLen = len(currentParts)
	}

	for i := 0; i < maxLen; i++ {
		var latestNum, currentNum int

		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &latestNum)
		}

		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &currentNum)
		}

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	return false
}

// FormatChangelog formats the release body for display.
func FormatChangelog(body string) string {
	lines := strings.Split(body, "\n")

	var formatted []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format markdown headers
		if strings.HasPrefix(line, "##") {
			line = "- " + strings.TrimPrefix(line, "##")
		} else if strings.HasPrefix(line, "#") {
			line = "* " + strings.TrimPrefix(line, "#")
		} else if strings.HasPrefix(line, "*") || strings.HasPrefix(line, "-") {
			line = "  • " + strings.TrimLeft(line, "*- ")
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

// PerformUpdate downloads and installs the new version.
func PerformUpdate(release *GitHubRelease) error {
	// Find the appropriate asset for current OS/arch
	assetURL := findAssetURL(release)
	if assetURL == "" {
		return fmt.Errorf("no compatible binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download to temporary file
	tempFile, err := downloadBinary(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tempFile)

	// Replace current binary
	if err := replaceBinary(tempFile); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	return nil
}

func findAssetURL(release *GitHubRelease) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Common naming patterns for releases
	patterns := []string{
		fmt.Sprintf("jotr_%s_%s", osName, arch),
		fmt.Sprintf("jotr-%s-%s", osName, arch),
		fmt.Sprintf("jotr_%s_%s.tar.gz", osName, arch),
		fmt.Sprintf("jotr-%s-%s.tar.gz", osName, arch),
	}

	for _, asset := range release.Assets {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(pattern)) {
				return asset.BrowserDownloadURL
			}
		}
	}

	return ""
}

func downloadBinary(url string) (string, error) {
	// Use longer timeout for downloads
	client := &http.Client{Timeout: 10 * time.Minute}

	var resp *http.Response

	var err error

	// Retry logic for downloads
	maxRetries := 2
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to download after %d attempts: %w", maxRetries, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "jotr-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy download to temp file with progress (for large files)
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	// Make executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		_ = os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	return tempFile.Name(), nil
}

func replaceBinary(newBinaryPath string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return err
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return err
	}

	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	err = copyFile(newBinaryPath, currentExe)
	if err == nil {
		_ = os.Remove(backupPath)
		return nil
	}

	renamePath := currentExe + ".old"
	if err := os.Rename(currentExe, renamePath); err != nil {
		_ = os.Remove(backupPath)
		return fmt.Errorf("failed to rename in-use binary: %w", err)
	}

	if err := copyFile(newBinaryPath, currentExe); err != nil {
		_ = os.Remove(currentExe)
		_ = os.Rename(renamePath, currentExe)
		_ = os.Remove(backupPath)
		return fmt.Errorf("failed to copy new binary after rename: %w", err)
	}

	_ = os.Remove(renamePath)
	_ = os.Remove(backupPath)

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}
