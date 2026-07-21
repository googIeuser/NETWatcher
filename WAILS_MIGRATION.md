# NetWatcher Wails migration

The release-candidate architecture lives in `next/`. The stable NetWatcher 2.2.7 Win32 application remains at the repository root until RC1 is manually approved on Windows.

## Stack

- Go monitoring, storage, reporting and Windows integration backend
- Wails v2 desktop bridge
- React + TypeScript + Vite frontend
- WebView2 on Windows

## Feature status

| Area | Status |
| --- | --- |
| Live dashboard and responsive layout | Implemented |
| Ping, TCP, HTTP/HTTPS target checks | Implemented |
| Jitter, packet loss, uptime, P95 and quality score | Implemented |
| Confirmation cycles and outage classification | Implemented |
| Version 2 settings and log-history migration | Implemented and tested |
| CSV logging and retention | Implemented |
| Statistics and target breakdown | Implemented |
| Outage History | Implemented |
| HTML and ISP evidence reports | Implemented |
| Diagnostics ZIP export | Implemented |
| Target add/edit/remove | Implemented |
| Turkish/English and light/dark UI | Implemented |
| Tray, Windows startup and notifications | Implemented; Windows validation required |
| Update checks | Implemented |
| EXE and NSIS installer workflow | Implemented; Windows validation required |
| Access Mode / GoodbyeDPI | Intentionally absent |

## Data compatibility

- Settings: `%APPDATA%\NetWatcher\settings.json`
- Logs: `Documents\NetWatcherLogs`
- NetWatcher 2.x `samples_*.csv`: read by RC1
- NetWatcher 2.x `outages.csv`: read but never modified by RC1
- NetWatcher 3 measurements: `measurements_*.csv`
- NetWatcher 3 outages: `outages_v3.csv`
- NetWatcher 3 events: `events_v3.csv`

## Branch and release policy

Development belongs on `refactor/wails-ui`. Do not replace the stable release until RC1 has been tested on Windows for monitoring, resize/DPI behavior, sleep/wake, tray open/exit, startup registration, notification delivery, report generation, legacy history import and installer upgrade/uninstall behavior.
