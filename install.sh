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

# ── Download to temp file ─────────────────────────────────────────────────────
TMP=$(mktemp)
curl -fsSL "$URL" -o "$TMP"
chmod +x "$TMP"

# ── Install ───────────────────────────────────────────────────────────────────
# Prefer /usr/local/bin; fall back to ~/.local/bin (no sudo, works in piped bash)
LOCAL_BIN="$HOME/.local/bin"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
  FINAL="${INSTALL_DIR}/${BINARY_NAME}"
else
  mkdir -p "$LOCAL_BIN"
  mv "$TMP" "${LOCAL_BIN}/${BINARY_NAME}"
  FINAL="${LOCAL_BIN}/${BINARY_NAME}"

  # Add ~/.local/bin to PATH in shell rc files if not already present
  for RC in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [ -f "$RC" ] && ! grep -q '\.local/bin' "$RC"; then
      echo '' >> "$RC"
      echo '# Added by gita-cli installer' >> "$RC"
      echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$RC"
    fi
  done

  echo ""
  echo "Note: installed to $LOCAL_BIN (no sudo required)"
  echo "If 'gita' is not found after install, run:  source ~/.bashrc"
fi

echo ""
echo "✓ gita installed → $FINAL"
echo ""
echo "Run:  gita"
