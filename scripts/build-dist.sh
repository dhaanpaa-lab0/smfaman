#!/usr/bin/env bash
set -euo pipefail

if ! command -v goreleaser >/dev/null 2>&1; then
  echo "goreleaser not found. Install it from https://goreleaser.com/install/" >&2
  exit 1
fi

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

goreleaser release --snapshot --clean
