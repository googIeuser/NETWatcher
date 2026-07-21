# NetWatcher 3.0.0

This folder contains the stable Wails + React + TypeScript NetWatcher desktop application. The Go monitoring and diagnostics engine remains local, while the interface is responsive and component based.

## Included functionality

- Gateway, Cloudflare, Google and custom-target monitoring
- Ping, TCP, HTTP and HTTPS checks
- Responsive live dashboard and 5-minute, 30-minute, 1-hour and 24-hour graph ranges
- Latency, jitter, packet loss, uptime, P95 and quality scoring
- Confirmed local-network, full-outage, partial-access and high-latency incidents
- Persistent measurements, events and outage history
- Statistics and per-target breakdowns
- Outage History
- Printable HTML report and ISP evidence report
- Diagnostics ZIP export and log-folder access
- Add, edit and remove custom targets
- Turkish and English interface
- Light and dark themes
- Log retention
- Windows outage and recovery notifications
- Windows startup, start-minimised and close-to-tray behavior
- Native Windows tray menu
- Manual and automatic GitHub release checks
- Wails/NSIS Windows installer
- Existing `%APPDATA%\NetWatcher\settings.json` compatibility
- Existing NetWatcher 2.x `samples_*.csv` and `outages.csv` history import

Access Mode and GoodbyeDPI are not included.

## Run locally

Requirements: Go 1.23+, Node.js, Wails v2.12.0 and WebView2.

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
cd next
wails doctor
wails dev
```

## Validate

From `next`:

```powershell
cd frontend
npm ci
npm run build
cd ..
go test ./internal/...
go vet ./internal/...
```

## Build installer

From the repository root:

```powershell
.\scripts\build-wails.ps1 -Installer
```

The installer is written to `next\build\bin`. Uninstalling preserves `%APPDATA%\NetWatcher` settings and `Documents\NetWatcherLogs`.
