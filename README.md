# NetWatcher

NetWatcher is a lightweight Windows connection monitor and local diagnostics utility. It records latency, jitter, packet loss and outages, separates local-network failures from wider internet failures, and produces modern evidence reports that can be shared with an ISP or regulator.

[Download the latest release](../../releases/latest)

![NetWatcher dark interface](docs/netwatcher-dark.png)

## Features

- Default gateway, Cloudflare, Google and custom-target monitoring
- Ping, TCP and HTTP/HTTPS target checks
- Live latency graph with 5-minute, 30-minute, 1-hour and 24-hour ranges
- Rolling packet-loss, jitter and connection-quality scoring
- Local-network, ISP-outage, partial-access and high-latency classification
- Target Manager for adding, editing, renaming and removing custom targets
- Modern Statistics, Outage History and ISP Evidence HTML pages
- 24-hour, 7-day and 30-day evidence-report ranges
- CSV logs, printable HTML reports and one-click ZIP export
- Configurable automatic log retention
- Windows outage and recovery notifications
- Full tray quick menu and start-with-Windows support
- Light and dark themes
- Automatic GitHub release checks
- Per-user one-click installation with an integrated uninstaller
- No telemetry, advertising or account requirement

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

An HTTP or HTTPS target checks an actual web endpoint instead of sending ping:

```text
https://example.com/health
http://192.168.1.10/status
```

## Privacy

Monitoring, statistics and report generation happen locally. NetWatcher does not upload measurements, IP addresses or log files. When enabled, update checks contact GitHub's public Releases API. See [PRIVACY.md](PRIVACY.md).

## Installation

1. Open the [latest release](../../releases/latest).
2. Download `NetWatcher_Setup_<version>.exe`.
3. Double-click it and approve the installation.

The application installs for the current Windows user. Logs are stored in:

```text
Documents\NetWatcherLogs
```

Because community builds may not be code-signed, Windows SmartScreen can show an unknown-publisher warning. Official signed releases can be produced by configuring the signing secrets described in [docs/RELEASING.md](docs/RELEASING.md).

## Build from source

Requirements: Go 1.23 or newer and Windows 10/11 for runtime testing.

```powershell
./scripts/build.ps1
```

The output is written to `dist/NetWatcher_Setup_2.2.7.exe`.

## Release process

Push a semantic-version tag such as `v2.2.7`, or run the **Release** workflow manually with version `2.2.7`. GitHub Actions tests the project, builds the Windows setup executable, optionally signs it, generates a SHA-256 checksum, and creates a GitHub Release. See [docs/RELEASING.md](docs/RELEASING.md).

## Contributing and security

Bug reports and focused pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before contributing. Report security issues privately as described in [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
