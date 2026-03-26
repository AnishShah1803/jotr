package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

// TaskService provides task management operations.
type TaskService struct {
	lockManager utils.FileLockManager
}

type TaskServiceOption func(*TaskService)

func WithLockManager(lm utils.FileLockManager) TaskServiceOption {
	return func(s *TaskService) { s.lockManager = lm }
}

func NewTaskService(opts ...TaskServiceOption) *TaskService {
	s := &TaskService{
		lockManager: utils.NewDefaultFileLockManager(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SyncOptions contains options for syncing tasks.
type SyncOptions struct {
	DiaryPath   string
	TodoPath    string
	StatePath   string
	LockTimeout time.Duration
	DryRun      bool
	DailyPath   string // optional: pre-calculated daily note path to use instead of building from DiaryPath
}

// acquireSyncLocks acquires locks on state, todo, and daily note files in the correct order.
// Returns a slice of lock handles that must be released in reverse order.
// Lock order: state file → todo file → daily note
func (s *TaskService) acquireSyncLocks(statePath, todoPath, notePath string, timeout time.Duration) ([]utils.LockHandle, error) {
	var locks []utils.LockHandle

	// Acquire lock order: state file → todo file → daily note
	// This ordering must be consistent to prevent deadlocks

	// Lock state file
	if statePath != "" {
		lockHandle, err := s.lockManager.LockFile(statePath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				s.lockManager.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on state file: %w", err)
		}
		locks = append(locks, lockHandle)
	}

	// Lock todo file
	if todoPath != "" {
		lockHandle, err := s.lockManager.LockFile(todoPath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				s.lockManager.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on todo file: %w", err)
		}
		locks = append(locks, lockHandle)
	}

	// Lock daily note (if provided)
	if notePath != "" {
		lockHandle, err := s.lockManager.LockFile(notePath, timeout)
		if err != nil {
			// Release any already acquired locks
			for _, l := range locks {
				s.lockManager.UnlockFile(l)
			}
			return nil, fmt.Errorf("failed to acquire lock on daily note: %w", err)
		}
		locks = append(locks, lockHandle)
	}

	return locks, nil
}

// isLockTimeoutError checks if an error is a lock timeout error.
// This is used to provide user-friendly error messages for sync operations.
func (s *TaskService) isLockTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, utils.ErrLockTimeout)
}

// SyncTasks performs bidirectional sync between daily notes and todo list.
func (s *TaskService) SyncTasks(ctx context.Context, opts SyncOptions) (*state.SyncResult, error) {
	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}
	notePath := opts.DailyPath
	if notePath == "" {
		today := time.Now()
		notePath = notes.BuildDailyNotePath(opts.DiaryPath, today)
	}
	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, notePath, lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another sync operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			s.lockManager.UnlockFile(locks[i])
		}
	}()

	return s.syncTasksWithLocks(ctx, opts, locks)
}

// syncTasksWithLocks performs the sync logic with pre-acquired locks (state + todo).
// The caller must own locks[0] (state file) and locks[1] (todo file), already acquired.
// This avoids deadlock when called from CreateTask/UpdateTask that already hold these locks.
func (s *TaskService) syncTasksWithLocks(ctx context.Context, opts SyncOptions, locks []utils.LockHandle) (*state.SyncResult, error) {
	result := &state.SyncResult{
		StatePath: opts.StatePath,
		TodoPath:  opts.TodoPath,
	}

	notePath := opts.DailyPath
	if notePath == "" {
		today := time.Now()
		notePath = notes.BuildDailyNotePath(opts.DiaryPath, today)
	}
	result.DailyPath = notePath

	if notePath != "" && !utils.FileExists(notePath) {
		return nil, fmt.Errorf("daily note doesn't exist: %s", notePath)
	}

	// Read all data AFTER locks are held to prevent race conditions
	var dailyTasks []tasks.Task
	if notePath != "" {
		var err error
		dailyTasks, err = tasks.ReadTasks(ctx, notePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read daily note: %w", err)
		}
	}

	var tasksNeedingIDWriteback []tasks.Task
	for i := range dailyTasks {
		hadNoID := dailyTasks[i].ID == ""
		tasks.EnsureTaskID(&dailyTasks[i])
		// Strip the ID comment from the text so state stores clean text.
		// The ID is already in task.ID; keeping it in Text causes spurious
		// "modified" detections on every subsequent sync.
		dailyTasks[i].Text = strings.TrimSpace(tasks.StripTaskID(dailyTasks[i].Text))
		if hadNoID {
			tasksNeedingIDWriteback = append(tasksNeedingIDWriteback, dailyTasks[i])
		}
	}

	// Write newly-generated IDs back to the daily note so future syncs can match tasks.
	if !opts.DryRun && len(tasksNeedingIDWriteback) > 0 {
		if err := s.writeNewTaskIDs(notePath, tasksNeedingIDWriteback); err != nil {
			return nil, fmt.Errorf("failed to write new task IDs to daily note: %w", err)
		}
	}

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if todoState.NeedsMigration() && utils.FileExists(opts.TodoPath) {
		existingTasks, err := tasks.ReadTasks(ctx, opts.TodoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing tasks during migration: %w", err)
		}
		if len(existingTasks) > 0 {
			todoState.MigrateFromMarkdown(existingTasks, "migration")
		}
	}

	var todoTasks []tasks.Task
	if utils.FileExists(opts.TodoPath) {
		todoTasks, _ = tasks.ReadTasks(ctx, opts.TodoPath)
	}

	if notePath != "" {
		for _, todoTask := range todoTasks {
			if todoTask.ID != "" {
				continue
			}
			cleanTodoText := strings.TrimSpace(tasks.StripTaskID(todoTask.Text))
			if cleanTodoText == "" {
				continue
			}
			alreadyInDaily := false
			for _, dailyTask := range dailyTasks {
				if strings.TrimSpace(tasks.StripTaskID(dailyTask.Text)) == cleanTodoText {
					alreadyInDaily = true
					break
				}
			}
			if alreadyInDaily {
				continue
			}
			if err := s.appendTaskToDailyNote(ctx, notePath, todoTask); err != nil {
				return nil, fmt.Errorf("failed to promote todo task to daily note: %w", err)
			}
		}
		if updatedDailyTasks, err := tasks.ReadTasks(ctx, notePath); err == nil {
			dailyTasks = updatedDailyTasks
			for i := range dailyTasks {
				tasks.EnsureTaskID(&dailyTasks[i])
				dailyTasks[i].Text = strings.TrimSpace(tasks.StripTaskID(dailyTasks[i].Text))
			}
		}
	}

	result.TasksRead = len(dailyTasks) + len(todoTasks)

	syncResult := todoState.BidirectionalSync(dailyTasks, todoTasks, notePath)

	result.Conflicts = syncResult.Conflicts
	result.ConflictsDetail = syncResult.ConflictsDetail
	if len(syncResult.Conflicts) > 0 {
		return result, nil
	}

	if !opts.DryRun {
		if syncResult.StateUpdated {
			if opts.StatePath != "" {
				if err := todoState.Write(opts.StatePath); err != nil {
					return nil, fmt.Errorf("failed to write state file: %w", err)
				}
			}
		}

		if syncResult.TodoChanged {
			if err := s.writeTodoFileFromState(opts.TodoPath, todoState, true); err != nil {
				return nil, fmt.Errorf("failed to write todo file: %w", err)
			}
		}

		if syncResult.DailyChanged {
			sourceFiles := make(map[string]bool)
			for _, taskID := range syncResult.ChangedTaskIDs {
				if taskState, exists := todoState.Tasks[taskID]; exists {
					switch taskState.Source {
					case "", "todo-list", "merged":
						if notePath != "" {
							sourceFiles[notePath] = true
						}
					case "deletion-detected":
						continue
					default:
						sourceFiles[taskState.Source] = true
					}
				}
			}

			for sourceFile := range sourceFiles {
				sourceTasks, err := tasks.ReadTasks(ctx, sourceFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read source file %s: %w", sourceFile, err)
				}

				if err := s.updateDailyNoteFromState(sourceFile, sourceTasks, todoState); err != nil {
					return nil, fmt.Errorf("failed to update daily note %s: %w", sourceFile, err)
				}
			}
		}
	}

	result.TasksFromDaily = syncResult.AppliedDaily
	result.TasksFromTodo = syncResult.AppliedTodo
	result.DeletedTasks = syncResult.Deleted
	result.DeletedTaskIDs = syncResult.DeletedTaskIDs
	result.ChangedTaskIDs = syncResult.ChangedTaskIDs

	result.AddedFromDaily = syncResult.AddedFromDaily
	result.UpdatedFromDaily = syncResult.UpdatedFromDaily
	result.AddedFromTodo = syncResult.AddedFromTodo
	result.UpdatedFromTodo = syncResult.UpdatedFromTodo
	result.DeletedTasksDetail = syncResult.DeletedTasksDetail

	return result, nil
}

func (s *TaskService) updateDailyNoteFromState(notePath string, dailyTasks []tasks.Task, todoState *state.TodoState) error {
	noteContent, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note: %w", err)
	}

	// Build O(1) lookup map: ID → task
	taskByID := make(map[string]tasks.Task, len(dailyTasks))
	for _, dt := range dailyTasks {
		if dt.ID != "" {
			taskByID[dt.ID] = dt
		}
	}

	lines := strings.Split(string(noteContent), "\n")
	var updatedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if this line is a task checkbox
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			matched := false
			if id := tasks.ExtractTaskID(line); id != "" {
				if _, inDailyTasks := taskByID[id]; inDailyTasks {
					if stateTask, exists := todoState.Tasks[id]; exists {
						updatedLines = append(updatedLines, s.formatTaskLine(stateTask))
						matched = true
					}
				}
			}
			if !matched {
				updatedLines = append(updatedLines, line)
			}
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	content := strings.Join(updatedLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := notes.WriteNote(context.Background(), notePath, content); err != nil {
		return fmt.Errorf("failed to write daily note: %w", err)
	}

	return nil
}

// writeNewTaskIDs writes newly-generated task IDs back to the daily note file.
// This ensures that tasks without IDs in the file get their generated IDs persisted,
// so future syncs (e.g. propagating completion status) can match them by ID.
func (s *TaskService) writeNewTaskIDs(notePath string, newIDTasks []tasks.Task) error {
	noteContent, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note: %w", err)
	}

	// Build a map from clean task text → task for O(1) lookup.
	// task.Text already has the ID appended (set by EnsureTaskID), so strip it to get the
	// original text that still appears in the file.
	taskByCleanText := make(map[string]tasks.Task, len(newIDTasks))
	for _, t := range newIDTasks {
		clean := strings.TrimSpace(tasks.StripTaskID(t.Text))
		taskByCleanText[clean] = t
	}

	lines := strings.Split(string(noteContent), "\n")
	var updatedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		updated := false
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			// Only rewrite lines that don't already have an ID.
			if tasks.ExtractTaskID(line) == "" {
				var lineText string
				if strings.HasPrefix(trimmed, "- [ ] ") {
					lineText = strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ] "))
				} else {
					lineText = strings.TrimSpace(strings.TrimPrefix(trimmed, "- [x] "))
				}
				if t, ok := taskByCleanText[lineText]; ok {
					updatedLines = append(updatedLines, s.formatTaskLine(func() state.TaskState {
						return state.TaskState{
							ID:        t.ID,
							Text:      tasks.StripTaskID(t.Text),
							Completed: t.Completed,
						}
					}()))
					updated = true
				}
			}
		}
		if !updated {
			updatedLines = append(updatedLines, line)
		}
	}

	content := strings.Join(updatedLines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return notes.WriteNote(context.Background(), notePath, content)
}

func (s *TaskService) formatTaskLine(stateTask state.TaskState) string {
	var sb strings.Builder
	if stateTask.Completed {
		sb.WriteString("- [x] ")
	} else {
		sb.WriteString("- [ ] ")
	}

	text := tasks.StripCompletedTag(tasks.StripTaskID(stateTask.Text))
	sb.WriteString(text)

	if stateTask.Priority != "" && !strings.Contains(text, "[P") {
		sb.WriteString(" [")
		sb.WriteString(stateTask.Priority)
		sb.WriteString("]")
	}

	for _, tag := range stateTask.Tags {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag == "" {
			continue
		}
		if !strings.Contains(text, "#"+tag) {
			sb.WriteString(" #")
			sb.WriteString(tag)
		}
	}

	if stateTask.ID != "" {
		sb.WriteString(fmt.Sprintf(" <!-- id: %s -->", stateTask.ID))
	}

	if stateTask.Completed && stateTask.CompletedDate != "" {
		sb.WriteString(fmt.Sprintf(" @completed(%s)", stateTask.CompletedDate))
	}

	return sb.String()
}

// writeTodoFileFromState generates and writes the todo markdown file from state.
func (s *TaskService) writeTodoFileFromState(todoPath string, todoState *state.TodoState, includeCompleted bool) error {
	var content strings.Builder
	content.WriteString("# To-Do List\n\n")

	var tasksToWrite []state.TaskState
	if includeCompleted {
		for _, ts := range todoState.Tasks {
			tasksToWrite = append(tasksToWrite, ts)
		}
	} else {
		tasksToWrite = todoState.GetActiveTasks()
	}

	sections := make(map[string][]state.TaskState)
	for _, task := range tasksToWrite {
		var section string
		// If task is completed and has a CompletedDate, use that as the section
		if task.Completed && task.CompletedDate != "" {
			section = task.CompletedDate
		} else {
			section = strings.TrimSpace(task.CreatedDate)
			if section == "" {
				section = "Tasks"
			}
		}
		sections[section] = append(sections[section], task)
	}

	var sectionNames []string
	for name := range sections {
		sectionNames = append(sectionNames, name)
	}

	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	sort.Slice(sectionNames, func(i, j int) bool {
		dateI := dateRegex.MatchString(sectionNames[i])
		dateJ := dateRegex.MatchString(sectionNames[j])

		if dateI && dateJ {
			return sectionNames[i] > sectionNames[j]
		}
		return dateI && !dateJ
	})

	for _, sectionName := range sectionNames {
		content.WriteString(fmt.Sprintf("## %s\n\n", sectionName))
		for _, task := range sections[sectionName] {
			content.WriteString(s.formatTaskLine(task) + "\n")
		}
		content.WriteString("\n")
	}

	if err := notes.WriteNote(context.Background(), todoPath, content.String()); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	return nil
}

func (s *TaskService) WriteTodoFileFromState(todoPath string, todoState *state.TodoState, includeCompleted bool) error {
	return s.writeTodoFileFromState(todoPath, todoState, includeCompleted)
}

// ArchiveOptions contains options for archiving tasks.
type ArchiveOptions struct {
	TodoPath    string
	StatePath   string
	BaseDir     string
	LockTimeout time.Duration // Timeout for acquiring file locks (default: 10s)
}

// ArchiveResult contains the result of an archive operation.
type ArchiveResult struct {
	ArchivePath    string
	ArchivedCount  int
	RemainingCount int
}

type CreateTaskOptions struct {
	DiaryPath   string
	TodoPath    string
	StatePath   string
	Text        string
	Section     string
	Priority    string
	Tags        []string
	LockTimeout time.Duration
}

type CreateTaskResult struct {
	TaskID   string
	TaskText string
}

type UpdateTaskOptions struct {
	DiaryPath   string
	TodoPath    string
	StatePath   string
	TaskID      string
	Text        string
	Section     string
	Priority    string
	Tags        []string
	LockTimeout time.Duration
}

type UpdateTaskResult struct {
	TaskID string
}

type PruneOptions struct {
	DiaryPath   string
	TodoPath    string
	StatePath   string
	LockTimeout time.Duration
	DryRun      bool
}

type PruneResult struct {
	RemovedTaskIDs  []string
	RemovedCount    int
	DuplicateCount  int
	OrphanCount     int
	DailyTaskCount  int
	TodoTaskCount   int
	RemainingCount  int
	PrunedStatePath string
	PrunedTodoPath  string
	OrphanedTasks   []state.TaskChangeDetail
	DuplicateTasks  []state.TaskChangeDetail
}

// ArchiveTasks moves completed tasks to an archive file using state as source of truth.
func (s *TaskService) ArchiveTasks(ctx context.Context, opts ArchiveOptions) (*ArchiveResult, error) {
	result := &ArchiveResult{}

	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}
	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, "", lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another archive operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			s.lockManager.UnlockFile(locks[i])
		}
	}()

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if todoState.NeedsMigration() && utils.FileExists(opts.TodoPath) {
		existingTasks, err := tasks.ReadTasks(ctx, opts.TodoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing tasks during migration: %w", err)
		}
		if len(existingTasks) > 0 {
			todoState.MigrateFromMarkdown(existingTasks, "migration")
		}
	}

	completedTasks := todoState.GetCompletedTasks()
	activeTasks := todoState.GetActiveTasks()

	if len(completedTasks) == 0 {
		result.ArchivedCount = 0
		result.RemainingCount = len(activeTasks)
		return result, nil
	}

	now := time.Now()

	archiveDir := filepath.Join(opts.BaseDir, "Archive")
	if err := notes.EnsureDir(archiveDir); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}

	archiveFile := filepath.Join(archiveDir, fmt.Sprintf("archive-%s.md", now.Format("2006-01")))

	var archiveContent string
	if utils.FileExists(archiveFile) {
		content, err := os.ReadFile(archiveFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read archive: %w", err)
		}
		archiveContent = string(content)
	} else {
		archiveContent = fmt.Sprintf("# Archive - %s\n\n", now.Format("January 2006"))
	}

	archiveContent += fmt.Sprintf("\n## Archived on %s\n\n", now.Format("2006-01-02"))
	for _, task := range completedTasks {
		archiveContent += fmt.Sprintf("- [x] %s\n", task.Text)
	}

	if err := notes.WriteNote(context.Background(), archiveFile, archiveContent); err != nil {
		return nil, fmt.Errorf("failed to write archive file: %w", err)
	}

	if err := s.writeTodoFileFromState(opts.TodoPath, todoState, false); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	todoState.MarkArchived()
	if opts.StatePath != "" {
		if err := todoState.Write(opts.StatePath); err != nil {
			return nil, fmt.Errorf("failed to write state file: %w", err)
		}
	}

	result.ArchivedCount = len(completedTasks)
	result.RemainingCount = len(activeTasks)
	result.ArchivePath = archiveFile

	return result, nil
}

func (s *TaskService) CreateTask(ctx context.Context, opts CreateTaskOptions) (*CreateTaskResult, error) {
	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}

	dailyPath := notes.BuildDailyNotePath(opts.DiaryPath, time.Now())

	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, dailyPath, lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another task create operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			s.lockManager.UnlockFile(locks[i])
		}
	}()

	text := strings.TrimSpace(opts.Text)
	if text == "" {
		return nil, fmt.Errorf("task text is required")
	}
	priority := normalizePriority(opts.Priority)

	if len(opts.Tags) > 0 {
		for _, tag := range opts.Tags {
			tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
			if tag == "" {
				continue
			}
			if !strings.Contains(text, "#"+tag) {
				text += " #" + tag
			}
		}
	}

	if priority != "" && !strings.Contains(text, "[P") {
		text = strings.TrimSpace(fmt.Sprintf("%s [%s]", text, priority))
	}

	newTask := tasks.Task{
		Text:      text,
		Section:   strings.TrimSpace(opts.Section),
		Priority:  priority,
		Tags:      normalizeTaskTags(opts.Tags),
		Completed: false,
	}
	tasks.EnsureTaskID(&newTask)

	if !utils.FileExists(dailyPath) {
		if err := notes.CreateDailyNote(ctx, dailyPath, []string{"Tasks"}, time.Now()); err != nil {
			return nil, fmt.Errorf("failed to create daily note: %w", err)
		}
	}
	if err := s.appendTaskToDailyNote(ctx, dailyPath, newTask); err != nil {
		return nil, err
	}

	if _, err := s.syncTasksWithLocks(ctx, SyncOptions{
		DiaryPath: opts.DiaryPath,
		TodoPath:  opts.TodoPath,
		StatePath: opts.StatePath,
		DailyPath: dailyPath,
	}, locks); err != nil {
		return nil, err
	}

	return &CreateTaskResult{TaskID: newTask.ID, TaskText: newTask.Text}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, opts UpdateTaskOptions) (*UpdateTaskResult, error) {
	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}

	editSourcePath := opts.TodoPath
	syncDailyPath := ""

	if statePath, err := state.Read(opts.StatePath); err == nil {
		if existing, ok := statePath.Tasks[opts.TaskID]; ok {
			if existing.Source != "" && existing.Source != "todo-list" && existing.Source != "merged" {
				editSourcePath = existing.Source
				syncDailyPath = editSourcePath
			}
		}
	}

	lockSourcePath := ""
	if editSourcePath != opts.TodoPath {
		lockSourcePath = editSourcePath
	}

	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, lockSourcePath, lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another task edit operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			s.lockManager.UnlockFile(locks[i])
		}
	}()

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	existing, ok := todoState.Tasks[opts.TaskID]
	if !ok {
		return nil, fmt.Errorf("task %s not found", opts.TaskID)
	}

	priority := normalizePriority(opts.Priority)
	updatedTask := tasks.Task{
		Text:      strings.TrimSpace(opts.Text),
		Section:   strings.TrimSpace(opts.Section),
		Priority:  priority,
		Tags:      normalizeTaskTags(opts.Tags),
		ID:        opts.TaskID,
		Completed: existing.Completed,
	}
	if updatedTask.Text == "" {
		updatedTask.Text = existing.Text
	}
	if updatedTask.Section == "" {
		updatedTask.Section = existing.Section
	}
	if updatedTask.Priority == "" {
		updatedTask.Priority = existing.Priority
	}
	if len(updatedTask.Tags) == 0 {
		updatedTask.Tags = existing.Tags
	}

	if err := s.replaceTaskLineInFile(ctx, editSourcePath, state.TaskState{
		Text:      updatedTask.Text,
		Section:   updatedTask.Section,
		Priority:  updatedTask.Priority,
		Tags:      updatedTask.Tags,
		ID:        updatedTask.ID,
		Completed: updatedTask.Completed,
		Source:    editSourcePath,
	}); err != nil {
		return nil, err
	}

	if _, err := s.syncTasksWithLocks(ctx, SyncOptions{
		DiaryPath: opts.DiaryPath,
		TodoPath:  opts.TodoPath,
		StatePath: opts.StatePath,
		DailyPath: syncDailyPath,
	}, locks); err != nil {
		return nil, err
	}

	return &UpdateTaskResult{TaskID: opts.TaskID}, nil
}

func (s *TaskService) PruneTasks(ctx context.Context, opts PruneOptions) (*PruneResult, error) {
	lockTimeout := opts.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = 10 * time.Second
	}

	locks, err := s.acquireSyncLocks(opts.StatePath, opts.TodoPath, "", lockTimeout)
	if err != nil {
		if s.isLockTimeoutError(err) {
			return nil, fmt.Errorf("another prune operation is in progress. Please try again in a few seconds")
		}
		return nil, err
	}
	defer func() {
		if locks == nil {
			return
		}
		for i := len(locks) - 1; i >= 0; i-- {
			s.lockManager.UnlockFile(locks[i])
		}
	}()

	todoState, err := state.Read(opts.StatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if todoState.NeedsMigration() && utils.FileExists(opts.TodoPath) {
		existingTasks, err := tasks.ReadTasks(ctx, opts.TodoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing tasks during migration: %w", err)
		}
		if len(existingTasks) > 0 {
			todoState.MigrateFromMarkdown(existingTasks, "migration")
		}
	}

	allNotes, err := notes.FindNotes(ctx, opts.DiaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find daily notes: %w", err)
	}

	dailyPaths := make([]string, 0)
	for _, p := range allNotes {
		if isDailyNotePath(p) {
			dailyPaths = append(dailyPaths, p)
		}
	}

	allDailyTasks := make([]tasks.Task, 0)
	for _, dailyPath := range dailyPaths {
		tasksOnNote, err := tasks.ReadTasks(ctx, dailyPath)
		if err != nil {
			continue
		}
		for _, task := range tasksOnNote {
			if task.ID != "" {
				allDailyTasks = append(allDailyTasks, task)
			}
		}
	}

	result := &PruneResult{
		DailyTaskCount:  len(allDailyTasks),
		TodoTaskCount:   len(todoState.Tasks),
		PrunedStatePath: opts.StatePath,
		PrunedTodoPath:  opts.TodoPath,
	}

	keepIDs := make(map[string]bool)
	seenDailyIDs := make(map[string]bool)
	seenDailyTexts := make(map[string]bool)
	for _, task := range allDailyTasks {
		cleanText := strings.TrimSpace(tasks.StripTaskID(task.Text))
		if task.ID != "" {
			seenDailyIDs[task.ID] = true
			keepIDs[task.ID] = true
		}
		if cleanText != "" {
			seenDailyTexts[cleanText] = true
		}
	}

	removedIDs := make([]string, 0)
	orphanedTasks := make([]state.TaskChangeDetail, 0)
	duplicateTasks := make([]state.TaskChangeDetail, 0)
	seenTodoTexts := make(map[string]string)
	for id, taskState := range todoState.Tasks {
		cleanText := strings.TrimSpace(tasks.StripTaskID(taskState.Text))
		if !seenDailyIDs[id] {
			if taskState.Source == "" || taskState.Source == "todo-list" || taskState.Source == "merged" {
				continue
			}
			removedIDs = append(removedIDs, id)
			orphanedTasks = append(orphanedTasks, state.TaskChangeDetail{ID: id, Text: cleanText, Change: "deleted", Details: "missing from daily notes"})
			result.OrphanCount++
			continue
		}
		if cleanText != "" {
			if firstID, ok := seenTodoTexts[cleanText]; ok && firstID != id {
				removedIDs = append(removedIDs, id)
				duplicateTasks = append(duplicateTasks, state.TaskChangeDetail{ID: id, Text: cleanText, Change: "deleted", Details: fmt.Sprintf("duplicate of %s", firstID)})
				result.DuplicateCount++
				continue
			}
			seenTodoTexts[cleanText] = id
		}
	}

	if opts.DryRun {
		result.RemovedTaskIDs = removedIDs
		result.RemovedCount = len(removedIDs)
		result.OrphanedTasks = orphanedTasks
		result.DuplicateTasks = duplicateTasks
		result.RemainingCount = len(todoState.Tasks) - len(removedIDs)
		return result, nil
	}

	for _, id := range removedIDs {
		todoState.RemoveTask(id)
	}
	result.RemovedTaskIDs = removedIDs
	result.RemovedCount = len(removedIDs)
	result.OrphanedTasks = orphanedTasks
	result.DuplicateTasks = duplicateTasks
	result.RemainingCount = len(todoState.Tasks)

	if err := todoState.Write(opts.StatePath); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}
	if err := s.writeTodoFileFromState(opts.TodoPath, todoState, false); err != nil {
		return nil, fmt.Errorf("failed to write todo file: %w", err)
	}

	return result, nil
}

func isDailyNotePath(path string) bool {
	base := filepath.Base(path)
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-[A-Za-z]{3}\.md$`).MatchString(base)
}

func (s *TaskService) replaceTaskLineInFile(ctx context.Context, path string, taskState state.TaskState) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read task source file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	updated := false
	for i, line := range lines {
		if tasks.ExtractTaskID(line) == taskState.ID {
			lines[i] = s.formatTaskLine(taskState)
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("task %s not found in source file", taskState.ID)
	}

	newContent := strings.Join(lines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	return notes.WriteNote(ctx, path, newContent)
}

func (s *TaskService) appendTaskToDailyNote(ctx context.Context, notePath string, task tasks.Task) error {
	if !utils.FileExists(notePath) {
		if err := notes.CreateDailyNote(ctx, notePath, []string{"Tasks"}, time.Now()); err != nil {
			return fmt.Errorf("failed to create daily note: %w", err)
		}
	}

	contentBytes, err := os.ReadFile(notePath)
	if err != nil {
		return fmt.Errorf("failed to read daily note: %w", err)
	}
	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		lines = []string{}
	}

	if len(lines) == 0 {
		lines = append(lines, fmt.Sprintf("# %s", time.Now().Format("2006-01-02-Mon")), "")
	}

	insertAt := -1
	targetSection := "Tasks"
	if strings.TrimSpace(task.Section) != "" {
		targetSection = strings.TrimSpace(task.Section)
	}

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "## "+targetSection) {
			insertAt = i + 1
			for insertAt < len(lines) && strings.TrimSpace(lines[insertAt]) == "" {
				insertAt++
			}
			break
		}
	}

	if insertAt == -1 {
		lastTaskIdx := -1
		for i := len(lines) - 1; i >= 0; i-- {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
				lastTaskIdx = i
				break
			}
		}
		if lastTaskIdx >= 0 {
			insertAt = lastTaskIdx + 1
		} else {
			insertAt = len(lines)
		}
	}

	taskLine := "- [ ] " + strings.TrimSpace(tasks.StripTaskID(task.Text))
	priority := normalizePriority(task.Priority)
	if priority != "" && !strings.Contains(taskLine, "[P") {
		taskLine = strings.TrimSpace(taskLine + " [" + priority + "]")
	}
	for _, tag := range task.Tags {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag != "" && !strings.Contains(taskLine, "#"+tag) {
			taskLine += " #" + tag
		}
	}
	taskLine += fmt.Sprintf(" <!-- id: %s -->", task.ID)

	updated := make([]string, 0, len(lines)+2)
	updated = append(updated, lines[:insertAt]...)
	updated = append(updated, taskLine)
	updated = append(updated, lines[insertAt:]...)

	content = strings.Join(updated, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return notes.WriteNote(ctx, notePath, content)
}

func normalizePriority(priority string) string {
	priority = strings.TrimSpace(strings.ToUpper(priority))
	if priority == "" {
		return ""
	}
	priority = strings.TrimPrefix(priority, "P")
	if len(priority) != 1 || priority < "0" || priority > "3" {
		return ""
	}
	return "P" + priority
}

// GetAllTasks reads all tasks from a file.
func (s *TaskService) GetAllTasks(ctx context.Context, todoPath string) ([]tasks.Task, error) {
	return tasks.ReadTasks(ctx, todoPath)
}

// GetTaskSummary returns a summary of tasks grouped by priority.
func (s *TaskService) GetTaskSummary(ctx context.Context, todoPath string) (*tasks.Task, error) {
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	completed := false
	pendingTasks := tasks.FilterTasks(allTasks, &completed, "", "")

	_ = tasks.GroupByPriority(pendingTasks)

	if len(pendingTasks) > 0 {
		return &pendingTasks[0], nil
	}

	return nil, nil
}

// GetTaskStats returns statistics about tasks.
func (s *TaskService) GetTaskStats(ctx context.Context, todoPath string) (*TaskStats, error) {
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	stats := &TaskStats{
		Total:      len(allTasks),
		ByPriority: tasks.GroupByPriority(allTasks),
		BySection:  tasks.GroupBySection(allTasks),
	}

	_, completed, pending := tasks.CountTasks(allTasks)
	stats.Completed = completed
	stats.Pending = pending

	if stats.Total > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	// Count overdue tasks
	for _, task := range allTasks {
		if tasks.IsOverdue(task) {
			stats.Overdue++
		}
	}

	return stats, nil
}

// TaskStats contains task statistics.
type TaskStats struct {
	ByPriority     map[string][]tasks.Task
	BySection      map[string][]tasks.Task
	Total          int
	Completed      int
	Pending        int
	Overdue        int
	CompletionRate float64
}

// GetTasksAndStats reads tasks from a file once and returns both the tasks and their statistics.
// This eliminates the double-read pattern when both data are needed.
func (s *TaskService) GetTasksAndStats(ctx context.Context, todoPath string) ([]tasks.Task, *TaskStats, error) {
	allTasks, err := tasks.ReadTasks(ctx, todoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	stats := &TaskStats{
		Total:      len(allTasks),
		ByPriority: tasks.GroupByPriority(allTasks),
		BySection:  tasks.GroupBySection(allTasks),
	}

	_, completed, pending := tasks.CountTasks(allTasks)
	stats.Completed = completed
	stats.Pending = pending

	if stats.Total > 0 {
		stats.CompletionRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	// Count overdue tasks
	for _, task := range allTasks {
		if tasks.IsOverdue(task) {
			stats.Overdue++
		}
	}

	return allTasks, stats, nil
}

func normalizeTaskTags(tags []string) []string {
	var result []string
	seen := make(map[string]bool)
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		result = append(result, tag)
	}
	return result
}
