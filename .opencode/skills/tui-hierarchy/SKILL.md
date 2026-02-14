---
name: tui-hierarchy
description: Patterns for scaling TUIs by nesting models and delegating messages.
license: MIT
compatibility: opencode
---

## Best Practices
1. **Root Delegation:** The root model should act as a router.
2. **Standard Message Forwarding:** Always forward `tea.WindowSizeMsg` to all active children so they can recalculate their internal Lipgloss styles.
3. **Return Pattern:** When a child model needs to "close" or signal the parent, return a custom `tea.Msg` (e.g., `EditorFinishedMsg`) which the parent catches.
4. **Pointer Receivers:** Use pointer receivers (`*model`) for sub-components to ensure state updates persist when passed between functions.
5. **State Machine Patterns:** For complex multi-stage flows, use the bubbletea-mvu skill to implement proper state machine patterns with enum-based session states.
