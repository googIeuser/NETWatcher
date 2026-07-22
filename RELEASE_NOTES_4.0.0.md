# NetWatcher 4.0.0

NetWatcher 4 is a complete Windows desktop rewrite using a Flutter interface and a Rust monitoring core.

## Highlights

- Responsive dashboard designed for narrow and wide Windows layouts
- Smooth page, card, metric, graph and light/dark theme transitions
- Desktop hover feedback and clickable mouse pointers
- Ping, TCP, HTTP and HTTPS target monitoring
- Latency, packet loss, jitter and quality measurements
- Measurement logging and outage history
- Standard HTML reports
- 1, 7 and 30-day ISP Evidence Reports with print/save-to-PDF support
- Diagnostics ZIP exports containing summaries and raw CSV records
- Existing NetWatcher settings and log directory compatibility

## Downloads

- `NetWatcher_Setup_4.0.0.exe`: standard Windows installer
- `NetWatcher_4.0.0_Windows_Portable.zip`: portable package
- Matching `.sha256` files for integrity verification

## Data location

Settings remain under `%APPDATA%\NetWatcher`.

Measurements, outages and reports remain under:

`%USERPROFILE%\Documents\NetWatcherLogs`

## Upgrade notes

NetWatcher 3.0.0 remains available from previous GitHub Releases. Installing NetWatcher 4 does not delete existing settings, measurements or generated reports.

The Windows installer is unsigned unless code-signing is configured separately, so Microsoft Defender SmartScreen may display a publisher warning.
