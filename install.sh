#!/bin/bash
set -e

# GPTCode CLI Installer
# Downloads the latest pre-built binary from GitHub releases

VERSION="${GPTCODE_VERSION:-latest}"
INSTALL_DIR="${GPTCODE_INSTALL_DIR:-$HOME/.local/bin}"
RELEASES_REPO="gptcode-cloud/cli-releases"

echo "🚀 GPTCode CLI Installer"
echo ""

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  mingw*|msys*|cygwin*)
    OS="windows"
    ;;
  *)
    echo "❌ Unsupported OS: $OS"
    exit 1
    ;;
esac

echo "📦 Detected: $OS/$ARCH"

# Get latest version if not specified
if [ "$VERSION" = "latest" ]; then
  echo "🔍 Fetching latest version..."
  VERSION=$(curl -sSL "https://raw.githubusercontent.com/${RELEASES_REPO}/main/LATEST" 2>/dev/null || echo "")
  
  if [ -z "$VERSION" ]; then
    # Fallback to GitHub API
    VERSION=$(curl -sSL "https://api.github.com/repos/${RELEASES_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  fi
  
  if [ -z "$VERSION" ]; then
    echo "❌ Could not determine latest version"
    exit 1
  fi
fi

echo "📥 Downloading GPTCode $VERSION..."

# Construct download URL
VERSION_NUM="${VERSION#v}"
if [ "$OS" = "windows" ]; then
  ARCHIVE="gptcode_${VERSION_NUM}_${OS}_${ARCH}.zip"
else
  ARCHIVE="gptcode_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
fi

DOWNLOAD_URL="https://github.com/${RELEASES_REPO}/releases/download/${VERSION}/${ARCHIVE}"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download archive
echo "   URL: $DOWNLOAD_URL"
curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE"

if [ ! -s "$TMP_DIR/$ARCHIVE" ]; then
  echo "❌ Download failed or file is empty"
  exit 1
fi

# Extract
echo "📂 Extracting..."
cd "$TMP_DIR"
if [ "$OS" = "windows" ]; then
  unzip -q "$ARCHIVE"
else
  tar -xzf "$ARCHIVE"
fi

# Find binary
BINARY=$(find . -name 'gptcode*' -type f ! -name '*.tar.gz' ! -name '*.zip' | head -1)

if [ -z "$BINARY" ]; then
  echo "❌ Binary not found in archive"
  exit 1
fi

# Install
echo "📁 Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
chmod +x "$BINARY"
mv "$BINARY" "$INSTALL_DIR/gptcode"

# Verify
if [ -x "$INSTALL_DIR/gptcode" ]; then
  echo ""
  echo "✅ GPTCode installed successfully!"
  echo ""
  echo "   Location: $INSTALL_DIR/gptcode"
  echo "   Version:  $("$INSTALL_DIR/gptcode" --version 2>/dev/null || echo "$VERSION")"
  echo ""
  
  # Check if in PATH
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "⚠️  $INSTALL_DIR is not in your PATH"
    echo ""
    echo "   Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "   export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
  fi
  
  echo "🎉 Run 'gptcode --help' to get started!"
else
  echo "❌ Installation failed"
  exit 1
fi
