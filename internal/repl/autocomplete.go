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
	registry       map[string]*CommandDef
}

func NewAutocomplete(rootCmd *cobra.Command) *Autocomplete {
	a := &Autocomplete{
		commands:       make(map[string]bool),
		aliases:        make(map[string]string),
		subCommands:    make(map[string][]string),
		actionCommands: make(map[string][]string),
		paramCommands:  make(map[string][]string),
		registry:       registryByName(),
	}

	for _, def := range commandRegistry {
		if def.Subcommands != nil {
			a.actionCommands[def.Name] = def.Subcommands
		}
		if len(def.Params) > 0 {
			tokens := make([]string, 0, len(def.Params))
			for _, p := range def.Params {
				if p.Kind == ParamFlag {
					tokens = append(tokens, p.Name)
				} else {
					tokens = append(tokens, p.Name+"=")
				}
			}
			a.paramCommands[def.Name] = tokens
		}
	}

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
		if parentPath == "" {
			a.commandNames = append(a.commandNames, name)
		}

		if parentPath != "" {
			a.subCommands[parentPath] = append(a.subCommands[parentPath], name)
		}

		for _, alias := range subCmd.Aliases {
			a.aliases[alias] = name
		}

		if subCmd.HasSubCommands() {
			a.buildIndex(subCmd, fullPath)
		}
	}
}

func (a *Autocomplete) resolve(name string) string {
	if canonical, ok := a.aliases[name]; ok {
		return canonical
	}
	return name
}

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

func (a *Autocomplete) GetParamValueCompletionsFromRegistry(cmdPath string, paramName string, partial string) []string {
	def, ok := a.registry[cmdPath]
	if !ok {
		return nil
	}
	for _, p := range def.Params {
		if p.Name != paramName {
			continue
		}
		switch p.Kind {
		case ParamDirPath:
			return a.getFilesystemPathCompletions(partial, true)
		case ParamFilePath:
			return a.getFilesystemPathCompletions(partial, false)
		case ParamEnum:
			var matches []string
			for _, v := range p.Values {
				if strings.HasPrefix(v, partial) {
					matches = append(matches, v)
				}
			}
			return matches
		}
		return nil
	}
	return nil
}

func (a *Autocomplete) getFilesystemPathCompletions(partial string, dirsOnly bool) []string {
	expandedPartial := partial
	if strings.HasPrefix(partial, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expandedPartial = filepath.Join(homeDir, partial[1:])
		}
	}

	searchDir := expandedPartial
	prefix := ""
	if idx := strings.LastIndex(expandedPartial, "/"); idx != -1 {
		searchDir = expandedPartial[:idx]
		prefix = expandedPartial[idx+1:]
	} else if expandedPartial != "" {
		searchDir = "."
		prefix = expandedPartial
	} else {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			searchDir = homeDir
		} else {
			searchDir = "."
		}
		prefix = ""
	}

	if info, err := os.Stat(searchDir); err != nil || !info.IsDir() {
		searchDir = "."
		if idx := strings.LastIndex(expandedPartial, "/"); idx != -1 {
			prefix = expandedPartial[idx+1:]
		} else {
			prefix = expandedPartial
		}
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	var completions []string
	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}

		if dirsOnly && !entry.IsDir() {
			continue
		}

		var completion string
		if searchDir == "." {
			completion = name
		} else {
			completion = filepath.Join(searchDir, name)
		}

		if strings.HasPrefix(partial, "~") {
			if homeDir, err := os.UserHomeDir(); err == nil {
				completion = "~" + strings.TrimPrefix(completion, homeDir)
			}
		}

		if entry.IsDir() {
			completion += "/"
		}

		completions = append(completions, completion)
	}

	return completions
}
