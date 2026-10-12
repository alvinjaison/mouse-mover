#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

# Build first.
bash build/build.sh

VERSION="${1:-portable}"
ZIP="dist/MouseMover-${VERSION}.zip"

echo "==> Packaging $ZIP..."
cd dist
# Use Go's archive/zip — if zip is available, use that instead.
if command -v zip >/dev/null 2>&1; then
  zip -j "../$ZIP" MouseMover.exe ../README.md
else
  # Fallback: create zip using python3
  python3 -c "
import zipfile, sys
with zipfile.ZipFile(sys.argv[1], 'w', zipfile.ZIP_DEFLATED) as zf:
    zf.write('MouseMover.exe')
    zf.write('../README.md', 'README.md')
" "../$ZIP"
fi
cd ..

echo "==> Packaged: $ZIP"
ls -lh "$ZIP"
