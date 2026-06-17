# Axon

**Axon** is an open-source remote input tool by [Alenia Studios](mailto:contact.aleniastudios@gmail.com). It turns any device with a browser or the Android app into a wireless touchpad, keyboard, and scroll wheel for your PC.

## Structure

```
Axon/
├── server/    # Go server (Linux/Windows/macOS)
├── android/   # Android client (Kotlin + Jetpack Compose)
└── LICENSE    # GPL v3
```

## Server

The server runs on your PC and exposes a local web interface + WebSocket endpoint.

### Build & Run

```bash
cd server
go build -o axon .
./axon
```

Scan the QR code shown in the terminal or open the displayed URL on your phone's browser.

### Supported Platforms

| Platform | Touchpad | Keyboard | Bluetooth |
|----------|----------|----------|-----------|
| Linux    | ✅       | ✅       | ✅        |
| Windows  | ✅       | ✅       | 🔜        |
| macOS    | ✅       | ✅       | 🔜        |

## Android

Native client built with Jetpack Compose. Connects via Wi-Fi or Bluetooth.

### Build

```bash
cd android
./gradlew assembleDebug
```

## License

This project is licensed under the **GNU General Public License v3 (GPL v3)**. See [LICENSE](LICENSE) for details.

---

**Alenia Studios** — contact.aleniastudios@gmail.com
