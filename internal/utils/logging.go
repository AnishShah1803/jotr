package utils

import (
	"fmt"
	"os"
	"time"
)

// Verbose logging utility for jotr
// This provides simple debug output when --verbose flag is used

var isVerbose = false

// SetVerbose enables or disables verbose output
func SetVerbose(verbose bool) {
	isVerbose = verbose
}

// VerboseLog prints debug information when verbose mode is enabled
func VerboseLog(format string, args ...interface{}) {
	if isVerbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stderr, "[%s] DEBUG: %s\n", timestamp, fmt.Sprintf(format, args...))
	}
}

// VerboseLogError prints error details when verbose mode is enabled
func VerboseLogError(operation string, err error) {
	if isVerbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stderr, "[%s] ERROR in %s: %v\n", timestamp, operation, err)
	}
}
