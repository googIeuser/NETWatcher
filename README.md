# NetWatcher

NetWatcher is a lightweight Windows connection monitor and local diagnostics utility built with a **Flutter desktop interface** and a **Rust monitoring core**.

It continuously measures latency, jitter and packet loss, distinguishes local-network problems from wider internet failures, keeps local outage history and creates reports that can be shared with an ISP or regulator.

**Current version:** `5.2.0`

[Download the latest release](../../releases/latest) · [Changelog](CHANGELOG.md) · [Privacy](PRIVACY.md)

> NetWatcher works locally. It does not require an account and does not upload your measurements or log files.

## Highlights

Version 5.2.0 adds translucent glass styling to the console header, connection status and chart target chooser. See [RELEASE_NOTES_5.2.0.md](RELEASE_NOTES_5.2.0.md) for details.

Version 5.1.3 fixes overlapping network-map targets by giving each row stable spacing. See [RELEASE_NOTES_5.1.3.md](RELEASE_NOTES_5.1.3.md) for details.

Version 5.1.2 expands the network map to show up to ten targets, with a compact list at narrow widths. See [RELEASE_NOTES_5.1.2.md](RELEASE_NOTES_5.1.2.md) for details.

Version 5.1.1 makes latency history easier to read from the first samples, adds a target chooser for crowded charts, and removes the navigation icon backgrounds while keeping their tooltips. See [RELEASE_NOTES_5.1.1.md](RELEASE_NOTES_5.1.1.md) for details.

Version 5.1.0 introduces the network-map overview, a compact side navigation and a matching topology icon. The Windows shortcuts now use the same app identity and versioned icon as desktop notifications. See [RELEASE_NOTES_5.1.0.md](RELEASE_NOTES_5.1.0.md) for details.

Version 5.0.2 gives the Windows app, tray, shortcuts and installer a new lime-and-ink icon matching the control-desk interface. See [RELEASE_NOTES_5.0.2.md](RELEASE_NOTES_5.0.2.md) for details.

Version 5.0.1 fixes the latency chart legend for targets that do not respond and brings generated HTML/evidence reports into the new control-desk visual style. See [RELEASE_NOTES_5.0.1.md](RELEASE_NOTES_5.0.1.md) for details.

Version 5.0.0 introduces a new network control desk: top navigation, a signal-led overview, a four-column telemetry strip and a warm paper/ink visual system across reports, targets, history and settings. See [RELEASE_NOTES_5.0.0.md](RELEASE_NOTES_5.0.0.md) for details.

Version 4.0.8 polishes the refreshed interface with clearer report actions, a compact target form and a neutral target count badge. See [RELEASE_NOTES_4.0.8.md](RELEASE_NOTES_4.0.8.md) for details.

Version 4.0.7 makes the Windows interface easier to scan: a clearer monitoring state, compact metrics and target rows, and report actions that fit in common desktop windows. See [RELEASE_NOTES_4.0.7.md](RELEASE_NOTES_4.0.7.md) for details.

Version 4.0.6 reduces the work needed to load recent measurement history and includes reliability fixes for active outages, HTTP checks, notifications and core connection errors. See [RELEASE_NOTES_4.0.6.md](RELEASE_NOTES_4.0.6.md) for details.

Version 4.0.5 fixes a Rust core compilation error. See [RELEASE_NOTES_4.0.5.md](RELEASE_NOTES_4.0.5.md) for its release summary.

Version 4.0.4 introduced:

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
- Optional Windows notifications for outages and recoveries
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
NetWatcher_Setup_5.2.0.exe
```

The installer creates the normal Windows installation and uninstallation entries.

### Portable package

```text
NetWatcher_5.2.0_Windows_Portable.zip
```

Extract the complete ZIP before running `netwatcher.exe`. The Flutter application and `netwatcher_core.exe` must remain together in the extracted folder.

### Verify downloads

Each installer and portable ZIP is published with a matching `.sha256` file.

PowerShell example:

```powershell
(Get-FileHash .\NetWatcher_Setup_5.2.0.exe -Algorithm SHA256).Hash.ToLower()
Get-Content .\NetWatcher_Setup_5.2.0.exe.sha256
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
.\scripts\build-stable-release.ps1 -Version "5.2.0"
```

Generated files are written to:

```text
dist\
```

Expected release assets:

```text
NetWatcher_Setup_5.2.0.exe
NetWatcher_Setup_5.2.0.exe.sha256
NetWatcher_5.2.0_Windows_Portable.zip
NetWatcher_5.2.0_Windows_Portable.zip.sha256
```

## Contributing and security

Bug reports and focused pull requests are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before contributing.

Report security issues privately as described in [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
