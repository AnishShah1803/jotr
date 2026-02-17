package cmd

import (
	"os"

	"github.com/AnishShah1803/jotr/internal/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages for jotr",
	Long: `Generate man pages for all jotr commands.

This command generates man pages in the roff format, which can be viewed
with the 'man' command or converted to other formats.

By default, pages are written to the current directory.`,
	Example: `  jotr man > jotr.1                    # Output to stdout
  jotr man ./man_pages           # Write to man_pages directory
  jotr man --dir /usr/share/man  # System man directory`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		if err := os.MkdirAll(dir, constants.FilePermDir); err != nil {
			return err
		}

		header := &doc.GenManHeader{
			Title:   "jotr",
			Section: "1",
		}

		if err := doc.GenManTree(rootCmd, header, dir); err != nil {
			return err
		}

		cmd.Printf("Man pages generated in: %s\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
}
