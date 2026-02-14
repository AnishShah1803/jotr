---
name: tui-responsive-math
description: Best practices for calculating dynamic terminal dimensions and avoiding layout overflows.
---

## Rules of Arithmetic
1. **The Border Tax:** When a model has a border, you MUST subtract 2 from the width and 2 from the height (`width - 2`, `height - 2`) before passing those dimensions to child components (like viewports or text inputs).
2. **Viewport Safety:** Never set a Viewport height to a negative number. Always use `max(0, availableHeight)` to prevent panics during extreme terminal shrinking.
3. **Join Balance:** When using `lipgloss.JoinHorizontal`, ensure the total width of all joined elements does not exceed `msg.Width`. If it does, the terminal will wrap the line and break the UI.

## Example
```go
func (m *model) setDimensions(w, h int) {
    borderSize := 2
    m.contentWidth = w - borderSize
    m.contentHeight = h - borderSize - footerHeight
    
    // Always validate to prevent panics
    if m.contentHeight < 0 { m.contentHeight = 0 }
}
