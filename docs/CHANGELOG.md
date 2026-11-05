# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.6.8] - 2026-11-04

### ✔️ Fixed
- **Windows 10 Console QuickEdit Freeze & Non-Blocking Logger:** Disabled `ENABLE_QUICK_EDIT_MODE` on `windows.Stdin` and implemented a non-blocking console logger so clicking the Windows terminal window or console I/O never pauses the WebSocket reader loop (`click`, `mousedown`, `mouseup`, `type`, `key`).
- **Windows 10 Exact Binary Layout for `SendInput`:** Unified mouse and keyboard `INPUT` structures into a strict 40-byte binary layout (`rawInput [40]byte`) and batched keyboard keydown/keyup events atomically.

## [1.6.7] - 2026-11-04

### ✔️ Fixed
- **Windows 10 Input Simulation (`SendInput`):** Fixed 64-bit `INPUT` struct alignment in `simulator_windows.go` (40 bytes instead of 48 bytes), restoring mouse movement, clicks, scroll, and pinch-to-zoom on Windows.
- **Windows Native Bluetooth RFCOMM:** Implemented real Winsock2 `AF_BTH` RFCOMM socket and SDP service registration (`00001101-0000-1000-8000-00805F9B34FB`) in `bluetooth_windows.go`.
- **Windows Wi-Fi & IP Detection:** Added outbound LAN IP detection and virtual adapter filtering (`vEthernet`, `WSL`, `VirtualBox`, `VMware`, `Docker`) plus disabled `SO_REUSEADDR` port hijacking on Windows.

## [1.6.6] - 2026-08-23

### 🚀 Added / Improved
- Alenia Studios coding standards compliance (Clean Code).
- Internal version bump in Go and Android.

## [1.6.3] - 2026-06-29

### ✔️ Fixed
- **ADB Warning:** Restored a clean, single-line ADB warning message. The verbose system error output is now suppressed; only a concise `exit status 1` notice is shown when ADB is installed but no device is connected.
- **macOS Input Simulation:** Separated the Linux-specific input simulator (uinput/xdotool) from macOS. macOS now uses its own native simulator backed by `cliclick` (installable via Homebrew), preventing runtime failures on Apple Silicon and Intel Macs.

### ✅ Added
- **Troubleshooting Section (README):** Added a clear explanation of the ADB startup warning with platform-specific installation instructions for Android Platform Tools.

## [1.6.2] - 2026-06-29

### ✔️ Fixed
- **ADB Warnings:** Silenced verbose ADB connectivity warnings and avoided invoking ADB if it is not installed in the system PATH, preventing terminal clutter.

## [1.6.1] - 2026-06-29

### ✔️ Fixed
- **Windows Terminal Rendering:** Enabled Virtual Terminal Processing for Windows CMD/PowerShell to correctly render ANSI colors and ASCII art.

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
