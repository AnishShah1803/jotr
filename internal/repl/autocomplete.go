package repl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
)

type Autocomplete struct {
	commands       map[string]bool
	commandNames   []string
	aliases        map[string]string
	subCommands    map[string][]string
	actionCommands map[string][]string
	paramCommands  map[string][]string
}

func NewAutocomplete(rootCmd *cobra.Command) *Autocomplete {
	a := &Autocomplete{
		commands:       make(map[string]bool),
		aliases:        make(map[string]string),
		subCommands:    make(map[string][]string),
		actionCommands: make(map[string][]string),
		paramCommands:  make(map[string][]string),
	}

	a.actionCommands["note"] = []string{"create", "open", "list", "rename", "delete", "move"}

	a.actionCommands["index"] = []string{"rebuild", "sync", "status"}
	a.actionCommands["alias"] = []string{"add", "remove", "list", "resolve"}
	a.actionCommands["schedule"] = []string{"add", "list", "delete"}
	a.actionCommands["shortcut"] = []string{"add", "remove", "list"}
	a.actionCommands["tags"] = []string{"list", "find", "stats"}

	a.actionCommands["git"] = []string{"status", "commit", "history", "diff"}
	a.actionCommands["bulk"] = []string{"rename", "tag"}

	a.actionCommands["links"] = []string{"outgoing", "backlinks"}
	a.actionCommands["frontmatter"] = []string{"get", "set", "remove", "list"}
	a.actionCommands["daily"] = []string{"append", "prepend", "read"}
	a.actionCommands["files"] = []string{}
	a.actionCommands["wordcount"] = []string{}
	a.actionCommands["outline"] = []string{}

	a.paramCommands["read"] = []string{"file=", "path=", "lines=", "format="}
	a.paramCommands["daily"] = []string{"date=", "open="}
	a.paramCommands["note"] = []string{"name=", "template="}
	a.paramCommands["note create"] = []string{"path=", "file="}
	a.paramCommands["note rename"] = []string{"file=", "name="}
	a.paramCommands["note delete"] = []string{"<query>", "--force"}
	a.paramCommands["note move"] = []string{"file=", "to="}
	a.paramCommands["search"] = []string{"query=", "path=", "limit=", "case", "format="}
	a.paramCommands["search context"] = []string{"query=", "path=", "limit="}
	a.paramCommands["tags"] = []string{"name="}
	a.paramCommands["capture"] = []string{"content="}
	a.paramCommands["links"] = []string{}
	a.paramCommands["frontmatter"] = []string{}
	a.paramCommands["daily append"] = []string{"content="}
	a.paramCommands["daily prepend"] = []string{"content="}
	a.paramCommands["files"] = []string{"folder=", "ext=", "total"}
	a.paramCommands["wordcount"] = []string{"file=", "path=", "words", "characters"}
	a.paramCommands["outline"] = []string{"file=", "path=", "format=", "total"}
	a.paramCommands["configure"] = []string{"--base-dir=", "--diary-dir=", "--todo-file=", "--pdp-file="}

	a.buildIndex(rootCmd, "")
	sort.Strings(a.commandNames)

	for parent := range a.subCommands {
		sort.Strings(a.subCommands[parent])
	}

	return a
}

func (a *Autocomplete) buildIndex(cmd *cobra.Command, parentPath string) {
	for _, subCmd := range cmd.Commands() {
		name := subCmd.Name()
		if name == "" || subCmd.Hidden {
			continue
		}

		var fullPath string
		if parentPath == "" {
			fullPath = name
		} else {
			fullPath = parentPath + " " + name
		}

		a.commands[name] = true
		a.commandNames = append(a.commandNames, name)

		if parentPath != "" {
			a.subCommands[parentPath] = append(a.subCommands[parentPath], name)
		}

		for _, alias := range subCmd.Aliases {
			a.aliases[alias] = name
			a.commands[alias] = true
			if parentPath != "" {
				a.subCommands[parentPath] = append(a.subCommands[parentPath], alias)
			}
		}

		if subCmd.HasSubCommands() {
			a.buildIndex(subCmd, fullPath)
		}
	}
}

// resolve returns the canonical command name for a given alias, or the name itself
// if it is not registered as an alias.
func (a *Autocomplete) resolve(name string) string {
	if canonical, ok := a.aliases[name]; ok {
		return canonical
	}
	return name
}

// resolvePath resolves the first word of a (possibly multi-word) command path.
// e.g. "n create" → "note create", "tags" → "tags".
func (a *Autocomplete) resolvePath(path string) string {
	if idx := strings.Index(path, " "); idx != -1 {
		return a.resolve(path[:idx]) + path[idx:]
	}
	return a.resolve(path)
}

func (a *Autocomplete) GetAllCommands() []string {
	c := make([]string, len(a.commandNames))
	copy(c, a.commandNames)
	return c
}

func (a *Autocomplete) IsCommand(name string) bool {
	return a.commands[name]
}

func (a *Autocomplete) Complete(input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return input
	}

	lastPart := parts[len(parts)-1]
	if len(parts) == 1 {
		return a.completeCommand(lastPart)
	}

	if strings.HasSuffix(input, " ") {
		return input
	}
	return input + " "
}

func (a *Autocomplete) completeCommand(partial string) string {
	var matches []string
	exactMatch := false

	for cmd := range a.commands {
		if cmd == partial {
			exactMatch = true
		}
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, cmd)
		}
	}

	if exactMatch {
		return partial + " "
	}

	if len(matches) == 0 {
		return partial
	}

	sort.Strings(matches)

	if len(matches) == 1 {
		return matches[0] + " "
	}

	prefix := longestCommonPrefix(matches)
	if len(prefix) > len(partial) {
		return prefix
	}

	return partial
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

func (a *Autocomplete) GetCompletions(partial string) []string {
	var matches []string
	for _, cmd := range a.commandNames {
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

func (a *Autocomplete) GetSubCommands(parent string) []string {
	parent = a.resolvePath(parent)
	var subs []string

	if subList, ok := a.subCommands[parent]; ok {
		subs = subList
	} else if subList, ok := a.actionCommands[parent]; ok {
		subs = subList
	}

	if subs == nil {
		return nil
	}

	var fullCommands []string
	for _, sub := range subs {
		fullCommands = append(fullCommands, parent+" "+sub)
	}
	return fullCommands
}

func (a *Autocomplete) GetSubCommandCompletions(parent, partial string) []string {
	parent = a.resolvePath(parent)
	var subNames []string

	if subList, ok := a.subCommands[parent]; ok {
		subNames = subList
	} else if subList, ok := a.actionCommands[parent]; ok {
		subNames = subList
	}

	var matches []string
	for _, sub := range subNames {
		if strings.HasPrefix(sub, partial) {
			matches = append(matches, parent+" "+sub)
		}
	}

	if params, ok := a.paramCommands[parent]; ok {
		for _, p := range params {
			if strings.HasPrefix(p, partial) {
				matches = append(matches, parent+" "+p)
			}
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return matches
}

func (a *Autocomplete) IsParamCommand(cmdName string) bool {
	cmdName = a.resolve(cmdName)
	_, ok := a.paramCommands[cmdName]
	return ok
}

func (a *Autocomplete) GetParamCompletions(cmdName string) []string {
	cmdName = a.resolve(cmdName)
	if params, ok := a.paramCommands[cmdName]; ok {
		var fullParams []string
		for _, p := range params {
			fullParams = append(fullParams, cmdName+" "+p)
		}
		return fullParams
	}
	return nil
}

// GetParamsForCommand returns the raw param tokens (e.g. ["path=", "file="]) for
// a given command path, or nil if none are registered.
func (a *Autocomplete) GetParamsForCommand(cmdName string) []string {
	cmdName = a.resolvePath(cmdName)
	if params, ok := a.paramCommands[cmdName]; ok {
		return params
	}
	return nil
}

func (a *Autocomplete) GetParamValueCompletions(paramName string, partial string, cfg *config.LoadedConfig) []string {
	if paramName == "path" || paramName == "file" {
		return a.GetPathCompletions(cfg, partial)
	}
	if paramName == "format" {
		options := []string{"pretty", "raw"}
		var matches []string
		for _, opt := range options {
			if strings.HasPrefix(opt, partial) {
				matches = append(matches, opt)
			}
		}
		return matches
	}
	return nil
}

func (a *Autocomplete) GetPathCompletions(cfg *config.LoadedConfig, partial string) []string {
	if cfg == nil {
		return nil
	}

	basePath := cfg.Paths.BaseDir
	searchPath := partial

	if idx := strings.LastIndex(partial, "/"); idx != -1 {
		searchPath = partial[:idx]
	}

	fullPath := filepath.Join(basePath, searchPath)
	entries, err := filepath.Glob(fullPath + "/*")
	if err != nil {
		return nil
	}

	var completions []string
	for _, entry := range entries {
		relPath, err := filepath.Rel(basePath, entry)
		if err != nil {
			continue
		}
		if info, err := os.Stat(entry); err == nil && info.IsDir() {
			completions = append(completions, relPath+"/")
		} else {
			completions = append(completions, relPath)
		}
	}

	return completions
}
