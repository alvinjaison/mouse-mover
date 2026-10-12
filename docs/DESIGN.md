# Mouse Mover — Requirements & Technical Design

## Part 1 — Requirements

### Summary

Mouse Mover is a lightweight Windows 10/11 desktop utility, written in Go, that
periodically performs a tiny, configurable cursor movement to keep a session
active, then returns the cursor to its original position. It runs as a standard
(non-administrator) user, supports portable execution and a per-user install
under `%LOCALAPPDATA%`, exposes Start/Stop controls through a main window and a
system tray, and persists settings as JSON. The entire application is pure Go
built with `CGO_ENABLED=0` so it cross-compiles to `windows/amd64` from the
Fedora/WSL development host (which has no C compiler and no mingw-w64).

### Functional Requirements

**FR-1 Main window**
- FR-1.1 Display the application name ("Mouse Mover") and current status
  (Running / Stopped).
- FR-1.2 Provide Start and Stop buttons; Start is disabled while running, Stop
  is disabled while stopped.
- FR-1.3 Offer movement-interval presets: 30 seconds, 1 minute, 2 minutes,
  5 minutes, and a Custom value entered in seconds.
- FR-1.4 Offer a movement-pattern selector from a fixed enum (e.g. Nudge,
  Jiggle, Square).
- FR-1.5 Display the last movement time and the elapsed running time, updated
  live while running.

**FR-2 Mouse movement engine**
- FR-2.1 Capture the current cursor coordinates before each move.
- FR-2.2 Perform a minimal, pattern-driven relative movement.
- FR-2.3 Return the cursor to its original coordinates after the movement.
- FR-2.4 Clamp every intermediate and final coordinate to the valid virtual
  screen bounds so the cursor never leaves the usable desktop.
- FR-2.5 Never move the mouse while stopped.
- FR-2.6 Work correctly across multiple monitors and mixed resolutions,
  including monitors positioned at negative virtual-screen coordinates.
- FR-2.7 Drive scheduling with a cancellable goroutine plus a `time.Ticker`.
- FR-2.8 Guarantee that a second Start while already running cannot create a
  duplicate movement loop.
- FR-2.9 Provide an immediate Stop that cancels in-flight scheduling promptly.

**FR-3 System tray**
- FR-3.1 Support minimizing the main window to the system tray.
- FR-3.2 Provide a tray menu with Show, Start, Stop, and Exit.
- FR-3.3 Indicate clearly in the tray whether movement is active (icon/tooltip
  state, and enabled/disabled menu items).
- FR-3.4 On Exit, stop the engine cleanly and release the tray icon and window
  resources before the process terminates.

**FR-4 Settings**
- FR-4.1 Persist settings to JSON under the current user's application-data
  directory: `%LOCALAPPDATA%\MouseMover\settings.json`.
- FR-4.2 Validate all interval and pattern settings before accepting or saving.
- FR-4.3 Restore saved settings on next launch; fall back to defaults when the
  file is missing or invalid.

**FR-5 No-administrator operation**
- FR-5.1 Run entirely as a standard Windows user.
- FR-5.2 Require no UAC elevation, no Windows service, and no machine-wide
  install.
- FR-5.3 Support portable execution from any user-writable folder.
- FR-5.4 Provide a per-user install option under `%LOCALAPPDATA%`.
- FR-5.5 Never write to protected directories (e.g. `Program Files`) and never
  make `HKLM` registry changes.
- FR-5.6 Request no unnecessary Windows permissions; the executable manifest
  requests `asInvoker` (no elevation).

**FR-6 Safety and transparency**
- FR-6.1 Remain visible to Task Manager and security software (no hiding, no
  process-name obfuscation).
- FR-6.2 Make no attempt to bypass enterprise security policy or access control.
- FR-6.3 Keep the running state visible in both window and tray.
- FR-6.4 Simulate only the configured mouse movement; never synthesize keyboard
  input or otherwise interfere with user input.
- FR-6.5 Provide an immediate Stop reachable from both window and tray.

**FR-7 Packaging & docs**
- FR-7.1 Produce a release `.exe` for `windows/amd64`.
- FR-7.2 Produce a portable ZIP containing the executable and README.
- FR-7.3 Provide a per-user installer that does not request elevation.
- FR-7.4 Document GUI/packaging dependencies and build/test instructions.

**FR-8 Project quality**
- FR-8.1 Clean, modular layout separating UI, engine, config, platform API, and
  packaging.
- FR-8.2 Proper cancellation and resource cleanup throughout.
- FR-8.3 Graceful handling of Windows API errors.
- FR-8.4 Unit tests for config validation and movement scheduling; the Linux-
  buildable packages (engine, config) must compile and test on the Fedora host.
- FR-8.5 Document limitations where Windows/enterprise policy prevents input
  injection.

### Non-Functional Requirements

- NFR-1 **Toolchain**: Pure Go, `CGO_ENABLED=0`, cross-compiled to
  `windows/amd64` from Linux. No Fyne, no CGO/OpenGL GUI, no mingw-w64.
- NFR-2 **Footprint**: Single self-contained `.exe` with no external runtime
  (no DLLs beyond the Windows system DLLs loaded at runtime).
- NFR-3 **Responsiveness**: Stop takes effect within one scheduling tick worst
  case and immediately cancels the context; the UI thread must never block on
  the engine.
- NFR-4 **Portability of the test suite**: `go test ./...` must pass on Linux
  for every non-`//go:build windows` package.

### Acceptance Criteria

1. `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./cmd/mousemover` succeeds
   on the Fedora/WSL host and produces a runnable `.exe`.
2. `go test ./...` passes on Linux, exercising `internal/engine` and
   `internal/config` (plus the `!windows` stubs of `internal/winapi` and
   `internal/ui`).
3. `Config.Validate` rejects an interval below the minimum, above the maximum,
   and a pattern value outside the enum, and accepts each of the five preset
   configurations.
4. Loading a missing or corrupt `settings.json` yields the documented defaults
   without error to the caller.
5. Saving then loading a config returns an equal config (round-trip).
6. Calling `Engine.Start` twice without an intervening `Stop` results in exactly
   one active goroutine/ticker (verified by a test observing move counts).
7. `Engine.Stop` cancels the context and stops the ticker; a fake cursor
   controller records no further moves after Stop returns.
8. With a fake controller, each scheduled tick records: a GetCursorPos, a move,
   and a SetCursorPos back to the captured origin, in that order.
9. Coordinates passed to the fake controller are always within the configured
   virtual-screen bounds, including when the origin sits near a boundary.
10. The engine exposes last-move time and elapsed running time, and both are
    `nil`/zero before the first run.
11. The built `.exe` manifest declares `requestedExecutionLevel level="asInvoker"`
    (no elevation prompt) — verified by inspecting the embedded manifest.
12. `build/package.sh` produces a ZIP containing `MouseMover.exe` and `README.md`.
13. The per-user installer copies the `.exe` under `%LOCALAPPDATA%\MouseMover`
    and creates a Start-menu/desktop shortcut without triggering UAC.

### Out of Scope

- macOS and Linux GUI targets (the GUI/tray and Win32 bindings are Windows-only;
  Linux is a build/test host only).
- Keyboard simulation or any input injection beyond the configured mouse move.
- Running as a service, scheduled task, or at machine scope.
- Defeating UIPI, Secure Desktop, session-lock, or group-policy restrictions.
- Auto-update, telemetry, or network features.
- Code signing (documented as a recommended follow-up, not implemented here).

---

## Part 2 — Technical Design

### Overview

Mouse Mover is a single Go module, `github.com/jarvis/mouse-mover`, producing one
executable. It is layered into platform bindings (`internal/winapi`), a
platform-agnostic scheduling engine (`internal/engine`), JSON settings
(`internal/config`), and a Windows-only GUI/tray front end (`internal/ui`), wired
together in `cmd/mousemover`. The hard constraint shaping every decision is that
the build must be **pure Go with `CGO_ENABLED=0`** so it cross-compiles from the
Fedora/WSL host, which has no C toolchain. This rules out Fyne and every
CGO/OpenGL GUI library. The GUI and tray are therefore implemented directly
against the Win32 API through `golang.org/x/sys/windows` using `syscall`-style
`LazyDLL`/`NewProc` calls — no cgo. To keep the engine and config packages
testable on Linux, all Windows-only code lives behind `//go:build windows` with
`//go:build !windows` stub counterparts, and the engine depends on a
`CursorController` interface rather than on `winapi` directly.

### Technology Stack (locked)

- **Language**: Go (module `github.com/jarvis/mouse-mover`, built with the
  installed Go 1.26.x toolchain). `go.mod` declares `go 1.23` as the minimum
  language version so the module is not pinned to a bleeding-edge toolchain.
- **Build mode**: `CGO_ENABLED=0`, `GOOS=windows`, `GOARCH=amd64`.
- **Windows API access**: `golang.org/x/sys/windows` (pure Go) plus
  `windows.NewLazySystemDLL` for `user32.dll` / `shell32.dll` procedures not
  already wrapped. **No cgo, no third-party GUI framework.**
- **GUI/tray**: hand-written Win32 window + `Shell_NotifyIcon` tray icon driven
  by a single message loop on a locked OS thread. No third-party systray
  dependency is used; if a future change needs one it MUST build with
  `CGO_ENABLED=0` for `windows/amd64` (this excludes `getlantern/systray` and
  any OpenGL-based toolkit).
- **Settings format**: JSON via the standard-library `encoding/json`.
- **Testing**: standard-library `testing`.

### Module & Package Layout

```
github.com/jarvis/mouse-mover
├── go.mod                       module github.com/jarvis/mouse-mover
├── go.sum
├── cmd/
│   └── mousemover/
│       ├── main.go              //go:build windows  — real entry point
│       └── main_stub.go         //go:build !windows — prints "windows only", exit 0
├── internal/
│   ├── winapi/                  CGO-free Win32 bindings
│   │   ├── winapi.go            shared types/consts (build-tag free)
│   │   ├── cursor_windows.go    //go:build windows  — GetCursorPos/SetCursorPos/SendInput/metrics
│   │   └── cursor_stub.go       //go:build !windows — stub returning ErrUnsupported
│   ├── engine/                  platform-agnostic scheduler (builds on all OSes)
│   │   ├── engine.go
│   │   ├── controller.go        CursorController interface + Point/Rect types
│   │   ├── pattern.go           movement patterns
│   │   └── engine_test.go       Linux-runnable unit tests with a fake controller
│   ├── config/                  JSON settings (builds on all OSes)
│   │   ├── config.go
│   │   ├── paths_windows.go     //go:build windows  — %LOCALAPPDATA% resolution
│   │   ├── paths_other.go       //go:build !windows — XDG/temp fallback for tests
│   │   └── config_test.go       validation + round-trip tests
│   └── ui/                      Win32 window + tray
│       ├── ui_windows.go        //go:build windows  — window, controls, message loop
│       ├── tray_windows.go      //go:build windows  — Shell_NotifyIcon tray
│       └── ui_stub.go           //go:build !windows — Run() returns ErrUnsupported
├── build/
│   ├── build.sh                 cross-compile the release .exe
│   ├── package.sh               portable ZIP (exe + README)
│   └── install-user.ps1         non-elevated per-user installer
├── docs/
│   └── DESIGN.md                this document
└── README.md
```

`internal/engine` and `internal/config` have **no** Windows-only imports in their
build-tag-free files, so `go test ./...` runs them on Linux. `winapi` and `ui`
compile on Linux only through their `!windows` stubs.

### `internal/winapi` — CGO-free Win32 bindings

The package exposes a small surface used by the UI to construct a real cursor
controller and to query screen metrics. Shared, OS-independent declarations live
in `winapi.go`:

```go
package winapi

type Point struct{ X, Y int32 }
type Rect  struct{ Left, Top, Right, Bottom int32 } // half-open: Right/Bottom exclusive
```

`cursor_windows.go` (`//go:build windows`) binds the following with
`windows.NewLazySystemDLL` + `.NewProc`:

- `user32.GetCursorPos(*POINT) BOOL` → `GetCursorPos() (Point, error)`.
- `user32.SetCursorPos(int, int) BOOL` → `SetCursorPos(p Point) error`.
- `user32.SendInput(uint, *INPUT, int) UINT` → `SendMouseMove(dx, dy int32, absolute bool) error`,
  packing a single `INPUT` with `type=INPUT_MOUSE` and `MOUSEINPUT`
  `dwFlags = MOUSEEVENTF_MOVE` (relative) or `MOUSEEVENTF_MOVE|MOUSEEVENTF_ABSOLUTE|MOUSEEVENTF_VIRTUALDESK`
  for absolute positioning in the virtual-screen coordinate space.
- Virtual-screen metrics via `user32.GetSystemMetrics(int) int`:
  `VirtualScreenRect() Rect` built from `SM_XVIRTUALSCREEN (76)`,
  `SM_YVIRTUALSCREEN (77)`, `SM_CXVIRTUALSCREEN (78)`, `SM_CYVIRTUALSCREEN (79)`.

**Positioning decision.** Two mechanisms exist for cursor placement: `SendInput`
and `SetCursorPos`. The requirement explicitly names `SendInput` (synthetic input
is what keeps a session "active"; `SetCursorPos` alone may not reset idle timers).
Therefore the engine's **movement** step uses `SendInput`, and the **restore**
step also uses `SendInput` in absolute/virtual-desktop mode to return to the exact
origin pixel. `SetCursorPos` is bound and exported as a fallback used only if
`SendInput` reports zero events inserted; this fallback is logged. Absolute
`SendInput` coordinates are normalized to the `0..65535` range **relative to the
virtual screen** so multi-monitor and negative-origin layouts resolve correctly:
`nx = round((x - vLeft) * 65535 / (vWidth - 1))`, likewise for `ny`.

**Error handling in winapi.** Each wrapped call checks the BOOL/UINT return.
On failure it wraps `windows.GetLastError()` into a Go `error`
(`fmt.Errorf("GetCursorPos: %w", err)`). These errors are **recoverable** at the
engine level (a single tick fails, is logged at WARN, and the loop continues);
they are never fatal to the process. `SendInput` returning fewer events than
requested is treated as a failure for that tick. The package defines
`var ErrUnsupported = errors.New("winapi: not supported on this platform")`
returned by every stub function in `cursor_stub.go` so the `!windows` build links.

`winapi` provides a constructor `NewCursorController() engine.CursorController`
(windows build) adapting the raw calls to the engine interface. To avoid an
import cycle, the interface is **defined in `engine`** and `winapi` imports
`engine` (one-way). The `!windows` stub constructor is not needed because the
engine is only wired to a real controller from `cmd/mousemover` under the
`windows` build; Linux tests inject a fake implementing the same interface.

### `internal/engine` — cancellable scheduler

The engine owns scheduling, the running guard, timing bookkeeping, and move
orchestration. It depends only on an interface, so it is fully unit-testable on
Linux.

```go
package engine

type Point struct{ X, Y int32 }
type Rect  struct{ Left, Top, Right, Bottom int32 } // Right/Bottom exclusive

type CursorController interface {
    GetPos() (Point, error)
    MoveTo(p Point) error          // absolute placement (SendInput abs, SetCursorPos fallback)
    VirtualScreen() (Rect, error)  // current virtual-screen bounds
}

type Pattern int
const (
    PatternNudge  Pattern = iota // move +Δ then back
    PatternJiggle                // +Δ, -Δ, back to origin
    PatternSquare                // trace a tiny square, back to origin
)

type Status struct {
    Running   bool
    LastMove  time.Time   // zero if never
    StartedAt time.Time   // zero if stopped
    Elapsed   time.Duration
    MoveCount uint64
}
```

`Engine` struct fields: `mu sync.Mutex`, `running bool`, `cancel context.CancelFunc`,
`done chan struct{}`, `ctrl CursorController`, `interval time.Duration`,
`pattern Pattern`, `delta int32` (pixels, default 2), `lastMove time.Time`,
`startedAt time.Time`, `moveCount uint64`, and an injectable `now func() time.Time`
and `newTicker func(time.Duration) *time.Ticker` (defaults to `time.Now` /
`time.NewTicker`) so tests control time without real sleeps. A `logger` field
(a small `Logger` interface with `Warnf`) lets tests capture logs.

**Start (duplicate-guard).** `Start(interval, pattern)` takes `mu`, and if
`running` is already true it returns immediately (idempotent no-op) — this is the
guarantee against duplicate loops (AC-6). Otherwise it validates nothing itself
(config validation happens in `config`; the UI only ever passes validated
values, and the engine defensively clamps interval to `>= time.Second`), sets
`running = true`, records `startedAt = now()`, creates
`ctx, cancel := context.WithCancel(context.Background())`, a fresh
`done := make(chan struct{})`, stores `cancel`/`done`, then launches exactly one
`go e.loop(ctx, done, interval, pattern)` and releases `mu`.

**loop.** Creates `t := e.newTicker(interval)`, `defer t.Stop()`,
`defer close(done)`. Selects on `ctx.Done()` (return) and `t.C` (call
`e.tick(pattern)`). `tick` reads the virtual-screen rect, captures origin via
`GetPos`, computes the pattern's point sequence, clamps every point to the rect
(see below), issues `MoveTo` for each step, then `MoveTo(origin)` to restore. On
any controller error the tick logs WARN and returns without aborting the loop.
On success it takes `mu`, sets `lastMove = now()`, increments `moveCount`,
releases `mu`.

**Stop (immediate).** `Stop()` takes `mu`; if not running, returns. Otherwise
calls `cancel()`, clears `running`, grabs the `done` channel, releases `mu`, then
`<-done` with a short guard so Stop blocks only until the loop goroutine exits.
Because `cancel()` fires before the next tick is processed and the loop selects on
`ctx.Done()`, Stop is immediate (AC-7). After Stop, `startedAt` is zeroed.

**Clamping (FR-2.4 / AC-9).** Given virtual rect `R` (half-open) and a desired
point `p`, clamp: `x = min(max(p.X, R.Left), R.Right-1)`,
`y = min(max(p.Y, R.Top), R.Bottom-1)`. Clamping applies to each intermediate
pattern point *and* to the restore target, so even a corner origin stays on
screen. The default `delta` of 2px keeps movement minimal and nearly invisible.

**Status/timing (FR-1.5 / AC-10).** `Status()` takes `mu` and returns a `Status`
snapshot; `Elapsed = now() - startedAt` when running else zero; `LastMove` is zero
until the first successful tick. The UI polls `Status()` on a 1s timer to refresh
labels — the engine never calls into the UI (keeps layering one-directional and
avoids UI-thread/engine deadlocks).

**Testability.** With a fake `CursorController` recording calls plus the injected
`now`/`newTicker`, tests cover: single-loop-on-double-Start, no-moves-after-Stop,
tick call ordering (GetPos → MoveTo steps → MoveTo origin), boundary clamping,
and WARN-on-controller-error without loop death. All run on Linux, no real time
elapses.

### `internal/config` — JSON settings

```go
package config

type Config struct {
    IntervalSeconds int    `json:"interval_seconds"`
    Pattern         string `json:"pattern"`      // "nudge" | "jiggle" | "square"
    DeltaPixels     int    `json:"delta_pixels"` // 1..20, default 2
    StartMinimized  bool   `json:"start_minimized"`
    AutoStart       bool   `json:"auto_start"`   // begin moving on launch
}

const (
    MinIntervalSeconds = 5
    MaxIntervalSeconds = 24 * 60 * 60 // 86400
    MinDelta           = 1
    MaxDelta           = 20
)

func Default() Config // {IntervalSeconds:60, Pattern:"nudge", DeltaPixels:2}
```

**Validation (FR-4.2 / AC-3).** `Validate() error` returns a descriptive error
when: `IntervalSeconds < MinIntervalSeconds` or `> MaxIntervalSeconds`;
`Pattern` not in the enum set; `DeltaPixels` out of `[MinDelta,MaxDelta]`. The
five UI presets (30/60/120/300 and a custom value in range) all pass. Errors are
returned to the UI, which shows them inline and refuses to Start/Save; invalid
input is **recoverable** (user corrects it), never fatal.

**Paths.** `paths_windows.go` resolves the settings path from
`os.Getenv("LOCALAPPDATA")`, falling back to
`%USERPROFILE%\AppData\Local` if the env var is empty, then
`filepath.Join(base, "MouseMover", "settings.json")`. It creates the
`MouseMover` directory with `os.MkdirAll(dir, 0o755)` (user-writable, no admin).
`paths_other.go` (`//go:build !windows`) returns a temp/XDG path so config tests
run on Linux.

**Load (FR-4.3 / AC-4).** `Load() (Config, error)`:
- File missing (`os.IsNotExist`) → return `Default(), nil` (not an error).
- Read or `json.Unmarshal` error, or post-unmarshal `Validate()` failure →
  log WARN, return `Default(), nil`. A corrupt file never propagates an error to
  the caller and never crashes; defaults are used (AC-4). The corrupt file is
  left untouched (not deleted) for user inspection.

**Save (FR-4.1 / AC-5).** `Save(c Config) error`: `Validate()` first (return the
error if invalid, do not write), `MkdirAll` the dir, marshal with
`json.MarshalIndent`, write to a temp file in the same dir, then `os.Rename` over
`settings.json` (atomic replace to avoid truncation on crash). Returns the IO
error if any step fails — Save failure is surfaced to the UI as a non-fatal
message. Round-trip `Save` then `Load` yields an equal config (AC-5).

### `internal/ui` — Win32 window + system tray (CGO-free)

All UI code is `//go:build windows`; `ui_stub.go` provides
`func Run(cfg config.Config, eng *engine.Engine) error { return ErrUnsupported }`
for the `!windows` build so the module links on Linux.

**Threading model.** Win32 requires that window creation and the message pump
run on one thread, and that thread must be the one that created the window.
`Run` calls `runtime.LockOSThread()` at the top (never unlocks — the process
exits when the loop ends). A single message loop (`GetMessage`/`TranslateMessage`/
`DispatchMessage`) on this locked thread drives both the main window and the tray
icon. The engine runs on its own goroutine(s); the UI and engine communicate only
through the engine's thread-safe methods (`Start`/`Stop`/`Status`) and never share
Win32 handles across goroutines.

**Window construction.** Using `user32.dll` procedures via `NewLazySystemDLL`:
`RegisterClassExW` (with a `WNDPROC` callback created via
`windows.NewCallback`), `CreateWindowExW` for the main window and for each child
control (static labels, buttons, a combo box for interval presets, an edit box
for custom seconds, a combo box for pattern). Controls are standard Win32
`BUTTON`/`STATIC`/`COMBOBOX`/`EDIT` classes — no custom rendering, no OpenGL.
`windows.NewCallback` is pure Go (no cgo) and is the sanctioned way to pass a Go
function as a Win32 callback.

**Controls → requirements.**
- App name + status label (FR-1.1): two `STATIC` controls; status text updated
  from a 1s `WM_TIMER` that reads `eng.Status()`.
- Start/Stop buttons (FR-1.2): `BUTTON` controls; `WM_COMMAND` handler calls
  `eng.Start(...)` / `eng.Stop()`; enabled state toggled from `Status().Running`.
- Interval presets + custom (FR-1.3): `COMBOBOX` with the four presets plus
  "Custom"; selecting Custom enables the seconds `EDIT`. On Start the UI builds a
  `Config`, calls `cfg.Validate()`, and on error shows a `MessageBoxW` and aborts
  Start.
- Pattern selector (FR-1.4): `COMBOBOX` bound to the pattern enum.
- Last-move + elapsed (FR-1.5): two `STATIC` labels refreshed by the `WM_TIMER`.

**Tray (FR-3) via `shell32.Shell_NotifyIcon`.** Build a `NOTIFYICONDATA`
(`cbSize`, `hWnd`, `uID`, `uFlags = NIF_ICON|NIF_MESSAGE|NIF_TIP`,
`uCallbackMessage = WM_APP+1`, `hIcon`, tooltip). `NIM_ADD` on startup,
`NIM_MODIFY` to update icon/tooltip when state changes, `NIM_DELETE` on exit.
The callback message arrives in the same WNDPROC; on `WM_RBUTTONUP` the handler
builds a popup `HMENU` (`CreatePopupMenu`/`AppendMenuW`) with **Show, Start, Stop,
Exit**, calls `SetForegroundWindow(hWnd)` then `TrackPopupMenu`, and routes the
chosen command id through the same `WM_COMMAND` logic. Start/Stop menu items are
greyed via `MF_GRAYED` to match `Status().Running` (FR-3.3). The tray icon and
tooltip reflect active/inactive state (e.g. tooltip "Mouse Mover — Running"
vs "Stopped"), updated whenever state changes and on the 1s timer.

**Minimize to tray (FR-3.1).** The WNDPROC intercepts `WM_SYSCOMMAND` with
`SC_MINIMIZE` (and the close button → `WM_CLOSE`) to `ShowWindow(hWnd, SW_HIDE)`
instead of taskbar-minimizing, keeping the app reachable only from the tray. The
tray "Show" command and a double-click (`WM_LBUTTONDBLCLK`) call
`ShowWindow(SW_SHOW)` + `SetForegroundWindow`.

**Clean exit (FR-3.4 / FR-8.2).** "Exit" (menu) and real quit:
`eng.Stop()` (blocks until the loop goroutine exits), `Shell_NotifyIcon(NIM_DELETE)`,
`DestroyWindow`, then `PostQuitMessage(0)` so `GetMessage` returns 0 and the loop
unwinds. Because the window-close button maps to hide (not quit), the only quit
paths are tray Exit and `WM_DESTROY`, both of which stop the engine first.

**Icons.** A minimal application icon is embedded via a `.syso` resource compiled
into the binary (see Packaging) so no icon file ships separately; `hIcon` is
loaded with `LoadIconW`/`LoadImageW` from the module's own instance handle. If the
icon resource is absent, the code falls back to `LoadIconW(0, IDI_APPLICATION)`
so the tray never fails to appear.

**UI error handling.** All Win32 calls that can fail (`CreateWindowExW == 0`,
`RegisterClassExW == 0`, `Shell_NotifyIcon` returning FALSE) are checked:
registration/window-creation failures are **fatal** (return an error from `Run`
that `main` logs and exits non-zero — the app cannot function without a window);
tray `NIM_ADD` failure is **recoverable** (log WARN, continue window-only). All
such events are logged to a log file under `%LOCALAPPDATA%\MouseMover\log.txt`
(best-effort; logging failure is swallowed) and never silently ignored.

### `cmd/mousemover` — composition root

`main.go` (`//go:build windows`): `config.Load()` → construct a `winapi`
cursor controller → `engine.New(ctrl, logger)` → if `cfg.AutoStart` call
`eng.Start(cfg.Interval(), cfg.PatternEnum())` → `ui.Run(cfg, eng)`. A top-level
`defer eng.Stop()` guarantees the engine is stopped even on an unexpected `Run`
return. Fatal errors from `Run` are logged and the process exits with code 1.
`main_stub.go` (`//go:build !windows`) prints "Mouse Mover is Windows-only" and
exits 0, so `go run ./cmd/mousemover` on Linux is harmless and `go vet ./...`
passes.

### Build, packaging & install (`build/`)

**`build.sh`** (run inside WSL): sets
`CGO_ENABLED=0 GOOS=windows GOARCH=amd64` and runs
`go build -ldflags "-H windowsgui -s -w" -o dist/MouseMover.exe ./cmd/mousemover`.
`-H windowsgui` produces a GUI subsystem binary (no console window). An embedded
manifest requesting `asInvoker` and an application icon are provided through a
`cmd/mousemover/rsrc_windows_amd64.syso` resource. **Decision on resource
generation:** rather than depending on the C-based `windres` (unavailable, no
mingw), the `.syso` is produced by the pure-Go tool
`github.com/akavel/rsrc` (or `github.com/tc-hib/go-winres`), run once on the
Linux host (`go run github.com/akavel/rsrc -manifest app.manifest -ico app.ico -o cmd/mousemover/rsrc_windows_amd64.syso`),
and the generated `.syso` is committed so routine builds need no extra tool. The
`app.manifest` declares `requestedExecutionLevel level="asInvoker" uiAccess="false"`
(FR-5.6, AC-11) and the Windows 10/11 `compatibility` GUIDs.

**`package.sh`**: builds via `build.sh`, then zips `MouseMover.exe` + `README.md`
into `dist/MouseMover-portable-<version>.zip` using Go's `archive/zip` through a
tiny helper (`go run ./build/zip`) so no `zip` binary is required on the host.

**`install-user.ps1`** (per-user, non-elevated — FR-5.4, AC-13): copies
`MouseMover.exe` to `$env:LOCALAPPDATA\MouseMover\`, creates a Start-menu shortcut
in `$env:APPDATA\Microsoft\Windows\Start Menu\Programs\` via `WScript.Shell`, and
optionally a desktop shortcut. It writes nothing to `Program Files` or `HKLM`,
sets no `#Requires -RunAsAdministrator`, and prints uninstall instructions
(delete the folder and shortcut). Uninstall is a documented manual step (remove
`%LOCALAPPDATA%\MouseMover` and the shortcut) — no registry uninstall entry is
created, keeping the footprint purely per-user.

### No-admin compliance (FR-5 / FR-6)

Nothing in the design requires elevation: all writes target `%LOCALAPPDATA%`
(settings, log) or the portable folder; the manifest requests `asInvoker`; there
is no service, scheduled task, driver, or `HKLM` access. The app only calls
documented user-level `user32`/`shell32` APIs. It does not rename its process,
inject into others, or hook global input — it is fully visible to Task Manager and
AV (FR-6.1/6.2). Only mouse movement is synthesized, never keyboard input
(FR-6.4), and Stop is reachable from window and tray at all times (FR-6.5).

### Error-handling summary (per operation)

| Operation | Failure condition | Recoverable? | Caller receives | Logged |
|---|---|---|---|---|
| `winapi.GetPos/MoveTo/SendMouseMove` | BOOL/UINT fail (`GetLastError`) | yes (tick-local) | wrapped `error` | WARN |
| `engine.tick` | controller error mid-tick | yes (skip tick) | — (loop continues) | WARN |
| `engine.Start` when running | duplicate start | yes (no-op) | silently ignored | DEBUG |
| `config.Load` missing file | `os.IsNotExist` | yes | `Default(), nil` | — |
| `config.Load` corrupt/invalid | unmarshal/validate fail | yes | `Default(), nil` | WARN |
| `config.Save` invalid | `Validate` fails | yes | validation `error` | — |
| `config.Save` IO | write/rename fail | yes | IO `error` | WARN |
| `ui` RegisterClass/CreateWindow | returns 0 | **fatal** | `error` from `Run` | ERROR |
| `ui` tray `NIM_ADD` | returns FALSE | yes | — (window-only) | WARN |

### Input validation summary (external inputs)

| Input | Source | Required | Type/limits | On failure |
|---|---|---|---|---|
| interval preset | combo box | yes | one of 30/60/120/300s | default to 60s |
| custom interval | edit box | only if Custom | int 5..86400 s | MessageBox, abort Start |
| movement pattern | combo box | yes | enum nudge/jiggle/square | default nudge |
| delta pixels | settings file | no | int 1..20 | clamp/default to 2 |
| `settings.json` | disk | no | valid JSON + Validate | use defaults |
| `LOCALAPPDATA` | env | no | non-empty path | fall back to `%USERPROFILE%\AppData\Local` |

### Invariants & ownership

- **Single active loop** — owned by `engine` via `mu` + `running` guard. Only the
  engine mutates `running`; the UI never starts goroutines.
- **Cursor stays on the virtual screen** — owned by `engine.tick` clamping; the
  UI and winapi do no clamping, keeping one source of truth.
- **No move while stopped** — owned by `engine`: the loop goroutine only exists
  between Start and Stop, and `cancel()` precedes any further tick processing.
- **All writes are user-scoped** — owned by `config.paths` + installer; no other
  package chooses write locations.
- **One UI thread** — owned by `ui.Run` via `LockOSThread`; all Win32 handle use
  stays on that thread.

### Testability

- **Unit (Linux, `go test ./...`)**: `engine` with a fake `CursorController` and
  injected clock/ticker covers scheduling, duplicate-Start, immediate-Stop,
  tick ordering, boundary clamping, and error-skip (AC-6..10). `config` covers
  validation of all presets and out-of-range values, missing/corrupt-file Load,
  and Save→Load round-trip (AC-3..5). The `!windows` stubs make the whole module
  build and vet on Linux (AC-2).
- **Integration (real Windows, manual/CI-on-Windows)**: cross-compiled `.exe`
  verified for window + tray appearance, Start/Stop, minimize-to-tray, clean
  Exit, and actual cursor nudging via `SendInput`. These are **not** runnable on
  the Linux host; the README documents the manual checklist. The manifest's
  `asInvoker` level (AC-11) is verifiable by inspecting the built binary.
- A design note: because the only truly untestable-on-Linux parts are the thin
  Win32 wrappers and the message loop, the untestable surface is deliberately
  minimized — all logic lives in the Linux-testable `engine`/`config`.

### Limitations (FR-8.5)

- **UIPI / integrity levels**: synthetic input from a medium-integrity process
  cannot reach higher-integrity windows; movement may have no effect over an
  elevated window in the foreground.
- **Secure Desktop / UAC prompt / Ctrl-Alt-Del / lock screen**: `SendInput` does
  not reach the Secure Desktop; movement pauses there and resumes afterward. On a
  locked session the cursor is not visible/affected.
- **Group policy / endpoint-security / RDP-without-session**: enterprise policy or
  anti-cheat/EDR software may block or flag synthetic input; the app does not and
  must not attempt to bypass these (FR-6.2).
- **Idle-timer semantics**: using `SendInput` maximizes the chance the OS/session
  treats the movement as genuine activity, but some corporate idle detectors key
  off other signals and may ignore it.
- **Verification scope**: interactive GUI, tray behavior, and real `SendInput`
  effects are only verifiable on actual Windows; CI on the Linux host validates
  compilation and all engine/config logic but cannot assert end-to-end input
  injection.
- **No code signing**: the unsigned `.exe` may trigger SmartScreen on first run;
  signing is a recommended follow-up and out of scope here.

### Assumptions

- The target is `windows/amd64`; `arm64` is not built (can be added later with the
  same source via `GOARCH=arm64`).
- `golang.org/x/sys/windows` is an acceptable dependency (pure Go, standard for
  Win32 access without cgo); it will be added to `go.mod`.
- The pure-Go `.syso` generator (`akavel/rsrc` or `tc-hib/go-winres`) is run once
  on the Linux host and its output committed, so day-to-day builds need no C tools.
