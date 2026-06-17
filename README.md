<div align="center">

# ⚡ Axon

**Turn any device into a wireless touchpad, keyboard & scroll wheel for your PC.**

[![Build](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml/badge.svg)](https://github.com/Kaia-Alenia/Axon/actions/workflows/build.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Android](https://img.shields.io/badge/Android-Kotlin-3DDC84?logo=android&logoColor=white)](https://developer.android.com/)
[![Jetpack Compose](https://img.shields.io/badge/Jetpack_Compose-4285F4?logo=jetpackcompose&logoColor=white)](https://developer.android.com/jetpack/compose)
[![GitHub Sponsors](https://img.shields.io/badge/Sponsor-♥-ea4aaa?logo=githubsponsors)](https://github.com/sponsors/Kaia-Alenia)

<br/>

An open-source remote input tool by **Alenia Studios**. Run the server on your PC, connect from your phone's browser or the native Android app — instant wireless control.

</div>

---

## ✨ Features

- 🖱️ **Touchpad** — Smooth cursor movement with multi-touch support
- ⌨️ **Keyboard** — Full keyboard input forwarding
- 📜 **Scroll Wheel** — Dedicated scroll zone
- 📱 **QR Connect** — Scan and connect instantly
- 🔗 **Wi-Fi & Bluetooth** — Dual connectivity
- 🖥️ **Cross-Platform Server** — Linux, Windows & macOS
- 🌐 **Web Client** — No app install needed, works from any browser
- 📲 **Native Android App** — Jetpack Compose client

## 📂 Project Structure

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
└── README.md
```

## 🚀 Quick Start

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

## 🖥️ Platform Support

| Platform | Touchpad | Keyboard | Scroll | Bluetooth |
|:---------|:--------:|:--------:|:------:|:---------:|
| Linux    | ✅       | ✅       | ✅     | ✅        |
| Windows  | ✅       | ✅       | ✅     | 🔜        |
| macOS    | ✅       | ✅       | ✅     | 🔜        |

## 🏗️ Tech Stack

| Component | Technology |
|:----------|:-----------|
| Server    | Go 1.25, Gorilla WebSocket, go-qrcode |
| Web Client | Vanilla HTML/CSS/JS (embedded via `embed.FS`) |
| Android   | Kotlin, Jetpack Compose, Material 3 |
| Protocol  | WebSocket (control) + UDP (low-latency input) |

## 🤝 Contributing

Contributions are welcome! Feel free to open issues and pull requests.

1. Fork the repository
2. Create your branch (`git checkout -b feature/awesome`)
3. Commit your changes (`git commit -m 'Add awesome feature'`)
4. Push to branch (`git push origin feature/awesome`)
5. Open a Pull Request

## 📄 License

This project is licensed under the **GNU General Public License v3 (GPL v3)** — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Alenia Studios** · [contact.aleniastudios@gmail.com](mailto:contact.aleniastudios@gmail.com)

</div>
