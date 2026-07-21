# NetWatcher Wails migration

The Wails + React + TypeScript migration was approved for the stable NetWatcher 3.0.0 release after Windows validation. The stable application source lives in `next/`; the former Win32 2.2.7 source remains at the repository root for historical compatibility during the transition.

## Stack

- Go monitoring, storage, reporting and Windows integration backend
- Wails v2 desktop bridge
- React + TypeScript + Vite frontend
- WebView2 on Windows

## Stable feature status

| Area | Status |
| --- | --- |
| Live dashboard and responsive layout | Stable |
| Ping, TCP, HTTP/HTTPS target checks | Stable |
| Jitter, packet loss, uptime, P95 and quality score | Stable |
| Confirmation cycles and outage classification | Stable |
| Version 2 settings and log-history migration | Stable |
| CSV logging and retention | Stable |
| Statistics and target breakdown | Stable |
| Outage History | Stable |
| HTML and ISP evidence reports | Stable |
| Diagnostics ZIP export | Stable |
| Target add/edit/remove | Stable |
| Turkish/English and light/dark UI | Stable |
| Tray, Windows startup and notifications | Stable |
| Update checks | Stable |
| EXE and NSIS installer workflow | Stable |
| Access Mode / GoodbyeDPI | Intentionally absent |

## Data compatibility

- Settings: `%APPDATA%\NetWatcher\settings.json`
- Logs: `Documents\NetWatcherLogs`
- NetWatcher 2.x `samples_*.csv`: imported
- NetWatcher 2.x `outages.csv`: read but never modified
- NetWatcher 3 measurements: `measurements_*.csv`
- NetWatcher 3 outages: `outages_v3.csv`
- NetWatcher 3 events: `events_v3.csv`
