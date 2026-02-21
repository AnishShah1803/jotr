package state

// SyncResult represents the result of a sync operation between daily notes and todo list.
// This is a unified type used across both the state package (for sync operations)
// and the services package (for external API responses).
type SyncResult struct {
	// File paths (populated by service layer)
	StatePath string `json:"state_path,omitempty"`
	DailyPath string `json:"daily_path,omitempty"`
	TodoPath  string `json:"todo_path,omitempty"`

	// Change tracking flags (set by state layer)
	StateUpdated bool `json:"state_updated,omitempty"`
	DailyChanged bool `json:"daily_changed,omitempty"`
	TodoChanged  bool `json:"todo_changed,omitempty"`

	// Counters (set by both layers)
	TasksRead      int `json:"tasks_read,omitempty"`
	TasksFromDaily int `json:"tasks_from_daily,omitempty"`
	TasksFromTodo  int `json:"tasks_from_todo,omitempty"`
	DeletedTasks   int `json:"deleted_tasks,omitempty"`
	AppliedDaily   int `json:"applied_daily,omitempty"`
	AppliedTodo    int `json:"applied_todo,omitempty"`
	Deleted        int `json:"deleted,omitempty"`
	Skipped        int `json:"skipped,omitempty"`

	// Task ID tracking
	DeletedTaskIDs []string `json:"deleted_task_ids,omitempty"`
	ChangedTaskIDs []string `json:"changed_task_ids,omitempty"`

	// Conflicts
	Conflicts       map[string]string `json:"conflicts,omitempty"`
	ConflictsDetail []ConflictDetail  `json:"conflicts_detail,omitempty"`

	// Detailed change tracking for CLI reporting
	AddedFromDaily     []TaskChangeDetail `json:"added_from_daily,omitempty"`
	UpdatedFromDaily   []TaskChangeDetail `json:"updated_from_daily,omitempty"`
	AddedFromTodo      []TaskChangeDetail `json:"added_from_todo,omitempty"`
	UpdatedFromTodo    []TaskChangeDetail `json:"updated_from_todo,omitempty"`
	DeletedTasksDetail []TaskChangeDetail `json:"deleted_tasks_detail,omitempty"`
}
