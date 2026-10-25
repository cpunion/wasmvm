#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
compiler="${1:-go}"
case "$compiler" in
  go)
    go version
    CGO_ENABLED=0 go test -count=1 -timeout=10m ./...
    ;;
  llgo)
    llgo version
    llgo test -count=1 -timeout=10m ./...
    ;;
  *)
    printf 'usage: %s [go|llgo]\n' "$0" >&2
    exit 2
    ;;
esac
