package utils

import "errors"

// ---- Common Error Variables ----
// These error variables provide consistent, identifiable error types across the codebase.
// Use errors.Is() to check for specific errors in downstream code.

// Configuration Errors
var (
	ErrConfigNotFound         = errors.New("config file not found")
	ErrConfigReadFailed       = errors.New("failed to read config")
	ErrConfigParseFailed      = errors.New("failed to parse config")
	ErrConfigValidationFailed = errors.New("config validation failed")
	ErrConfigWriteFailed      = errors.New("failed to write config")
	ErrConfigBackupFailed     = errors.New("failed to backup config")
)

// Path/Directory Errors
var (
	ErrBaseDirRequired             = errors.New("base_dir is required in config")
	ErrDiaryDirRequired            = errors.New("diary_dir is required in config")
	ErrTodoFilePathRequired        = errors.New("todo_file_path is required in config")
	ErrDailyNotePatternRequired    = errors.New("daily_note_pattern is required")
	ErrDailyNoteDirPatternRequired = errors.New("daily_note_dir_pattern is required")
	ErrPatternPlaceholderMissing   = errors.New("pattern must contain %s placeholder")
	ErrPathNotDirectory            = errors.New("path exists but is not a directory")
	ErrCannotAccessDirectory       = errors.New("cannot access directory")
	ErrCannotCreateDirectory       = errors.New("cannot create directory")
	ErrCannotCreateAtRoot          = errors.New("cannot create directory at root")
)

// Permission Errors
var (
	ErrNoWritePermission       = errors.New("no write permission")
	ErrInsufficientDiskSpace   = errors.New("insufficient disk space")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

// Editor Errors
var (
	ErrEditorCommandEmpty  = errors.New("editor command is empty")
	ErrEditorNotFound      = errors.New("editor not found in PATH")
	ErrEditorNotExecutable = errors.New("editor is not executable")
)

// Note Errors
var (
	ErrNoteAlreadyExists = errors.New("note already exists")
	ErrNoteNotFound      = errors.New("note not found")
	ErrInvalidEditor     = errors.New("invalid editor")
)

// Task Errors
var (
	ErrTaskNoteNotFound = errors.New("task note doesn't exist")
	ErrTaskReadFailed   = errors.New("failed to read tasks")
	ErrTaskWriteFailed  = errors.New("failed to write tasks")
	ErrTaskLockFailed   = errors.New("failed to acquire lock")
)

// Archive Errors
var (
	ErrArchiveCreateFailed = errors.New("failed to create archive directory")
	ErrArchiveReadFailed   = errors.New("failed to read archive")
	ErrArchiveWriteFailed  = errors.New("failed to write archive")
)

// GitHub/Update Errors
var (
	ErrGitHubAPIFailed   = errors.New("GitHub API request failed")
	ErrUpdateCheckFailed = errors.New("failed to check for updates")
)

// Search Errors
var (
	ErrNoNotesFound        = errors.New("no notes found")
	ErrSearchQueryRequired = errors.New("search query required")
	ErrSearchFailed        = errors.New("search failed")
)

// User Input Errors
var (
	ErrNoteNameRequired = errors.New("note name is required")
	ErrInvalidSelection = errors.New("invalid selection")
)

// ---- Helper Functions ----

// WrapConfigError wraps a config-related error with additional context.
func WrapConfigError(err error, context string) error {
	if err == nil {
		return nil
	}
	return &configError{err: err, context: context}
}

// configError wraps a config error with additional context.
type configError struct {
	err     error
	context string
}

func (e *configError) Error() string {
	return e.context + ": " + e.err.Error()
}

func (e *configError) Unwrap() error {
	return e.err
}

// IsConfigError checks if the error is a configuration-related error.
func IsConfigError(err error) bool {
	return errors.Is(err, ErrConfigNotFound) ||
		errors.Is(err, ErrConfigReadFailed) ||
		errors.Is(err, ErrConfigParseFailed) ||
		errors.Is(err, ErrConfigValidationFailed) ||
		errors.Is(err, ErrBaseDirRequired)
}
