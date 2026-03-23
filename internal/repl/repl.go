package repl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

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

	if _, _, err := m.rootCmd.Find(args); err != nil {
		return fmt.Sprintf("Command not found: %s\nType 'help' for available commands.", args[0])
	}

	// Capture Cobra output separately from any direct stdout/stderr writes.
	var outBuf, errBuf bytes.Buffer

	// Set output on the original root command before cloning it so that
	// subcommands, which traverse the parent chain to resolve their writer,
	// write into our buffers rather than os.Stdout/os.Stderr.
	m.rootCmd.SetOut(&outBuf)
	m.rootCmd.SetErr(&errBuf)
	defer func() {
		m.rootCmd.SetOut(nil)
		m.rootCmd.SetErr(nil)
	}()

	// Use a cloned root command with Run/RunE disabled to avoid
	// re-entering the REPL when no subcommand is selected.
	tmpRoot := *m.rootCmd
	tmpRoot.Run = nil
	tmpRoot.RunE = nil
	tmpRoot.SetArgs(args)

	// Reset any changed flags between REPL commands to prevent previous values
	// from leaking into the next execution.  Find the target on the clone so the
	// reset operates on the same command set that Execute will use.
	subCmd, _, _ := tmpRoot.Find(args)
	resetFlags := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				if setErr := f.Value.Set(f.DefValue); setErr != nil {
					fmt.Fprintf(&errBuf, "warning: could not reset flag %q: %v\n", f.Name, setErr)
				}
				f.Changed = false
			}
		})
	}
	resetFlags(tmpRoot.Flags())
	if subCmd != nil && subCmd != &tmpRoot {
		resetFlags(subCmd.Flags())
		resetFlags(subCmd.InheritedFlags())
	}

	directOutput, execErr := captureProcessOutput(func() error {
		return tmpRoot.Execute()
	})

	var result strings.Builder

	if outBuf.Len() > 0 {
		result.WriteString(outBuf.String())
	}

	if errBuf.Len() > 0 {
		result.WriteString(errBuf.String())
	}
	if directOutput != "" {
		result.WriteString(directOutput)
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

func captureProcessOutput(fn func() error) (output string, err error) {
	origStdout := os.Stdout
	origStderr := os.Stderr

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var stdoutErr error
	var stderrErr error
	var wg sync.WaitGroup

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return "", err
	}
	defer stdoutR.Close()

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		_ = stdoutW.Close()
		return "", err
	}
	defer stderrR.Close()

	os.Stdout = stdoutW
	os.Stderr = stderrW
	defer func() {
		_ = stdoutW.Close()
		_ = stderrW.Close()
		wg.Wait()
		os.Stdout = origStdout
		os.Stderr = origStderr

		combined := stdoutBuf.String() + stderrBuf.String()
		if stdoutErr != nil || stderrErr != nil {
			captureErr := fmt.Errorf("capture output: stdout=%v stderr=%v", stdoutErr, stderrErr)
			if err != nil {
				err = errors.Join(err, captureErr)
			} else {
				err = captureErr
			}
		}
		output = combined
	}()
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, stdoutErr = io.Copy(&stdoutBuf, stdoutR)
	}()
	go func() {
		defer wg.Done()
		_, stderrErr = io.Copy(&stderrBuf, stderrR)
	}()

	err = fn()
	return
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
	// Set JOTR_REPL_MODE environment variable
	origReplMode := os.Getenv("JOTR_REPL_MODE")
	os.Setenv("JOTR_REPL_MODE", "true")
	defer os.Setenv("JOTR_REPL_MODE", origReplMode)

	m := NewModel(ctx, cfg, rootCmd)

	p := tea.NewProgram(
		&m,
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}
