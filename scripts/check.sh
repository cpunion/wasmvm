#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
compiler="${1:-go}"
case "$compiler" in
  go)
    go version
    python3 scripts/core2_corpus.py
    python3 scripts/core3_corpus.py
    CGO_ENABLED=0 go test -count=1 -timeout=10m ./...
    ;;
  llgo)
    llgo version
    python3 scripts/core2_corpus.py
    python3 scripts/core3_corpus.py
    llgo test -O3 -lto=full -count=1 -timeout=10m ./...
    ;;
  *)
    printf 'usage: %s [go|llgo]\n' "$0" >&2
    exit 2
    ;;
esac
