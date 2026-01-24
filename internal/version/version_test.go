package version

import (
	"strings"
	"testing"
)

func TestVersion_Format(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion should not return empty string")
	}

	if !strings.HasPrefix(v, "v") && v != "dev" {
		t.Errorf("Version should start with 'v', got: %s", v)
	}
}

func TestIsNewerVersion_Basic(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"1.0.0", "1.1.0", true},
		{"1.1.0", "1.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"2.0.0", "1.9.9", false},
		{"1.9.9", "2.0.0", true},
	}

	for _, tt := range tests {
		result := isNewerVersion(tt.current, tt.latest)
		if result != tt.expected {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
				tt.current, tt.latest, result, tt.expected)
		}
	}
}

func TestIsNewerVersion_WithVPrefix(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v1.0.0", "v1.1.0", true},
		{"v1.1.0", "v1.0.0", false},
		{"1.0.0", "v1.1.0", true},
		{"v1.0.0", "1.1.0", true},
	}

	for _, tt := range tests {
		result := isNewerVersion(tt.current, tt.latest)
		if result != tt.expected {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
				tt.current, tt.latest, result, tt.expected)
		}
	}
}

func TestGetCurrentVersion(t *testing.T) {
	v := getCurrentVersion()
	if v == "" {
		t.Error("getCurrentVersion should not return empty string")
	}

	if !strings.HasPrefix(v, "v") {
		t.Errorf("Version should start with 'v', got: %s", v)
	}
}

func TestGetLatestReleaseVersion_Format(t *testing.T) {
	v, err := getLatestReleaseVersion()
	if err != nil {
		if err.Error() == "GitHub API returned status 404" {
			t.Skip("GitHub API not available, skipping test")
		}

		t.Errorf("getLatestReleaseVersion should not error: %v", err)
	}

	if v != "" && !strings.HasPrefix(v, "v") {
		t.Errorf("Version should start with 'v', got: %s", v)
	}
}

func TestVersion_Constant(t *testing.T) {
	if Version == "dev" {
		t.Skip("Skipping version constant test in development mode (Version is 'dev')")
	}

	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	if Version != "1.2.0" {
		t.Errorf("Expected Version to be 1.2.0, got: %s", Version)
	}
}
