# NetWatcher 3.0.0

NetWatcher 3.0.0 replaces the manually positioned Win32 interface with a responsive Wails + React + TypeScript frontend while retaining the local Go monitoring and diagnostics engine.

## Highlights

- Fully redesigned responsive Windows desktop interface
- Live Dashboard with target status and selectable graph history
- Gateway, Cloudflare, Google and custom-target monitoring
- Ping, TCP, HTTP and HTTPS checks
- Latency, jitter, packet loss, uptime, P95 and quality scoring
- Local-network, full-outage, partial-access and high-latency classification
- Persistent measurements, event history and outage history
- Statistics and per-target breakdowns
- Printable HTML connection reports and ISP evidence reports
- Diagnostics ZIP export containing local logs and summaries
- Add, edit and remove custom targets
- Turkish and English interface
- Light and dark themes
- Native Windows tray menu
- Start with Windows, start minimised and close-to-tray behavior
- Outage and recovery notifications
- Manual and automatic GitHub release checks
- Stable NSIS installer with WebView2 runtime handling

## Compatibility

NetWatcher 3.0.0 reads the existing `%APPDATA%\NetWatcher\settings.json` configuration and imports NetWatcher 2.x `samples_*.csv` and `outages.csv` history. New outage records use `outages_v3.csv`, leaving the older file untouched.

## Removed

Access Mode and GoodbyeDPI functionality are not included.
