---
name: tui-polish
description: Final polish for TUIs, including alt-screen management and Unicode safety.
---

## Best Practices
1. **AltScreen Mode:** Always use `tea.WithAltScreen()` for full-screen apps to prevent polluting the user's terminal scrollback history.
2. **Unicode/Emoji Safety:** Lipgloss can miscalculate width for double-width characters (emojis). Use `runewidth.StringWidth()` for manual calculations if Lipgloss's default `.Width()` fails.
3. **Help Bubble:** Use `bubbles/help` to automatically generate a keybinding footer. Do not hardcode help strings.
4. **Graceful Exit:** Ensure `tea.Quit` is the only way to exit so the terminal is properly restored from raw mode.
