# jotr

#### [Install](#installation) · [Configure](#configuration) · [Wiki](https://github.com/AnishShah1803/jotr/wiki)

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
jotr                         # Launch interactive REPL
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

⚠️ **Note**: This will overwrite any existing jotr installation at `/usr/local/bin/jotr`

### For Developers

```bash
# Development build (includes dev mode)
git clone https://github.com/AnishShah1803/jotr
cd jotr
make dev

# Production build (user release)
make build  # ⚠️  Will overwrite existing binary
```

### Other Installation Methods

- **Pre-built binaries**: Download from [releases](https://github.com/AnishShah1803/jotr/releases) (recommended for users)
- **Go install**: `go install github.com/AnishShah1803/jotr@latest`
- **Build from source**: See [INSTALL.md](INSTALL.md) for detailed instructions

**⚠️ Build Options**:

- **For users**: Use pre-built binaries or `make install`
- **For developers**: Use `make dev` for development builds
- **Production build**: `make build` (local binary, no system changes)

### Requirements

- **Runtime**: None! Single binary with no dependencies
- **Build**: Go 1.21+ (optional)
- **Recommended**: nvim, fzf, git, graphviz
- **Installation Path**: `/usr/local/bin/jotr` (may overwrite existing)

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
jotr note rename "old name" "new name"  # Rename a note
jotr note move "my note" work           # Move note to a subfolder
jotr note delete "old note"             # Delete a note
jotr note delete "old note" --force     # Delete without confirmation
jotr search "authentication"            # Find notes mentioning auth

# Work with tags and links
jotr tags find work          # Find all work-related notes
jotr links outgoing MyNote   # Show outgoing wiki-links in a note
jotr links backlinks MyNote  # Show all notes that link to MyNote
jotr graph                   # Visualize note relationships
```

### Daily Notes

```bash
jotr daily                              # Open today's daily note
jotr daily append "Standup notes here"  # Append text without opening editor
jotr daily prepend "Top of mind: ..."   # Prepend to today's note
jotr daily read                         # Print today's note to stdout
jotr daily --date 2024-01-15            # Work with a specific date
```

### Frontmatter

```bash
jotr frontmatter MyNote                   # Show all frontmatter fields
jotr frontmatter get MyNote status        # Get a specific field
jotr frontmatter set MyNote status=done   # Set a field
jotr frontmatter remove MyNote status     # Remove a field
```

### Stats and Structure

```bash
jotr files                        # List all notes
jotr files --folder work          # List notes in a subfolder
jotr files --ext md --total       # Count notes by extension
jotr wordcount --path work        # Word count across a folder
jotr wordcount --file MyNote      # Word and character count for one note
jotr outline --file MyNote        # Show heading structure of a note
jotr outline --path work --total  # Summarise headings across a folder
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

**Requirements:** Install `graphviz` package (see [Installation Guide](INSTALL.md))

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

jotr uses a JSON config at `~/.config/jotr/config.json`. Run `jotr configure` for interactive setup, or use flags for non-interactive configuration:

```bash
jotr configure --base-dir ~/Notes --diary-dir Diary --todo-file todo
```

See [config.template.json](config.template.json) for all configuration options.

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
| ------- | ----------- | ------- |
| `daily` | Create/open daily note | `d` |
| `note` | Create, open, list, rename, move, delete notes | `n` |
| `search` | Search across all notes | `find`, `grep` |
| `capture` | Quick capture to daily note | `cap` |
| `tags` | Manage tags | `tag` |
| `links` | Show outgoing links, backlinks, unresolved, orphans, deadends | |
| `aliases` | List and search note aliases | |
| `properties` | Manage typed note properties | `props` |
| `frontmatter` | View and edit note frontmatter (legacy) | `fm` |
| `files` | List notes with optional filters | |
| `wordcount` | Word and character counts | `wc` |
| `outline` | Show heading structure of notes | |
| `summary` | Show task summary | `sum` |
| `stats` | Show task statistics | `st` |
| `sync` | Sync tasks to todo list | `s` |
| `archive` | Archive completed tasks | `arc` |
| `streak` | Show daily note streak | |
| `calendar` | Show calendar view | `cal` |
| `template` | Manage templates | `tmpl` |
| `list` | List recent notes | `ls` |
| `recents` | List recently modified notes | |
| `bulk` | Bulk operations | |
| `bookmark` | Bookmark notes for quick access | `alias` |
| `configure` | Configuration wizard | `config`, `cfg` |
| `graph` | Generate graph visualization | |
| `version` | Show version | |

### Sub-commands

**`note`**

| Sub-command | Description |
| ----------- | ----------- |
| `note create [name]` | Create a new note |
| `note open [query]` | Fuzzy-find and open a note |
| `note list` | List all notes |
| `note rename [query] [new-name]` | Rename a note (interactive if args omitted) |
| `note move [query] [folder]` | Move a note to a folder (interactive if args omitted) |
| `note delete [query]` | Delete a note (use `--permanent` to skip trash) |
| `note random` | Open a random note |
| `note unique` | List notes with unique content sizes |

**`daily`**

| Sub-command | Description |
| ----------- | ----------- |
| `daily` | Open today's daily note in editor |
| `daily append [text]` | Append text to today's note |
| `daily prepend [text]` | Prepend text to today's note |
| `daily read` | Print today's note to stdout |

**`links`**

| Sub-command | Description |
| ----------- | ----------- |
| `links outgoing [note]` | Show wiki-links contained in a note |
| `links backlinks [note]` | Show all notes that link to a note |
| `links unresolved` | Show broken/unresolved wiki-links |
| `links orphans` | List notes with no incoming links |
| `links deadends` | List notes with no outgoing links |

**`aliases`**

| Sub-command | Description |
| ----------- | ----------- |
| `aliases list` | List all aliases in the vault |
| `aliases find [name]` | Find notes by alias (case-insensitive) |
| `aliases stats` | Show alias usage statistics |

**`properties`** (alias: `props`)

| Sub-command | Description |
| ----------- | ----------- |
| `properties list [note]` | Show all properties with types |
| `properties get [note] [key]` | Get the value of a specific property |
| `properties set [note] key=value` | Set property with auto type detection |
| `properties set [note] key:type=value` | Set with explicit type (text, list, number, checkbox, date, datetime) |
| `properties remove [note] [key]` | Remove a property |
| `properties stats` | Show vault-wide property statistics |
| `properties stats [property] [--counts]` | Show usage stats for specific property |

**`recents`**

| Flag | Description |
| ---- | ----------- |
| `--limit N` | Show top N recent notes (default: 10) |
| `--total` | Show count only |

**`frontmatter`** (alias: `fm`)

| Sub-command | Description |
| ----------- | ----------- |
| `frontmatter list [note]` | Show all frontmatter fields |
| `frontmatter get [note] [key]` | Get the value of a specific field |
| `frontmatter set [note] key=value` | Set or update a field |
| `frontmatter remove [note] [key]` | Remove a field |

**`files`**

| Flag | Description |
| ---- | ----------- |
| `--folder` | Filter by subfolder |
| `--ext` | Filter by file extension |
| `--total` | Print a total count instead of listing |

**`wordcount`** (alias: `wc`)

| Flag | Description |
| ---- | ----------- |
| `--file` | Target a single note by name |
| `--path` | Target all notes under a folder |
| `--words` | Show word count only |
| `--characters` | Show character count only |

**`outline`**

| Flag | Description |
| ---- | ----------- |
| `--file` | Target a single note by name |
| `--path` | Target all notes under a folder |
| `--format` | Output format (`pretty` or `raw`) |
| `--total` | Print heading counts instead of full outline |

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
- [Wiki](https://github.com/AnishShah1803/jotr/wiki)

## Support

- [Report Issues](https://github.com/AnishShah1803/jotr/issues)
- [Discussions](https://github.com/AnishShah1803/jotr/discussions)
- [Wiki](https://github.com/AnishShah1803/jotr/wiki)

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Made for developers and power users**
