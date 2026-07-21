## 3.0.0 - 2026-07-21

- Replaced the manually positioned Win32 interface with a responsive Wails, React and TypeScript frontend.
- Retained the Go monitoring, metrics, storage, reporting and Windows integration backend.
- Added stable Dashboard, Statistics, Outage History, Reports, Targets and Settings pages.
- Added a stable NSIS installer and NetWatcher 3 GitHub Release workflow.
- Preserved NetWatcher 2.x settings and local log-history compatibility.
- Access Mode and GoodbyeDPI are not included.

## 2.2.7 - 2026-07-21

- Removed Access Mode and all related controls, tray commands, proxy startup, and proxy test functionality.
- Restores an older NetWatcher-owned Windows proxy setting once during upgrade, then removes the legacy ownership state.
- Left monitoring, graphs, reports, targets, Settings, notifications, and tray behavior unchanged.

## 2.2.6

- Widened the Remove Target button so its complete text is visible.

## 2.2.5 - 2026-07-21

- Fixed main-window and Settings-window resize corruption.
- Added safe minimum window sizes and atomic relayout redraws.
- Prevented toolbar, checkbox, and footer controls from overlapping.
- Increased button widths so labels remain fully visible.
- Improved Settings card/background consistency.

# Changelog

## [2.2.4] - 2026-07-21

- Fixed clipped and overlapping controls in the main dashboard.
- Widened the report and history action columns for DPI-scaled displays.
- Rebuilt the Settings layout with larger cards, consistent spacing and fully visible labels.
- Prevented checkbox captions and section headings from drawing over each other.

## 2.2.3

- Fixed graph ranges so samples are positioned on the complete selected time window.
- Restored up to 24 hours of graph history from local CSV logs across application restarts.
- Added readable time-axis labels for 5-minute, 30-minute, 1-hour and 24-hour graphs.
- Redesigned Settings with modern sections, larger controls and complete untruncated labels.
- Removed ellipsis-based text shortening from application and installer drawing.


## 2.2.2

- Fixed spacing between the graph range selector and Outage History button.
- Modernized the main-window buttons with rounded owner-drawn styling.
- Added clearer primary, accent, secondary, and destructive button states.
- Kept the existing Win32 architecture and monitoring engine unchanged.

## 2.2.1

- Redesigned the dashboard connection panel for clearer, non-technical status information.
- Replaced the cramped multi-column table with readable per-target summaries.
- Explained connection quality using plain-language descriptions instead of an unexplained score.
- Renamed jitter to “Variation” on the main dashboard while retaining technical metrics in reports.
- Simplified the bottom status bar to Monitoring, Internet quality, Access Mode, samples, outages and downtime.
- Displayed local gateway health separately from overall internet quality.

## 2.2.0

- Added rolling jitter, packet-loss and connection-quality metrics to the dashboard.
- Added 5-minute, 30-minute, 1-hour and 24-hour graph ranges with downsampling.
- Added a Target Manager with add, edit, rename and remove operations.
- Added TCP and HTTP/HTTPS checks alongside ICMP ping targets.
- Added modern Outage History and 1/7/30-day ISP Evidence reports.
- Added a full tray quick menu for monitoring, reports, logs, export and Access Mode.
- Added configurable log retention and automatic cleanup.
- Added experimental browser/proxy-aware Access Mode using a local CONNECT proxy, DNS-over-HTTPS and first-flight TLS fragmentation without ICMP ping or a packet driver.
- Added safe restoration and stale-setting recovery for Windows proxy configuration.

## 2.1.3

- Removed the monitoring interval field from the main dashboard.
- Monitoring interval and timeout are now configured only in Settings.
- The Start button always uses the saved timing values from Settings.

## 2.1.2

- Removed the timeout field from the main dashboard; ping timeout is now configured only in Settings.
- Monitoring started from the dashboard uses the saved timeout value.
- Redesigned HTML reports with the same blue gradient, responsive cards, dark/light styling, and print-friendly layout as Statistics.

## 2.1.1

- Added an editable custom-target drop-down to make saved targets easy to select.
- Added a Remove Target button with confirmation.
- Default gateway, Cloudflare, and Google targets remain protected.
- Removed targets are deleted from settings and disappear from the graph immediately.

## 2.1.0

- Added 24-hour and 7-day statistics pages generated from local CSV logs.
- Added one-click ZIP export for logs and HTML reports.
- Added Windows notifications when an outage starts and when connectivity recovers.
- Added automatic update checks through GitHub Releases.
- Added update and notification preferences to Settings.
- Added a tray-menu command to check for updates manually.
- Added GitHub Actions CI, release packaging, optional code signing, and checksums.
- Added project documentation, privacy information, contribution guidance, and security policy.

## 2.0.7

- Minimize button sends the application to the notification area.
- Improved tray, startup, settings-dialog, installer, graph-legend, and monitoring-button behavior.
