#!/bin/bash
# Dark Storage CLI Installation Script
# Usage: curl -fsSL https://install.darkstorage.io | sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="afterdarksys/darkstorage-cli"
BINARY_NAME="darkstorage"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"

# Functions
info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Detect OS
detect_os() {
    OS="$(uname -s)"
    case "$OS" in
        Linux*)     OS='linux';;
        Darwin*)    OS='darwin';;
        MINGW*|MSYS*|CYGWIN*)     OS='windows';;
        *)          error "Unsupported operating system: $OS"; exit 1;;
    esac
}

# Detect Architecture
detect_arch() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64)     ARCH='amd64';;
        amd64)      ARCH='amd64';;
        arm64)      ARCH='arm64';;
        aarch64)    ARCH='arm64';;
        armv7l)     ARCH='arm';;
        *)          error "Unsupported architecture: $ARCH"; exit 1;;
    esac
}

# Get latest version
get_latest_version() {
    info "Fetching latest version..."

    if command -v curl &> /dev/null; then
        VERSION=$(curl -sL "$GITHUB_API" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "$GITHUB_API" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required"
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
        exit 1
    fi

    success "Latest version: $VERSION"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/darkstorage_${VERSION#v}_${OS}_${ARCH}.tar.gz"

    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/darkstorage_${VERSION#v}_${OS}_${ARCH}.zip"
    fi

    info "Downloading from: $DOWNLOAD_URL"

    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/darkstorage.tar.gz"

    if [ "$OS" = "windows" ]; then
        TMP_FILE="$TMP_DIR/darkstorage.zip"
    fi

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    fi

    if [ ! -f "$TMP_FILE" ]; then
        error "Failed to download binary"
        exit 1
    fi

    success "Downloaded successfully"
}

# Extract binary
extract_binary() {
    info "Extracting binary..."

    if [ "$OS" = "windows" ]; then
        unzip -q "$TMP_FILE" -d "$TMP_DIR"
    else
        tar -xzf "$TMP_FILE" -C "$TMP_DIR"
    fi

    if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
        error "Binary not found in archive"
        exit 1
    fi

    success "Extracted successfully"
}

# Install binary
install_binary() {
    info "Installing to $INSTALL_DIR..."

    # Check if install dir is writable
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo &> /dev/null; then
            sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
            sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
        else
            error "$INSTALL_DIR is not writable and sudo is not available"
            warn "Try running with: INSTALL_DIR=\$HOME/.local/bin $0"
            exit 1
        fi
    else
        mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    success "Installed to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    info "Verifying installation..."

    if command -v "$BINARY_NAME" &> /dev/null; then
        VERSION_OUTPUT=$("$BINARY_NAME" version 2>&1 || true)
        success "Installation verified!"
        echo ""
        echo "$VERSION_OUTPUT"
    else
        warn "$BINARY_NAME is installed but not in PATH"
        warn "Add $INSTALL_DIR to your PATH:"
        echo ""
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
        warn "Or create a symlink:"
        echo ""
        echo "  ln -s $INSTALL_DIR/$BINARY_NAME /usr/local/bin/$BINARY_NAME"
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  Dark Storage CLI installed successfully! 🚀${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Log in to your Dark Storage account:"
    echo -e "     ${BLUE}darkstorage login${NC}"
    echo ""
    echo "  2. Or use an API key:"
    echo -e "     ${BLUE}darkstorage login --key YOUR_API_KEY${NC}"
    echo ""
    echo "  3. Test it out:"
    echo -e "     ${BLUE}darkstorage whoami${NC}"
    echo -e "     ${BLUE}darkstorage ls${NC}"
    echo ""
    echo "Documentation: https://docs.darkstorage.io"
    echo "Get help:      darkstorage --help"
    echo ""
}

# Main
main() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Dark Storage CLI Installer${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    detect_os
    detect_arch

    info "Detected OS: $OS"
    info "Detected Architecture: $ARCH"
    echo ""

    get_latest_version
    download_binary
    extract_binary
    install_binary
    verify_installation
    print_next_steps
}

# Run
main
