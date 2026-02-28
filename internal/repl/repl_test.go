package repl

import (
	"path/filepath"
	"testing"
)

func newTestHistory(t *testing.T) *History {
	t.Helper()
	tmpDir := t.TempDir()
	h := &History{
		entries:  make([]string, 0),
		position: -1,
		filePath: filepath.Join(tmpDir, "test_history"),
	}
	return h
}

func TestParseInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "daily",
			expected: []string{"daily"},
		},
		{
			name:     "command with args",
			input:    "search hello world",
			expected: []string{"search", "hello", "world"},
		},
		{
			name:     "quoted argument",
			input:    `capture "hello world"`,
			expected: []string{"capture", "hello world"},
		},
		{
			name:     "single quoted argument",
			input:    `capture 'hello world'`,
			expected: []string{"capture", "hello world"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "extra spaces",
			input:    "daily   ",
			expected: []string{"daily"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInput(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseInput(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseInput(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestHistoryAdd(t *testing.T) {
	h := newTestHistory(t)

	h.Add("command1")
	h.Add("command2")
	h.Add("command3")

	entries := h.All()
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	if entries[0] != "command1" {
		t.Errorf("expected first entry to be 'command1', got %q", entries[0])
	}
}

func TestHistoryNavigation(t *testing.T) {
	h := newTestHistory(t)

	h.Add("command1")
	h.Add("command2")
	h.Add("command3")

	prev := h.Previous()
	if prev != "command3" {
		t.Errorf("expected previous to be 'command3', got %q", prev)
	}

	prev = h.Previous()
	if prev != "command2" {
		t.Errorf("expected previous to be 'command2', got %q", prev)
	}

	next := h.Next()
	if next != "command3" {
		t.Errorf("expected next to be 'command3', got %q", next)
	}
}

func TestAutocompleteComplete(t *testing.T) {
	a := &Autocomplete{
		commands: map[string]bool{
			"daily":     true,
			"dashboard": true,
			"d":         true,
			"search":    true,
			"sync":      true,
		},
		aliases: make(map[string]string),
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"da", "da"},
		{"dai", "daily "},
		{"s", "s"},
		{"search", "search "},
		{"xyz", "xyz"},
		{"daily", "daily "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := a.Complete(tt.input)
			if result != tt.expected {
				t.Errorf("Complete(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		strs     []string
		expected string
	}{
		{[]string{"daily", "dashboard", "d"}, "d"},
		{[]string{"search", "sync", "summary"}, "s"},
		{[]string{"daily", "daily"}, "daily"},
		{[]string{}, ""},
		{[]string{"abc", "def"}, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := longestCommonPrefix(tt.strs)
			if result != tt.expected {
				t.Errorf("longestCommonPrefix(%v) = %q, want %q", tt.strs, result, tt.expected)
			}
		})
	}
}

func TestGetCompletions(t *testing.T) {
	a := &Autocomplete{
		commands: map[string]bool{
			"calendar":  true,
			"capture":   true,
			"check":     true,
			"config":    true,
			"daily":     true,
			"dashboard": true,
			"search":    true,
			"sync":      true,
			"c":         true, // alias
		},
		commandNames: []string{"calendar", "capture", "check", "config", "daily", "dashboard", "search", "sync"},
		aliases:      make(map[string]string),
	}

	tests := []struct {
		input    string
		expected []string
	}{
		{"c", []string{"calendar", "capture", "check", "config"}},
		{"ca", []string{"calendar", "capture"}},
		{"cal", []string{"calendar"}},
		{"cale", []string{"calendar"}},
		{"s", []string{"search", "sync"}},
		{"se", []string{"search"}},
		{"x", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := a.GetCompletions(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("GetCompletions(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetCompletions(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}
