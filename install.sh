#!/bin/bash
# Dark Storage CLI - Local Installation Script
# Usage: ./install.sh [OPTIONS]
#
# Options:
#   --fresh     Clean build from scratch
#   --update    Update to latest and rebuild
#   --dev       Install with debug symbols
#   --help      Show this help message

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="darkstorage"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BUILD_FLAGS="-ldflags=-s -w"

# Parse arguments
FRESH=false
UPDATE=false
DEV=false

for arg in "$@"; do
    case $arg in
        --fresh)
            FRESH=true
            shift
            ;;
        --update)
            UPDATE=true
            shift
            ;;
        --dev)
            DEV=true
            BUILD_FLAGS="-gcflags=all=-N -l"
            shift
            ;;
        --help)
            echo "Dark Storage CLI - Installation Script"
            echo ""
            echo "Usage: ./install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --fresh     Clean build from scratch (removes binaries and caches)"
            echo "  --update    Update to latest git version and rebuild"
            echo "  --dev       Build with debug symbols (no optimization)"
            echo "  --help      Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
            echo ""
            echo "Examples:"
            echo "  ./install.sh                    # Normal install"
            echo "  ./install.sh --fresh            # Clean install"
            echo "  ./install.sh --update           # Update and reinstall"
            echo "  ./install.sh --dev              # Debug build"
            echo "  INSTALL_DIR=~/.local/bin ./install.sh   # Install to user directory"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $arg${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

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

# Check requirements
check_requirements() {
    info "Checking requirements..."

    if ! command -v go &> /dev/null; then
        error "Go is not installed"
        echo "Install Go from: https://go.dev/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    success "Go ${GO_VERSION} installed"

    if ! command -v git &> /dev/null; then
        error "Git is not installed"
        exit 1
    fi

    success "Git installed"
}

# Update repository
update_repo() {
    if [ "$UPDATE" = true ]; then
        info "Updating repository..."

        # Stash any local changes
        if ! git diff-index --quiet HEAD --; then
            warn "You have uncommitted changes. Stashing..."
            git stash
        fi

        # Pull latest
        git pull origin main
        success "Repository updated to latest version"
    fi
}

# Clean build
clean_build() {
    if [ "$FRESH" = true ]; then
        info "Cleaning previous builds..."

        # Remove binary
        if [ -f "$BINARY_NAME" ]; then
            rm "$BINARY_NAME"
            success "Removed old binary"
        fi

        # Clean go cache
        go clean -cache -modcache -testcache
        success "Cleaned Go caches"
    fi
}

# Download dependencies
download_deps() {
    info "Downloading dependencies..."
    go mod download
    success "Dependencies downloaded"
}

# Build binary
build_binary() {
    info "Building Dark Storage CLI..."

    if [ "$DEV" = true ]; then
        warn "Building with debug symbols (larger binary, slower execution)"
    fi

    # Get version info
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Build with version info
    go build $BUILD_FLAGS \
        -ldflags="-X github.com/darkstorage/cli/cmd.Version=${VERSION} \
                  -X github.com/darkstorage/cli/cmd.Commit=${COMMIT} \
                  -X github.com/darkstorage/cli/cmd.Date=${DATE} \
                  -X github.com/darkstorage/cli/cmd.BuiltBy=local" \
        -o "$BINARY_NAME" main.go

    if [ ! -f "$BINARY_NAME" ]; then
        error "Build failed"
        exit 1
    fi

    # Make executable
    chmod +x "$BINARY_NAME"

    # Get binary size
    if command -v du &> /dev/null; then
        SIZE=$(du -h "$BINARY_NAME" | cut -f1)
        success "Build complete (${SIZE})"
    else
        success "Build complete"
    fi
}

# Test binary
test_binary() {
    info "Testing binary..."

    if ./"$BINARY_NAME" version &> /dev/null; then
        success "Binary works correctly"

        # Show version info
        echo ""
        ./"$BINARY_NAME" version --verbose
        echo ""
    else
        error "Binary test failed"
        exit 1
    fi
}

# Install binary
install_binary() {
    info "Installing to $INSTALL_DIR..."

    # Check if directory exists
    if [ ! -d "$INSTALL_DIR" ]; then
        warn "$INSTALL_DIR does not exist"

        # Try to create it
        if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
            success "Created $INSTALL_DIR"
        else
            error "Cannot create $INSTALL_DIR"
            warn "Try: sudo mkdir -p $INSTALL_DIR"
            warn "Or:  INSTALL_DIR=~/.local/bin ./install.sh"
            exit 1
        fi
    fi

    # Check if writable
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo &> /dev/null; then
            info "Need sudo permissions for $INSTALL_DIR"
            sudo cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
            success "Installed with sudo"
        else
            error "$INSTALL_DIR is not writable and sudo is not available"
            warn "Try: INSTALL_DIR=~/.local/bin ./install.sh"
            exit 1
        fi
    else
        cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        success "Installed to $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Verify installation
verify_installation() {
    info "Verifying installation..."

    if command -v "$BINARY_NAME" &> /dev/null; then
        success "Installation verified!"

        # Get installed version
        INSTALLED_VERSION=$("$BINARY_NAME" version 2>&1)
        echo "  $INSTALLED_VERSION"
    else
        warn "$BINARY_NAME is installed but not in PATH"
        warn "Add $INSTALL_DIR to your PATH:"
        echo ""
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
        warn "Add this to your shell config (~/.bashrc, ~/.zshrc, etc.)"
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
    echo "  4. Get help:"
    echo -e "     ${BLUE}darkstorage --help${NC}"
    echo ""

    if [ "$DEV" = true ]; then
        warn "This is a DEBUG build (not optimized for production)"
        echo "  For production use: ./install.sh (without --dev)"
        echo ""
    fi
}

# Main installation
main() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Dark Storage CLI - Local Installation${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_requirements
    update_repo
    clean_build
    download_deps
    build_binary
    test_binary
    install_binary
    verify_installation
    print_next_steps
}

# Run
main
