# Contributing to jotr

Thank you for your interest in contributing to jotr!

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/AnishShah1803/jotr
cd jotr

# Build the project
make build

# Run tests
make test

# Run locally
./jotr --help
```

## Project Structure

```
jotr/
├── main.go              # Entry point
├── cmd/                 # Command implementations (Cobra)
│   ├── root.go         # Root command
│   ├── daily.go        # Daily note command
│   ├── dashboard.go    # TUI dashboard
│   └── ...             # 20 total commands
├── internal/           # Internal packages
│   ├── config/        # Configuration management
│   ├── notes/         # Note operations
│   ├── tasks/         # Task operations
│   └── tui/           # Bubbletea TUI components
├── go.mod             # Go module definition
└── go.sum             # Dependency checksums
```

## Making Changes

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**
4. **Test your changes**
   ```bash
   make test
   make build
   ./jotr <your-command>
   ```
5. **Commit with clear messages**
   ```bash
   git commit -m "Add: description of your change"
   ```
6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Create a Pull Request**

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

## Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Implement the command using Cobra
3. Register it in `cmd/root.go`
4. Add tests if applicable
5. Update documentation

Example:
```go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  "Detailed description",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

## Testing

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run specific test
go test ./internal/config -v
```

## Documentation

- Update README.md if adding major features
- Update wiki pages for user-facing changes
- Add inline comments for complex code
- Update CHANGELOG.md (if exists)

## Homebrew Formula

The `jotr.rb` file in the root is a Homebrew formula for package installation.

**Note:** This is written in Ruby (Homebrew's DSL), but jotr itself is Go.

To update after a release:
1. Build binaries for all platforms
2. Calculate SHA256 checksums
3. Update URLs and checksums in `jotr.rb`
4. Test: `brew install --build-from-source ./jotr.rb`

See [ABOUT_JOTR_RB.md](../ABOUT_JOTR_RB.md) for details.

## Release Process

See [RELEASE.md](RELEASE.md) for the complete release process.

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Check the [FAQ](../jotr.wiki/FAQ.md)

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

---

Thank you for contributing! 🎉

