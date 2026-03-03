package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
)

var wordcountCmdFlags = struct {
	file      string
	path      string
	wordsOnly bool
	charsOnly bool
}{}

func init() {
	WordcountCmd.Flags().StringVar(&wordcountCmdFlags.file, "file", "", "note name to count")
	WordcountCmd.Flags().StringVar(&wordcountCmdFlags.path, "path", "", "file path to count")
	WordcountCmd.Flags().BoolVar(&wordcountCmdFlags.wordsOnly, "words", false, "show only word count")
	WordcountCmd.Flags().BoolVar(&wordcountCmdFlags.charsOnly, "characters", false, "show only character count")
}

var WordcountCmd = &cobra.Command{
	Use:   "wordcount [note-name]",
	Short: "Count words and characters in notes",
	Long: `Count words and characters in one or all notes.

Examples:
  jotr wordcount                    # Count all notes
  jotr wordcount MyNote             # Count a specific note
  jotr wordcount --file MyNote      # Same with flag
  jotr wordcount --words            # Show only word counts
  jotr wordcount --characters       # Show only character counts`,
	Aliases: []string{"wc"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		noteName := wordcountCmdFlags.file
		if len(args) > 0 {
			noteName = args[0]
		}

		if wordcountCmdFlags.path != "" {
			return countFile(wordcountCmdFlags.path, wordcountCmdFlags.wordsOnly, wordcountCmdFlags.charsOnly)
		}

		if noteName != "" {
			notePath, err := pickNoteByName(cmd.Context(), cfg, noteName)
			if err != nil {
				return err
			}
			return countFile(notePath, wordcountCmdFlags.wordsOnly, wordcountCmdFlags.charsOnly)
		}

		allNotes, err := notes.FindNotes(cmd.Context(), cfg.Paths.BaseDir)
		if err != nil {
			return err
		}

		var totalWords, totalChars int
		for _, np := range allNotes {
			content, err := os.ReadFile(np)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", np, err)
				continue
			}
			words := countWords(string(content))
			chars := utf8.RuneCount(content)
			totalWords += words
			totalChars += chars

			relPath, _ := filepath.Rel(cfg.Paths.BaseDir, np)
			if wordcountCmdFlags.wordsOnly {
				fmt.Printf("%6d  %s\n", words, relPath)
			} else if wordcountCmdFlags.charsOnly {
				fmt.Printf("%6d  %s\n", chars, relPath)
			} else {
				fmt.Printf("%6d words  %6d chars  %s\n", words, chars, relPath)
			}
		}

		fmt.Println()
		if wordcountCmdFlags.wordsOnly {
			fmt.Printf("%6d  total\n", totalWords)
		} else if wordcountCmdFlags.charsOnly {
			fmt.Printf("%6d  total\n", totalChars)
		} else {
			fmt.Printf("%6d words  %6d chars  total\n", totalWords, totalChars)
		}

		return nil
	},
}

func countFile(path string, wordsOnly, charsOnly bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	words := countWords(string(content))
	chars := utf8.RuneCount(content)

	name := filepath.Base(path)
	if wordsOnly {
		fmt.Printf("%d words in %s\n", words, name)
	} else if charsOnly {
		fmt.Printf("%d characters in %s\n", chars, name)
	} else {
		fmt.Printf("%s: %d words, %d characters\n", name, words, chars)
	}
	return nil
}

func countWords(s string) int {
	inWord := false
	count := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else {
			if !inWord {
				count++
			}
			inWord = true
		}
	}
	return count
}
