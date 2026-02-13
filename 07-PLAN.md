# Plan 07: Multi-Date Note Support

## Problem
Current sync only works with today's daily note.

## Desired Behavior
Sync tasks across ALL daily notes, not just today.

## Use Cases

1. **Complete overdue task:**
   - Task created in 2026-02-01 daily note
   - User completes it in todo list on 2026-02-06
   - System finds and updates the 2026-02-01 daily note

2. **Update task text:**
   - User edits task in 2026-02-03 daily note
   - Changes propagate to todo list
   - Other daily notes with same task (by ID) also updated

3. **Retroactive entry:**
   - User creates daily note for past date
   - Tasks sync bidirectionally with todo list

## Implementation

### Task Discovery

- [ ] Scan entire diary directory for all daily notes
- [ ] Parse each note and extract tasks with IDs
- [ ] Build ID → file path mapping in state

### Date Range Support

Allow specifying date range:
```bash
jotr sync                    # Sync today only (default)
jotr sync --date 2026-02-01  # Sync specific date
jotr sync --all              # Sync all dates
jotr sync --since 7d         # Sync last 7 days
```

### State Tracking

State should track task locations:
```json
{
  "tasks": {
    "abc123": {
      "id": "abc123",
      "text": "Review proposal",
      "source": "/path/to/2026-02-01-Sat.md",
      "createdDate": "2026-02-01"
    }
  }
}
```

### Implementation Tasks

1. [ ] Create `FindDailyNoteByDate()` function
2. [ ] Create `FindDailyNoteByTaskID()` function
3. [ ] Modify `BidirectionalSync()` to:
   - [ ] Accept date parameter (optional, default today)
   - [ ] Load tasks from specified daily note
   - [ ] When todo changes, find and update correct daily note by ID
4. [ ] Add CLI flags:
   - [ ] `--date YYYY-MM-DD` - Sync specific date
   - [ ] `--all` - Sync all dates
   - [ ] `--since Nd` - Sync last N days
5. [ ] Handle missing daily notes gracefully
6. [ ] Performance optimization for large diary directories

### Todo List Section Management

When task is completed:
1. [ ] Identify completion date
2. [ ] Move task from original date section to completion date section
3. [ ] If completion date section doesn't exist, create it
4. [ ] Maintain chronological order (newest first)

### Edge Cases

- [ ] Task exists in multiple daily notes (shouldn't happen with IDs, but handle it)
- [ ] Daily note file missing but referenced in state
- [ ] Date parsing errors
- [ ] Leap years and timezone handling
- [ ] Very large diary directories (1000+ notes)

### Performance Considerations

- [ ] Cache parsed daily notes during sync
- [ ] Only read files that have changed (check modification time)
- [ ] Batch file operations
- [ ] Consider indexing for very large diaries
