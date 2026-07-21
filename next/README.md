# NetWatcher 3.0.0 RC1

This folder contains the feature-complete Wails + React + TypeScript release candidate. The stable 2.2.7 Win32 application remains at the repository root while RC1 is validated on Windows.

## Included functionality

- Gateway, Cloudflare, Google and custom target monitoring
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
- Turkish and English UI
- Light and dark themes
- Log retention
- Windows outage/recovery notifications
- Windows startup, start-minimised and close-to-tray behavior
- Native Windows tray menu
- Manual and automatic GitHub release checks
- Wails/NSIS Windows installer workflow
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

## Build executable

From the repository root:

```powershell
.\scripts\build-wails.ps1
```

The executable is written to `next\build\bin`.

## Build installer

Install NSIS and run:

```powershell
.\scripts\build-wails.ps1 -Installer
```

The installer is written to `next\build\bin`. The installer preserves `%APPDATA%\NetWatcher` settings and `Documents\NetWatcherLogs` when uninstalling.

## Manual Windows RC checklist

1. Start and stop monitoring; verify Gateway, Cloudflare and Google updates.
2. Resize the window from the minimum size through maximised mode at the Windows DPI scale in use.
3. Add, edit and remove Ping, TCP and HTTPS targets.
4. Verify 5m, 30m, 1h and 24h graphs after restarting the app.
5. Open Statistics and Outage History for every available range.
6. Generate standard HTML, ISP evidence and diagnostics ZIP exports.
7. Verify light/dark mode and Turkish/English text.
8. Test close-to-tray, tray open/start/stop/logs/exit and Windows startup.
9. Test outage/recovery notifications.
10. Install and uninstall the NSIS package; confirm settings and logs are retained.

## Automated validation completed

The React production build, focused Go tests, legacy 2.x log compatibility tests, race-enabled storage/config/statistics tests, Go vet and a Windows/amd64 source compile check passed in the preparation environment. The actual Wails window, WebView2, tray, notifications and NSIS installer still require the manual Windows checklist above.
