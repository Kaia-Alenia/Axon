# Axon Roadmap

This document outlines the future vision and development plans for Axon. The goal is to maintain a clear guide of planned features, from immediate improvements to major platform expansions.

## Short Term (Immediate Improvements & Stability)

- **Security & Privacy:**
  - [ ] Implement a PIN/Password system for secure pairing (prevent unwanted connections on public networks).
  - [ ] End-to-End Encryption (E2EE) for keystroke transmission.
- **UI/UX Improvements (Client):**
  - [ ] Full support for dynamic themes (Material You) and manual dark/light mode toggle.
  - [ ] Custom profiles for touchpad sensitivity and acceleration.
- **New Control Modes:**
  - [ ] **Presentation Mode:** Simplified interface with giant next/previous buttons for slides (PowerPoint/Keynote) and a virtual laser pointer.
  - [ ] **Dedicated Media Controls:** Specific screen for playback control (Play/Pause, skip track, advanced volume controls).

## Medium Term (Evolution & Expansion)

- **Cross-Platform Client Support:**
  - [ ] **iOS Client:** Develop the version for iPhone/iPad (considering Kotlin Multiplatform or native Swift).
- **Graphical Interface for the Server:**
  - [ ] **System Tray Application:** Create a minimalist GUI for the PC server (Windows/Mac/Linux) to toggle the server, view connected devices, and configure ports without using the terminal.
- **Advanced Inputs:**
  - [ ] **Gamepad Mode (Gyroscope):** Use the phone's motion sensors (gyroscope/accelerometer) as a steering wheel for racing games or flight simulators on PC.
  - [ ] **Voice Dictation:** Use the phone's speech-to-text capabilities to type text directly onto the PC at high speeds.

## Long Term (Future Vision)

- **Expansion to Smart TVs (Android TV / Google TV):**
  - [ ] **Axon TV Server:** Adapt the server to run natively on Google TV televisions.
  - [ ] Use the *AccessibilityService* API to inject mouse movements and clicks into the TV interface.
  - [ ] Allow users to search for movies or type long passwords on the TV comfortably from their mobile keyboard.
- **Lightweight Screen Mirroring:**
  - [ ] Cast the PC screen to the phone at very low FPS (to avoid impacting latency) so the PC can be controlled even if the monitor is off or in another room.
- **Automation Integration:**
  - [ ] Shortcuts or *Intents* to allow apps like Tasker or Siri Shortcuts to execute macros on the PC (e.g., "Turn off PC" from a voice command to the phone).
