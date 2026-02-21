package tui

import (
	"time"

	"github.com/AnishShah1803/jotr/internal/tasks"
)

type tickMsg time.Time

type editorFinishedMsg struct{ err error }

type dataLoadedMsg struct {
	notes            []string
	tasks            []tasks.Task
	streak           int
	totalNotes       int
	totalTasks       int
	completedTasks   int
	editorConfigured bool
	editorFallback   bool
}

type previewLoadedMsg []byte

type updateCheckMsg struct {
	err       error
	version   string
	hasUpdate bool
}

type errorMsg struct {
	err       error
	retryable bool
}

type editorFallbackMsg struct{}

type editorOpenAttemptMsg struct {
	useShellFallback bool
}

func newErrorMsg(err error, retryable bool) errorMsg {
	return errorMsg{err: err, retryable: retryable}
}
