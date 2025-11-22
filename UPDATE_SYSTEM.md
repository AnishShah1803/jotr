# Update System for jotr

This document describes the update system implemented for jotr.

## Features

### CLI Update Commands
- `jotr update` - Check for and install updates
- `jotr update --check` - Only check for updates, don't install
- `jotr update --force` - Force update even if same version
- `jotr --update` - Quick update flag

### TUI Update Integration
- Press `u` in the TUI to check for updates
- Update notifications shown in the footer when available
- Status messages for update progress

## How It Works

1. **Version Checking**: Queries GitHub releases API for the latest version
2. **Changelog Display**: Shows formatted release notes before updating
3. **Binary Download**: Downloads the appropriate binary for your OS/architecture
4. **Self-Replacement**: Safely replaces the current binary with backup/restore
5. **Cross-Platform**: Supports Linux, macOS, and Windows

## Setup for Your Repository

1. **Update Repository URL**: Change the GitHub repo URL in `internal/updater/updater.go`:
   ```go
   const githubRepo = "your-username/jotr"
   ```

2. **Release Workflow**: Use the provided GitHub Actions workflow (`.github-workflows-release.yml.example`) to automatically build and release binaries.

3. **Binary Naming**: Ensure your releases follow the naming pattern:
   - `jotr-linux-amd64`
   - `jotr-darwin-amd64` 
   - `jotr-windows-amd64.exe`
   - etc.

4. **Version Tagging**: Use semantic versioning tags (e.g., `v1.0.0`, `v1.1.0`)

## Usage Examples

### CLI Usage
```bash
# Check for updates
jotr update --check

# Install updates
jotr update

# Quick update on startup
jotr --update
```

### TUI Usage
```
1. Launch jotr (opens TUI)
2. Press 'u' to check for updates
3. Update status shown in footer
4. If update available, exit and run 'jotr update'
```

## Error Handling

The update system handles various scenarios:
- Network connectivity issues
- GitHub API rate limiting  
- Missing binaries for platform
- Permissions issues during binary replacement
- Backup/restore on update failure

## Security Considerations

- Downloads are from official GitHub releases only
- Binary checksums could be added for additional security
- Backup is created before replacement
- Failed updates are automatically rolled back