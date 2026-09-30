# NetWatcher

A local Windows connection monitor with a Flutter desktop interface and a Rust monitoring core.

**Current version:** `5.1.4`

[Download for Windows](https://github.com/googIeuser/NETWatcher/releases/latest) · [Changelog](CHANGELOG.md) · [Release notes](docs/releases/5.1.4.md)

## Features

- Monitor the default gateway, Cloudflare, Google and custom targets using ping, TCP or HTTP/HTTPS.
- Track latency, packet loss, jitter and connection quality, with a network map showing up to ten targets.
- View per-target graphs for the last 5 minutes, 30 minutes, hour or 24 hours.
- Keep local outage history and generate HTML reports, ISP evidence reports and diagnostics ZIPs.
- Use light or dark mode, tray controls, startup options and outage notifications.

NetWatcher works locally, requires no account and does not upload measurements or log files. Read the [privacy policy](docs/PRIVACY.md).

## Installation

Choose an installer or portable ZIP from the [latest release](https://github.com/googIeuser/NETWatcher/releases/latest).

- **Installer:** run `NetWatcher_Setup_5.1.4.exe`.
- **Portable:** extract the complete `NetWatcher_5.1.4_Windows_Portable.zip`, then run `netwatcher.exe`. Keep the extracted files together.

Both packages include a `.sha256` file. To verify the installer in PowerShell:

```powershell
(Get-FileHash .\NetWatcher_Setup_5.1.4.exe -Algorithm SHA256).Hash.ToLower()
Get-Content .\NetWatcher_Setup_5.1.4.exe.sha256
```

The two hashes should match. Unsigned builds may show a Windows SmartScreen unknown-publisher warning.

## Custom targets

Add targets from the **Targets** page:

| Format | Check |
| --- | --- |
| `example.com` or `1.1.1.1` | ICMP ping |
| `tcp://example.com:443` | TCP connection |
| `https://example.com/health` | HTTP/HTTPS endpoint |

## Local data

Settings are stored in `%APPDATA%\NetWatcher`. Measurements, outages and reports are stored in `Documents\NetWatcherLogs`.

## Development

| Directory | Purpose |
| --- | --- |
| `flutter_app/` | Desktop interface and UI tests |
| `rust_core/` | Monitoring, storage, reporting and core tests |
| `scripts/` | Development, testing and packaging tools |
| `installer/` | Windows installer definition |
| `docs/` | Project guides and current release notes |

The app starts `netwatcher_core.exe` locally and communicates through JSON commands over standard input and output.

[Build from source](docs/BUILDING.md) · [Contributing](docs/CONTRIBUTING.md) · [Release process](docs/RELEASING.md) · [Security](docs/SECURITY.md) · [Code of conduct](docs/CODE_OF_CONDUCT.md)

## License

MIT — see [LICENSE](LICENSE).
