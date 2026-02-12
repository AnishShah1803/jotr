# Plan 02: ID-Based Task Tracking (Not Date-Section Based)

## Problem
Currently tasks are tracked by their date section in the todo list. If a task from 2026-02-01 is completed on 2026-02-06, it stays in the 2026-02-01 section.

## Desired Behavior
- Track tasks by their unique ID across all files
- When completed, move task to the completion date's section in todo list
- Update the original daily note with completion status
- Add `@completed(YYYY-MM-DD)` tag to the task

## Example Flow

```markdown
# Daily Note 2026-02-01.md
## Tasks
- [ ] Review project proposal <!-- id: abc123 -->
```

User completes it in todo.md on 2026-02-06:

```markdown
# todo.md
## 2026-02-06
- [x] Review project proposal <!-- id: abc123 -->

## 2026-02-01
[Task no longer here - moved to completion date]
```

```markdown
# Daily Note 2026-02-01.md
## Tasks
- [x] Review project proposal <!-- id: abc123 --> @completed(2026-02-06)
```

---

## Implementation Tasks

### Phase 1: State Schema — Add `CreatedDate` and `CompletedDate` fields

> The `TaskState` struct in `internal/state/state.go` already has `Source string` (daily note path), `CreatedAt time.Time`, and `CompletedAt time.Time`. We need to add string date fields that are human-readable and used for section placement.

- [x] **1.1** In `internal/state/state.go`, add `CreatedDate string` field to `TaskState` struct (JSON tag: `"createdDate,omitempty"`). This stores `"2026-02-01"` format — the date section the task was originally created under.
- [x] **1.2** In `internal/state/state.go`, add `CompletedDate string` field to `TaskState` struct (JSON tag: `"completedDate,omitempty"`). This stores `"2026-02-06"` format — the date section the task should move to upon completion.
- [x] **1.3** In `state.AddTask()` (`internal/state/state.go` line ~81), when a task is first added (the `else` branch at line ~98), populate `CreatedDate` from the task's `Section` field if it matches `YYYY-MM-DD` format, otherwise set to `time.Now().Format("2006-01-02")`.
- [x] **1.4** In `state.AddTask()`, when a task transitions to completed (`task.Completed && ts.CompletedAt.IsZero()` at line ~102), also set `CompletedDate` to `time.Now().Format("2006-01-02")`.
- [x] **1.5** In `state.AddTask()`, when updating an existing task (the `if existing, ok` branch at line ~95), preserve `CreatedDate` and `CompletedDate` from the existing state.
- [x] **1.6** Verify: run `go build ./...` — must compile cleanly.

### Phase 2: Task Section Reassignment on Completion

> When `writeTodoFileFromState()` in `internal/services/task_service.go` writes the todo file, completed tasks should appear under their `CompletedDate` section, not their original `Section`/`CreatedDate` section.

- [x] **2.1** In `TaskService.writeTodoFileFromState()` (`internal/services/task_service.go` line ~305), modify the section-grouping logic: when building the `sections` map, if a task is completed AND has a non-empty `CompletedDate`, use `CompletedDate` as the section key instead of `task.Section`.
- [x] **2.2** Ensure the existing date-based sort logic (lines ~341-356) still works — `CompletedDate` values are `YYYY-MM-DD` strings and will naturally sort correctly with the existing `dateRegex` comparator.
- [x] **2.3** Verify: run `go build ./...` — must compile cleanly.

### Phase 3: Write `@completed(date)` Tag Back to Daily Notes

> When a task is completed (detected during sync), `updateDailyNoteFromState()` must append `@completed(YYYY-MM-DD)` to the task line in the originating daily note.

- [x] **3.1** In `TaskService.formatTaskLine()` (`internal/services/task_service.go` line ~287), if `stateTask.Completed` is true AND `stateTask.CompletedDate` is non-empty, append ` @completed(YYYY-MM-DD)` after the ID comment but before the newline. Example output: `- [x] Review project proposal <!-- id: abc123 --> @completed(2026-02-06)`.
- [x] **3.2** In `tasks.StripTaskID()` or a new companion function `StripCompletedTag()` in `internal/tasks/tasks.go`, add a function to strip `@completed(YYYY-MM-DD)` from task text for clean display. Pattern: `\s*@completed\(\d{4}-\d{2}-\d{2}\)`.
- [x] **3.3** In `tasks.ParseTasks()` (`internal/tasks/tasks.go` line ~31), extract `@completed(date)` tag during parsing so it's not treated as part of the task text. Store it somewhere accessible (either strip it from `Text` or add a field to `Task`).
- [x] **3.4** Verify: run `go build ./...` — must compile cleanly.

### Phase 4: Handle Completion Detected from Todo List (todo.md → daily note)

> When `BidirectionalSync` detects a task was completed in todo.md, we need to propagate the completion back to the originating daily note, not just today's daily note.

- [ ] **4.1** In `state.BidirectionalSync()` (`internal/state/state.go` line ~378), when applying a `todoChange` where the task transitions from incomplete→complete, set `CompletedDate` on the `NewTask` to `time.Now().Format("2006-01-02")`.
- [ ] **4.2** In `TaskService.SyncTasks()` (`internal/services/task_service.go` line ~108), after sync completes, if `syncResult.DailyChanged` is true, identify which daily notes need updating by examining state tasks that changed. For each changed task, look up `Source` (the original daily note path) and update that file — not just today's note.
- [ ] **4.3** Refactor `updateDailyNoteFromState()` to accept a specific file path and a filtered list of tasks relevant to that file (currently it only updates `notePath` which is today's daily note). It may need to be called multiple times — once per distinct `Source` path among changed tasks.
- [ ] **4.4** Verify: run `go build ./...` — must compile cleanly.

### Phase 5: Handle Completion Detected from Daily Note (daily note → todo.md)

> When a task is completed directly in a daily note, sync should propagate that to todo.md and move it to the completion date section.

- [ ] **5.1** Confirm that the existing `CompareWithDailyNotes()` already detects completion changes (it compares `stateTask.Completed` vs `sourceTask.Completed` in `isTaskModified()` at line ~303). If a task goes `false→true`, it should be detected as `Modified`. **No code change expected — just verify this path works.**
- [ ] **5.2** In `state.applyChange()` (`internal/state/state.go` line ~442), when applying a change where `NewTask.Completed` is true and the existing task was not completed, set `CompletedDate` to `time.Now().Format("2006-01-02")`.
- [ ] **5.3** Verify that `writeTodoFileFromState()` (modified in Phase 2) correctly places the newly-completed task under the `CompletedDate` section in todo.md.
- [ ] **5.4** Verify: run `go build ./...` — must compile cleanly.

### Phase 6: Preserve Task ID Through All Write Operations

> Currently `writeTodoFileFromState()` writes `- [x] {text}` without the ID comment. It must include the ID.

- [ ] **6.1** In `TaskService.writeTodoFileFromState()` (`internal/services/task_service.go` line ~365), change the task line format from `fmt.Sprintf("- %s %s\n", checkbox, task.Text)` to include the task ID: `fmt.Sprintf("- %s %s <!-- id: %s -->\n", checkbox, task.Text, task.ID)`. Only append the ID comment if `task.ID != ""`.
- [ ] **6.2** Verify that `tasks.ParseTasks()` correctly re-extracts the ID on the next read (it uses `ExtractTaskID` which matches `<!-- id: [a-f0-9]{8} -->`). Confirm the round-trip works: write→read→write produces stable output.
- [ ] **6.3** Verify: run `go build ./...` — must compile cleanly.

### Phase 7: Tests

- [ ] **7.1** In `internal/state/bidirectional_sync_test.go`, add a test: task created under section `"2026-02-01"`, completed via todo change → verify `CompletedDate` is set, `CreatedDate` is `"2026-02-01"`.
- [ ] **7.2** Add a test: task completed in daily note → verify `CompletedDate` is set on the state task after `BidirectionalSync`.
- [ ] **7.3** Add a test for `writeTodoFileFromState()`: given a state with a task that has `Section="2026-02-01"` and `CompletedDate="2026-02-06"` and `Completed=true`, verify the output places the task under the `## 2026-02-06` section, not `## 2026-02-01`.
- [ ] **7.4** Add a test for `formatTaskLine()`: verify completed task with `CompletedDate` produces `@completed(YYYY-MM-DD)` tag.
- [ ] **7.5** Add a test: round-trip test — write todo file with IDs, read it back, verify IDs are preserved.
- [ ] **7.6** Run full test suite: `go test ./...` — all tests must pass.

### Phase 8: Build Verification

- [ ] **8.1** Run `make dev` — must succeed.
- [ ] **8.2** Manual smoke test: create a daily note with a task, run `jotr sync`, verify task appears in todo.md with ID. Mark it complete in todo.md, run `jotr sync` again, verify it moves to today's date section and the original daily note gets `@completed(date)` tag.
