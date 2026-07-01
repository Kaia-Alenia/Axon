<div align="center">

<img src="assets/logo.png" alt="Axon Logo" width="600">

# Axon

**Turn any device into a high-performance wireless touchpad, keyboard, and scroll wheel for your PC.**

[![Build](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml/badge.svg)](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Android](https://img.shields.io/badge/Android-Kotlin-3DDC84?logo=android&logoColor=white)](https://developer.android.com/)
[![Jetpack Compose](https://img.shields.io/badge/Jetpack_Compose-4285F4?logo=jetpackcompose&logoColor=white)](https://developer.android.com/jetpack/compose)
[![GitGem](https://gitgem.org/api/badge/github/Kaia-Alenia/Axon.svg)](https://gitgem.org/github/Kaia-Alenia/Axon)

[![Patreon](https://img.shields.io/badge/Patreon-alenia__studios-F96854?logo=patreon&logoColor=white)](https://www.patreon.com/alenia_studios)
[![Ko-Fi](https://img.shields.io/badge/Ko--Fi-alenia__studios-F16061?logo=ko-fi&logoColor=white)](https://ko-fi.com/alenia_studios)
[![PayPal](https://img.shields.io/badge/PayPal-Donate-00457C?logo=paypal&logoColor=white)](https://www.paypal.com/ncp/payment/TCCYMCFSVMV8E)
[![Itch.io](https://img.shields.io/badge/Itch.io-Alenia_Studios-FA5C5C?logo=itch.io&logoColor=white)](https://alenia-studios.itch.io/)

<br/>

An open-source, ultra-low-latency remote input tool by **AXON**. Run the server on your computer, connect from your phone's native Android app or browser, and experience seamless, native-feeling control.

</div>

---

## Overview and Key Features

Axon is engineered from the ground up to provide a flawless, lag-free peripheral experience. Unlike conventional remote-control apps, Axon uses optimized networking protocols (UDP and low-level RFCOMM Bluetooth) to achieve unprecedented performance.

- **Ultra-Low Latency (1ms - 3ms):** Thanks to our optimized Bluetooth and USB/ADB implementations, input delay is practically non-existent. It feels exactly like using a high-end native hardware mouse.
- **Dynamic Port Allocation:** Robust server architecture automatically manages port conflicts. If standard ports are in use, it seamlessly falls back to dynamic allocation without crashing.
- **Advanced Gestures:** Pinch-to-zoom, multi-touch scrolling, and double-tap-and-hold to drag (perfect for text selection or dragging windows).
- **Hardware Integration:** Utilize native physical volume buttons on your Android device to control system volume.
- **Cross-Platform Compatibility:** Runs flawlessly on Linux, Windows, and macOS with native input simulation.
- **Zero-Setup Pairing:** Scan the QR code presented in your terminal and connect instantly.
- **Triple Connectivity Options:** Wi-Fi, Bluetooth, and USB/ADB modes to suit any networking environment.

## Downloads & Installation

To use Axon, you need to set up the Client (on your phone) and the Server (on your computer).

### 1. Android App (Client)

1. Download the latest APK from the releases page: [axon-android-debug.apk](https://github.com/kaia-alenia/axon/releases/latest/download/axon-android-debug.apk)
2. Transfer the file to your Android device (if downloaded on a computer).
3. Open the file on your device. You may need to grant your file manager permission to "Install from Unknown Sources".
4. Follow the on-screen prompts to complete the installation.

### 2. Server (Computer)

We provide three distinct methods to install the Axon server, catering to all user preferences.

#### Option A: One-Liner Global Installation (Recommended)

This is the fastest and easiest method. It automatically detects your operating system, downloads the correct binary, and installs it globally so you can simply type `axon` in any terminal to start the server.

**For Linux and macOS (or Git Bash on Windows):**
Open your terminal and run:
```bash
curl -fsSL https://raw.githubusercontent.com/Kaia-Alenia/Axon/main/install.sh | bash
```

**For Windows (PowerShell):**
Open PowerShell as Administrator and run:
```powershell
iwr -useb https://raw.githubusercontent.com/Kaia-Alenia/Axon/main/install.ps1 | iex
```

Once installed, you can launch the server from anywhere:
```bash
axon                # Start server with default ports
axon --help         # Show all available commands and flags
```

#### Option B: Direct Binary Download

If you prefer not to use an automated script, you can download the standalone executable for your specific system:

- **Windows:** [axon-windows-amd64.exe](https://github.com/kaia-alenia/axon/releases/latest/download/axon-windows-amd64.exe)
- **Linux (x86_64):** [axon-linux-amd64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-linux-amd64)
- **Linux (ARM64):** [axon-linux-arm64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-linux-arm64)
- **Linux (ARM):** [axon-linux-arm](https://github.com/kaia-alenia/axon/releases/latest/download/axon-linux-arm)
- **macOS (Intel):** [axon-darwin-amd64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-darwin-amd64)
- **macOS (Apple Silicon):** [axon-darwin-arm64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-darwin-arm64)

*Note for Linux and macOS users: You will need to make the file executable before running it. Open a terminal and run `chmod +x filename`, then run it with `./filename`.*

#### Option C: Build from Source

For developers who wish to modify the code or compile it locally:

1. Ensure you have Go 1.25 or higher installed.
2. Clone the repository:
```bash
git clone https://github.com/Kaia-Alenia/Axon.git
cd Axon/server
```
3. Build the server:
```bash
go build -o axon .
```
4. Run the server:
```bash
./axon
```

## Connecting and Using Axon

Once the server is running on your computer, a QR code and connection details will be displayed in your terminal.

1. Open the Axon app on your Android device.
2. Select your preferred connection method:
   - **USB/ADB (Highly Recommended):** Provides the absolute lowest latency (virtually instantaneous). Requires USB Debugging enabled on your phone.
   - **Bluetooth (Recommended):** Delivers phenomenal 1ms-3ms latency, mimicking the feel of a premium wireless native mouse. Supported across Windows, macOS, and Linux.
   - **Wi-Fi:** A reliable, high-speed alternative if Bluetooth or USB are unavailable. Connect by scanning the QR code or manually entering the IP address.

## Server Commands and Configuration

The Axon server is designed to be highly robust. If a requested port is occupied by another application, it will dynamically fall back to an available port to ensure it always starts successfully.

Available flags:

```bash
# Run with default ports (6969 for TCP/WebSocket, 6970 for UDP)
axon

# Use custom TCP/WebSocket port
axon --port 7000

# Use custom UDP port for the fast-input protocol
axon --udp-port 7001

# Auto-allocate all ports (let the OS choose available ports)
axon --port 0 --udp-port 0

# Display version information
axon --version
```

## Troubleshooting

### ADB Warning on Startup

When starting the server, you may see the following message:

```
[ADB] Warning: failed to setup adb reverse for port 6969: exit status 1
```

**This is not an error.** Axon fully supports Wi-Fi and Bluetooth without ADB. This warning only appears when ADB is installed on your system but no Android device is connected via USB at the time the server starts.

- **If you only use Wi-Fi or Bluetooth:** Ignore this message completely.
- **If you want USB/ADB mode (lowest possible latency):** You need to install Android Platform Tools and connect your phone with USB Debugging enabled.

#### Installing ADB (Android Platform Tools)

**Linux:**
```bash
sudo apt install adb
```

**macOS (Homebrew):**
```bash
brew install android-platform-tools
```

**Windows:** Download the [Android Platform Tools ZIP](https://developer.android.com/tools/releases/platform-tools) from Google, extract it, and add the folder to your system `PATH`.

After installing ADB, connect your Android device via USB, enable USB Debugging in Developer Options, and restart Axon. The warning will disappear once a device is detected.

### Viewing Your Phone on Your PC (scrcpy)

If you want to view and control your Android device screen directly from your PC (which pairs perfectly with Axon over ADB), the tool you are looking for is **scrcpy** (Screen Copy).

**Installation:**
- **Windows:** You can download it from their [official GitHub repository](https://github.com/Genymobile/scrcpy) or install via scoop (`scoop install scrcpy`).
- **macOS:** `brew install scrcpy`
- **Linux:** `sudo apt install scrcpy`

**Usage:**
Once installed and your device is connected via ADB, simply run this command in your terminal:
```bash
scrcpy
```

## Contributing

We welcome contributions from the community. Please review our [Contributing Guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before getting started.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/advanced-gestures`)
3. Commit your changes (`git commit -m 'Add advanced gesture support'`)
4. Push to the branch (`git push origin feature/advanced-gestures`)
5. Open a Pull Request

## Security

If you discover a vulnerability, please report it responsibly. Refer to our [Security Policy](SECURITY.md) for detailed reporting procedures.

## License

This project is strictly licensed under the **GNU General Public License v3 (GPL v3)**. Please see the [LICENSE](LICENSE) file for comprehensive details. Assets and music (where applicable) are licensed under the Alenia Studios Standard (CC BY 4.0 + Additional Terms).

---

<div align="center">

**AXON** · [contact.aleniastudios@gmail.com](mailto:contact.aleniastudios@gmail.com)

</div>
