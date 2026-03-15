# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [2.0.1] - 2026-03-15

### Changed
- Centralize version management via Makefile + ldflags build-time injection
- Pass ldflags to `go run` targets for consistent version output
- Replace .NET Visual Studio .gitignore with Go-appropriate rules
- Improve README tables formatting

### Fixed
- Goreleaser ldflags targeting non-existent `main.version` variable

## [2.0.0] - 2026-03-15

### Added
- Cross-platform Go port with CLI + GUI dual mode
- Windows, macOS, Linux support (amd64 + arm64)
- AES-256-GCM credential encryption with machine-derived key
- Turkish/English real-time localization (65+ keys)
- System tray integration with never-close window architecture
- Intelligent auto-reconnection with configurable intervals
- Platform-specific auto-start (Windows registry, macOS LaunchAgent, Linux XDG)
- Toast notification system
- OLED dark theme
- 193 test functions across 8 packages
- Goreleaser + GitHub Actions release pipeline
