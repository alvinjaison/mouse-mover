#!/usr/bin/env bash
set -euo pipefail
cd /home/jarvis/repos/mouse-mover

# Stage everything except dist/ and temp files
git add -A
git reset HEAD dist/ 2>/dev/null || true
git status
echo "---"
git -c user.name="jarvis" -c user.email="jarvis@local" commit -m "feat: Mouse Mover for Windows and macOS

- Cancellable movement engine with duration auto-stop and configurable interval
- CGO-free Win32 GUI with native window, system tray, and Shell_NotifyIcon
- macOS menu bar app using Core Graphics and Cocoa (CGO)
- JSON settings persistence (%LOCALAPPDATA% on Windows, ~/Library/Application Support on macOS)
- Multi-monitor support with virtual-screen clamping
- Movement patterns: Nudge, Jiggle, Square
- Custom Duration and Custom Interval UI options
- Per-user installer (no admin), portable ZIP packaging
- Unit tests for engine scheduling and config validation
- Cross-compile from WSL Fedora: CGO_ENABLED=0 GOOS=windows GOARCH=amd64"
echo "COMMIT_RC=$?"
git log --oneline -3
