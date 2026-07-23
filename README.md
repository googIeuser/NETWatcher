# NetWatcher

NetWatcher is a lightweight Windows connection monitor and local diagnostics utility built with a **Flutter desktop interface** and a **Rust monitoring core**.

It continuously measures latency, jitter and packet loss, distinguishes local-network problems from wider internet failures, keeps local outage history and creates reports that can be shared with an ISP or regulator.

**Current version:** `4.0.4`

[Download the latest release](../../releases/latest) · [Changelog](CHANGELOG.md) · [Privacy](PRIVACY.md)

> NetWatcher works locally. It does not require an account and does not upload your measurements or log files.

## Highlights in 4.0.4

- Restored real per-target latency history from locally stored measurements
- Restored 5-minute, 30-minute, 1-hour and 24-hour graph ranges
- Added brighter graph series, thicker lines and latest-sample markers
- Changed the latency axis to clear, rounded millisecond intervals
- Added a confirmation-protected action for deleting saved outage history
- Preserves an outage that is still active when saved history is cleared
- Records both monitoring start and monitoring stop actions in Recent events
- Restored start-with-Windows, start-minimized and automatic-monitoring controls
- Improved responsive layouts for common Windows desktop sizes

See [RELEASE_NOTES_4.0.4.md](RELEASE_NOTES_4.0.4.md) for the complete release summary.

## Features

### Live connection monitoring

- Monitors the default gateway, Cloudflare, Google and user-defined targets
- Supports ICMP ping, TCP and HTTP/HTTPS checks
- Displays average latency, packet loss, jitter, sample count and connection quality
- Classifies failures as:
  - Local network failure
  - Internet outage
  - Partial access
  - High latency / degraded connection
- Keeps a Recent events timeline for monitoring and connection-state changes

### Latency history and statistics

- Real per-target latency history restored from local measurement logs
- Selectable history ranges:
  - Last 5 minutes
  - Last 30 minutes
  - Last hour
  - Last 24 hours
- Readable rounded millisecond axis
- Distinct high-contrast target colors
- Latest-sample markers and graph glow
- Automatic downsampling for large graph histories
- Target-by-target Statistics page

### Outage History

- Filter by:
  - Last 24 hours
  - Last 7 days
  - Last 30 days
  - Last year
  - All time
- Shows active and resolved incidents
- Displays start time, end time, duration and diagnostic details
- Summarizes incident count, active incidents, total downtime and longest incident
- Refreshes outage data while the application is running
- Saved history can be deleted after confirmation
- An outage currently in progress remains visible after saved history is cleared

### Reports and exports

NetWatcher generates all reports locally:

- **HTML report** — connection measurements, target summaries and completed outages
- **ISP Evidence Report** — availability, packet loss, latency, jitter and outage evidence
- **Diagnostics ZIP** — settings, current snapshot, calculated summaries, outages and original CSV logs

Reports are saved under:

```text
Documents\NetWatcherLogs\Reports
```

### Windows desktop integration

- Native notification-area icon
- Open NetWatcher from the tray
- Start or stop monitoring from the tray menu
- Keep running in the notification area when the window is closed
- Start NetWatcher with Windows
- Start minimized after Windows login
- Start monitoring automatically
- Light and dark themes
- Responsive layouts for desktop and compact window sizes

## Custom target formats

A plain hostname or IP address uses ICMP ping:

```text
1.1.1.1
example.com
```

A TCP target checks whether the specified port accepts a connection:

```text
tcp://example.com:443
tcp://192.168.1.10:22
```

An HTTP or HTTPS target checks a web endpoint:

```text
https://example.com/health
http://192.168.1.10/status
```

Default targets are managed by NetWatcher. Custom targets can be added and removed from the **Targets** page.

## Installation

Open the [latest release](../../releases/latest) and choose one of the Windows packages.

### Installer

```text
NetWatcher_Setup_4.0.4.exe
```

The installer creates the normal Windows installation and uninstallation entries.

### Portable package

```text
NetWatcher_4.0.4_Windows_Portable.zip
```

Extract the complete ZIP before running `netwatcher.exe`. The Flutter application and `netwatcher_core.exe` must remain together in the extracted folder.

### Verify downloads

Each installer and portable ZIP is published with a matching `.sha256` file.

PowerShell example:

```powershell
(Get-FileHash .\NetWatcher_Setup_4.0.4.exe -Algorithm SHA256).Hash.ToLower()
Get-Content .\NetWatcher_Setup_4.0.4.exe.sha256
```

The two hash values should match.

Community builds may show a Windows SmartScreen unknown-publisher warning when they are not code-signed.

## Local data and privacy

Settings are stored at:

```text
%APPDATA%\NetWatcher\settings.json
```

Measurements, events and outage records are stored at:

```text
Documents\NetWatcherLogs
```

NetWatcher does not require an account, does not contain advertising and does not upload measurements, target addresses or report files. See [PRIVACY.md](PRIVACY.md) for more information.

## Architecture

```text
flutter_app/        Flutter Windows interface
rust_core/          Rust monitoring, storage and reporting core
scripts/            Test, preparation and release build scripts
installer/          Inno Setup installer definition
dist/               Generated release assets
```

The Flutter application starts `netwatcher_core.exe` locally and communicates with it through a small JSON command interface over standard input and output.

## Build from source

### Requirements

- Windows 10 or Windows 11, x64
- Flutter stable with Windows desktop support
- Dart SDK compatible with `>=3.4.0 <4.0.0`
- Rust stable
- Visual Studio 2022 with **Desktop development with C++**
- Inno Setup 6 for installer creation
- PowerShell

### Prepare and test

From the repository root:

```powershell
.\scripts\prepare-flutter-windows.ps1
.\scripts\test-rust-flutter.ps1
```

### Build installer and portable assets

```powershell
.\scripts\build-stable-release.ps1 -Version "4.0.4"
```

Generated files are written to:

```text
dist\
```

Expected release assets:

```text
NetWatcher_Setup_4.0.4.exe
NetWatcher_Setup_4.0.4.exe.sha256
NetWatcher_4.0.4_Windows_Portable.zip
NetWatcher_4.0.4_Windows_Portable.zip.sha256
```

## Contributing and security

Bug reports and focused pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before contributing.

Report security issues privately as described in [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
