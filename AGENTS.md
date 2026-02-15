# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-14
**Stack:** Go 1.21+, Cobra CLI, Bubble Tea TUI

## STRUCTURE

```
jotr/
├── main.go           # Thin entry → cmd.Execute()
├── cmd/              # CLI commands (cobra)
│   ├── root.go       # 30 subcommands, global flags
│   ├── note/         # daily, create, open, list
│   ├── task/         # sync, summary, archive
│   ├── search/       # search, tags, links
│   ├── system/       # configure, update, version
│   ├── templatecmd/  # template operations
│   ├── util/         # utility commands
│   └── visual/       # visualization commands
├── internal/
│   ├── tui/          # Bubble Tea dashboard
│   ├── config/       # Config loading + migration
│   ├── tasks/        # Task parsing, sync logic
│   ├── templates/    # Template rendering
│   ├── services/     # Business logic
│   ├── utils/        # Shared utilities
│   ├── notes/        # Note operations
│   ├── options/      # CLI option handling
│   ├── output/       # Output formatting
│   ├── state/        # Application state
│   ├── updater/      # Self-update logic
│   ├── version/      # Version info
│   └── testhelpers/  # Test utilities
└── test/             # Integration tests
```

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add CLI command | `cmd/<name>/` + register in `cmd/root.go` init() |
| Modify TUI | `internal/tui/` |
| Config changes | `internal/config/config.go` |
| Task logic | `internal/tasks/tasks.go` |
| Note operations | `internal/notes/notes.go` |
| Shared helpers | `internal/utils/` |
| Test helpers | `internal/testhelpers/` |

## CONVENTIONS

**Build:**

- `make dev` for testing (NOT `make install`)
- Version: `YEAR_IN_DEV.MONTH.0` (year - 2025)

**Code Style:**

- 40+ linters via golangci-lint (complexity ≤15)
- `go fmt` required before commit
- Pre-commit hooks enforce linting

**Commands:**

- Export `var Cmd = &cobra.Command{}` from sub-package
- Import + add in `cmd/root.go` init()
- Flags: use helpers from `cmd/options.go`

**Testing:**

- `testhelpers.NewTestFS(t)` for file setup
- `testhelpers.ExecuteCommand()` for CLI tests
- Table-driven tests preferred

## ANTI-PATTERNS

- **DO NOT** commit `.ralph-*` files
- **DO NOT** create styles in TUI `View()` (allocate once)
- **DO NOT** mutate model from goroutines (use `tea.Cmd`)
- **DO NOT** compile regex in loops (hoist to package `var`)
- **DO NOT** use `filepath.Walk` (use `filepath.WalkDir`)

## COMMANDS

```bash
make dev           # Development build
make test          # Run tests
make test-race     # Race detection
make lint          # golangci-lint
make fmt           # Format code
```

## NOTES

- Dev command only compiles with `//go:build dev` tag
- Config migration auto-removes deprecated fields
- TUI has dedicated skills: `bubbletea-mvu`, `lipgloss-layout`, `tui-*`
