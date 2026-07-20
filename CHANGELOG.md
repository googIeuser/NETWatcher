# Changelog

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
