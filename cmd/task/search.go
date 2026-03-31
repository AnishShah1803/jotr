package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AnishShah1803/jotr/internal/config"
	"github.com/AnishShah1803/jotr/internal/state"
	"github.com/AnishShah1803/jotr/internal/tasks"
)

var taskSearchFilters []string

func init() {
	TaskSearchCmd.Flags().StringArrayVarP(
		&taskSearchFilters,
		"filter",
		"f",
		nil,
		"Repeatable filter in kind:value form (priority:P1, tag:ops, section:Backlog, status:pending|completed|all, source:Diary)",
	)
}

var TaskSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks",
	Long: `Search tasks by query text with optional repeatable filters.

Examples:
  jotr task search "release"
  jotr task search "api" --filter priority:P1 --filter tag:backend
  jotr task search "cleanup" -f status:completed -f source:Diary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadWithContext(cmd.Context(), "")
		if err != nil {
			return err
		}

		query, filters := resolveTaskSearchCriteria(args, taskSearchFilters)
		return RunTaskSearch(cmd.Context(), cfg, query, filters)
	},
}

func resolveTaskSearchCriteria(args, rawFilters []string) (string, []string) {
	filters := make([]string, 0, len(rawFilters)+len(args))
	filters = append(filters, rawFilters...)

	queryParts := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		candidate := strings.TrimSpace(args[i])
		if candidate == "" {
			continue
		}

		if strings.ContainsAny(candidate, " \t\n\r") {
			queryParts = append(queryParts, candidate)
			continue
		}

		kind, ok := taskSearchFilterKind(candidate)
		if ok && isSupportedTaskSearchFilterKind(kind) {
			filters = append(filters, candidate)
			continue
		}

		if isSupportedTaskSearchFilterKind(strings.ToLower(candidate)) && i+1 < len(args) {
			next := strings.TrimSpace(args[i+1])
			if next != "" && !strings.ContainsAny(next, " \t\n\r") {
				filters = append(filters, candidate+":"+next)
				i++
				continue
			}
		}

		queryParts = append(queryParts, candidate)
	}

	query := strings.TrimSpace(strings.Join(queryParts, " "))
	return query, filters
}

func taskSearchFilterKind(raw string) (string, bool) {
	sep := ":"
	if !strings.Contains(raw, sep) {
		if strings.Contains(raw, "=") {
			sep = "="
		} else {
			return "", false
		}
	}

	parts := strings.SplitN(raw, sep, 2)
	if len(parts) != 2 {
		return "", false
	}

	kind := strings.ToLower(strings.TrimSpace(parts[0]))
	if kind == "" {
		return "", false
	}

	return kind, true
}

func isSupportedTaskSearchFilterKind(kind string) bool {
	switch kind {
	case "priority", "p", "tag", "t", "section", "s", "source", "src", "status", "st":
		return true
	default:
		return false
	}
}

type taskSearchQueryFilters struct {
	priorities      map[string]bool
	tags            map[string]bool
	sections        map[string]bool
	sources         []string
	includePending  bool
	includeComplete bool
}

func RunTaskSearch(ctx context.Context, cfg *config.LoadedConfig, query string, rawFilters []string) error {
	query = strings.TrimSpace(query)

	parsedFilters, err := parseTaskSearchFilters(rawFilters)
	if err != nil {
		return err
	}

	if query == "" && !hasTaskSearchFilters(rawFilters) {
		return fmt.Errorf("search query or at least one filter is required")
	}

	taskList, err := loadTasks(ctx, cfg, true)
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	stateData, err := state.Read(cfg.StatePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	matched := make([]tasks.Task, 0, len(taskList))
	for _, task := range taskList {
		source := resolveTaskSource(task, stateData, cfg.TodoPath)
		if query != "" && !matchesTaskQuery(task, query) {
			continue
		}
		if !matchesTaskFilters(task, source, parsedFilters) {
			continue
		}
		matched = append(matched, task)
	}

	if len(matched) == 0 {
		fmt.Println("No matching tasks found")
		return nil
	}

	for i, task := range matched {
		source := resolveTaskSource(task, stateData, cfg.TodoPath)
		fmt.Printf("%d. %s\n", i+1, displaySearchTaskLine(task))

		contextParts := make([]string, 0, 3)
		if task.Section != "" {
			contextParts = append(contextParts, task.Section)
		}
		contextParts = append(contextParts, "source: "+renderSourceContext(cfg.Paths.BaseDir, source))
		if task.Line > 0 {
			contextParts = append(contextParts, fmt.Sprintf("L%d", task.Line))
		}

		fmt.Printf("   ↳ %s\n", strings.Join(contextParts, " · "))
	}

	return nil
}

func hasTaskSearchFilters(rawFilters []string) bool {
	for _, filter := range rawFilters {
		if strings.TrimSpace(filter) != "" {
			return true
		}
	}

	return false
}

func parseTaskSearchFilters(rawFilters []string) (taskSearchQueryFilters, error) {
	filters := taskSearchQueryFilters{
		priorities:      make(map[string]bool),
		tags:            make(map[string]bool),
		sections:        make(map[string]bool),
		sources:         nil,
		includePending:  true,
		includeComplete: false,
	}

	hasStatusFilter := false

	for _, filter := range rawFilters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}

		sep := ":"
		if !strings.Contains(filter, sep) {
			if strings.Contains(filter, "=") {
				sep = "="
			} else {
				return filters, fmt.Errorf("invalid filter %q: expected kind:value", filter)
			}
		}

		parts := strings.SplitN(filter, sep, 2)
		if len(parts) != 2 {
			return filters, fmt.Errorf("invalid filter %q: expected kind:value", filter)
		}

		kind := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if value == "" {
			return filters, fmt.Errorf("invalid filter %q: value cannot be empty", filter)
		}

		switch kind {
		case "priority", "p":
			filters.priorities[strings.ToUpper(normalizePriority(value))] = true
		case "tag", "t":
			filters.tags[strings.ToLower(strings.TrimPrefix(value, "#"))] = true
		case "section", "s":
			filters.sections[strings.ToLower(value)] = true
		case "source", "src":
			filters.sources = append(filters.sources, strings.ToLower(value))
		case "status", "st":
			hasStatusFilter = true
			if !applyStatusFilterValue(&filters, strings.ToLower(value)) {
				return filters, fmt.Errorf("invalid status filter %q: expected pending, completed, or all", value)
			}
		default:
			return filters, fmt.Errorf("unsupported filter kind %q", kind)
		}
	}

	if len(filters.sources) > 1 {
		sort.Strings(filters.sources)
	}

	if hasStatusFilter {
		if !filters.includePending && !filters.includeComplete {
			return filters, fmt.Errorf("status filter excludes all tasks")
		}
	}

	return filters, nil
}

func applyStatusFilterValue(filters *taskSearchQueryFilters, value string) bool {
	if filters == nil {
		return false
	}

	if filters.includePending && !filters.includeComplete {
		filters.includePending = false
		filters.includeComplete = false
	}

	switch value {
	case "pending", "open", "todo":
		filters.includePending = true
		return true
	case "completed", "done", "closed":
		filters.includeComplete = true
		return true
	case "all":
		filters.includePending = true
		filters.includeComplete = true
		return true
	default:
		return false
	}
}

func matchesTaskQuery(task tasks.Task, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return false
	}

	searchBlob := strings.ToLower(strings.Join([]string{
		task.Text,
		task.ID,
		task.Priority,
		task.Section,
		strings.Join(task.Tags, " "),
	}, " "))

	return strings.Contains(searchBlob, query)
}

func matchesTaskFilters(task tasks.Task, source string, filters taskSearchQueryFilters) bool {
	if task.Completed {
		if !filters.includeComplete {
			return false
		}
	} else if !filters.includePending {
		return false
	}

	if len(filters.priorities) > 0 {
		if !filters.priorities[strings.ToUpper(task.Priority)] {
			return false
		}
	}

	if len(filters.sections) > 0 {
		if !filters.sections[strings.ToLower(task.Section)] {
			return false
		}
	}

	if len(filters.tags) > 0 {
		tagMatch := false
		for _, tag := range task.Tags {
			if filters.tags[strings.ToLower(tag)] {
				tagMatch = true
				break
			}
		}
		if !tagMatch {
			return false
		}
	}

	if len(filters.sources) > 0 {
		sourceLower := strings.ToLower(source)
		matched := false
		for _, pattern := range filters.sources {
			if strings.Contains(sourceLower, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func resolveTaskSource(task tasks.Task, stateData *state.TodoState, todoPath string) string {
	if task.ID == "" || stateData == nil {
		return todoPath
	}

	taskState, ok := stateData.Tasks[task.ID]
	if !ok {
		return todoPath
	}

	if taskState.Source == "" || taskState.Source == "todo-list" || taskState.Source == "merged" {
		return todoPath
	}

	return taskState.Source
}

func renderSourceContext(baseDir, source string) string {
	if source == "" {
		return "unknown"
	}

	if filepath.IsAbs(source) {
		if rel, err := filepath.Rel(baseDir, source); err == nil {
			return rel
		}
	}

	return source
}
