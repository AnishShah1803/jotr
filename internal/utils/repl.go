package utils

import "os"

// IsReplMode returns true if the application is running in REPL mode.
// In REPL mode, stdin is not available for interactive prompts.
func IsReplMode() bool {
	return os.Getenv("JOTR_REPL_MODE") == "true"
}
