#!/bin/bash
set -e

# ============================================================================
# KK CLI Installer
# Usage: curl -sSL https://get.kkengine.com/cli | bash
# ============================================================================

# Configuration
REPO="kkauto-net/kk-install"
BINARY="kk"
INSTALL_DIR="/usr/local/bin"

# Colors and formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Symbols
CHECK="✓"
CROSS="✗"
ARROW="→"
INFO="•"

# ----------------------------------------------------------------------------
# Helper Functions
# ----------------------------------------------------------------------------

print_header() {
    echo ""
    echo -e "${CYAN}${BOLD}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}${BOLD}║                      KK CLI Installer                        ║${NC}"
    echo -e "${CYAN}${BOLD}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${BLUE}${ARROW}${NC} $1"
}

print_success() {
    echo -e "${GREEN}${CHECK}${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

print_error() {
    echo -e "${RED}${CROSS}${NC} $1"
}

print_info() {
    echo -e "${INFO} $1"
}

# ----------------------------------------------------------------------------
# System Detection
# ----------------------------------------------------------------------------

detect_platform() {
    print_step "Detecting platform..."

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    print_info "OS: ${BOLD}$OS${NC}"
    print_info "Architecture: ${BOLD}$ARCH${NC}"
}

# ----------------------------------------------------------------------------
# Version Check
# ----------------------------------------------------------------------------

get_latest_version() {
    print_step "Checking latest version..."

    if command -v jq &> /dev/null; then
        LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.tag_name')
    else
        LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$LATEST" ]; then
        print_error "Failed to fetch version. Please check your network connection."
        exit 1
    fi

    # Validate version format (must be vX.Y.Z)
    if [[ ! "$LATEST" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format: $LATEST"
        exit 1
    fi

    print_info "Latest version: ${BOLD}${GREEN}$LATEST${NC}"
}

# ----------------------------------------------------------------------------
# Download and Verify
# ----------------------------------------------------------------------------

download_binary() {
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/kkcli_${LATEST#v}_${OS}_${ARCH}.tar.gz"
    CHECKSUM_URL="https://github.com/$REPO/releases/download/$LATEST/checksums.txt"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download checksum file
    print_step "Downloading checksum..."
    if ! curl -sL "$CHECKSUM_URL" -o "$TMP_DIR/checksums.txt"; then
        print_warning "Could not download checksum file. Skipping verification."
    fi

    # Download binary
    print_step "Downloading binary..."
    print_info "URL: $DOWNLOAD_URL"
    if ! curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/kkcli.tar.gz"; then
        print_error "Failed to download binary."
        exit 1
    fi
    print_success "Download complete"
}

verify_checksum() {
    if [ ! -f "$TMP_DIR/checksums.txt" ]; then
        return 0
    fi

    print_step "Verifying checksum..."
    cd "$TMP_DIR"
    CHECKSUM_FILE="kkcli_${LATEST#v}_${OS}_${ARCH}.tar.gz"

    # Get expected checksum
    EXPECTED=$(grep "$CHECKSUM_FILE" checksums.txt | awk '{print $1}')
    if [ -z "$EXPECTED" ]; then
        print_warning "Checksum not found for $CHECKSUM_FILE"
        cd - > /dev/null
        return 0
    fi

    # Calculate actual checksum
    if command -v sha256sum &> /dev/null; then
        ACTUAL=$(sha256sum kkcli.tar.gz | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        ACTUAL=$(shasum -a 256 kkcli.tar.gz | awk '{print $1}')
    else
        print_warning "No checksum tool available. Skipping verification."
        cd - > /dev/null
        return 0
    fi

    # Compare
    if [ "$EXPECTED" = "$ACTUAL" ]; then
        print_success "Checksum verified"
    else
        print_error "Checksum mismatch!"
        print_info "Expected: $EXPECTED"
        print_info "Actual:   $ACTUAL"
        exit 1
    fi

    cd - > /dev/null
}

# ----------------------------------------------------------------------------
# Installation
# ----------------------------------------------------------------------------

install_binary() {
    print_step "Extracting archive..."
    tar -xz -f "$TMP_DIR/kkcli.tar.gz" -C "$TMP_DIR"

    print_step "Installing to $INSTALL_DIR..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
        chmod 755 "$INSTALL_DIR/$BINARY"
    else
        sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
        sudo chown root:root "$INSTALL_DIR/$BINARY"
        sudo chmod 755 "$INSTALL_DIR/$BINARY"
    fi
}

verify_installation() {
    print_step "Verifying installation..."

    if command -v $BINARY &> /dev/null; then
        print_success "Installation successful!"
        echo ""
        echo -e "${CYAN}────────────────────────────────────────────────────────────────${NC}"
        echo ""
        $BINARY --version
        echo ""
        echo -e "${CYAN}────────────────────────────────────────────────────────────────${NC}"
        echo ""
        echo -e "${BOLD}Get started:${NC}"
        echo -e "  ${GREEN}\$${NC} kk init"
        echo ""
        echo -e "${BOLD}Documentation:${NC}"
        echo -e "  https://docs.kkauto.net"
        echo ""
    else
        print_error "Installation failed. Please try again."
        exit 1
    fi
}

# ----------------------------------------------------------------------------
# Main
# ----------------------------------------------------------------------------

main() {
    print_header
    detect_platform
    get_latest_version
    download_binary
    verify_checksum
    install_binary
    verify_installation
}

main "$@"
