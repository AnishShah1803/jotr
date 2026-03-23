package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/notes"
	"github.com/AnishShah1803/jotr/internal/services"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
	"github.com/AnishShah1803/jotr/internal/utils"
)

var (
	taskBulletRegex = regexp.MustCompile(`^(\s*[-*+]\s*)\[([ xX])\]\s*(.*)$`)
	taskTagRegex    = regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
)

var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `List, complete, edit, archive, and inspect tasks in the todo file.`,
}

var TaskAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Add a task",
	Long: `Create a new task in the todo list.

If no text is provided, the command prompts for it outside REPL mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return addTask(cmd.Context(), cfg, args)
	},
}

var TaskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List pending tasks",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showTaskList(cmd.Context(), cfg, false)
	},
}

var TaskCompleteCmd = &cobra.Command{
	Use:   "complete [task numbers]",
	Short: "Complete one or more tasks",
	Long: `Mark selected tasks complete by their displayed numbers.

If no task numbers are provided, the command shows the current task list and prompts for numbers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return completeTasks(cmd.Context(), cfg, args)
	},
}

var TaskEditCmd = &cobra.Command{
	Use:   "edit [task number]",
	Short: "Open a task for editing",
	Long: `Show the current task list and open the todo file in the configured editor.

If no task number is provided, the command prompts for one after rendering the list.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return editTask(cmd.Context(), cfg, args)
	},
}

var TaskArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive completed tasks from prior months",
	Long:  `Archive completed tasks whose date section falls before the current month.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return archiveMonthlyTasks(cmd.Context(), cfg)
	},
}

var taskPruneDryRun bool

var TaskPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune tasks missing from daily notes",
	Long: `Remove tasks from the todo list that do not appear in any daily note, and deduplicate tasks with the same ID or same text.

Use --dry-run to preview what would be removed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return pruneTasks(cmd.Context(), cfg)
	},
}

var TaskStatsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show monthly completed task count",
	Long:    `Show how many tasks were completed in the current month.`,
	Aliases: []string{"st"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}
		return showMonthlyTaskStats(cmd.Context(), cfg)
	},
}

func init() {
	TaskCmd.AddCommand(TaskListCmd)
	TaskCmd.AddCommand(TaskAddCmd)
	TaskCmd.AddCommand(TaskCompleteCmd)
	TaskCmd.AddCommand(TaskEditCmd)
	TaskCmd.AddCommand(TaskArchiveCmd)
	TaskCmd.AddCommand(TaskPruneCmd)
	TaskCmd.AddCommand(TaskStatsCmd)
	TaskPruneCmd.Flags().BoolVar(&taskPruneDryRun, "dry-run", false, "Show what would be removed without making changes")
}

func loadTasks(ctx context.Context, cfg *config.LoadedConfig, includeCompleted bool) ([]tasks.Task, error) {
	allTasks, err := tasks.ReadTasks(ctx, cfg.TodoPath)
	if err != nil {
		return nil, err
	}

	if includeCompleted {
		return orderTasksBySection(allTasks), nil
	}

	completed := false
	return orderTasksBySection(tasks.FilterTasks(allTasks, &completed, "", "")), nil
}

func loadVisibleTaskList(ctx context.Context, cfg *config.LoadedConfig) ([]tasks.Task, error) {
	return loadTasks(ctx, cfg, false)
}

func orderTasksBySection(taskList []tasks.Task) []tasks.Task {
	if len(taskList) == 0 {
		return taskList
	}

	groups := make(map[string][]tasks.Task)
	firstSeen := make(map[string]int)
	dateSections := make(map[string]time.Time)

	for i, task := range taskList {
		section := task.Section
		if section == "" {
			section = "Tasks"
		}
		groups[section] = append(groups[section], task)
		if _, ok := firstSeen[section]; !ok {
			firstSeen[section] = i
		}
		if isDateSection(section) {
			if parsed, err := time.Parse("2006-01-02", section); err == nil {
				dateSections[section] = parsed
			}
		}
	}

	sectionNames := make([]string, 0, len(groups))
	for name := range groups {
		sectionNames = append(sectionNames, name)
	}

	sort.SliceStable(sectionNames, func(i, j int) bool {
		left, right := sectionNames[i], sectionNames[j]
		leftIsDate := isDateSection(left)
		rightIsDate := isDateSection(right)

		if leftIsDate && rightIsDate {
			return dateSections[left].After(dateSections[right])
		}
		if leftIsDate != rightIsDate {
			return leftIsDate
		}
		return firstSeen[left] < firstSeen[right]
	})

	ordered := make([]tasks.Task, 0, len(taskList))
	for _, section := range sectionNames {
		ordered = append(ordered, groups[section]...)
	}

	return ordered
}

func isDateSection(section string) bool {
	_, err := time.Parse("2006-01-02", section)
	return err == nil
}

func stripTaskTags(text string) string {
	cleaned := taskTagRegex.ReplaceAllString(text, "")
	return strings.Join(strings.Fields(cleaned), " ")
}

func taskDate(task tasks.Task) (time.Time, bool) {
	if task.CompletedDate != "" {
		if parsed, err := time.Parse("2006-01-02", task.CompletedDate); err == nil {
			return parsed, true
		}
	}
	if isDateSection(task.Section) {
		if parsed, err := time.Parse("2006-01-02", task.Section); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func displayTaskLine(task tasks.Task) string {
	status := "○"
	if task.Completed {
		status = "✓"
	}

	text := stripTaskTags(strings.TrimSpace(tasks.StripCompletedTag(tasks.StripTaskID(task.Text))))
	parts := []string{status}
	if task.Priority != "" {
		parts = append(parts, fmt.Sprintf("[%s]", task.Priority))
	}
	if text != "" {
		parts = append(parts, text)
	}
	if len(task.Tags) > 0 {
		tags := make([]string, 0, len(task.Tags))
		for _, tag := range task.Tags {
			tags = append(tags, "#"+tag)
		}
		parts = append(parts, fmt.Sprintf("tags: %s", strings.Join(tags, " ")))
	}
	if task.CompletedDate != "" {
		parts = append(parts, fmt.Sprintf("done: %s", task.CompletedDate))
	}

	return strings.Join(parts, "  ")
}

func showTaskList(ctx context.Context, cfg *config.LoadedConfig, includeCompleted bool) error {
	taskList, err := loadTasks(ctx, cfg, includeCompleted)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	if len(taskList) == 0 {
		fmt.Println("No tasks to show")
		return nil
	}

	currentSection := ""
	for i, task := range taskList {
		section := task.Section
		if section == "" {
			section = "Tasks"
		}
		if section != currentSection {
			if i > 0 {
				fmt.Println()
			}
			currentSection = section
			if isDateSection(section) {
				fmt.Printf("## %s\n", section)
			} else {
				fmt.Printf("## %s\n", section)
			}
		}
		fmt.Printf("%d. %s\n", i+1, displayTaskLine(task))
	}

	return nil
}

func addTask(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	text := strings.TrimSpace(strings.Join(args, " "))
	if text == "" {
		if utils.IsReplMode() {
			return fmt.Errorf("task text must be provided in REPL mode")
		}
		value, err := utils.PromptUserRequired("Task text: ")
		if err != nil {
			return err
		}
		text = strings.TrimSpace(value)
	}
	if text == "" {
		return fmt.Errorf("task text is required")
	}

	section, err := promptOptional("Section (press enter for Tasks): ")
	if err != nil {
		return err
	}
	priority, err := promptOptional("Priority (P0-P3, press enter to skip): ")
	if err != nil {
		return err
	}
	tagsRaw, err := promptOptional("Tags (comma-separated, press enter to skip): ")
	if err != nil {
		return err
	}
	var tags []string
	if tagsRaw != "" {
		for _, part := range strings.Split(tagsRaw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	taskService := services.NewTaskService()
	result, err := taskService.CreateTask(ctx, services.CreateTaskOptions{
		DiaryPath: cfg.DiaryPath,
		TodoPath:  cfg.TodoPath,
		StatePath: cfg.StatePath,
		Text:      text,
		Section:   section,
		Priority:  strings.TrimSpace(priority),
		Tags:      tags,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Added task %s\n", result.TaskID)
	return nil
}

func promptOptional(prompt string) (string, error) {
	if utils.IsReplMode() {
		return "", nil
	}
	return utils.PromptUser(prompt)
}

func promptForSelection(prompt string) ([]int, error) {
	input, err := utils.PromptUserRequired(prompt)
	if err != nil {
		return nil, err
	}
	return parseSelection(input)
}

func parseSelection(input string) ([]int, error) {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == ';'
	})
	if len(fields) == 0 {
		return nil, fmt.Errorf("no task numbers provided")
	}

	seen := make(map[int]bool)
	var selection []int
	for _, field := range fields {
		if field == "" {
			continue
		}
		number, err := strconv.Atoi(field)
		if err != nil || number <= 0 {
			return nil, fmt.Errorf("invalid task number %q", field)
		}
		if !seen[number] {
			seen[number] = true
			selection = append(selection, number)
		}
	}

	sort.Ints(selection)
	return selection, nil
}

func resolveSelection(args []string, prompt string) ([]int, error) {
	if len(args) > 0 {
		return parseSelection(strings.Join(args, " "))
	}

	if utils.IsReplMode() {
		return nil, fmt.Errorf("task numbers are required in REPL mode")
	}

	return promptForSelection(prompt)
}

func completeTasks(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	taskList, err := loadTasks(ctx, cfg, false)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}
	if len(taskList) == 0 {
		fmt.Println("No pending tasks to complete")
		return nil
	}

	selection, err := resolveSelection(args, "Task numbers to complete: ")
	if err != nil {
		if len(args) == 0 {
			if err := showTaskList(ctx, cfg, false); err != nil {
				return err
			}
		}
		return err
	}

	if len(args) == 0 {
		if err := showTaskList(ctx, cfg, false); err != nil {
			return err
		}
	}

	selected := make(map[int]bool)
	selectedTasks := make(map[int]tasks.Task, len(selection))
	for _, n := range selection {
		if n < 1 || n > len(taskList) {
			return fmt.Errorf("task number %d is out of range", n)
		}
		selected[n] = true
		picked := taskList[n-1]
		if picked.ID == "" {
			picked.ID = tasks.GenerateTaskID(picked.Text)
		}
		picked.Completed = true
		if picked.CompletedDate == "" {
			picked.CompletedDate = time.Now().Format("2006-01-02")
		}
		selectedTasks[n] = picked
	}

	content, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	completedDate := time.Now().Format("2006-01-02")

	for i, task := range taskList {
		if !selected[i+1] {
			continue
		}
		lineIdx := task.Line - 1
		if lineIdx < 0 || lineIdx >= len(lines) {
			continue
		}
		lines[lineIdx] = markTaskLineComplete(lines[lineIdx], selectedTasks[i+1], completedDate)
	}

	updated := strings.Join(lines, "\n")
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}

	if err := notes.WriteNote(ctx, cfg.TodoPath, updated); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	if err := markTasksCompletedInState(cfg.StatePath, selectedTasks); err != nil {
		return err
	}

	fmt.Printf("Completed %d task(s)\n", len(selected))
	return nil
}

func markTaskLineComplete(line string, task tasks.Task, completedDate string) string {
	matches := taskBulletRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return line
	}

	prefix := matches[1]
	body := strings.TrimSpace(matches[3])
	body = strings.TrimSpace(tasks.StripCompletedTag(tasks.StripTaskID(body)))
	if body == "" {
		body = strings.TrimSpace(matches[3])
	}

	if task.ID != "" {
		body = strings.TrimSpace(body + fmt.Sprintf(" <!-- id: %s -->", task.ID))
	}

	body = strings.TrimSpace(tasks.StripCompletedTag(body))
	return fmt.Sprintf("%s[x] %s @completed(%s)", prefix, body, completedDate)
}

func editTask(ctx context.Context, cfg *config.LoadedConfig, args []string) error {
	taskList, err := loadTasks(ctx, cfg, true)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}
	if len(taskList) == 0 {
		fmt.Println("No tasks to edit")
		return nil
	}

	selection, err := resolveSelection(args, "Task number to edit: ")
	if err != nil {
		if len(args) == 0 {
			if err := showTaskList(ctx, cfg, true); err != nil {
				return err
			}
		}
		return err
	}
	if len(selection) == 0 {
		return fmt.Errorf("no task number selected")
	}
	if len(selection) > 1 {
		return fmt.Errorf("edit expects a single task number")
	}

	n := selection[0]
	if n < 1 || n > len(taskList) {
		return fmt.Errorf("task number %d is out of range", n)
	}

	selected := taskList[n-1]
	fmt.Printf("Editing task %d: %s\n", n, displayTaskLine(selected))
	editPath := cfg.TodoPath
	stateData, err := state.Read(cfg.StatePath)
	if err == nil {
		if taskState, ok := stateData.Tasks[selected.ID]; ok {
			if taskState.Source != "" && taskState.Source != "todo-list" && taskState.Source != "merged" {
				editPath = taskState.Source
			}
		}
	}
	return notes.OpenInEditorWithContext(ctx, editPath)
}

func showMonthlyTaskStats(ctx context.Context, cfg *config.LoadedConfig) error {
	taskList, err := loadTasks(ctx, cfg, true)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	now := time.Now()
	completedThisMonth := 0
	for _, task := range taskList {
		if !task.Completed {
			continue
		}
		if task.CompletedDate == "" {
			continue
		}
		if date, err := time.Parse("2006-01-02", task.CompletedDate); err == nil {
			if date.Year() == now.Year() && date.Month() == now.Month() {
				completedThisMonth++
			}
		}
	}

	fmt.Printf("Completed this month: %d\n", completedThisMonth)
	return nil
}

func archiveMonthlyTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	content, err := os.ReadFile(cfg.TodoPath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}

	taskList := tasks.ParseTasks(string(content))
	if len(taskList) == 0 {
		fmt.Println("No tasks to archive")
		return nil
	}

	currentMonthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var archived []tasks.Task
	activeLines := strings.Split(string(content), "\n")
	archivedLineSet := make(map[int]bool)

	for _, task := range taskList {
		if !task.Completed {
			continue
		}
		date, ok := taskDate(task)
		if !ok || !date.Before(currentMonthStart) {
			continue
		}
		archived = append(archived, task)
		lineIdx := task.Line - 1
		if lineIdx >= 0 && lineIdx < len(activeLines) {
			archivedLineSet[lineIdx] = true
		}
	}

	if len(archived) == 0 {
		fmt.Println("No completed tasks from prior months to archive")
		return nil
	}

	stateData, err := state.Read(cfg.StatePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var remaining []string
	for i, line := range activeLines {
		if archivedLineSet[i] {
			continue
		}
		remaining = append(remaining, line)
	}
	updatedTodo := strings.Join(remaining, "\n")
	updatedTodo = strings.TrimRight(updatedTodo, "\n") + "\n"
	if err := notes.WriteNote(ctx, cfg.TodoPath, updatedTodo); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	for _, task := range archived {
		if task.ID != "" {
			stateData.RemoveTask(task.ID)
		}
	}
	stateData.MarkArchived()
	if err := stateData.Write(cfg.StatePath); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	now := time.Now()
	archiveDir := filepath.Join(cfg.Paths.BaseDir, "Archive")
	if err := notes.EnsureDir(archiveDir); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}
	archivePath := filepath.Join(archiveDir, fmt.Sprintf("archive-%s.md", now.Format("2006-01")))

	archiveContent := ""
	if utils.FileExists(archivePath) {
		data, err := os.ReadFile(archivePath)
		if err != nil {
			return fmt.Errorf("failed to read archive file: %w", err)
		}
		archiveContent = string(data)
	} else {
		archiveContent = fmt.Sprintf("# Archive - %s\n\n", now.Format("January 2006"))
	}

	archiveContent += fmt.Sprintf("\n## Archived on %s\n\n", now.Format("2006-01-02"))
	for _, task := range archived {
		archiveContent += formatArchivedTaskLine(task) + "\n"
	}
	archiveContent = strings.TrimRight(archiveContent, "\n") + "\n"

	if err := notes.WriteNote(ctx, archivePath, archiveContent); err != nil {
		return fmt.Errorf("failed to write archive file: %w", err)
	}

	fmt.Printf("Archived %d task(s) to %s\n", len(archived), archivePath)
	return nil
}

func pruneTasks(ctx context.Context, cfg *config.LoadedConfig) error {
	taskService := services.NewTaskService()
	result, err := taskService.PruneTasks(ctx, services.PruneOptions{
		DiaryPath: cfg.DiaryPath,
		TodoPath:  cfg.TodoPath,
		StatePath: cfg.StatePath,
		DryRun:    taskPruneDryRun,
	})
	if err != nil {
		return err
	}

	if taskPruneDryRun {
		fmt.Printf("Would remove %d task(s)\n", result.RemovedCount)
		printPruneDetails(result)
		return nil
	}

	fmt.Printf("Pruned %d task(s)\n", result.RemovedCount)
	printPruneDetails(result)
	return nil
}

func printPruneDetails(result *services.PruneResult) {
	if len(result.OrphanedTasks) > 0 {
		fmt.Println("  Orphaned:")
		for _, task := range result.OrphanedTasks {
			fmt.Printf("    - %s (%s)\n", task.ID, task.Text)
		}
	}
	if len(result.DuplicateTasks) > 0 {
		fmt.Println("  Duplicates:")
		for _, task := range result.DuplicateTasks {
			fmt.Printf("    - %s (%s)\n", task.ID, task.Text)
		}
	}
}

func formatArchivedTaskLine(task tasks.Task) string {
	text := strings.TrimSpace(tasks.StripCompletedTag(tasks.StripTaskID(task.Text)))
	line := fmt.Sprintf("- [x] %s", text)
	if task.ID != "" {
		line += fmt.Sprintf(" <!-- id: %s -->", task.ID)
	}
	if task.CompletedDate != "" {
		line += fmt.Sprintf(" @completed(%s)", task.CompletedDate)
	}
	return line
}

func markTasksCompletedInState(statePath string, selectedTasks map[int]tasks.Task) error {
	stateData, err := state.Read(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	for _, task := range selectedTasks {
		completedTask := task
		completedTask.Completed = true
		if completedTask.CompletedDate == "" {
			completedTask.CompletedDate = time.Now().Format("2006-01-02")
		}
		stateData.AddTask(completedTask, "todo-list")
	}

	if statePath != "" {
		if err := stateData.Write(statePath); err != nil {
			return fmt.Errorf("failed to write state file: %w", err)
		}
	}

	return nil
}
