# Contributing to Axon

First off, **thank you** for considering contributing to Axon! Every contribution helps make this project better for everyone.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Style Guide](#style-guide)

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [contact.aleniastudios@gmail.com](mailto:contact.aleniastudios@gmail.com).

## How Can I Contribute?

### Reporting Bugs

- Use the [Bug Report](https://github.com/Kaia-Alenia/Axon/issues/new?template=bug_report.md) issue template.
- Include your OS, Go version, and steps to reproduce.
- Attach logs or screenshots if applicable.

### Suggesting Features

- Use the [Feature Request](https://github.com/Kaia-Alenia/Axon/issues/new?template=feature_request.md) issue template.
- Explain the problem your feature would solve.
- Describe your proposed solution.

### Code Contributions

1. Look for issues labeled [`good first issue`](https://github.com/Kaia-Alenia/Axon/labels/good%20first%20issue) or [`help wanted`](https://github.com/Kaia-Alenia/Axon/labels/help%20wanted).
2. Comment on the issue to let others know you're working on it.
3. Fork, branch, code, and submit a PR.

## Getting Started

### Prerequisites

**Server (Go):**
- Go 1.25+
- Linux: `xdotool` package for input simulation
- Windows/macOS: No additional dependencies

**Android:**
- Android Studio or JDK 17+
- Android SDK 36

### Development Setup

```bash
# Clone
git clone https://github.com/Kaia-Alenia/Axon.git
cd Axon

# Server
cd server
go build -o axon .
./axon

# Android
cd android
./gradlew assembleDebug
```

## Pull Request Process

1. **Fork** the repository and create your branch from `main`.
2. **Follow** the [style guide](#style-guide) below.
3. **Test** your changes locally before submitting.
4. **Update** documentation if you changed any public API or behavior.
5. **Fill out** the PR template completely.
6. **Wait** for a review — we aim to respond within 48 hours.

## Style Guide

### Go (Server)

- Follow standard `gofmt` formatting.
- Use meaningful variable names.
- Keep functions focused and short.
- Platform-specific code goes in `_linux.go`, `_windows.go`, or `_darwin.go` files.

### Kotlin (Android)

- Follow [Kotlin coding conventions](https://kotlinlang.org/docs/coding-conventions.html).
- Use Jetpack Compose best practices.
- Keep composables small and reusable.

### Commits

- Use clear, descriptive commit messages.
- Use present tense: "Add feature" not "Added feature".
- Reference issues when applicable: `Fix #42`.

---

Thank you for helping make Axon better!
