---
name: bubbletea-mvu
description: Enforces Model-View-Update (MVU) patterns and state machine logic for Bubble Tea apps.
license: MIT
compatibility: opencode
---

## What I do
- Guide the implementation of the Elm Architecture in Go.
- Enforce the "State Machine" pattern for multi-step TUIs.
- Prevent common concurrency anti-patterns (e.g., manual goroutines instead of tea.Cmd).

## Best Practices
1. **Event Loop Speed:** Keep `Update()` and `View()` pure and fast. Offload ANY I/O (network, file system, database) to a `tea.Cmd`.
2. **State Machine:** For multi-stage apps, use an `enum` for `sessionState` and a `switch` in `Update` and `View` to delegate to sub-models.
3. **Avoid Manual Concurrency:** Never start a goroutine inside `Update`. Use `tea.Cmd` and send messages back to the loop.
4. **Window Resizing:** Always handle `tea.WindowSizeMsg` to dynamically update component dimensions.

## Example: State Delegation
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
    }
    // Delegate to active component
    switch m.state {
    case stateList:
        return m.updateList(msg)
    case stateForm:
        return m.updateForm(msg)
    }
    return m, nil
}
