package constants

import "os"

// File permission constants for consistent file operations across the codebase.

const (
	// FilePerm0644 is the default permission for regular files (rw-r--r--).
	// Use for notes, configs, templates, and other non-sensitive files.
	FilePerm0644 = os.FileMode(0644)

	// FilePerm0600 is the permission for sensitive files (rw-------).
	// Use for locks, credentials, and other files that should only be
	// accessible by the owner.
	FilePerm0600 = os.FileMode(0600)
)
