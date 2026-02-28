package repl

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type Autocomplete struct {
	commands       map[string]bool
	commandNames   []string
	aliases        map[string]string
	subCommands    map[string][]string
	actionCommands map[string][]string
}

func NewAutocomplete(rootCmd *cobra.Command) *Autocomplete {
	a := &Autocomplete{
		commands:       make(map[string]bool),
		aliases:        make(map[string]string),
		subCommands:    make(map[string][]string),
		actionCommands: make(map[string][]string),
	}

	// Commands that use args instead of sub-commands (action-style)
	a.actionCommands["note"] = []string{"create", "open", "list"}
	a.actionCommands["n"] = []string{"create", "open", "list"}
	a.actionCommands["index"] = []string{"rebuild", "sync", "status"}
	a.actionCommands["alias"] = []string{"add", "remove", "list", "resolve"}
	a.actionCommands["schedule"] = []string{"add", "list", "delete"}
	a.actionCommands["shortcut"] = []string{"add", "remove", "list"}
	a.actionCommands["tags"] = []string{"list", "find", "stats"}
	a.actionCommands["tag"] = []string{"list", "find", "stats"}
	a.actionCommands["git"] = []string{"status", "commit", "history", "diff"}
	a.actionCommands["bulk"] = []string{"rename", "tag"}

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

		// Use full parent path as key to support arbitrary depth without ambiguity
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

func (a *Autocomplete) GetAllCommands() []string {
	c := make([]string, len(a.commandNames))
	copy(c, a.commandNames)
	return c
}

// IsCommand reports whether name is a registered command or alias.
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
	var subNames []string

	if subList, ok := a.subCommands[parent]; ok {
		subNames = subList
	} else if subList, ok := a.actionCommands[parent]; ok {
		subNames = subList
	}

	if subNames == nil {
		return nil
	}

	var matches []string
	for _, sub := range subNames {
		if strings.HasPrefix(sub, partial) {
			matches = append(matches, parent+" "+sub)
		}
	}
	return matches
}
