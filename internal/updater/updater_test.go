package updater

import (
	"testing"
)

func TestCheckForUpdates(t *testing.T) {
	tests := []struct {
		name        string
		current     string
		latest      string
		expectedHas bool
	}{
		{
			name:        "newer version available",
			current:     "1.0.0",
			latest:      "1.1.0",
			expectedHas: true,
		},
		{
			name:        "same version",
			current:     "1.1.0",
			latest:      "1.1.0",
			expectedHas: false,
		},
		{
			name:        "already latest",
			current:     "1.2.0",
			latest:      "1.1.0",
			expectedHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has := isNewerVersion(tt.latest, tt.current)
			if has != tt.expectedHas {
				t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tt.latest, tt.current, has, tt.expectedHas)
			}
		})
	}
}

func TestValidateBeforeUpdate(t *testing.T) {
	t.Run("config exists", func(t *testing.T) {
		err := validateBeforeUpdate()
		if err != nil {
			t.Errorf("validateBeforeUpdate() returned error: %v", err)
		}
	})
}

func TestFindAssetURL(t *testing.T) {
	release := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "jotr-linux-amd64", BrowserDownloadURL: "http://example.com/linux"},
			{Name: "jotr-darwin-amd64", BrowserDownloadURL: "http://example.com/darwin"},
		},
	}

	url := findAssetURL(release)
	if url != "http://example.com/linux" {
		t.Errorf("findAssetURL() = %q; want %q", url, "http://example.com/linux")
	}
}

func TestGitHubRelease(t *testing.T) {
	release := &GitHubRelease{
		TagName:    "v1.1.0",
		Name:       "Release v1.1.0",
		Body:       "## Changes\n- New features",
		Draft:      false,
		Prerelease: false,
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "jotr-linux-amd64", BrowserDownloadURL: "http://example.com/linux"},
		},
	}

	if release.TagName != "v1.1.0" {
		t.Errorf("TagName = %q; want %q", release.TagName, "v1.1.0")
	}

	if release.Draft {
		t.Error("Expected release to not be a draft")
	}
}

func TestAsset(t *testing.T) {
	asset := Asset{
		Name: "jotr-linux-amd64",
		URL:  "http://example.com/linux",
	}

	if asset.Name != "jotr-linux-amd64" {
		t.Errorf("Name = %q; want %q", asset.Name, "jotr-linux-amd64")
	}

	if asset.URL != "http://example.com/linux" {
		t.Errorf("URL = %q; want %q", asset.URL, "http://example.com/linux")
	}
}

func TestGlobalVariables(t *testing.T) {
	if githubRepo != "AnishShah1803/jotr" {
		t.Errorf("githubRepo = %q; want %q", githubRepo, "AnishShah1803/jotr")
	}
}
