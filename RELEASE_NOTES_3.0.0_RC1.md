# NetWatcher 3.0.0 Release Candidate 1

NetWatcher 3 RC1 replaces the manually positioned Win32 interface with a responsive Wails + React + TypeScript frontend while retaining the local Go monitoring and diagnostics engine.

## Feature-complete migration

- Responsive Dashboard with live target status and selectable graph history
- Gateway, Cloudflare, Google and custom-target monitoring
- Ping, TCP, HTTP and HTTPS checks
- Latency, jitter, packet loss, uptime, P95 and quality scoring
- Confirmed local-network, full-outage, partial-access and high-latency incidents
- Persistent measurement, event and outage history
- Statistics and per-target breakdowns
- Outage History with active and completed incidents
- Printable HTML connection reports
- ISP evidence reports
- Diagnostics ZIP export containing local logs and summaries
- Add, edit and remove custom targets
- Turkish and English interface
- Light and dark themes
- Log retention controls
- Native Windows tray menu
- Start with Windows, start minimised and close-to-tray behavior
- Outage and recovery notifications
- Manual and automatic GitHub release checks
- Windows executable and NSIS installer build workflow

## Compatibility

The release candidate reads the existing `%APPDATA%\NetWatcher\settings.json` configuration and imports NetWatcher 2.x `samples_*.csv` and `outages.csv` history. New outage records are written to `outages_v3.csv`, so the older semicolon-delimited file is never modified or corrupted.

## Intentionally excluded

Access Mode and GoodbyeDPI functionality are not included.

## Validation status

The clean React production build, focused Go tests, race-enabled storage/config/statistics tests, Go vet, legacy-log compatibility tests and a Windows/amd64 source compile check passed in the preparation environment. The actual Wails window, WebView2 runtime, native tray, Windows notifications and NSIS installer must still be manually exercised on Windows before this branch replaces the stable 2.2.7 release.
