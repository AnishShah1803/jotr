---
name: lipgloss-layout
description: Best practices for styling TUIs with Lipgloss, focusing on layouts and adaptive colors.
license: MIT
compatibility: opencode
---

## What I do
- Manage complex TUI layouts using JoinHorizontal and JoinVertical.
- Enforce the use of AdaptiveColors for light/dark mode support.
- Calculate layout arithmetic safely to avoid terminal wrapping bugs.

## Best Practices
1. **Adaptive Colors:** Always use `lipgloss.AdaptiveColor{Light: "...", Dark: "..."}` to ensure accessibility across terminal themes.
2. **Layout Arithmetic:** Use `lipgloss.Width()` to measure rendered strings before joining them. Subtract padding/borders from the total `WindowSizeMsg` width before setting component widths.
3. **Performance:** Do not create new styles inside the `View()` function. Define them as global or member variables to avoid unnecessary allocations per frame.
4. **String Building:** For complex views, use `strings.Builder` or `lipgloss.JoinVertical` instead of standard string concatenation (+).

## Example: Fixed Header + Flexible Body
```go
func (m model) View() string {
    header := headerStyle.Width(m.width).Render("My App")
    body := bodyStyle.Width(m.width).Height(m.height - 3).Render(m.content)
    return lipgloss.JoinVertical(lipgloss.Left, header, body)
}
