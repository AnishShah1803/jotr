package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cterm "github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
)

func PromptPathWithCompletion(prompt string, r *bufio.Reader) (string, error) {
	if IsReplMode() {
		return "", fmt.Errorf("cannot prompt in REPL mode: provide path as argument")
	}

	fmt.Print(prompt)

	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		input, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(input), nil
	}

	return readPathRaw(prompt)
}

func readPathRaw(prompt string) (string, error) {
	fd := os.Stdin.Fd()
	oldState, err := cterm.MakeRaw(fd)
	if err != nil {
		return readLineFallback()
	}
	defer cterm.Restore(fd, oldState) //nolint:errcheck

	termWidth, termHeight, err := cterm.GetSize(fd)
	if err != nil || termWidth < 20 {
		termWidth = 80
	}
	if termHeight < 5 {
		termHeight = 24
	}

	var buf []byte

	repaint := func() {
		fmt.Printf("\r\033[2K%s%s", prompt, string(buf))
	}

	// redrawAll erases from the current line downward and redraws the prompt
	// line plus the completion grid (if any labels are set). This is the only
	// place that writes to the terminal for grid updates — no cursor-up tricks,
	// no reserved-space logic.
	var cycleMatches []string
	var cycleLabels []string
	var cycleIndex int
	cycleActive := false

	redrawAll := func(selectedIdx int) {
		// Erase from cursor (prompt line) to end of screen.
		fmt.Print("\r\033[J")
		// Redraw the prompt + current input.
		fmt.Printf("%s%s", prompt, string(buf))

		if len(cycleLabels) == 0 {
			return
		}

		maxLen := 0
		for _, l := range cycleLabels {
			if len(l) > maxLen {
				maxLen = len(l)
			}
		}
		colWidth := maxLen + 2
		cols := termWidth / colWidth
		if cols < 1 {
			cols = 1
		}
		rows := (len(cycleLabels) + cols - 1) / cols

		maxRows := termHeight - 1
		if rows > maxRows {
			rows = maxRows
		}

		// Print each grid row on a new line below the prompt.
		for row := 0; row < rows; row++ {
			fmt.Print("\r\n")
			for col := 0; col < cols; col++ {
				idx := row*cols + col
				if idx >= len(cycleLabels) {
					break
				}
				label := cycleLabels[idx]
				padded := label + strings.Repeat(" ", colWidth-len(label))
				if idx == selectedIdx {
					fmt.Printf("\033[7m%s\033[0m", padded)
				} else {
					fmt.Print(padded)
				}
			}
		}
		// Move cursor back up to the prompt line.
		fmt.Printf("\033[%dA\r%s%s", rows, prompt, string(buf))
	}

	clearGrid := func() {
		// Erase from prompt line downward, redraw prompt only.
		fmt.Printf("\r\033[J%s%s", prompt, string(buf))
	}

	stdin := os.Stdin

	for {
		b := make([]byte, 1)
		_, err := stdin.Read(b)
		if err != nil {
			return "", err
		}

		ch := b[0]

		if ch != '\t' && cycleActive {
			clearGrid()
			cycleActive = false
			cycleMatches = nil
			cycleLabels = nil
		}

		switch ch {
		case '\r', '\n':
			clearGrid()
			fmt.Print("\r\n")
			return strings.TrimSpace(string(buf)), nil

		case 127, 8:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				repaint()
			}

		case '\t':
			if !cycleActive {
				matches, _ := getPathMatches(string(buf))
				if len(matches) == 0 {
					break
				}
				// If there's a single match and it's a directory, auto-expand into
				// its children. Keep drilling until we have 0 or 2+ children.
				emptyDir := false
				for len(matches) == 1 && strings.HasSuffix(matches[0], "/") {
					buf = []byte(matches[0])
					children, _ := getPathMatches(string(buf))
					if len(children) == 0 {
						// Empty directory — show the path as-is.
						repaint()
						emptyDir = true
						break
					}
					matches = children
				}
				if emptyDir {
					break
				}
				if len(matches) == 1 {
					// Single non-directory match — complete and stop.
					buf = []byte(matches[0])
					repaint()
					break
				}
				cycleMatches = matches
				cycleLabels = make([]string, len(matches))
				for i, m := range matches {
					cycleLabels[i] = filepath.Base(strings.TrimSuffix(m, "/")) + "/"
				}
				cycleIndex = 0
				cycleActive = true
				buf = []byte(cycleMatches[0])
				redrawAll(0)
			} else {
				cycleIndex = (cycleIndex + 1) % len(cycleMatches)
				buf = []byte(cycleMatches[cycleIndex])
				redrawAll(cycleIndex)
			}

		case 3:
			clearGrid()
			fmt.Print("\r\n")
			return "", fmt.Errorf("interrupted")

		case 27:
			seq := make([]byte, 2)
			if _, err := stdin.Read(seq); err != nil {
				// EOF or read error consuming escape sequence — ignore
				break
			}

		default:
			if ch >= 32 {
				buf = append(buf, ch)
				fmt.Printf("%c", ch)
			}
		}
	}
}

func readLineFallback() (string, error) {
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func getPathMatches(input string) (matches []string, base string) {
	prefix := input
	expandedPrefix := prefix
	if strings.HasPrefix(prefix, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expandedPrefix = filepath.Join(homeDir, prefix[1:])
		}
	}

	var searchDir, partial string
	if strings.HasSuffix(expandedPrefix, "/") || expandedPrefix == "" {
		searchDir = expandedPrefix
		if searchDir == "" {
			// Default to home directory when input is empty.
			homeDir, err := os.UserHomeDir()
			if err == nil {
				searchDir = homeDir
			} else {
				searchDir = "."
			}
		}
		partial = ""
	} else {
		searchDir = filepath.Dir(expandedPrefix)
		partial = filepath.Base(expandedPrefix)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(partial)) {
			var newPath string
			if searchDir == "." {
				newPath = name + "/"
			} else {
				newPath = filepath.Join(searchDir, name) + "/"
			}
			if strings.HasPrefix(prefix, "~") {
				homeDir, _ := os.UserHomeDir()
				newPath = "~" + strings.TrimPrefix(newPath, homeDir)
			}
			matches = append(matches, newPath)
		}
	}

	return matches, searchDir
}
