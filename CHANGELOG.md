# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.0] - 2026-06-29

### ✔️ Fixed
- **Windows Compilation:** Resolved a critical cross-platform build error related to `SO_REUSEADDR` by implementing platform-specific socket control logic.

### ✅ Added

**Connectivity & Input**
- **Low-Latency Bluetooth:** Introduced super fast, low-latency Bluetooth support specifically optimized for Windows and macOS.
- **Advanced Gestures:** Added new intuitive gestures for zooming.
- **Hardware Integration:** Volume can now be controlled directly using hardware volume up/down buttons.
- **Auto-Connect:** Seamless pairing and connection via QR code.
- **Versatile Connectivity:** Full support for both Wi-Fi and Bluetooth modes.

**Architecture & Clients**
- **Go Server:** Powerful backend utilizing WebSocket + UDP protocols for ultra-low latency input transmission.
- **Embedded Web Client:** Full support for touchpad, scroll, keyboard, and mouse buttons directly from the browser.
- **Native Android Client:** Built from the ground up using Jetpack Compose for a native, fluid experience.
- **Input Simulation:** Reliable cross-platform input simulation supporting Linux, Windows, and macOS.

**Project & Infrastructure**
- **CI/CD Pipeline:** Automated build and release workflows powered by GitHub Actions.
- **Community Standards:** Added Contributing guide, Code of Conduct, and Security Policy to foster open-source collaboration.
