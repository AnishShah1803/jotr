---
name: tui-transitions
description: Managing async state transitions (Loading -> Data -> Error) to prevent UI flickering.
license: MIT
compatibility: opencode
---

## Best Practices
1. **The Spinner Pattern:** While a `tea.Cmd` is running, show a `spinner.Model`. Don't just show a blank screen.
2. **Clear on Success:** When data arrives, explicitly reset any "loading" or "error" flags.
3. **Debouncing:** For "search-as-you-type" features, use `tea.Tick` (from `github.com/charmbracelet/bubbletea`) for debouncing instead of `time.Tick`, which doesn't work well with Bubble Tea's event loop.
4. **Error Recovery:** Always provide a "Retry" keybind (e.g., `r`) when an async command fails.

## Example
```go
func (m model) View() string {
    if m.err != nil {
        return errorStyle.Render("Error: " + m.err.Error() + "\nPress 'r' to retry")
    }
    if m.loading {
        return m.spinner.View() + " Fetching data..."
    }
    return m.mainContent()
}
