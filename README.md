# jotr

#### [Install](#installation) · [Configure](#configuration) · [Docs](../jotr.wiki/Home.md)

[![Latest Release](https://img.shields.io/github/v/release/AnishShah1803/jotr?style=for-the-badge&logo=starship&color=C9CBFF&logoColor=D9E0EE&labelColor=302D41&include_prerelease&sort=semver)](https://github.com/AnishShah1803/jotr/releases/latest)
[![Last Commit](https://img.shields.io/github/last-commit/AnishShah1803/jotr?style=for-the-badge&logo=starship&color=8bd5ca&logoColor=D9E0EE&labelColor=302D41)](https://github.com/AnishShah1803/jotr/pulse)
[![License](https://img.shields.io/github/license/AnishShah1803/jotr?style=for-the-badge&logo=starship&color=ee99a0&logoColor=D9E0EE&labelColor=302D41)](https://github.com/AnishShah1803/jotr/blob/main/LICENSE)
[![Stars](https://img.shields.io/github/stars/AnishShah1803/jotr?style=for-the-badge&logo=starship&color=c69ff5&logoColor=D9E0EE&labelColor=302D41)](https://github.com/AnishShah1803/jotr/stargazers)

**A lightning-fast command-line note-taking and task management system built for developers and power users.**

Lightning fast | Lightweight | Single binary | Zero dependencies

Stop juggling multiple tools for notes, tasks, and daily planning. jotr unifies your workflow into a single, powerful CLI that starts instantly and gets out of your way.

---

## Features

**Built for Speed** - Sub-5ms startup time, concurrent operations, and instant search across thousands of notes

**Developer-First** - Designed by developers, for developers. Integrates seamlessly with your terminal workflow

**Beautiful Interface** - Interactive TUI dashboard with fuzzy finding, task management, and calendar views

**Smart Task Tracking** - Unique task IDs enable cross-note task tracking and intelligent sync

**Graph Visualization** - Generate visual maps of your notes and their relationships using Graphviz

## Quick Start

**One-line install:**

```bash
curl -fsSL https://raw.githubusercontent.com/AnishShah1803/jotr/main/install.sh | bash
```

**Or clone and build:**

```bash
# Clone and install
git clone https://github.com/AnishShah1803/jotr
cd jotr  
make install
```

**Set up and start using:**

```bash
# Set up your workspace
jotr configure

# Start taking notes
jotr daily                    # Open today's daily note
jotr capture "Important idea" # Quick capture to daily note  
jotr dashboard               # Launch interactive TUI
```

**That's it!** You're ready to streamline your note-taking workflow.

## Installation

### Quick Install

```bash
# Clone and build
git clone https://github.com/AnishShah1803/jotr
cd jotr  
make install
```

### Other Installation Methods

- **Pre-built binaries**: Download from [releases](https://github.com/AnishShah1803/jotr/releases)
- **Go install**: `go install github.com/AnishShah1803/jotr@latest`
- **Build from source**: See [INSTALL.md](INSTALL.md) for detailed instructions

### Requirements

- **Runtime**: None! Single binary with no dependencies
- **Build**: Go 1.21+ (optional)
- **Recommended**: nvim, fzf, git, graphviz

## Core Features

### Smart Daily Notes
Automatic daily note creation with customizable templates, task sections, and streak tracking. Never lose track of your daily planning again.

### Instant Search  
Lightning-fast full-text search across all your notes. Find anything in milliseconds, no matter how large your note collection grows.

### Intelligent Task Management
Every task gets a unique ID for precise tracking across notes. Smart sync prevents duplicates while maintaining task relationships.

### Interactive Dashboard
Beautiful terminal interface with fuzzy finding, calendar views, and real-time task statistics. All the power of a GUI in your terminal.

### Visual Knowledge Mapping
Generate graph visualizations of your notes to discover hidden connections and patterns in your thinking.

## Usage Examples

### Daily Workflow
```bash
# Start your day
jotr daily                    # Open today's note
jotr summary                  # Review pending tasks

# Capture ideas throughout the day  
jotr capture "API design thoughts"
jotr capture "Meeting notes - discuss Q4 planning"

# End of day review
jotr stats                    # Check productivity metrics
jotr archive                  # Archive completed tasks
```

### Note Management
```bash
# Create and organize notes
jotr note create "Project Architecture"
jotr note open               # Fuzzy search and open any note
jotr search "authentication" # Find notes mentioning auth

# Work with tags and links
jotr tags find work          # Find all work-related notes
jotr graph                   # Visualize note relationships
```

### Task Tracking
```bash
# Smart task management
jotr sync                    # Sync tasks to your todo list
jotr summary                 # View task overview
jotr streak                  # Check daily note consistency
```

## Advanced Features

### Task ID System

Every task gets a unique identifier for precise cross-note tracking:

```markdown
- [ ] Review project proposal <!-- id: abc123 -->
- [x] Update documentation <!-- id: def456 -->
```

**Benefits:**
- **Automatic ID generation** - New tasks get unique IDs automatically
- **Cross-note tracking** - Reference the same task across multiple notes  
- **Smart sync** - The `sync` command uses IDs to avoid duplicates
- **Manual ID support** - Assign custom IDs when needed

### Graph Visualization

Generate visual maps of your knowledge base:

```bash
jotr graph                   # Generate DOT graph and open in default viewer
```

**Features:**
- **DOT syntax output** - Standard Graphviz format
- **Auto-sanitization** - Handles special characters safely  
- **Link detection** - Shows note relationships
- **Visual clustering** - Groups related content

**Requirements:** Install `graphviz` package (see [installation guide](INSTALL.md))

### Daily Note Templates

Automatic daily note structure with customizable sections:

```markdown
## Tasks

### Todo
- [ ] New task <!-- id: generated-id -->

### In Progress  
- [ ] Ongoing task <!-- id: another-id -->

### Done
- [x] Completed task <!-- id: done-id -->
```

Task sections respect your configuration and provide consistent organization.

## Configuration

jotr uses a JSON config at `~/.config/jotr/config.json`. Run `jotr configure` for interactive setup, or see [config.template.json](config.template.json) for all options.

Example configuration:
```json
{
  "paths": {
    "base_dir": "/Users/you/Documents/Notes",
    "diary_dir": "Diary",
    "todo_file_path": "todo"
  },
  "format": {
    "task_section": "Tasks",
    "daily_note_sections": ["Notes", "Meetings"]
  }
}
```

## Command Reference

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
| `graph` | Generate graph visualization | |
| `version` | Show version | |

## Performance & Architecture

**Performance:**
- Startup time: ~5ms (instant)  
- Memory usage: ~15MB (lightweight)
- Binary size: ~15MB (self-contained)
- Build time: ~2 seconds

**Architecture:**
- Built with Go for speed and reliability
- Concurrent operations (search, sync, etc.)
- Single binary deployment
- Cross-platform (macOS, Linux, Windows)

**Dependencies:**
- cobra (CLI framework)
- viper (configuration)  
- bubbletea (TUI framework)
- lipgloss (terminal styling)

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch  
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

For development:
```bash
make build        # Build binary
make test         # Run tests  
make fmt          # Format code
make clean        # Clean artifacts
```

## Documentation

- [Installation Guide](INSTALL.md)
- [Release Guide](RELEASE.md)  
- [Wiki](../jotr.wiki/Home.md)
- [FAQ](../jotr.wiki/FAQ.md)

## Support

- [Report Issues](https://github.com/AnishShah1803/jotr/issues)
- [Discussions](https://github.com/AnishShah1803/jotr/discussions)
- [Documentation](../jotr.wiki/Home.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Made for developers and power users**