# Mouse Mover

A lightweight Windows 10/11 desktop utility that periodically moves the mouse cursor by a tiny amount to keep your session active, then returns it to the original position. Runs as a standard user — no administrator privileges required.

## Features

- **Start/Stop** control from a main window and system tray
- **Configurable intervals**: 30s, 1m, 2m, 5m, or custom (5–86400s)
- **Movement patterns**: Nudge, Jiggle, Square
- **System tray** with minimize-to-tray, Show/Start/Stop/Exit menu
- **Persistent settings** saved as JSON under `%LOCALAPPDATA%\MouseMover\`
- **Multi-monitor aware**: clamps cursor to virtual-screen bounds
- **No admin**: no UAC, no service, no machine-wide install
- **Portable**: run from any folder, or install per-user

## Build

The project cross-compiles from Linux (WSL Fedora) to Windows. Go 1.23+ required.

```bash
# From the project root (inside WSL):
export PATH=$PATH:/usr/local/go/bin:/usr/lib/golang/bin

# Run tests (Linux):
go test ./...

# Build the Windows exe:
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -ldflags "-H windowsgui -s -w" \
  -o dist/MouseMover.exe ./cmd/mousemover

# Or use the build script:
bash build/build.sh

# Package portable ZIP:
bash build/package.sh
```

## Usage

### Portable

Copy `MouseMover.exe` to any user-writable folder and double-click to run.

### Per-User Install

From PowerShell (no admin needed):

```powershell
powershell -ExecutionPolicy Bypass -File build\install-user.ps1
```

This copies the exe to `%LOCALAPPDATA%\MouseMover\` and creates a Start Menu shortcut.

### Using the App

1. Select an interval (or enter a custom value in seconds)
2. Choose a movement pattern
3. Click **Start**
4. The app moves the cursor by a few pixels at each interval, then restores it
5. Minimize to tray — the app keeps running in the background
6. Click **Stop** or use the tray menu to stop

## Uninstall

1. Delete `%LOCALAPPDATA%\MouseMover\` (exe + settings)
2. Delete the Start Menu shortcut: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Mouse Mover.lnk`

## Project Structure

```
cmd/mousemover/       Entry point (Windows exe + Linux stub)
internal/
  engine/             Cancellable movement scheduler (platform-agnostic)
  config/             JSON settings + validation
  winapi/             CGO-free Win32 cursor bindings (SendInput, GetCursorPos, etc.)
  ui/                 Native Win32 window + system tray
build/                Build, package, and install scripts
docs/                 Design documentation
```

## Troubleshooting

- **Mouse doesn't move**: Some enterprise policies, UIPI (User Interface Privilege Isolation), Secure Desktop (UAC prompt, lock screen), or group policy settings can block synthetic input injection. Mouse Mover cannot bypass these by design.
- **No tray icon**: If the system tray is full, Windows may hide the icon. Check the overflow area.
- **Settings not saved**: Ensure `%LOCALAPPDATA%\MouseMover\` is writable.

## Limitations

- **Windows only**: The GUI, tray, and cursor movement use Win32 APIs.
- **UIPI**: Cannot inject input into elevated (admin) windows from a standard-user process.
- **Secure Desktop**: Does not work on the lock screen, UAC prompt, or Ctrl+Alt+Del screen.
- **Enterprise policy**: Group Policy may restrict `SendInput` for standard users.
- **Session lock**: Mouse movement has no effect while the session is locked.
- **No keyboard simulation**: Only mouse movement is performed; keyboard input is never synthesized.

## License

MIT
