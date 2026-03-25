package repl

// ParamKind describes how a parameter's value should be completed.
type ParamKind int

const (
	// ParamString accepts free-form text; no value completions offered.
	ParamString ParamKind = iota
	// ParamFlag is a bare flag with no = suffix (e.g. "total", "words", "case").
	ParamFlag
	// ParamFilePath triggers filesystem completion (files and directories).
	ParamFilePath
	// ParamDirPath triggers filesystem completion restricted to directories only.
	ParamDirPath
	// ParamEnum offers a fixed set of values for completion.
	ParamEnum
)

// ParamDef describes a single parameter accepted by a command or subcommand.
type ParamDef struct {
	// Name is the parameter key as typed in the REPL (without the trailing =).
	// For bare flags (ParamFlag), Name is the full token (e.g. "total").
	Name string
	Kind ParamKind
	// Values is the list of allowed values; only used when Kind == ParamEnum.
	Values []string
}

// CommandDef describes a REPL command (or "command subcommand" path) and the
// parameters it accepts. Subcommands are listed separately so that the
// autocomplete engine can offer them after the parent command is typed.
type CommandDef struct {
	// Name is the full command path as it appears in the REPL input,
	// e.g. "note", "note create", "daily append".
	Name string
	// Subcommands lists the immediate child subcommand names (not full paths).
	// Leave nil for leaf commands.
	Subcommands []string
	// Params lists the parameters this command path accepts.
	Params []ParamDef
}

// commandRegistry is the single source of truth for all REPL commands,
// their subcommands, and their accepted parameters.
//
// To add a new command:
//  1. Add a CommandDef entry here.
//  2. Implement the corresponding Cobra command in cmd/.
//  3. Nothing else needs to change — autocomplete derives everything from this table.
var commandRegistry = []CommandDef{
	// ── Notes ────────────────────────────────────────────────────────────────
	{
		Name:        "note",
		Subcommands: []string{"create", "open", "list", "rename", "delete", "move"},
		Params:      []ParamDef{{Name: "name", Kind: ParamString}, {Name: "template", Kind: ParamString}},
	},
	{
		Name:   "note create",
		Params: []ParamDef{{Name: "path", Kind: ParamFilePath}, {Name: "file", Kind: ParamFilePath}},
	},
	{
		Name:   "note rename",
		Params: []ParamDef{{Name: "file", Kind: ParamFilePath}, {Name: "name", Kind: ParamString}},
	},
	{
		Name:   "note delete",
		Params: []ParamDef{{Name: "--force", Kind: ParamFlag}, {Name: "<query>", Kind: ParamFlag}},
	},
	{
		Name:   "note move",
		Params: []ParamDef{{Name: "file", Kind: ParamFilePath}, {Name: "to", Kind: ParamString}},
	},

	// ── Daily ────────────────────────────────────────────────────────────────
	{
		Name:        "daily",
		Subcommands: []string{"append", "prepend", "read"},
		Params:      []ParamDef{{Name: "date", Kind: ParamString}, {Name: "open", Kind: ParamString}},
	},
	{
		Name:   "daily append",
		Params: []ParamDef{{Name: "content", Kind: ParamString}},
	},
	{
		Name:   "daily prepend",
		Params: []ParamDef{{Name: "content", Kind: ParamString}},
	},

	// ── Search ───────────────────────────────────────────────────────────────
	{
		Name: "search",
		Params: []ParamDef{
			{Name: "query", Kind: ParamString},
			{Name: "path", Kind: ParamFilePath},
			{Name: "limit", Kind: ParamString},
			{Name: "case", Kind: ParamFlag},
			{Name: "format", Kind: ParamEnum, Values: []string{"pretty", "raw"}},
		},
	},
	{
		Name: "search context",
		Params: []ParamDef{
			{Name: "query", Kind: ParamString},
			{Name: "path", Kind: ParamFilePath},
			{Name: "limit", Kind: ParamString},
		},
	},

	// ── Capture ──────────────────────────────────────────────────────────────
	{
		Name:   "capture",
		Params: []ParamDef{{Name: "content", Kind: ParamString}},
	},

	// ── Read ─────────────────────────────────────────────────────────────────
	{
		Name: "read",
		Params: []ParamDef{
			{Name: "file", Kind: ParamFilePath},
			{Name: "path", Kind: ParamFilePath},
			{Name: "lines", Kind: ParamString},
			{Name: "format", Kind: ParamEnum, Values: []string{"pretty", "raw"}},
		},
	},

	// ── Tags ─────────────────────────────────────────────────────────────────
	{
		Name:        "tags",
		Subcommands: []string{"list", "find", "stats"},
		Params:      []ParamDef{{Name: "name", Kind: ParamString}},
	},

	// ── Links ────────────────────────────────────────────────────────────────
	{
		Name:        "links",
		Subcommands: []string{"outgoing", "backlinks"},
	},

	// ── Frontmatter ──────────────────────────────────────────────────────────
	{
		Name:        "frontmatter",
		Subcommands: []string{"get", "set", "remove", "list"},
	},

	// ── Files ────────────────────────────────────────────────────────────────
	{
		Name: "files",
		Params: []ParamDef{
			{Name: "folder", Kind: ParamString},
			{Name: "ext", Kind: ParamString},
			{Name: "total", Kind: ParamFlag},
		},
	},

	// ── Word count ───────────────────────────────────────────────────────────
	{
		Name: "wordcount",
		Params: []ParamDef{
			{Name: "file", Kind: ParamFilePath},
			{Name: "path", Kind: ParamFilePath},
			{Name: "words", Kind: ParamFlag},
			{Name: "characters", Kind: ParamFlag},
		},
	},

	// ── Outline ──────────────────────────────────────────────────────────────
	{
		Name: "outline",
		Params: []ParamDef{
			{Name: "file", Kind: ParamFilePath},
			{Name: "path", Kind: ParamFilePath},
			{Name: "format", Kind: ParamEnum, Values: []string{"pretty", "raw"}},
			{Name: "total", Kind: ParamFlag},
		},
	},

	// ── Index ────────────────────────────────────────────────────────────────
	{
		Name:        "index",
		Subcommands: []string{"rebuild", "sync", "status"},
	},

	// ── Task ────────────────────────────────────────────────────────────────
	{
		Name:        "task",
		Subcommands: []string{"add", "list", "search", "complete", "edit", "archive", "prune", "stats", "sync"},
	},
	{
		Name: "task search",
		Params: []ParamDef{
			{Name: "query", Kind: ParamString},
			{Name: "status", Kind: ParamEnum, Values: []string{"pending", "completed", "all"}},
			{Name: "priority", Kind: ParamEnum, Values: []string{"P0", "P1", "P2", "P3"}},
			{Name: "tag", Kind: ParamString},
			{Name: "section", Kind: ParamString},
		},
	},

	// ── Alias ────────────────────────────────────────────────────────────────
	{
		Name:        "alias",
		Subcommands: []string{"add", "remove", "list", "resolve"},
	},

	// ── Schedule ─────────────────────────────────────────────────────────────
	{
		Name:        "schedule",
		Subcommands: []string{"add", "list", "delete"},
	},

	// ── Shortcut ─────────────────────────────────────────────────────────────
	{
		Name:        "shortcut",
		Subcommands: []string{"add", "remove", "list"},
	},

	// ── Git ──────────────────────────────────────────────────────────────────
	{
		Name:        "git",
		Subcommands: []string{"status", "commit", "history", "diff"},
	},

	// ── Bulk ─────────────────────────────────────────────────────────────────
	{
		Name:        "bulk",
		Subcommands: []string{"rename", "tag"},
	},

	// ── Configure ────────────────────────────────────────────────────────────
	{
		Name: "configure",
		Params: []ParamDef{
			{Name: "base-dir", Kind: ParamDirPath},
			{Name: "diary-dir", Kind: ParamFilePath},
			{Name: "todo-file", Kind: ParamFilePath},
			{Name: "pdp-file", Kind: ParamFilePath},
		},
	},
}

// registryByName returns a map from command path → *CommandDef for O(1) lookup.
func registryByName() map[string]*CommandDef {
	m := make(map[string]*CommandDef, len(commandRegistry))
	for i := range commandRegistry {
		m[commandRegistry[i].Name] = &commandRegistry[i]
	}
	return m
}
