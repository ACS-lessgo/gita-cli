#!/bin/bash
set -e

REPO="ACS-lessgo/gita-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="gita"

# ── Detect OS and arch ────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ASSET="gita-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading gita-cli for ${OS}/${ARCH}..."
echo "URL: $URL"

# ── Download to temp file ─────────────────────────────────────────────────────
TMP=$(mktemp)
curl -fsSL "$URL" -o "$TMP"
chmod +x "$TMP"

# ── Install (use sudo only if needed) ────────────────────────────────────────
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo ""
echo "✓ gita installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "Run:  gita"
