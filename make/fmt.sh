#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd -- "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root_dir"

mapfile -t gofiles < <(find . -name '*.go' -not -path './vendor/*')
if [[ ${#gofiles[@]} -gt 0 ]]; then
    gofmt -w "${gofiles[@]}"
fi
