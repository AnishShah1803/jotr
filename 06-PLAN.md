# Plan 06: Detailed Sync Reporting

## Problem
Current sync only reports basic counts. Users need visibility into what changed.

## Desired Behavior
Show detailed report of sync operations:
- What tasks were added/updated/deleted
- Where changes came from
- Where changes were applied
- Conflicts detected

## Report Format

### Success Output
```
✓ Sync completed successfully

From Daily Notes:
  + Added: "Review project proposal" (id: abc123)
  ~ Updated: "Call mom" - marked complete (id: def456)

From Todo List:
  ~ Updated: "Buy groceries" - changed priority to P1 (id: ghi789)

Conflicts (not synced):
  ! "Write report" - Text differs:
      Daily: "Write Q4 report"
      Todo:  "Write quarterly report"

Summary:
  4 tasks checked
  2 tasks added/updated from daily notes
  1 task updated from todo list
  1 conflict requires manual resolution
```

### Dry Run Output
```
⚠ DRY RUN - No changes made

Planned Changes:
  From Daily Notes → Todo:
    + Add "Review project proposal"
    ~ Update "Call mom" (mark complete)
  
  From Todo → Daily Notes:
    ~ Update "Buy groceries" (priority P1)
```

---

## Implementation Tasks

### Phase 1: Data Structures

- [x] **1.1** Create new `TaskChangeDetail` type in `internal/state/state.go`:
  ```go
  type TaskChangeDetail struct {
      ID      string
      Text    string
      Change  string // "added", "updated", "deleted"
      From    string // previous value (for updates)
      To      string // new value (for updates)
      Details string // "marked complete", "priority changed to P1", etc.
  }
  ```

- [x] **1.2** Create new `ConflictDetail` type in `internal/state/state.go`:
  ```go
  type ConflictDetail struct {
      ID        string
      TextDaily string
      TextTodo  string
      Reason    string
  }
  ```

- [x] **1.3** Extend `state.SyncResult` (line 430 in `internal/state/state.go`) with detailed tracking fields:
  ```go
  type SyncResult struct {
      // ... existing fields ...
      
      // Detailed change tracking
      AddedFromDaily   []TaskChangeDetail
      UpdatedFromDaily []TaskChangeDetail
      AddedFromTodo    []TaskChangeDetail
      UpdatedFromTodo  []TaskChangeDetail
      DeletedTasks     []TaskChangeDetail
      ConflictsDetail  []ConflictDetail
  }
  ```

- [x] **1.4** Extend `services.SyncResult` (line 38 in `internal/services/task_service.go`) with same detailed fields to expose to CLI layer (renamed `DeletedTasks` to `DeletedTasksDetail` to avoid conflict with existing int field)

### Phase 2: BidirectionalSync Implementation

- [x] **2.1** Update `TodoState.BidirectionalSync()` (line 445 in `internal/state/state.go`) to populate the new detailed tracking fields:
  - Track when tasks are added from daily notes → add to `AddedFromDaily`
  - Track when tasks are updated from daily notes → add to `UpdatedFromDaily`
  - Track when tasks are added from todo list → add to `AddedFromTodo`
  - Track when tasks are updated from todo list → add to `UpdatedFromTodo`
  - Track deleted tasks → add to `DeletedTasks`
  - Populate `ConflictsDetail` with full conflict information

- [ ] **2.2** Update `TaskService.SyncTasks()` (line 111 in `internal/services/task_service.go`) to map detailed fields from state result to service result

### Phase 3: CLI Flags

- [ ] **3.1** Wire up `--dry-run` flag to sync command in `cmd/task/sync.go`:
  - Add `AddDryRunFlag(SyncCmd)` in init function
  - Retrieve flag in `syncTasks()` function
  - Pass dry-run option to `TaskService.SyncTasks()` (may need to add to `SyncOptions`)

- [ ] **3.2** Add `--quiet` flag to sync command (use existing `OutputOption` pattern from `internal/options/options.go` or add directly):
  ```go
  syncCmd.Flags().Bool("quiet", false, "Suppress normal output")
  ```

- [ ] **3.3** Add `--json` flag to sync command:
  ```go
  syncCmd.Flags().Bool("json", false, "Output in JSON format")
  ```

- [ ] **3.4** Add `--verbose` flag support (may already exist via AddVerboseFlag):
  - Add `AddVerboseFlag(SyncCmd)` if needed for verbose sync output

- [ ] **3.5** Update `services.SyncOptions` to include dry-run, quiet, json, verbose fields if needed

### Phase 4: Report Formatter

- [ ] **4.1** Create `FormatSyncReport()` function in `cmd/task/sync.go` or new file:
  - Takes `services.SyncResult` and output options
  - Returns formatted string based on output mode (default, quiet, json, verbose)

- [ ] **4.2** Implement default output format with:
  - Task additions (prefix `+`)
  - Task updates with details (prefix `~`)
  - Task deletions (prefix `-`)
  - Conflicts (prefix `!`)
  - Summary section with counts

- [ ] **4.3** Implement quiet mode: show only summary counts

- [ ] **4.4** Implement JSON mode: output complete result as JSON using `json.MarshalIndent(result, "", "  ")`

- [ ] **4.5** Implement verbose mode: include unchanged tasks in output

### Phase 5: Color-Coded Output

- [ ] **5.1** Add color support (check if already exists in codebase):
  - Search for existing color library usage (e.g., "github.com/fatih/color" or similar)
  - If not exists, add dependency or use existing formatting

- [ ] **5.2** Apply colors:
  - Green (`+`) for added tasks
  - Yellow/Blue (`~`) for updated tasks
  - Red (`!`) for conflicts
  - Gray for deleted tasks

- [ ] **5.3** Make colors respect `--no-color` flag or auto-detect terminal (check existing pattern in codebase)

### Phase 6: Dry-Run Implementation

- [ ] **6.1** Implement dry-run logic in `TaskService.SyncTasks()`:
  - When dry-run is true, run sync but don't write any files
  - Return same result structure with all planned changes

- [ ] **6.2** Add dry-run header to output:
  ```
  ⚠ DRY RUN - No changes made
  ```

### Phase 7: Integration & Testing

- [ ] **7.1** Update `syncTasks()` in `cmd/task/sync.go` to:
  - Read all flags (dry-run, quiet, json, verbose)
  - Call `TaskService.SyncTasks()` with options
  - Call `FormatSyncReport()` with result and output options
  - Print formatted output

- [ ] **7.2** Test: sync with no changes → shows "Everything is in sync" message

- [ ] **7.3** Test: sync with daily note changes only → shows added/updated from daily notes

- [ ] **7.4** Test: sync with todo list changes only → shows added/updated from todo list

- [ ] **7.5** Test: sync with conflicts → shows conflict details

- [ ] **7.6** Test: sync with deletions → shows deleted tasks

- [ ] **7.7** Test: dry-run mode → shows planned changes without applying

- [ ] **7.8** Test: --json flag → outputs valid JSON

- [ ] **7.9** Test: --quiet flag → shows only summary

---

## Reference: Existing Code Patterns

### Flag Handling (from cmd/options.go)
```go
func AddDryRunFlag(cmd *cobra.Command) {
    cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
}
```

### OutputOption Pattern (from internal/options/options.go)
```go
type OutputOption struct {
    CountOnly bool
    FilesOnly bool
    PathOnly  bool
    Quiet     bool
    JSON      bool
}

func (o *OutputOption) AddFlags(cmd *cobra.Command) {
    cmd.Flags().BoolVar(&o.CountOnly, "count", false, "Show only the count of matches")
    cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Suppress normal output")
    cmd.Flags().BoolVar(&o.JSON, "json", false, "Output in JSON format")
}
```

### JSON Output Pattern (from various cmd files)
```go
data, err := json.MarshalIndent(result, "", "  ")
// use data...
```

### Current SyncResult (state - line 430)
```go
type SyncResult struct {
    StateUpdated   bool
    DailyChanged   bool
    TodoChanged    bool
    Conflicts      map[string]string
    AppliedDaily   int
    AppliedTodo    int
    Deleted        int
    Skipped        int
    ChangedTaskIDs []string
    DeletedTaskIDs []string
}
```

### Current SyncResult (services - line 38)
```go
type SyncResult struct {
    StatePath      string
    DailyPath      string
    TodoPath       string
    TasksRead      int
    TasksFromDaily int
    TasksFromTodo  int
    DeletedTasks   int
    DeletedTaskIDs []string
    Conflicts      map[string]string
}
```
