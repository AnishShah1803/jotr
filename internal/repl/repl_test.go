package repl

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
			"daily":  true,
			"d":      true,
			"search": true,
			"sync":   true,
		},
		aliases: make(map[string]string),
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"da", "daily "},
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
		{[]string{"daily", "d"}, "d"},
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
			"calendar": true,
			"capture":  true,
			"check":    true,
			"config":   true,
			"daily":    true,
			"search":   true,
			"sync":     true,
			"c":        true, // alias
		},
		commandNames: []string{"calendar", "capture", "check", "config", "daily", "search", "sync"},
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

func TestGetSubCommands(t *testing.T) {
	a := &Autocomplete{
		commands: map[string]bool{
			"template": true,
			"note":     true,
			"n":        true,
		},
		commandNames: []string{"template", "note"},
		aliases:      map[string]string{"n": "note"},
		subCommands: map[string][]string{
			"template": {"create", "delete", "edit", "list"},
		},
		actionCommands: map[string][]string{
			"note": {"create", "list", "open"},
			"n":    {"create", "list", "open"},
		},
	}

	tests := []struct {
		parent   string
		expected []string
	}{
		{"template", []string{"template create", "template delete", "template edit", "template list"}},
		{"note", []string{"note create", "note list", "note open"}},
		{"n", []string{"n create", "n list", "n open"}},
		{"unknown", nil},
		{"template create", nil},
	}

	for _, tt := range tests {
		t.Run(tt.parent, func(t *testing.T) {
			result := a.GetSubCommands(tt.parent)
			if len(result) != len(tt.expected) {
				t.Errorf("GetSubCommands(%q) = %v, want %v", tt.parent, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetSubCommands(%q)[%d] = %q, want %q", tt.parent, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestGetSubCommandCompletions(t *testing.T) {
	a := &Autocomplete{
		commands: map[string]bool{
			"template": true,
			"note":     true,
		},
		commandNames: []string{"template", "note"},
		aliases:      make(map[string]string),
		subCommands: map[string][]string{
			"template": {"create", "delete", "edit", "list"},
		},
		actionCommands: map[string][]string{
			"note": {"create", "list", "open"},
		},
	}

	tests := []struct {
		parent   string
		partial  string
		expected []string
	}{
		{"template", "c", []string{"template create"}},
		{"template", "l", []string{"template list"}},
		{"template", "e", []string{"template edit"}},
		{"template", "", []string{"template create", "template delete", "template edit", "template list"}},
		{"note", "c", []string{"note create"}},
		{"note", "l", []string{"note list"}},
		{"unknown", "c", nil},
		{"template create", "x", nil},
	}

	for _, tt := range tests {
		t.Run(tt.parent+"/"+tt.partial, func(t *testing.T) {
			result := a.GetSubCommandCompletions(tt.parent, tt.partial)
			if len(result) != len(tt.expected) {
				t.Errorf("GetSubCommandCompletions(%q, %q) = %v, want %v", tt.parent, tt.partial, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetSubCommandCompletions(%q, %q)[%d] = %q, want %q", tt.parent, tt.partial, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestIsCommand(t *testing.T) {
	a := &Autocomplete{
		commands: map[string]bool{
			"template": true,
			"note":     true,
			"n":        true,
		},
		commandNames:   []string{"template", "note"},
		aliases:        map[string]string{"n": "note"},
		subCommands:    make(map[string][]string),
		actionCommands: make(map[string][]string),
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"template", true},
		{"note", true},
		{"n", true},
		{"unknown", false},
		{"template create", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := a.IsCommand(tt.name)
			if result != tt.expected {
				t.Errorf("IsCommand(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "jotr",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	echo := &cobra.Command{
		Use: "echo",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, _ := cmd.Flags().GetString("msg")
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	echo.Flags().String("msg", "default", "message to echo")

	dual := &cobra.Command{
		Use: "dual",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "stdout-line")
			fmt.Fprintln(cmd.ErrOrStderr(), "stderr-line")
			return nil
		},
	}

	root.AddCommand(echo, dual)
	return root
}

func newTestModel(t *testing.T, rootCmd *cobra.Command) *Model {
	t.Helper()
	return &Model{
		rootCmd:      rootCmd,
		history:      newTestHistory(t),
		autocomplete: NewAutocomplete(rootCmd),
		width:        80,
		height:       24,
	}
}

func TestExecuteCommand_UnknownCommand(t *testing.T) {
	root := newTestRootCmd()
	m := newTestModel(t, root)

	out := m.executeCommand("unknown")
	if out == "" {
		t.Error("expected non-empty output for unknown command")
	}
}

func TestExecuteCommand_CapturesStdout(t *testing.T) {
	root := newTestRootCmd()
	m := newTestModel(t, root)

	out := m.executeCommand("echo")
	if !strings.Contains(out, "default") {
		t.Errorf("expected output to contain 'default', got: %q", out)
	}
}

func TestExecuteCommand_CombinedStdoutStderr(t *testing.T) {
	root := newTestRootCmd()
	m := newTestModel(t, root)

	out := m.executeCommand("dual")
	if !strings.Contains(out, "stdout-line") {
		t.Errorf("expected output to contain 'stdout-line', got: %q", out)
	}
	if !strings.Contains(out, "stderr-line") {
		t.Errorf("expected output to contain 'stderr-line', got: %q", out)
	}
}

func TestExecuteCommand_FlagResetAcrossRuns(t *testing.T) {
	root := newTestRootCmd()
	m := newTestModel(t, root)

	// First run sets the flag to a custom value.
	out := m.executeCommand("echo --msg=custom")
	if !strings.Contains(out, "custom") {
		t.Errorf("expected first run output to contain 'custom', got: %q", out)
	}

	// Second run omits the flag; it must revert to the default, not reuse 'custom'.
	out = m.executeCommand("echo")
	if strings.Contains(out, "custom") {
		t.Errorf("expected second run output NOT to contain 'custom' (flag should be reset), got: %q", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("expected second run output to contain 'default', got: %q", out)
	}
}
