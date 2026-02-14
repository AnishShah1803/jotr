---
name: tui-concurrency
description: Rules for handling asynchronous I/O and preventing race conditions in Bubble Tea.
license: MIT
compatibility: opencode
---

## Best Practices
1. **No Direct Mutation:** NEVER change model fields from a goroutine. Use `tea.Cmd` to return a `tea.Msg` that the `Update` loop handles.
2. **Command Factory:** If you need to pass data to a command, use a function that returns a `tea.Cmd`:
   ```go
   func fetchData(id string) tea.Cmd {
       return func() tea.Msg {
           res := api.Get(id)
           return dataMsg(res)
       }
   }
3. **Sequential Execution**: Use `tea.Sequence(cmd1, cmd2)` when order matters (e.g., Log In -> Fetch Profile).

4. **Program.Send**: Only use `p.Send(msg)` if you are communicating from a truly external process (like an OS signal handler).
