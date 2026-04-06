# System Manager
A centralized application monitoring system designed for an Ubuntu Host natively running systemd apps and docker containers.

## Architecture
- **Agent:** Go-based daemon that discovers applications, reads CPU/RAM from `/proc`, handles SQLite logging and exposes a WebSocket interface.
- **Dashboard:** Flutter Linux Desktop App utilizing Riverpod state and WebSocket streaming for 60FPS UI.

## Getting Started

### 1. Build and Run the Go Agent
On your Ubuntu machine (or test environment), compile and run the agent:

```bash
cd agent
# Run go mod tidy to fetch dependencies
go mod tidy
# Build the sysagent binary
go build -o sysagent ./cmd/sysagent
# Run sysagent
./sysagent
```
*Note: The agent generates an auth token in the stdout when it starts. The Flutter UI is hardcoded in this template to use a mock token (`mock_jwt_for_ui_since_no_login_screen_yet`), but in production you will pass the token via CLI or environment to the UI.*

### 2. Build and Run the Flutter UI
On your Linux machine (or Mac running the dashboard pointing to a remote instance):
```bash
cd dashboard
flutter pub get
flutter run -d linux
```

## Features
- **Auto Discovery:** Monitors all running systemd services and docker containers.
- **Lightweight Metrics:** Reads straight from `/proc` minimizing CPU overhead.
- **Controls:** Ability to Start/Stop/Restart services visually directly from the desktop.
- **Log Streaming:** Real-time stdout/stderr from `journalctl` and `docker`.
- **Database:** Batches writes to SQLite caching 24h worth of history.
