package repl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
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
		},
	}

	tests := []struct {
		parent   string
		expected []string
	}{
		{"template", []string{"template create", "template delete", "template edit", "template list"}},
		{"note", []string{"note create", "note list", "note open"}},
		{"n", []string{"note create", "note list", "note open"}},
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

func TestCaptureProcessOutput(t *testing.T) {
	out, err := captureProcessOutput(func() error {
		fmt.Fprintln(os.Stdout, "stdout-line")
		fmt.Fprintln(os.Stderr, "stderr-line")
		return nil
	})
	if err != nil {
		t.Fatalf("expected no capture error, got %v", err)
	}
	if !strings.Contains(out, "stdout-line") {
		t.Fatalf("expected captured output to contain stdout-line, got %q", out)
	}
	if !strings.Contains(out, "stderr-line") {
		t.Fatalf("expected captured output to contain stderr-line, got %q", out)
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

func TestTranscriptCapsAtTenEntries(t *testing.T) {
	root := newTestRootCmd()
	m := newTestModel(t, root)

	for i := 1; i <= 12; i++ {
		m.appendTranscript(fmt.Sprintf("cmd%d", i), fmt.Sprintf("out%d", i))
	}

	if got := len(m.transcript); got != transcriptMaxEntries {
		t.Fatalf("expected transcript length %d, got %d", transcriptMaxEntries, got)
	}

	if m.transcript[0].command != "cmd3" || m.transcript[0].output != "out3" {
		t.Fatalf("expected oldest retained entry to be cmd3/out3, got %#v", m.transcript[0])
	}

	if m.transcript[9].command != "cmd12" || m.transcript[9].output != "out12" {
		t.Fatalf("expected newest retained entry to be cmd12/out12, got %#v", m.transcript[9])
	}
}

// newTestAutocomplete returns an Autocomplete pre-loaded with a small fixture
// suitable for testing resolve, resolvePath, GetParamCompletions,
// GetParamsForCommand, and GetPathCompletions.
func newTestAutocomplete() *Autocomplete {
	return &Autocomplete{
		commands: map[string]bool{
			"note": true, "n": true,
			"tags": true, "tag": true,
			"read": true, "daily": true,
		},
		commandNames: []string{"note", "tags", "read", "daily"},
		aliases:      map[string]string{"n": "note", "tag": "tags"},
		subCommands:  make(map[string][]string),
		actionCommands: map[string][]string{
			"note": {"create", "open", "list"},
			"tags": {"list", "find", "stats"},
		},
		paramCommands: map[string][]string{
			"note":        {"name=", "template="},
			"note create": {"path=", "file="},
			"tags":        {"name="},
			"read":        {"file=", "path=", "lines=", "format="},
			"daily":       {"date=", "open="},
		},
	}
}

func TestResolve(t *testing.T) {
	a := newTestAutocomplete()
	tests := []struct {
		input    string
		expected string
	}{
		{"n", "note"},
		{"tag", "tags"},
		{"note", "note"},
		{"unknown", "unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := a.resolve(tt.input)
			if got != tt.expected {
				t.Errorf("resolve(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	a := newTestAutocomplete()
	tests := []struct {
		input    string
		expected string
	}{
		{"n", "note"},
		{"tag", "tags"},
		{"n create", "note create"},
		{"tag find", "tags find"},
		{"note open", "note open"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := a.resolvePath(tt.input)
			if got != tt.expected {
				t.Errorf("resolvePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetParamCompletions(t *testing.T) {
	a := newTestAutocomplete()
	tests := []struct {
		cmd       string
		wantLen   int
		wantFirst string
	}{
		{"note", 2, "note name="},
		{"n", 2, "note name="}, // alias resolves to note
		{"tags", 1, "tags name="},
		{"tag", 1, "tags name="}, // alias resolves to tags
		{"daily", 2, "daily date="},
		{"unknown", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := a.GetParamCompletions(tt.cmd)
			if len(got) != tt.wantLen {
				t.Errorf("GetParamCompletions(%q) len = %d, want %d (got %v)", tt.cmd, len(got), tt.wantLen, got)
				return
			}
			if tt.wantLen > 0 && got[0] != tt.wantFirst {
				t.Errorf("GetParamCompletions(%q)[0] = %q, want %q", tt.cmd, got[0], tt.wantFirst)
			}
		})
	}
}

func TestGetParamsForCommand(t *testing.T) {
	a := newTestAutocomplete()
	tests := []struct {
		cmd  string
		want []string
	}{
		{"note", []string{"name=", "template="}},
		{"n", []string{"name=", "template="}}, // alias
		{"note create", []string{"path=", "file="}},
		{"n create", []string{"path=", "file="}}, // alias path
		{"tags", []string{"name="}},
		{"tag", []string{"name="}}, // alias
		{"unknown", nil},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := a.GetParamsForCommand(tt.cmd)
			if len(got) != len(tt.want) {
				t.Errorf("GetParamsForCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetParamsForCommand(%q)[%d] = %q, want %q", tt.cmd, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestGetPathCompletions(t *testing.T) {
	a := newTestAutocomplete()

	t.Run("nil config returns nil", func(t *testing.T) {
		got := a.GetPathCompletions(nil, "")
		if got != nil {
			t.Errorf("expected nil for nil config, got %v", got)
		}
	})

	// Set up a temporary directory tree:
	// <base>/notes/work.md
	// <base>/notes/personal.md
	// <base>/diary/ (directory)
	base := t.TempDir()
	notesDir := filepath.Join(base, "notes")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, "diary"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"notes/work.md", "notes/personal.md"} {
		if err := os.WriteFile(filepath.Join(base, f), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.LoadedConfig{}
	cfg.Paths.BaseDir = base

	t.Run("top-level listing", func(t *testing.T) {
		got := a.GetPathCompletions(cfg, "")
		sort.Strings(got)
		want := []string{"diary/", "notes/"}
		if len(got) != len(want) {
			t.Fatalf("GetPathCompletions top-level = %v, want %v", got, want)
		}
		for i, v := range got {
			if v != want[i] {
				t.Errorf("GetPathCompletions top-level[%d] = %q, want %q", i, v, want[i])
			}
		}
	})

	t.Run("subdir listing", func(t *testing.T) {
		got := a.GetPathCompletions(cfg, "notes/")
		sort.Strings(got)
		want := []string{"notes/personal.md", "notes/work.md"}
		if len(got) != len(want) {
			t.Fatalf("GetPathCompletions notes/ = %v, want %v", got, want)
		}
		for i, v := range got {
			if v != want[i] {
				t.Errorf("GetPathCompletions notes/[%d] = %q, want %q", i, v, want[i])
			}
		}
	})

	t.Run("no matches returns empty or nil", func(t *testing.T) {
		got := a.GetPathCompletions(cfg, "nonexistent/")
		if len(got) != 0 {
			t.Errorf("expected no completions for nonexistent path, got %v", got)
		}
	})
}
