# Plan 01: BidirectionalSync File Locking

## Context

The `BidirectionalSync` feature exists in `internal/state/state.go:376`. It performs bidirectional synchronization between:
- Daily notes (source of truth for new tasks)
- Todo list (aggregated view)
- State file (`.todo_state.json` - persistence layer)

**Current Flow:**
1. Read state file
2. Read daily note
3. Read todo file
4. Compute `BidirectionalSync` (detects conflicts, merges changes)
5. Write state file
6. Write todo file
7. Write daily note (with lock at line 194-198)

**Current Locking Status:**
- ✅ Daily note: Locked during write (`updateDailyNoteFromState`)
- ❌ State file: No locking
- ❌ Todo file: No locking during sync (only in `ArchiveTasks`)
- ❌ Multi-file atomicity: Not guaranteed

## The Problem

Two concurrent `jotr sync` operations:
```
Process A              Process B
   |                       |
   |-- Read state --------->|
   |                       |-- Read state (same data)
   |                       |
   |-- Compute changes     |-- Compute changes
   |                       |
   |-- Write state         |-- Write state (overwrites A's changes!)
   |                       |
   Result: A's changes are lost
```

## Implementation Tasks

### Phase 1: Add Locking to SyncTasks Operation

- [x] Add file locking at the start of `SyncTasks` in `internal/services/task_service.go`
  - Lock order: state file → todo file → daily note
  - Use 10-second timeout (consistent with existing `LockFile` usage)
  - Release all locks in reverse order using `defer`
  - Location: Before line 83 where `state.Read()` is called

- [x] Add lock acquisition helper function
  - Create `acquireSyncLocks(statePath, todoPath, notePath string) (locks []fileLock, error)`
  - Ensure consistent lock ordering to prevent deadlocks
  - Handle partial lock failure (release already-acquired locks on error)

### Phase 2: Atomic Multi-File Updates

- [x] Ensure atomic writes across all three files
  - State file write (line 108-112) should be inside the lock
  - Todo file write (line 115-119) should be inside the lock
  - Daily note write (line 121-125) already has lock, verify it uses the same lock

- [x] Add rollback mechanism on write failure
  - If any write fails, log the error but don't leave partial state
  - Consider the operation failed if any file can't be updated

### Phase 3: Handle Lock Timeout Gracefully

- [x] Add user-friendly error messages for lock timeouts
  - "Another sync operation is in progress. Please try again in a few seconds."
  - Log the actual error for debugging

 - [x] Add retry logic (optional enhancement)
  - Exponential backoff with max 3 retries
  - Only retry on lock timeout, not on other errors

### Phase 4: Testing

- [x] Add concurrent sync test in `bidirectional_sync_test.go`
  - Launch 2 goroutines calling `SyncTasks` simultaneously
  - Verify no data loss or corruption
  - Verify both tasks end up in state/todo files

- [x] Add lock timeout test
  - Verify proper error handling when lock can't be acquired

- [x] Verify lock ordering doesn't cause deadlocks
  - Test with rapid concurrent sync calls

## Lock Order (Critical)

Always acquire locks in this order to prevent deadlocks:
1. State file (`.todo_state.json.lock`)
2. Todo file (`todo.md.lock`)
3. Daily note file (`2026-02-11-Tue.md.lock`)

Always release in reverse order.

## Files to Modify

- `internal/services/task_service.go` - Add locking to `SyncTasks`
- `internal/services/task_service.go` - Add helper functions
- `internal/state/bidirectional_sync_test.go` - Add concurrent tests

## Reference

Existing lock usage:
- `task_service.go:194-198` - `updateDailyNoteFromState` locks daily note
- `task_service.go:365-369` - `ArchiveTasks` locks todo file
- `internal/utils/fileutils.go` - `LockFile`, `TryLockFile`, `UnlockFile` implementations
