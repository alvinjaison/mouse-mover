#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

echo "==> Building Mouse Mover for macOS..."
mkdir -p dist

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    GOARCH="arm64"
elif [ "$ARCH" = "x86_64" ]; then
    GOARCH="amd64"
else
    GOARCH="arm64"
fi

echo "    Architecture: $GOARCH"

CGO_ENABLED=1 GOOS=darwin GOARCH=$GOARCH \
  go build -ldflags "-s -w" \
  -o dist/MouseMover ./cmd/mousemover

echo "==> Built: dist/MouseMover"
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
