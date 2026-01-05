#!/bin/bash
set -e

# KK CLI Installer
# Usage: curl -sSL https://get.kkengine.com/cli | bash

REPO="kkauto-net/kk-install"
BINARY="kk"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
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
        echo "Kien truc khong ho tro: $ARCH"
        exit 1
        ;;
esac

# Get latest release
echo "Dang kiem tra phien ban moi nhat..."

# Use jq if available, fallback to grep/sed with validation
if command -v jq &> /dev/null; then
    LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.tag_name')
else
    LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

if [ -z "$LATEST" ]; then
    echo "Khong tim thay phien ban. Vui long kiem tra ket noi mang."
    exit 1
fi

# Validate version format (must be vX.Y.Z)
if [[ ! "$LATEST" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Dinh dang phien ban khong hop le: $LATEST"
    exit 1
fi

echo "Phien ban moi nhat: $LATEST"

# Download URLs
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/kkcli_${LATEST#v}_${OS}_${ARCH}.tar.gz"
CHECKSUM_URL="https://github.com/$REPO/releases/download/$LATEST/checksums.txt"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download checksum file
echo "Dang kiem tra checksum..."
if ! curl -sL "$CHECKSUM_URL" -o "$TMP_DIR/checksums.txt"; then
    echo "Warning: Khong tai duoc checksum file. Bo qua kiem tra checksum."
fi

# Download binary
echo "Dang tai tu: $DOWNLOAD_URL"
if ! curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/kkcli.tar.gz"; then
    echo "Tai binary that bai."
    exit 1
fi

# Verify checksum if available
if [ -f "$TMP_DIR/checksums.txt" ]; then
    cd "$TMP_DIR"
    CHECKSUM_FILE="kkcli_${LATEST#v}_${OS}_${ARCH}.tar.gz"

    if command -v sha256sum &> /dev/null; then
        # Get expected checksum from file
        EXPECTED=$(grep "$CHECKSUM_FILE" checksums.txt | awk '{print $1}')
        if [ -z "$EXPECTED" ]; then
            echo "Warning: Khong tim thay checksum cho $CHECKSUM_FILE trong file checksums.txt"
        else
            # Calculate actual checksum
            ACTUAL=$(sha256sum kkcli.tar.gz | awk '{print $1}')
            if [ "$EXPECTED" = "$ACTUAL" ]; then
                echo "Checksum hop le."
            else
                echo "Checksum khong khop!"
                echo "Expected: $EXPECTED"
                echo "Actual: $ACTUAL"
                exit 1
            fi
        fi
    elif command -v shasum &> /dev/null; then
        # Get expected checksum from file
        EXPECTED=$(grep "$CHECKSUM_FILE" checksums.txt | awk '{print $1}')
        if [ -z "$EXPECTED" ]; then
            echo "Warning: Khong tim thay checksum cho $CHECKSUM_FILE trong file checksums.txt"
        else
            # Calculate actual checksum
            ACTUAL=$(shasum -a 256 kkcli.tar.gz | awk '{print $1}')
            if [ "$EXPECTED" = "$ACTUAL" ]; then
                echo "Checksum hop le."
            else
                echo "Checksum khong khop!"
                echo "Expected: $EXPECTED"
                echo "Actual: $ACTUAL"
                exit 1
            fi
        fi
    else
        echo "Warning: Khong tim thay sha256sum hoac shasum. Bo qua kiem tra checksum."
    fi
    cd - > /dev/null
fi

# Extract after verification
tar -xz -f "$TMP_DIR/kkcli.tar.gz" -C "$TMP_DIR"

# Install
echo "Dang cai dat..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
    chmod 755 "$INSTALL_DIR/$BINARY"
else
    sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
    sudo chown root:root "$INSTALL_DIR/$BINARY"
    sudo chmod 755 "$INSTALL_DIR/$BINARY"
fi

# Verify
if command -v $BINARY &> /dev/null; then
    echo ""
    echo "Cai dat thanh cong!"
    echo ""
    $BINARY --version
    echo ""
    echo "Bat dau su dung: kk init"
else
    echo "Cai dat that bai. Vui long thu lai."
    exit 1
fi
