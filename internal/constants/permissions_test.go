package constants

import (
	"os"
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      os.FileMode
		expected os.FileMode
	}{
		{"FilePerm0644", FilePerm0644, os.FileMode(0644)},
		{"FilePerm0600", FilePerm0600, os.FileMode(0600)},
		{"FilePermDir", FilePermDir, os.FileMode(0755)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}
