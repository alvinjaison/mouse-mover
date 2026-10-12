#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_DIR"

echo "==> Building Mouse Mover for windows/amd64 (CGO_ENABLED=0)..."
mkdir -p dist

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -ldflags "-H windowsgui -s -w" \
  -o dist/MouseMover.exe ./cmd/mousemover

echo "==> Built: dist/MouseMover.exe"
ls -lh dist/MouseMover.exe
