package repl

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type Autocomplete struct {
	commands map[string]bool
	aliases  map[string]string
}

func NewAutocomplete(rootCmd *cobra.Command) *Autocomplete {
	a := &Autocomplete{
		commands: make(map[string]bool),
		aliases:  make(map[string]string),
	}

	a.buildIndex(rootCmd)
	return a
}

func (a *Autocomplete) buildIndex(cmd *cobra.Command) {
	for _, subCmd := range cmd.Commands() {
		name := subCmd.Name()
		a.commands[name] = true

		for _, alias := range subCmd.Aliases {
			a.aliases[alias] = name
			a.commands[alias] = true
		}

		if subCmd.HasSubCommands() {
			a.buildIndex(subCmd)
		}
	}
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
	for cmd := range a.commands {
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, cmd)
		}
	}
	sort.Strings(matches)
	return matches
}
