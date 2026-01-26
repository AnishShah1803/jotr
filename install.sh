#!/bin/bash

set -e

REPO_URL="https://github.com/AnishShah1803/jotr"
API_URL="https://api.github.com/repos/AnishShah1803/jotr/releases/latest"
BIN_DIR="${HOME}/.local/bin"
CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
CONFIG_DIR="$CONFIG_HOME/jotr"
TEMP_DIR="/tmp/jotr-install-$(date +%s)"

echo "Installing jotr..."
echo ""

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -s "$API_URL" | grep '"tag_name"' | cut -d'"' -f4
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$API_URL" | grep '"tag_name"' | cut -d'"' -f4
    else
        echo "latest"
    fi
}

get_checksum_from_api() {
    local version="$1"
    local binary_name="$2"
    local api_url="https://api.github.com/repos/AnishShah1803/jotr/releases/tags/${version}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s "$api_url" | grep -A 5 "\"name\": \"${binary_name}\"" | grep '"digest"' | cut -d'"' -f4 | cut -d':' -f2
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$api_url" | grep -A 5 "\"name\": \"${binary_name}\"" | grep '"digest"' | cut -d'"' -f4 | cut -d':' -f2
    fi
}

verify_checksum() {
    local file="$1"
    local expected="$2"
    
    if command -v sha256sum >/dev/null 2>&1; then
        actual=$(sha256sum "$file" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        actual=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        echo "Cannot verify checksum: no checksum tool available"
        return 0
    fi
    
    if [ "$actual" = "$expected" ]; then
        echo "Checksum verified"
        return 0
    else
        echo "❌ Checksum verification failed"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        return 1
    fi
}

mkdir -p "$TEMP_DIR"

LATEST_VERSION=$(get_latest_version)
echo "Latest version: $LATEST_VERSION"
echo ""

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
        echo "Supported: Linux, macOS, Windows (WSL)"
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
        echo "Supported: amd64, arm64"
        exit 1
        ;;
esac

echo "Detected platform: $OS_TYPE-$ARCH_TYPE"
echo ""

BINARY_NAME="jotr-${OS_TYPE}-${ARCH_TYPE}"
if [ "$OS_TYPE" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
    ARCHIVE_EXTENSION="zip"
else
    ARCHIVE_EXTENSION="tar.gz"
fi

# Use raw binary directly instead of archive when available
ARCHIVE_NAME="${BINARY_NAME}"
DOWNLOAD_URL="https://github.com/AnishShah1803/jotr/releases/download/${LATEST_VERSION}/${BINARY_NAME}"

echo "Downloading: $BINARY_NAME"
echo "URL: $DOWNLOAD_URL"
echo ""

if command -v curl >/dev/null 2>&1; then
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/$BINARY_NAME"; then
        echo "❌ Download failed"
        exit 1
    fi
elif command -v wget >/dev/null 2>&1; then
    if ! wget -q "$DOWNLOAD_URL" -O "$TEMP_DIR/$BINARY_NAME"; then
        echo "❌ Download failed"
        exit 1
    fi
else
    echo "❌ Neither curl nor wget available"
    exit 1
fi

echo "Download successful"

CHECKSUMS_FILE="checksums.txt"
CHECKSUMS_URL="https://github.com/AnishShah1803/jotr/releases/download/${LATEST_VERSION}/${CHECKSUMS_FILE}"

# Try to get checksum from checksums.txt file first
EXPECTED_CHECKSUM=""
if curl -fsSL "$CHECKSUMS_URL" -o "$TEMP_DIR/$CHECKSUMS_FILE" 2>/dev/null; then
    if [ -s "$TEMP_DIR/$CHECKSUMS_FILE" ]; then
        EXPECTED_CHECKSUM=$(grep "$BINARY_NAME" "$TEMP_DIR/$CHECKSUMS_FILE" | cut -d' ' -f1)
    fi
fi

# If checksums.txt didn't work, try to get it from the GitHub API
if [ -z "$EXPECTED_CHECKSUM" ]; then
    echo "Checksum not found in checksums.txt, trying GitHub API..."
    EXPECTED_CHECKSUM=$(get_checksum_from_api "$LATEST_VERSION" "$BINARY_NAME")
fi

if [ -n "$EXPECTED_CHECKSUM" ]; then
    if verify_checksum "$TEMP_DIR/$BINARY_NAME" "$EXPECTED_CHECKSUM"; then
        echo "Security verification passed"
    else
        echo "❌ Security verification failed"
        echo "  Expected: $EXPECTED_CHECKSUM"
        exit 1
    fi
else
    echo "⚠️  Could not retrieve checksum for verification"
    echo "  This may be a temporary issue with the checksums file"
    echo "  Proceeding without checksum verification (not recommended)"
fi

# Verify the binary was downloaded
FOUND_BINARY="$BINARY_NAME"
if [ ! -f "$TEMP_DIR/$FOUND_BINARY" ]; then
    echo "❌ Binary not found after download"
    echo "Expected: $TEMP_DIR/$FOUND_BINARY"
    ls -la "$TEMP_DIR"
    exit 1
fi

echo "Installing binary to $BIN_DIR..."

if [ -f "$BIN_DIR/jotr" ]; then
    echo "Existing jotr found, backing up..."
    cp "$BIN_DIR/jotr" "$BIN_DIR/jotr.backup.$(date +%s)"
fi

mkdir -p "$BIN_DIR"

mv "$TEMP_DIR/$FOUND_BINARY" "$BIN_DIR/jotr"
chmod +x "$BIN_DIR/jotr"
INSTALL_METHOD="user"

if ! echo "$PATH" | grep -q "$HOME/.local/bin"; then
    echo ""
    echo "⚠️  ~/.local/bin is not in your PATH"
    echo "Add it with:"
    echo "  echo 'export PATH=\"\$PATH:\$HOME/.local/bin\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
fi

echo "Binary installed to $BIN_DIR/jotr"

if ! "$BIN_DIR/jotr" version >/dev/null 2>&1; then
    echo "❌ Installation verification failed"
    exit 1
fi

mkdir -p "$CONFIG_DIR"

if [ ! -f "$CONFIG_DIR/config.json" ]; then
echo "Config directory created: $CONFIG_DIR"
echo "Run 'jotr configure' to set up your configuration"
else
    echo "Config directory exists: $CONFIG_DIR"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Installation Summary:"
echo "  Version: $LATEST_VERSION"
echo "  Binary: $BIN_DIR/jotr"
echo "  Config: $CONFIG_DIR/config.json"
echo "  Method: $INSTALL_METHOD"
echo ""

INSTALLED_VERSION=$("$BIN_DIR/jotr" version 2>/dev/null || echo "unknown")
echo "Verification: $INSTALLED_VERSION"

echo ""
echo "Quick Start:"
echo "  jotr configure          # Configuration wizard"
echo "  jotr daily              # Create daily note"
echo "  jotr capture \"idea\"     # Quick capture"
echo "  jotr --help             # Show all commands"
echo ""

echo "Documentation:"
echo "  Wiki: $REPO_URL/wiki"
echo "  Issues: $REPO_URL/issues"
echo "  Releases: $REPO_URL/releases"
echo ""

if [ "$INSTALL_METHOD" = "local" ]; then
    echo "Remember to start a new terminal or run:"
    echo "  export PATH=\"\$PATH:\$HOME/.local/bin\""
fi

