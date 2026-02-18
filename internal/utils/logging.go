package utils

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// ---- Legacy API (Backward Compatibility) ----

// logContext maintains global state for legacy API.
type logContext struct {
	verbose bool
	mu      sync.RWMutex
}

var globalLogContext = &logContext{verbose: false}

// SetVerbose enables or disables verbose output (legacy API).
func SetVerbose(verbose bool) {
	WithWLock(&globalLogContext.mu, func() {
		globalLogContext.verbose = verbose
	})
}

// SetVerboseWithContext sets verbose output with context (legacy API).
func SetVerboseWithContext(ctx context.Context, verbose bool) {
	WithWLock(&globalLogContext.mu, func() {
		globalLogContext.verbose = verbose
	})

	select {
	case <-ctx.Done():
		return
	default:
	}
}

// VerboseLog prints debug information when verbose mode is enabled (legacy API).
func VerboseLog(format string, args ...interface{}) {
	VerboseLogWithContext(context.Background(), format, args...)
}

// VerboseLogWithContext prints debug information when verbose mode is enabled with context (legacy API).
func VerboseLogWithContext(ctx context.Context, format string, args ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	WithRLock(&globalLogContext.mu, func() {
		if globalLogContext.verbose {
			timestamp := time.Now().Format("15:04:05")
			fmt.Fprintf(os.Stderr, "[%s] DEBUG: %s\n", timestamp, fmt.Sprintf(format, args...))
		}
	})
}

// VerboseLogError prints error details when verbose mode is enabled (legacy API).
func VerboseLogError(operation string, err error) {
	VerboseLogErrorWithContext(context.Background(), operation, err)
}

// VerboseLogErrorWithContext prints error details when verbose mode is enabled with context (legacy API).
func VerboseLogErrorWithContext(ctx context.Context, operation string, err error) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	WithRLock(&globalLogContext.mu, func() {
		if globalLogContext.verbose {
			timestamp := time.Now().Format("15:04:05")
			fmt.Fprintf(os.Stderr, "[%s] ERROR in %s: %v\n", timestamp, operation, err)
		}
	})
}

// PrintError prints an error message to stderr with consistent formatting.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
