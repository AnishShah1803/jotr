---
name: tui-compatibility
description: Handling cross-terminal differences, color profiles, and non-TTY environments.
license: MIT
compatibility: opencode
---

## Best Practices
1. **Color Profile Detection:** Use `lipgloss.ColorProfile()` to check if the terminal supports TrueColor (16 million colors), ANSI256, or only 16 colors.
2. **No-Color Support:** Respect the `NO_COLOR` environment variable. If set, use `lipgloss.NewRenderer(os.Stderr).SetHasDarkBackground(true)` with no colors.
3. **TTY Detection:** Check if `os.Stdout` is a TTY before starting `tea.NewProgram`. If not (e.g., in a CI/CD pipe), output a simplified plain-text version of the data.
4. **Unicode Fallbacks:** If a user is on a legacy Windows console, provide ASCII fallback characters for borders and spinners.
