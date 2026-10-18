#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

echo "==> Building Mouse Mover for macOS (universal: arm64 + amd64)..."
mkdir -p dist

# Build both architecture slices. The darwin build uses cgo (Core Graphics /
# Cocoa), so cross-compiling each slice needs the matching clang -arch flag.
echo "    Building arm64 slice..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC="clang -arch arm64" \
  go build -ldflags "-s -w" -o dist/MouseMover-arm64 ./cmd/mousemover

echo "    Building amd64 slice..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="clang -arch x86_64" \
  go build -ldflags "-s -w" -o dist/MouseMover-amd64 ./cmd/mousemover

# Combine into a single universal (fat) binary.
echo "    Combining into universal binary..."
lipo -create -output dist/MouseMover dist/MouseMover-arm64 dist/MouseMover-amd64
rm -f dist/MouseMover-arm64 dist/MouseMover-amd64

echo "==> Built: dist/MouseMover"
file dist/MouseMover
ls -lh dist/MouseMover

# Create a minimal .app bundle
APP="dist/MouseMover.app"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS"
mkdir -p "$APP/Contents/Resources"
cp dist/MouseMover "$APP/Contents/MacOS/MouseMover"

cat > "$APP/Contents/Info.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>Mouse Mover</string>
    <key>CFBundleDisplayName</key>
    <string>Mouse Mover</string>
    <key>CFBundleIdentifier</key>
    <string>com.jarvis.mouse-mover</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleExecutable</key>
    <string>MouseMover</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

echo "==> App bundle: $APP"
echo ""
echo "To install: drag MouseMover.app to /Applications"
echo "NOTE: On first run, macOS will ask for Accessibility permission"
echo "      (System Settings > Privacy & Security > Accessibility)"
