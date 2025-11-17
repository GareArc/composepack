#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${COMPOSEPACK_INSTALL_DIR:-/usr/local/bin}"
TARGET="$INSTALL_DIR/composepack"

if [[ -f "$TARGET" ]]; then
  rm -f "$TARGET"
  echo "Removed $TARGET"
else
  echo "composepack not found in $INSTALL_DIR"
fi
