# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### ✔️ Fixed
- **`axon --update` printed nothing on Linux:** the server log goes through a non-blocking writer with a background goroutine, and the updater called `os.Exit` right after logging its result, so the queued lines (e.g. "Already up to date") were lost when the process exited. The updater now flushes pending log output before exiting, on every path, and the same flush applies to the update helper and to server startup errors that exit. Also, the version check now runs before the install-directory writability probe, so an up-to-date install in a read-only directory (e.g. `/usr/local/bin`) correctly reports "Already up to date" instead of failing with permission denied.

## [1.7.2] - 2026-11-15

### ✔️ Fixed
- **Language switch back to Chinese (#11, bug):** the home-screen selector only toggled ES <-> EN, so once the user left Chinese they could never return to it. The selector now shows all three languages (EN / ES / 中文) and each is directly selectable; Chinese uses `Locale.SIMPLIFIED_CHINESE` (zh-Hans-CN) so the `values-zh` resources match reliably, and stored language tags such as `zh-CN` are normalised on load.
- **Delete key only deleted phone-typed text (#11):** the phone delete key was swallowed by the IME whenever the local text buffer was empty, so nothing reached the PC. Delete is now intercepted (`onKeyEvent`) in the touchpad and keyboard screens and always sends a real Backspace event to the PC, and both screens have a dedicated ⌫ Backspace button.
- **QR code too small and coloured (#12):** the server printed the connection QR with a rainbow ANSI colour gradient at a small size. It now prints a large plain black-on-white QR (full-size modules, 4-module quiet zone, no colour codes), and the scanner screen has a zoom slider (CameraX linear zoom) with a scan frame that grows with the zoom.
- **FastGithub breaks the mouse (#12):** the Windows simulator now checks the `SendInput` return value and logs a one-time warning when Windows rejects synthetic input. The usual cause is UIPI: the foreground app (e.g. FastGithub) runs as Administrator while Axon does not, so Windows silently drops the input. The server log now tells the user to run `axon.exe` as Administrator. This could not be reproduced on this Linux VM, so the underlying Windows behaviour is diagnosed, not verified on Windows.
- **Function keys and media keys silently dropped on Linux/macOS (#9, #10):** on Linux, keys the uinput map did not cover (F1–F4, F6–F12, mute, play/pause) were dropped whenever the uinput keyboard was available; they now fall back to xdotool. On macOS, F1–F12 are mapped to cliclick's lowercase names and volume keys are handled with AppleScript (relative volume, mute toggle, F8 for play/pause). Windows already handled all of these.

### 🚀 Added / Improved
- **Two-finger scroll vs pinch-zoom (#9):** the touchpad now classifies a two-finger gesture once — pan scrolls, only a real pinch zooms (Ctrl + / Ctrl −) — and sticks to that classification until the fingers lift, so a scroll can no longer turn into a zoom halfway through. Two-finger drag no longer moves the mouse cursor.
- **Two-finger tap = right-click (#11):** a quick two-finger tap on the touchpad sends a right-click; the gesture hint is shown on the touchpad.
- **Hideable scroll strip and fullscreen touchpad (#11):** the touchpad header has toggles to hide the scroll strip and to switch to a fullscreen touchpad mode (whole screen is the touchpad, no other buttons, with an exit button).
- **Volume and Douyin video buttons (#10):** the touchpad and keyboard media rows already had volume up/down/mute and Up/Down video buttons, and the phone hardware volume keys already sent volume events; the video buttons are now labelled "Video ↑ / Video ↓" (EN/ES) / "上一个视频 / 下一个视频" (ZH) so their Douyin purpose is clear, and the server now actually delivers them on Linux and macOS (see Fixed above).
- **Server self-update actually replaces the binary:** `axon --update` used to download the new binary as `axon.exe.new` and leave the manual delete/rename to the user — on Windows the rename always failed because the running exe is locked. It now checks the latest release via the GitHub API (and skips the download when already up to date), verifies the download (byte count against Content-Length and the release asset size, plus magic bytes), checks free disk space and install-dir writability first, and swaps the binary in with a backup (`axon.exe.old`): on Windows a detached helper process waits for the parent to exit, swaps the files, verifies the new binary actually launches (`--version` prints the new version), and only then deletes the backup — if the swap or launch check fails, the old binary is restored from backup. Every step is logged to `update.log` next to the exe. On Linux/macOS the swap happens directly and the process restarts into the new binary.
- **Android in-app updater:** the app had no updater at all — it never checked for a new version. It now checks the GitHub releases API once per launch (and on demand via a "Check for updates" button), shows an in-app update dialog with the new version, downloads the APK with progress into a stable location (`filesDir/updates/`), verifies its size and ZIP magic before installing, and installs it through the system installer via FileProvider — never opening GitHub in the browser. It handles the "Install unknown apps" permission (guides to Settings, retries after returning), user-cancelled installs (dismissed silently, file kept for a later retry), and incomplete/corrupt downloads (deleted, with a retry). The downloaded APK is only deleted once the new version is confirmed installed.

## [1.6.9] - 2026-11-05

### ✔️ Fixed
- **Windows 10 Background Execution on Console Minimize:** Configured `SetProcessInformation` to disable Power Throttling / EcoQoS, elevated priority to `ABOVE_NORMAL_PRIORITY_CLASS`, and set `SetThreadExecutionState` to prevent Windows from suspending networking threads (WebSocket / UDP / Bluetooth).
- **Console Buffer Lockup & Crash Prevention:** Integrated `IsIconic` minimization detection to suspend `os.Stdout` writes while minimized (preventing `conhost.exe` buffer reflow crashes and window terminations). Disabled `ENABLE_WINDOW_INPUT` and `ENABLE_MOUSE_INPUT`.
- **Connection Resilience:** Increased WebSocket ping write deadline from 3s to 10s to absorb window transition latencies. Added `--minimized` flag to launch directly minimized.

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
