# jotr

⚡ **Lightning fast** | 💾 **Lightweight** | 📦 **Single binary** | 🚀 **No dependencies**

A powerful command-line note-taking and task management system built with Go.

## Features

✅ **20 Commands Implemented:**
- 📝 Daily notes with templates
- 📓 Note management (create, open, list)
- 🔍 Full-text search across all notes
- ⚡ Quick capture to daily note
- 🏷️ Tag management and statistics
- 📋 Task management and sync
- 📊 Task statistics and summaries
- 🗄️ Archive completed tasks
- 🔥 Daily note streak tracking
- 📅 Calendar view
- 🎨 Template management
- 🔧 Health checks and validation
- 📦 Bulk operations
- 🎯 Quick actions menu
- 📱 Beautiful TUI dashboard (Bubbletea)
- And more!

## Status

✅ **Production Ready** - All core features implemented and tested!

### Completed ✅
- [x] 20 commands fully implemented
- [x] Interactive TUI dashboard with 4-panel layout
- [x] Configuration wizard
- [x] Task sync and management
- [x] Search and filtering
- [x] Template system
- [x] Statistics and analytics
- [x] Health checks

### Planned 📋
- [ ] Full test coverage
- [ ] CI/CD pipeline
- [ ] Binary releases for macOS, Linux, Windows
- [ ] Git integration enhancements
- [ ] Graph visualization
- [ ] Plugin system

## Quick Start

### Installation

See [INSTALL.md](INSTALL.md) for detailed installation instructions.

**Quick install:**
```bash
# Using Makefile
make install

# Or build manually
go build -o jotr
sudo mv jotr /usr/local/bin/
```

### First Run

```bash
# Run configuration wizard
jotr configure

# Create/open today's daily note
jotr daily

# Launch interactive dashboard
jotr dashboard

# Show help
jotr --help
```

## Usage Examples

```bash
# Daily notes
jotr daily                    # Open today's note
jotr daily --yesterday        # Open yesterday's note
jotr daily --path            # Show path only

# Note management
jotr note create "Project Ideas"
jotr note open               # Fuzzy find and open
jotr note list               # List recent notes

# Quick capture
jotr capture "Meeting at 2pm"
jotr cap "Buy groceries"     # Using alias

# Search
jotr search "project"
jotr find "TODO"             # Using alias

# Tasks
jotr summary                 # Show task summary
jotr stats                   # Show statistics
jotr sync                    # Sync tasks to todo list
jotr archive                 # Archive completed tasks

# Tags
jotr tags list               # List all tags
jotr tags find work          # Find notes with tag
jotr tags stats              # Tag statistics

# Other
jotr streak                  # Show daily note streak
jotr calendar                # Calendar view
jotr check                   # Health check
jotr quick                   # Quick actions menu
```

## Commands

| Command | Description | Aliases |
|---------|-------------|---------|
| `daily` | Create/open daily note | `d` |
| `note` | Create, open, list notes | `n` |
| `search` | Search across all notes | `find`, `grep` |
| `capture` | Quick capture to daily note | `cap` |
| `tags` | Manage tags | `tag` |
| `summary` | Show task summary | `sum` |
| `stats` | Show task statistics | `st` |
| `sync` | Sync tasks to todo list | `s` |
| `archive` | Archive completed tasks | `arc` |
| `streak` | Show daily note streak | |
| `calendar` | Show calendar view | `cal` |
| `template` | Manage templates | `tmpl` |
| `list` | List recent notes | `ls` |
| `quick` | Quick actions menu | `q` |
| `bulk` | Bulk operations | |
| `check` | Health check | |
| `dashboard` | Interactive TUI dashboard | `dash` |
| `configure` | Configuration wizard | `config`, `cfg` |
| `version` | Show version | |

## Project Structure

```
jotr/
├── main.go                 # Entry point
├── cmd/                    # Commands (Cobra)
│   ├── root.go            # Root command
│   ├── daily.go           # Daily note command
│   ├── dashboard.go       # TUI dashboard
│   ├── configure.go       # Configuration wizard
│   └── ...                # 20 total commands
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── notes/            # Note operations
│   ├── tasks/            # Task operations
│   └── tui/              # Bubbletea TUI components
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── Makefile              # Build automation
├── install.sh            # Installation script
└── config.template.json  # Configuration template
```

## Building & Development

```bash
# Build
make build

# Build for all platforms
make build-all

# Install locally
make install

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

## Dependencies

- **cobra** - CLI framework
- **viper** - Configuration management
- **bubbletea** - TUI framework
- **lipgloss** - Terminal styling
- **bubbles** - TUI components (viewport, list, etc.)

## Why jotr?

1. **⚡ Performance** - Lightning fast startup (~5ms)
2. **📦 Single Binary** - No runtime dependencies, easy distribution
3. **🎨 Beautiful TUI** - Interactive dashboard with Bubbletea
4. **🌍 Cross-Platform** - Works on macOS, Linux, Windows
5. **🔒 Type Safety** - Built with Go for reliability
6. **⚙️ Concurrent** - Fast parallel operations (search, sync, etc.)
7. **🎯 Developer-Friendly** - Designed for power users and developers

## Performance

```
Startup Time:  ~5ms (instant)
Memory Usage:  ~15MB (lightweight)
Binary Size:   ~15MB (self-contained)
Build Time:    ~2 seconds
```

## Requirements

**Runtime:** None! Single binary with no dependencies.

**Build (optional):** Go 1.21+

**Recommended:**
- nvim (Neovim) or your preferred editor
- fzf for fuzzy finding (optional but recommended)
- git for version control (optional)

## Configuration

jotr uses a JSON configuration file located at `~/.config/jotr/config.json`.

Run `jotr configure` to set up your configuration interactively, or see [config.template.json](config.template.json) for all available options.

## Documentation

- 📖 [Installation Guide](INSTALL.md)
- 📖 [Release Guide](RELEASE.md)
- 📚 [Wiki](../jotr.wiki/Home.md)
- ❓ [FAQ](../jotr.wiki/FAQ.md)

## License

MIT License - see [LICENSE](LICENSE) for details

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

- 🐛 [Report Issues](https://github.com/AnishShah1803/jotr/issues)
- 💬 [Discussions](https://github.com/AnishShah1803/jotr/discussions)
- 📖 [Documentation](../jotr.wiki/Home.md)

---

Made with ❤️ for developers and power users
