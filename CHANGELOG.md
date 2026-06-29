# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2026-06-29

### Fixed
- Fixed compilation error on Windows related to `SO_REUSEADDR` by adding platform-specific socket control logic.

### Added
- Go server with WebSocket + UDP protocol for low-latency input
- Embedded web client (touchpad, scroll, keyboard, mouse buttons)
- Native Android client with Jetpack Compose
- QR code auto-connect
- Wi-Fi and Bluetooth connectivity
- Cross-platform input simulation (Linux, Windows, macOS)
- CI/CD pipeline with GitHub Actions
- Community files (Contributing guide, Code of Conduct, Security Policy)
