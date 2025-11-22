#!/bin/bash

set -e  # Exit on error

REPO_URL="https://github.com/AnishShah1803/jotr"
RELEASE_URL="https://github.com/AnishShah1803/jotr/releases/latest/download"
BIN_DIR="/usr/local/bin"
CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
CONFIG_DIR="$CONFIG_HOME/jotr"

echo "🔧 Installing jotr..."
echo ""

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Darwin*)
        OS_TYPE="darwin"
        ;;
    Linux*)
        OS_TYPE="linux"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        OS_TYPE="windows"
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH_TYPE="amd64"
        ;;
    arm64|aarch64)
        ARCH_TYPE="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected: $OS_TYPE-$ARCH_TYPE"
echo ""

# Download pre-built binary
echo "Downloading pre-built binary..."
BINARY_NAME="jotr-${OS_TYPE}-${ARCH_TYPE}"
DOWNLOAD_URL="${RELEASE_URL}/${BINARY_NAME}"

echo "Downloading from: $DOWNLOAD_URL"
echo ""

# Try to download
if curl -fsSL "$DOWNLOAD_URL" -o /tmp/jotr 2>/dev/null; then
    echo "✓ Downloaded successfully"
else
    echo "❌ Failed to download pre-built binary"
    echo ""
    echo "Please try one of these alternatives:"
    echo "  1. Build from source: git clone $REPO_URL && cd jotr && make install"
    echo "  2. Download manually from: ${REPO_URL}/releases"
    echo "  3. Check if your platform is supported"
    exit 1
fi

# Install binary
echo "Installing to $BIN_DIR/jotr..."

# Check if we need sudo
if [ -w "$BIN_DIR" ]; then
    mv /tmp/jotr "$BIN_DIR/jotr"
    chmod +x "$BIN_DIR/jotr"
else
    echo "(This may require your password)"
    sudo mv /tmp/jotr "$BIN_DIR/jotr"
    sudo chmod +x "$BIN_DIR/jotr"
fi

# Create config directory
mkdir -p "$CONFIG_DIR"

# Copy template if config doesn't exist
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    echo "✓ Config directory created at: $CONFIG_DIR"
    echo "  Run 'jotr configure' to set up your configuration"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "Config location: $CONFIG_DIR/config.json"
echo "Binary location: $BIN_DIR/jotr"
echo ""
echo "Next steps:"
echo "  1. Run 'jotr configure' to set up your paths"
echo "  2. Start using: jotr daily, jotr sync, etc."
echo ""
echo "Quick start:"
echo "  jotr configure          # Run configuration wizard"
echo "  jotr daily              # Create/open daily note"
echo "  jotr note create        # Create a note"
echo "  jotr --help             # Show help"
echo ""
echo "Version:"
jotr version

