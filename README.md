# NetWatcher

NetWatcher is a lightweight Windows connection monitor and local diagnostics utility. It records latency, jitter, packet loss and outages, separates local-network failures from wider internet failures, and produces evidence reports that can be shared with an ISP or regulator.

[Download the latest release](../../releases/latest)

## NetWatcher 3

NetWatcher 3 uses a responsive Wails + React + TypeScript interface while retaining the local Go monitoring engine. Stable application source is located in `next/`.

## Features

- Default gateway, Cloudflare, Google and custom-target monitoring
- Ping, TCP and HTTP/HTTPS target checks
- Live latency graph with 5-minute, 30-minute, 1-hour and 24-hour ranges
- Rolling packet-loss, jitter and connection-quality scoring
- Local-network, ISP-outage, partial-access and high-latency classification
- Target management with add, edit, rename and remove operations
- Statistics, Outage History and ISP Evidence reports
- CSV logs, printable HTML reports and one-click diagnostics ZIP export
- Configurable automatic log retention
- Windows outage and recovery notifications
- Native tray menu and start-with-Windows support
- Turkish and English interface
- Light and dark themes
- Automatic GitHub release checks
- NSIS installer with WebView2 runtime handling
- No telemetry, advertising or account requirement

Access Mode and GoodbyeDPI are not included.

## Custom target formats

A plain host or IP uses ICMP ping:

```text
1.1.1.1
example.com
```

A TCP target checks whether a port accepts a connection:

```text
tcp://example.com:443
```

An HTTP or HTTPS target checks a web endpoint:

```text
https://example.com/health
http://192.168.1.10/status
```

## Privacy

Monitoring, statistics and report generation happen locally. NetWatcher does not upload measurements, IP addresses or log files. When enabled, update checks contact GitHub's public Releases API. See [PRIVACY.md](PRIVACY.md).

## Installation

1. Open the [latest release](../../releases/latest).
2. Download `NetWatcher_Setup_3.0.0.exe`.
3. Run the installer.

Settings are stored under `%APPDATA%\NetWatcher`. Logs are stored under:

```text
Documents\NetWatcherLogs
```

Community builds may show a Windows SmartScreen unknown-publisher warning when they are not code-signed.

## Build from source

Requirements: Go 1.23+, Node.js, Wails v2.12.0, WebView2 and Windows 10/11.

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
.\scripts\build-wails.ps1 -Installer
```

The output is written to `next\build\bin`.

## Release process

Run the **NetWatcher 3 Stable Release** workflow with version `3.0.0`. It validates the source version, builds the Wails application and NSIS installer, generates SHA-256 verification data and publishes the GitHub Release as the latest stable version.

## Contributing and security

Bug reports and focused pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before contributing. Report security issues privately as described in [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
