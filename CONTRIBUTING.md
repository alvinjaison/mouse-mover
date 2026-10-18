# Contributing to Mouse Mover

Thanks for your interest in contributing! Mouse Mover is a small, focused Go
project, and contributions of all kinds are welcome: bug reports, documentation
fixes, new movement patterns, platform improvements, and more.

## Ways to contribute

- **Report a bug** — open an issue using the Bug Report template.
- **Request a feature** — open an issue using the Feature Request template.
- **Improve docs** — typos, clarifications, and examples are all appreciated.
- **Write code** — pick up an open issue or propose a change in an issue first.

## Development setup

Mouse Mover is a single Go module (`github.com/jarvis/mouse-mover`). You need
Go 1.23 or newer.

```bash
git clone https://github.com/alvinjaison/mouse-mover.git
cd mouse-mover
go test ./...
```

The platform-agnostic packages (`internal/engine`, `internal/config`) build and
test on any OS. The GUI/tray code is platform-specific:

- **Windows**: native Win32 window + system tray, pure Go, `CGO_ENABLED=0`.
- **macOS**: Cocoa menu bar app, uses cgo (Core Graphics + AppKit).
- **Other (Linux)**: builds through `!windows && !darwin` stubs for testing only.

### Build

```bash
# Windows (cross-compile from any OS, no C toolchain needed):
bash build/build.sh

# macOS universal binary (run on a Mac with Xcode command line tools):
bash build/build-mac.sh
```

## Pull request process

1. **Open an issue first** for anything beyond a trivial fix, so we can agree on
   the approach before you invest time.
2. **Fork** the repo and create a topic branch: `git checkout -b fix/short-name`.
3. **Keep changes focused.** One logical change per PR.
4. **Run the checks locally** before pushing:
   ```bash
   gofmt -l .      # should print nothing
   go vet ./...
   go test ./...
   ```
5. **Write a clear PR description**: what changed, why, and how you tested it.
6. CI must pass. A maintainer will review and may request changes.

## Coding guidelines

- Format with `gofmt` (CI enforces this).
- Keep the layered architecture: UI → engine (via the `CursorController`
  interface) → platform bindings. The engine must stay platform-agnostic and
  testable without a real cursor.
- Platform-specific files use build tags (`//go:build windows`,
  `//go:build darwin`, `//go:build !windows && !darwin`).
- Add or update tests for engine and config changes. These run on every OS in
  CI.
- Prefer small, documented functions over cleverness.

## Reporting security issues

Please do **not** open a public issue for security vulnerabilities. See
[SECURITY.md](SECURITY.md) for how to report privately.

## Code of Conduct

By participating, you agree to abide by our
[Code of Conduct](CODE_OF_CONDUCT.md).

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE) that covers this project.
