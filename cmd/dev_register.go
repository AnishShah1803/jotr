//go:build dev

package cmd

func init() {
	rootCmd.AddCommand(devCmd)
}
