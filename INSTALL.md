# Installation Guide

## Quick Install

### macOS / Linux

```bash
# Using Make (recommended)
git clone https://github.com/AnishShah1803/jotr
cd jotr
make install

# Or using install script
curl -fsSL https://raw.githubusercontent.com/AnishShah1803/jotr/main/install.sh | bash
```

### Windows

```bash
# Download binary
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-windows-amd64.exe -o jotr.exe

# Move to a directory in your PATH
move jotr.exe C:\Windows\System32\
```

## Detailed Installation Methods

### Method 1: Using Make (Recommended for Developers)

**Prerequisites:** Go 1.21+

```bash
# Clone repository
git clone https://github.com/AnishShah1803/jotr
cd jotr

# Build and install
make install

# This will:
# 1. Build the binary
# 2. Copy to /usr/local/bin/jotr
# 3. Make it executable
```

**Other Make commands:**

```bash
make build       # Just build (creates ./jotr)
make build-all   # Build for all platforms
make clean       # Clean build artifacts
make test        # Run tests
make uninstall   # Remove installed binary
```

### Method 2: Using Install Script

**Prerequisites:** None (downloads or builds automatically)

```bash
# Download and run
curl -fsSL https://raw.githubusercontent.com/AnishShah1803/jotr/main/install.sh | bash

# Or download first, then run
curl -fsSL https://raw.githubusercontent.com/AnishShah1803/jotr/main/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

The install script will:
1. Detect your OS and architecture
2. Build from source (if Go is installed) or download pre-built binary
3. Install to `/usr/local/bin/jotr`
4. Create config directory at `~/.config/jotr/`

### Method 3: Download Pre-built Binary

**Prerequisites:** None

Download the appropriate binary for your platform:

**macOS:**
```bash
# Intel
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-darwin-amd64 -o jotr

# Apple Silicon
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-darwin-arm64 -o jotr

# Install
chmod +x jotr
sudo mv jotr /usr/local/bin/
```

**Linux:**
```bash
# x86_64
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-linux-amd64 -o jotr

# ARM64
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-linux-arm64 -o jotr

# Install
chmod +x jotr
sudo mv jotr /usr/local/bin/
```

**Windows:**
```bash
# Download
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-windows-amd64.exe -o jotr.exe

# Move to PATH (example)
move jotr.exe C:\Windows\System32\
```

### Method 4: Build from Source

**Prerequisites:** Go 1.21+

```bash
# Clone repository
git clone https://github.com/AnishShah1803/jotr
cd jotr

# Build
go build -o jotr

# Install
sudo mv jotr /usr/local/bin/

# Or just use locally
./jotr version
```

### Method 5: Using Go Install

**Prerequisites:** Go 1.21+

```bash
go install github.com/AnishShah1803/jotr@latest
```

This installs to `$GOPATH/bin/jotr` (usually `~/go/bin/jotr`).

Make sure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Post-Installation

### 1. Install Optional Dependencies

**For Graph Visualization (optional but recommended):**

```bash
# macOS
brew install graphviz

# Ubuntu/Debian
sudo apt-get install graphviz

# CentOS/RHEL/Fedora
sudo yum install graphviz
# or
sudo dnf install graphviz

# Arch Linux
sudo pacman -S graphviz

# Windows (using chocolatey)
choco install graphviz
```

### 2. Verify Installation

```bash
jotr version
```

You should see:
```
jotr version 1.0.0
```

### 3. Run Configuration Wizard

```bash
jotr configure
```

This will guide you through setting up:
- Base directory (where your notes are)
- Diary directory (for daily notes)
- Todo file path
- PDP file path (optional)

### 4. Run Health Check

```bash
jotr check
```

This verifies:
- Config file exists
- Directories are accessible
- Editor is configured
- Everything is working

### 5. Create Your First Note

```bash
jotr daily
```

## Configuration

The config file is located at: `~/.config/jotr/config.json`

Example configuration:

```json
{
  "paths": {
    "base_dir": "/Users/you/Documents/Notes",
    "diary_dir": "Diary",
    "todo_file_path": "todo",
    "pdp_file_path": "PDP"
  },
  "format": {
    "task_section": "Important Things",
    "capture_section": "Captured",
    "daily_note_sections": ["Notes", "Conversations/Activities"]
  }
}
```

## Updating

### Using Make

```bash
cd jotr
git pull
make install
```

### Using Install Script

```bash
./install.sh
```

The script will detect existing installation and update it.

### Manual Update

```bash
# Download new binary
curl -L https://github.com/AnishShah1803/jotr/releases/latest/download/jotr-[platform] -o jotr
chmod +x jotr
sudo mv jotr /usr/local/bin/
```

## Uninstalling

### Using Make

```bash
make uninstall
```

### Manual Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/jotr

# Optionally remove config (this deletes your configuration!)
rm -rf ~/.config/jotr/
```

## Troubleshooting

See [Getting Started](jotr.wiki/Getting-Started-Go.md) for troubleshooting tips.

