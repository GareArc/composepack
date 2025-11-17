#!/usr/bin/env bash
set -euo pipefail

REPO="${COMPOSEPACK_REPO:-GareArc/composepack}"
VERSION="${1:-latest}"
INSTALL_DIR="${COMPOSEPACK_INSTALL_DIR:-/usr/local/bin}"

if [[ "$VERSION" == "latest" ]]; then
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required to determine latest release" >&2
    exit 1
  fi
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required to parse GitHub API responses" >&2
    exit 1
  fi
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | python3 -c "import sys,json; print(json.load(sys.stdin)['tag_name'])")
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)
    ARCH=amd64
    ;;
  arm64|aarch64)
    ARCH=arm64
    ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

ASSET="composepack-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fSL "$URL" -o "$TMP_DIR/composepack"
chmod +x "$TMP_DIR/composepack"
install -m 0755 "$TMP_DIR/composepack" "$INSTALL_DIR/composepack"

echo "composepack ${VERSION} installed to ${INSTALL_DIR}/composepack"
