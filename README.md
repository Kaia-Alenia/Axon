<div align="center">

<img src="assets/logo.png" alt="Axon Logo" width="600">

# Axon

**Turn any device into a wireless touchpad, keyboard & scroll wheel for your PC.**

[![Build](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml/badge.svg)](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Android](https://img.shields.io/badge/Android-Kotlin-3DDC84?logo=android&logoColor=white)](https://developer.android.com/)
[![Jetpack Compose](https://img.shields.io/badge/Jetpack_Compose-4285F4?logo=jetpackcompose&logoColor=white)](https://developer.android.com/jetpack/compose)

[![Patreon](https://img.shields.io/badge/Patreon-alenia__studios-F96854?logo=patreon&logoColor=white)](https://www.patreon.com/alenia_studios)
[![Ko-Fi](https://img.shields.io/badge/Ko--Fi-alenia__studios-F16061?logo=ko-fi&logoColor=white)](https://ko-fi.com/alenia_studios)
[![PayPal](https://img.shields.io/badge/PayPal-Donate-00457C?logo=paypal&logoColor=white)](https://www.paypal.com/ncp/payment/TCCYMCFSVMV8E)
[![Itch.io](https://img.shields.io/badge/Itch.io-Alenia_Studios-FA5C5C?logo=itch.io&logoColor=white)](https://alenia-studios.itch.io/)

<br/>

An open-source remote input tool by **AXON**. Run the server on your PC, connect from your phone's browser or the native Android app — instant wireless control.

</div>

---

## Features

- Touchpad - Smooth cursor movement with multi-touch support
- Keyboard - Full keyboard input forwarding
- Scroll Wheel - Dedicated scroll zone
- QR Connect - Scan and connect instantly
- Wi-Fi & Bluetooth - Dual connectivity
- Cross-Platform Server - Linux, Windows & macOS
- Web Client - No app install needed, works from any browser
- Native Android App - Jetpack Compose client

## Project Structure

```
Axon/
├── server/          # Go server (Linux / Windows / macOS)
│   ├── main.go      # Entry point & WebSocket handler
│   ├── simulator*.go # Platform-specific input simulation
│   ├── bluetooth*.go # Platform-specific Bluetooth
│   └── web/         # Embedded web client (HTML/CSS/JS)
├── android/         # Native Android client (Kotlin)
│   └── app/         # Jetpack Compose application
├── .github/
│   ├── workflows/   # CI build pipelines
│   └── FUNDING.yml  # Sponsorship config
├── LICENSE          # GNU GPL v3
├── CONTRIBUTING.md  # How to contribute
├── CODE_OF_CONDUCT.md
├── SECURITY.md      # Vulnerability reporting
├── CHANGELOG.md     # Version history
└── README.md
```
## Downloads & Installation

To use Axon, you need to download two files: the app for your phone and the server for your computer.

### 1. Android App (Client)
Download the following file and install it on your mobile device:
* [axon-android-debug.apk](https://github.com/kaia-alenia/axon/releases/latest/download/axon-android-debug.apk)

### 2. Server (Computer)
Download the correct file for your computer's operating system to receive connections:
* **Windows:** [axon-windows-amd64.exe](https://github.com/kaia-alenia/axon/releases/latest/download/axon-windows-amd64.exe)
* **Linux:** [axon-linux-amd64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-linux-amd64)
* **macOS (Intel):** [axon-darwin-amd64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-darwin-amd64)
* **macOS (Apple Silicon):** [axon-darwin-arm64](https://github.com/kaia-alenia/axon/releases/latest/download/axon-darwin-arm64)

> **Note for Linux and macOS users:** After downloading the server file, remember to grant it execution permissions from the terminal before launching it:
> `chmod +x filename`

## Quick Start

### Server

```bash
cd server
go build -o axon .
./axon
```

A QR code will appear in your terminal. Scan it with your phone or open the displayed URL in any browser.

### Android App

```bash
cd android
./gradlew assembleDebug
```

The APK will be at `android/app/build/outputs/apk/debug/`.

## Connection Recommendations

To get the best experience and minimize latency, the following guidelines are suggested:

- **Local & Minimum Latency (Recommended)**: USB/ADB connection is recommended for an almost instantaneous response.
- **Wi-Fi vs Bluetooth**: It is recommended to use Wi-Fi over Bluetooth due to its higher speed and lower latency.
- **Bluetooth**: Bluetooth is only recommended in specific cases where no Wi-Fi network or USB cable is available.

## Platform Support

| Platform | Touchpad | Keyboard | Scroll | Bluetooth |
|:---------|:--------:|:--------:|:------:|:---------:|
| Linux    | Yes      | Yes      | Yes    | Yes       |
| Windows  | Yes      | Yes      | Yes    | Planned   |
| macOS    | Yes      | Yes      | Yes    | Planned   |

## Tech Stack

| Component | Technology |
|:----------|:-----------|
| Server    | Go 1.25, Gorilla WebSocket, go-qrcode |
| Web Client | Vanilla HTML/CSS/JS (embedded via `embed.FS`) |
| Android   | Kotlin, Jetpack Compose, Material 3 |
| Protocol  | WebSocket (control) + UDP (low-latency input) |

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before getting started.

1. Fork the repository
2. Create your branch (`git checkout -b feature/awesome`)
3. Commit your changes (`git commit -m 'Add awesome feature'`)
4. Push to branch (`git push origin feature/awesome`)
5. Open a Pull Request

## Security

Found a vulnerability? Please report it responsibly. See our [Security Policy](SECURITY.md) for details.

## License

This project is licensed under the **GNU General Public License v3 (GPL v3)** — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**AXON** · [contact.axon@gmail.com](mailto:contact.axon@gmail.com)

</div>
