package repl

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/AnishShah1803/jotr/internal/config"
)

func (m *Model) executeCommand(input string) string {
	args := parseInput(input)
	if len(args) == 0 {
		return ""
	}

	if args[0] == "exit" || args[0] == "quit" || args[0] == "q" {
		m.quitting = true
		return "Goodbye!"
	}

	_, _, err := m.rootCmd.Find(args)
	if err != nil {
		return fmt.Sprintf("Command not found: %s\nType 'help' for available commands.", args[0])
	}

	// Capture stdout and stderr using buffers
	var outBuf, errBuf bytes.Buffer

	// Use a cloned root command with Run/RunE disabled to avoid
	// re-entering the REPL when no subcommand is selected.
	tmpRoot := *m.rootCmd
	tmpRoot.Run = nil
	tmpRoot.RunE = nil
	tmpRoot.SetArgs(args)

	// Set output buffers on the cloned root command to capture all output
	// This is critical: tmpRoot.Execute() uses tmpRoot's output, not the subcommand's
	tmpRoot.SetOut(&outBuf)
	tmpRoot.SetErr(&errBuf)

	// Also set buffers on the original command for flags reset output
	tmpRoot.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			if setErr := f.Value.Set(f.DefValue); setErr != nil {
				fmt.Fprintf(&errBuf, "warning: could not reset flag %q: %v\n", f.Name, setErr)
			}
			f.Changed = false
		}
	})

	execErr := tmpRoot.Execute()

	var result strings.Builder

	if outBuf.Len() > 0 {
		result.WriteString(outBuf.String())
	}

	if errBuf.Len() > 0 {
		result.WriteString(errBuf.String())
	}

	if execErr != nil {
		if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
			result.WriteString("\n")
		}
		result.WriteString(fmt.Sprintf("Error: %v", execErr))
	}

	output := strings.TrimSpace(result.String())
	if output == "" {
		return ""
	}

	return output + "\n"
}

func parseInput(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range input {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func LaunchREPL(ctx context.Context, cfg *config.LoadedConfig, rootCmd *cobra.Command) error {
	m := NewModel(ctx, cfg, rootCmd)

	p := tea.NewProgram(
		&m,
	)

	_, err := p.Run()
	return err
}
