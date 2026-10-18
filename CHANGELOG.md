# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Open-source project files: `LICENSE` (MIT), `CONTRIBUTING.md`,
  `CODE_OF_CONDUCT.md`, `SECURITY.md`, issue/PR templates, and CI.
- macOS universal binary (Apple Silicon + Intel) in the build script.
- Documentation with Windows and macOS screenshots.

### Fixed

- macOS menu bar status item not appearing: the Cocoa UI now runs on the main
  OS thread (`runtime.LockOSThread`).
- macOS build no longer collides with the cross-platform stubs (build tags now
  exclude `darwin`).
- macOS config tests are hermetic (the settings path honors `XDG_CONFIG_HOME`).

## [0.1.0] - 2026-10-18

### Added

- Initial release.
- Windows 10/11 native Win32 window + system tray UI with Start/Stop.
- macOS menu bar app.
- Configurable run duration (auto-stop) and move interval.
- Movement patterns: Nudge, Jiggle, Square.
- Multi-monitor aware cursor clamping.
- Persistent JSON settings under the per-user application-data directory.

[Unreleased]: https://github.com/alvinjaison/mouse-mover/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/alvinjaison/mouse-mover/releases/tag/v0.1.0
