package repl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	taskcmd "github.com/AnishShah1803/jotr/cmd/task"
	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
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

func TestCommandRegistry_TaskIncludesSyncSubcommand(t *testing.T) {
	taskDef := (*CommandDef)(nil)
	for i := range commandRegistry {
		if commandRegistry[i].Name == "task" {
			taskDef = &commandRegistry[i]
			break
		}
	}

	if taskDef == nil {
		t.Fatalf("expected command registry to include task command")
	}

	if !slices.Contains(taskDef.Subcommands, "sync") {
		t.Fatalf("expected task subcommands to include sync, got %v", taskDef.Subcommands)
	}
}

func TestExecuteCommand_DispatchesTaskSyncSubcommand(t *testing.T) {
	root := &cobra.Command{
		Use:           "jotr",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	task := &cobra.Command{Use: "task"}
	syncInvoked := false
	taskSync := &cobra.Command{
		Use: "sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			syncInvoked = true
			fmt.Fprintln(cmd.OutOrStdout(), "task sync invoked")
			return nil
		},
	}
	task.AddCommand(taskSync)
	root.AddCommand(task)

	m := newTestModel(t, root)
	out := m.executeCommand("task sync")

	if !syncInvoked {
		t.Fatalf("expected task sync subcommand to be invoked")
	}
	if !strings.Contains(out, "task sync invoked") {
		t.Fatalf("expected output to include sync invocation marker, got %q", out)
	}
}

func TestTaskSyncIsRegisteredInTaskCommandTree(t *testing.T) {
	root := &cobra.Command{Use: "jotr"}
	root.AddCommand(taskcmd.TaskCmd)
	found, _, err := root.Find([]string{"task", "sync"})
	if err != nil {
		t.Fatalf("expected task sync to be found, got error: %v", err)
	}
	if found == nil || found.Name() != "sync" {
		t.Fatalf("expected to find sync subcommand, got %#v", found)
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

func TestRunTaskAddFromPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.LoadedConfig{
		DiaryPath: filepath.Join(tmpDir, "Diary"),
		TodoPath:  filepath.Join(tmpDir, "todo.md"),
		StatePath: filepath.Join(tmpDir, "state.json"),
	}

	todayPath := notes.BuildDailyNotePath(cfg.DiaryPath, time.Now())
	if err := os.MkdirAll(filepath.Dir(todayPath), 0750); err != nil {
		t.Fatalf("failed to create diary directory: %v", err)
	}
	if err := notes.WriteNote(context.Background(), todayPath, "# Today\n\n## Tasks\n\n"); err != nil {
		t.Fatalf("failed to create today's note: %v", err)
	}

	out, err := taskcmd.RunTaskAdd(context.Background(), cfg, "Write REPL prompt", "P1", []string{"repl", "task"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "Added task") {
		t.Fatalf("expected output to confirm task add, got %q", out)
	}
}

func TestExecuteCommand_TaskSearchFilterOnlyQueries(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.LoadedConfig{
		Config:    config.Config{},
		TodoPath:  filepath.Join(tmpDir, "todo.md"),
		StatePath: filepath.Join(tmpDir, "state.json"),
	}

	todoContent := `# To-Do List

## Tasks

- [ ] Repl pending p3 task [P3] #home
- [x] Repl completed p1 task [P1] #work @completed(2026-03-01)
`
	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	newSearchModel := func(t *testing.T) *Model {
		t.Helper()
		root := &cobra.Command{Use: "jotr", SilenceErrors: true, SilenceUsage: true}
		task := &cobra.Command{Use: "task"}
		var filters []string
		search := &cobra.Command{
			Use: "search [query]",
			RunE: func(cmd *cobra.Command, args []string) error {
				query := strings.TrimSpace(strings.Join(args, " "))
				return taskcmd.RunTaskSearch(context.Background(), cfg, query, filters)
			},
		}
		search.Flags().StringArrayVarP(&filters, "filter", "f", nil, "")
		task.AddCommand(search)
		root.AddCommand(task)
		return newTestModel(t, root)
	}

	t.Run("priority P3 filter only", func(t *testing.T) {
		m := newSearchModel(t)
		out := m.executeCommand("task search --filter priority=P3")

		if strings.Contains(out, "No matching tasks found") {
			t.Fatalf("expected visible P3 task in REPL filter-only search output, got: %q", out)
		}
		if !strings.Contains(out, "Repl pending p3 task") {
			t.Fatalf("expected P3 task in REPL filter-only search output, got: %q", out)
		}
		if strings.Contains(out, "Repl completed p1 task") {
			t.Fatalf("did not expect non-P3 task in REPL priority filter output, got: %q", out)
		}
	})

	t.Run("status all filter only", func(t *testing.T) {
		m := newSearchModel(t)
		out := m.executeCommand("task search --filter status=all")
		normalized := stripANSIEscapeCodes(out)

		if strings.Contains(normalized, "No matching tasks found") {
			t.Fatalf("expected visible tasks in REPL status=all filter-only search output, got: %q", normalized)
		}
		if !strings.Contains(normalized, "Repl pending p3 task") {
			t.Fatalf("expected pending task in REPL status=all output, got: %q", normalized)
		}
		if !strings.Contains(normalized, "Repl completed p1 task") {
			t.Fatalf("expected completed task in REPL status=all output, got: %q", normalized)
		}
		if strings.Contains(normalized, "done:") {
			t.Fatalf("expected completed date without done label in REPL output, got: %q", normalized)
		}
		if !strings.Contains(normalized, "2026-03-01") {
			t.Fatalf("expected completed date to remain visible in REPL output, got: %q", normalized)
		}
	})
}

func newTaskPromptModel(t *testing.T, todoContent string) *Model {
	t.Helper()

	tmpDir := t.TempDir()
	cfg := &config.LoadedConfig{
		TodoPath:  filepath.Join(tmpDir, "todo.md"),
		StatePath: filepath.Join(tmpDir, "state.json"),
	}

	if err := notes.WriteNote(context.Background(), cfg.TodoPath, todoContent); err != nil {
		t.Fatalf("failed to create todo file: %v", err)
	}

	root := &cobra.Command{Use: "jotr"}
	m := NewModel(context.Background(), cfg, root)
	return &m
}

func TestStartTaskPrompt_LoadsTaskListContextBeforePrompt(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Pending from repl
`)
	m.appendTranscript("search", "some unrelated output")

	started := m.startTaskPrompt("task complete")
	if !started {
		t.Fatalf("expected startTaskPrompt to start for task complete")
	}
	if !m.taskPromptActive {
		t.Fatalf("expected task prompt to become active")
	}
	if got := len(m.transcript); got < 2 {
		t.Fatalf("expected task list transcript to be appended, got %d entries", got)
	}

	last := m.transcript[len(m.transcript)-1]
	if last.command != transcriptTaskListPending {
		t.Fatalf("expected last transcript command %q, got %q", transcriptTaskListPending, last.command)
	}
	if strings.TrimSpace(last.output) == "" {
		t.Fatalf("expected task list output to be captured before prompt")
	}
}

func TestStartTaskPrompt_ReusesExistingTaskListContext(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Existing pending task
`)

	if err := m.ensureTaskListContext(false); err != nil {
		t.Fatalf("ensureTaskListContext failed: %v", err)
	}
	before := len(m.transcript)

	started := m.startTaskPrompt("task complete")
	if !started {
		t.Fatalf("expected startTaskPrompt to start for task complete")
	}
	if !m.taskPromptActive {
		t.Fatalf("expected task prompt to become active")
	}
	if got := len(m.transcript); got != before {
		t.Fatalf("expected existing task list context to be reused, transcript len before=%d after=%d", before, got)
	}
}

func TestStartTaskEditPrompt_LoadsTaskListContextBeforePrompt(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Pending edit task
- [x] Completed edit task @completed(2026-03-01)
`)
	m.appendTranscript("read", "unrelated output")

	started := m.startTaskEditPrompt("task edit")
	if !started {
		t.Fatalf("expected startTaskEditPrompt to start for task edit")
	}
	if !m.taskEditActive {
		t.Fatalf("expected task edit prompt to become active")
	}
	if got := len(m.transcript); got < 2 {
		t.Fatalf("expected task list transcript to be appended, got %d entries", got)
	}

	last := m.transcript[len(m.transcript)-1]
	if last.command != transcriptTaskListForEditor {
		t.Fatalf("expected last transcript command %q, got %q", transcriptTaskListForEditor, last.command)
	}
	if strings.TrimSpace(last.output) == "" {
		t.Fatalf("expected task list output to be captured before prompt")
	}

	normalized := stripANSIEscapeCodes(last.output)
	if !strings.Contains(normalized, "Completed edit task") {
		t.Fatalf("expected completed task text in REPL task-list context, got: %q", normalized)
	}
	if !strings.Contains(normalized, "2026-03-01") {
		t.Fatalf("expected completed date in REPL task-list context, got: %q", normalized)
	}
	if strings.Contains(normalized, "done:") {
		t.Fatalf("expected completed date without done label in REPL task-list context, got: %q", normalized)
	}
}

func TestStartTaskEditPrompt_ReusesExistingTaskListContext(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Existing editable task
`)

	if err := m.ensureTaskListContext(true); err != nil {
		t.Fatalf("ensureTaskListContext failed: %v", err)
	}
	before := len(m.transcript)

	started := m.startTaskEditPrompt("task edit")
	if !started {
		t.Fatalf("expected startTaskEditPrompt to start for task edit")
	}
	if !m.taskEditActive {
		t.Fatalf("expected task edit prompt to become active")
	}
	if got := len(m.transcript); got != before {
		t.Fatalf("expected existing task list context to be reused, transcript len before=%d after=%d", before, got)
	}
}

func TestTaskEditPrompt_AfterTaskList_DoesNotDuplicateList(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Pending edit task
- [x] Completed edit task
`)

	m.textInput.SetValue("task list")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("expected Model after entering task list command, got %T", next)
	}
	m = &updated

	before := len(m.transcript)

	m.textInput.SetValue("task edit")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after entering task edit command, got %T", next)
	}
	m = &updated

	if !m.taskEditActive {
		t.Fatalf("expected task edit prompt to be active")
	}
	if m.taskEditStep != 0 {
		t.Fatalf("expected task edit step 0, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Task number to edit: " {
		t.Fatalf("expected task number prompt, got %q", got)
	}
	if got := len(m.transcript); got != before {
		t.Fatalf("expected no duplicate task list transcript before task edit prompt, len before=%d after=%d", before, got)
	}
}

func TestTaskEditPrompt_EnterSubmitsEditedTextPriorityAndTags(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Original task text [P2] #alpha #beta <!-- id: abc12345 -->
`)

	todayPath := notes.BuildDailyNotePath(m.config.DiaryPath, time.Now())
	if err := os.MkdirAll(filepath.Dir(todayPath), 0o750); err != nil {
		t.Fatalf("failed to create diary directory: %v", err)
	}
	if err := notes.WriteNote(context.Background(), todayPath, `# Today

## Tasks

- [ ] Original task text [P2] #alpha #beta <!-- id: abc12345 -->
`); err != nil {
		t.Fatalf("failed to create today's note: %v", err)
	}

	todoState := state.NewTodoState()
	todoState.AddTask(tasks.Task{
		Text:      "Original task text [P2] #alpha #beta",
		ID:        "abc12345",
		Priority:  "P2",
		Tags:      []string{"alpha", "beta"},
		Completed: false,
	}, "todo-list")
	if err := todoState.Write(m.config.StatePath); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	m.textInput.SetValue("task edit")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("expected Model after entering task edit command, got %T", next)
	}
	m = &updated

	if !m.taskEditActive {
		t.Fatalf("expected task edit prompt to be active")
	}
	if m.taskEditStep != 0 {
		t.Fatalf("expected task edit step 0, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Task number to edit: " {
		t.Fatalf("expected task number prompt, got %q", got)
	}
	if got := m.textInput.Value(); got != "" {
		t.Fatalf("expected empty task number input, got %q", got)
	}

	m.textInput.SetValue("1")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after selecting task number, got %T", next)
	}
	m = &updated

	if m.taskEditStep != 1 {
		t.Fatalf("expected task edit step 1 after selecting task number, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Task text: " {
		t.Fatalf("expected task text prompt, got %q", got)
	}
	if got := m.textInput.Value(); got != "Original task text" {
		t.Fatalf("expected prefilled task text %q, got %q", "Original task text", got)
	}

	m.textInput.SetValue("Edited task text")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after submitting edited task text, got %T", next)
	}
	m = &updated

	if m.taskEditStep != 2 {
		t.Fatalf("expected task edit step 2 after submitting task text, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Priority (P0-P3, press enter to skip): " {
		t.Fatalf("expected task priority prompt, got %q", got)
	}
	if got := m.textInput.Value(); got != "2" {
		t.Fatalf("expected prefilled priority %q, got %q", "2", got)
	}

	m.textInput.SetValue("P1")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after submitting edited task priority, got %T", next)
	}
	m = &updated

	if m.taskEditStep != 3 {
		t.Fatalf("expected task edit step 3 after submitting task priority, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Tags (comma-separated, press enter to skip): " {
		t.Fatalf("expected task tags prompt, got %q", got)
	}
	if got := strings.TrimSpace(m.textInput.Value()); got == "" {
		t.Fatalf("expected prefilled tags value, got empty")
	}
	tagParts := strings.Split(m.textInput.Value(), ",")
	for i := range tagParts {
		tagParts[i] = strings.TrimSpace(tagParts[i])
	}
	sort.Strings(tagParts)
	if strings.Join(tagParts, ",") != "alpha,beta" {
		t.Fatalf("expected prefilled tags to contain alpha and beta, got %q", m.textInput.Value())
	}

	m.textInput.SetValue("repl, task")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after submitting edited task tags, got %T", next)
	}
	m = &updated

	if m.taskEditActive {
		t.Fatalf("expected task edit prompt to be cleared after submitting tags")
	}
	if got := m.textInput.Value(); got != "" {
		t.Fatalf("expected empty input buffer after completing task edit, got stale value %q", got)
	}
	if got := m.textInput.Prompt; got != defaultREPLPrompt {
		t.Fatalf("expected REPL prompt to be restored to %q, got %q", defaultREPLPrompt, got)
	}
	if got := m.textInput.Placeholder; got != defaultREPLPlaceholder {
		t.Fatalf("expected REPL placeholder to be restored to %q, got %q", defaultREPLPlaceholder, got)
	}

	for _, entry := range m.transcript {
		if entry.command == "task edit" && strings.Contains(entry.output, "task edit prompt is incomplete") {
			t.Fatalf("unexpected incomplete prompt error in transcript: %q", entry.output)
		}
	}

	tasksOnTodo, err := tasks.ReadTasks(context.Background(), m.config.TodoPath)
	if err != nil {
		t.Fatalf("failed to read todo tasks after edit submit: %v", err)
	}
	if len(tasksOnTodo) != 1 {
		t.Fatalf("expected 1 task in todo file after edit submit, got %d", len(tasksOnTodo))
	}
	if !strings.Contains(tasksOnTodo[0].Text, "Edited task text") {
		t.Fatalf("expected edited task text to contain %q, got %q", "Edited task text", tasksOnTodo[0].Text)
	}
	if got := tasksOnTodo[0].Priority; got != "P1" {
		t.Fatalf("expected edited task priority %q, got %q", "P1", got)
	}
	gotTags := append([]string(nil), tasksOnTodo[0].Tags...)
	sort.Strings(gotTags)
	if strings.Join(gotTags, ",") != "repl,task" {
		t.Fatalf("expected edited task tags %q, got %v", "repl,task", tasksOnTodo[0].Tags)
	}
}

func TestTaskEditPrompt_SelectsVisibleSortedTaskRow(t *testing.T) {
	m := newTaskPromptModel(t, `# To-Do List

## Tasks

- [ ] Low priority source-first task [P3] #backlog
- [ ] High priority source-second task [P0] #urgent
`)

	m.textInput.SetValue("task edit")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("expected Model after entering task edit command, got %T", next)
	}
	m = &updated

	if !m.taskEditActive {
		t.Fatalf("expected task edit prompt to be active")
	}
	if m.taskEditStep != 0 {
		t.Fatalf("expected task edit step 0, got %d", m.taskEditStep)
	}
	if len(m.transcript) == 0 {
		t.Fatalf("expected transcript to include rendered task list before prompt")
	}

	last := m.transcript[len(m.transcript)-1]
	if last.command != transcriptTaskListForEditor {
		t.Fatalf("expected task list context command %q, got %q", transcriptTaskListForEditor, last.command)
	}
	if strings.Contains(last.output, "## Tasks") {
		t.Fatalf("expected task list output without markdown Tasks heading, got: %q", last.output)
	}
	if strings.Contains(last.output, "- [ ]") || strings.Contains(last.output, "- [x]") {
		t.Fatalf("expected task list output without markdown checkboxes, got: %q", last.output)
	}
	if !strings.Contains(last.output, "#urgent") {
		t.Fatalf("expected high-priority task tags metadata in output, got: %q", last.output)
	}
	if !strings.Contains(last.output, "#backlog") {
		t.Fatalf("expected low-priority task tags metadata in output, got: %q", last.output)
	}

	firstVisible := "1. ○  [P0]  High priority source-second task"
	secondVisible := "2. ○  [P3]  Low priority source-first task"
	firstIdx := strings.Index(last.output, firstVisible)
	secondIdx := strings.Index(last.output, secondVisible)
	if firstIdx == -1 || secondIdx == -1 {
		t.Fatalf("expected sorted task list rows in output; missing first=%t second=%t output=%q", firstIdx != -1, secondIdx != -1, last.output)
	}
	if firstIdx > secondIdx {
		t.Fatalf("expected visible row 1 before row 2, got firstIdx=%d secondIdx=%d output=%q", firstIdx, secondIdx, last.output)
	}

	m.textInput.SetValue("2")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated, ok = next.(Model)
	if !ok {
		t.Fatalf("expected Model after selecting task number, got %T", next)
	}
	m = &updated

	if m.taskEditStep != 1 {
		t.Fatalf("expected task edit step 1 after selecting task number, got %d", m.taskEditStep)
	}
	if got := m.textInput.Prompt; got != "Task text: " {
		t.Fatalf("expected task text prompt, got %q", got)
	}
	if got := m.textInput.Value(); got != "Low priority source-first task" {
		t.Fatalf("expected prefilled text for visible row 2 %q, got %q", "Low priority source-first task", got)
	}

	assertOutputFragmentsInOrder(t, last.output, []string{
		"High priority source-second task",
		"Low priority source-first task",
	})
}

func assertOutputFragmentsInOrder(t *testing.T, output string, fragments []string) {
	t.Helper()

	lastIdx := -1
	for _, fragment := range fragments {
		idx := strings.Index(output, fragment)
		if idx == -1 {
			t.Fatalf("expected output to contain fragment %q, got: %q", fragment, output)
		}
		if idx < lastIdx {
			t.Fatalf("expected fragment %q to appear after previous fragments, got: %q", fragment, output)
		}
		lastIdx = idx
	}
}

var ansiEscapeCodePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSIEscapeCodes(output string) string {
	return ansiEscapeCodePattern.ReplaceAllString(output, "")
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
