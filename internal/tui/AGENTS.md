# TUI Dashboard (Bubble Tea)

**Framework:** Bubble Tea MVU + Lipgloss styling
**Files:** model.go, update.go, view.go

## ARCHITECTURE

```
Model (state) ──► Update(msg) ──► Model' ──► View() ──► string
     ▲                                        │
     └──────────── tea.Cmd ◄──────────────────┘
```

## KEY TYPES

| Type | File | Purpose |
|------|------|---------|
| `Model` | model.go | TUI state (notes, tasks, viewports, focus) |
| `keyMap` | model.go | Key bindings (Quit, Tab, Up, Down, Enter) |
| `panel` | model.go | Focus state (panelNotes, panelPreview, panelTasks, panelStats) |
| `tickMsg` | model.go | Timer for status message expiry |
| `dataLoadedMsg` | update.go | Async data load result |

## UPDATE FLOW

1. `tea.WindowSizeMsg` → Update dimensions, trigger `loadData()`
2. `tea.KeyMsg` → Navigation, refresh, open file
3. `dataLoadedMsg` → Populate notes/tasks/streak
4. `tickMsg` → Clear expired status messages

## VIEW RENDERING

**Layout (view.go):**
```
┌─────────────────────────────────┐
│ Header (ASCII art + status)     │
├─────────────────┬───────────────┤
│ Notes Panel     │ Preview Panel │
├─────────────────┼───────────────┤
│ Tasks Panel     │ Stats Panel   │
├─────────────────┴───────────────┤
│ Footer (help)                   │
└─────────────────────────────────┘
```

**Panel Cycle:** notes → preview → tasks → stats → notes

## ANTI-PATTERNS (CRITICAL)

| Rule | Violation | Fix |
|------|-----------|-----|
| No goroutine mutation | Direct field change from goroutine | Return `tea.Msg` via `tea.Cmd` |
| No style allocation in View | `lipgloss.NewStyle()` in View() | Define as package `var` |
| No I/O in Update/View | File read in Update() | Use `tea.Cmd` factory |

## SKILLS TO USE

| Skill | When |
|-------|------|
| `bubbletea-mvu` | State machine patterns, Update logic |
| `lipgloss-layout` | Styling, JoinHorizontal/Vertical |
| `tui-concurrency` | Async operations, tea.Cmd patterns |
| `tui-testing-debug` | Unit testing without terminal |

## TESTING

```go
// Use testhelpers for TUI tests
fs := testhelpers.NewTestFS(t)
cfg := testhelpers.NewConfigHelper(fs).CreateBasicConfig(t)
m := tui.NewModel(ctx, cfg)

// Simulate messages
m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
m, _ = m.Update(tickMsg{})
```

## DIMENSIONS

- Min for ASCII art: 50×40 (width×height)
- Header/footer height: 14 (large) / 3 (small)
- Panel padding: 4 chars horizontal
