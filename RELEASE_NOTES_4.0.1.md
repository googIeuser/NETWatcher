# NetWatcher 4.0.1

NetWatcher 4.0.1 is a Windows desktop maintenance release focused on app identity, outage diagnostics and tray reliability.

## Fixed

- Replaced the default Flutter icon with the NetWatcher icon for the executable, taskbar, shortcuts, installer and system tray.
- Restored close-to-tray behavior when **Keep NetWatcher running in the notification area when the window closes** is enabled.
- Added a working tray menu with Open NetWatcher, Start/Stop monitoring and Exit actions.
- Restored the application window from a left-click on the tray icon.
- Kept the Rust monitoring core running while the window is hidden in the notification area.

## Outage History improvements

- Replaced the single outage-count message with a detailed incident history.
- Added incident category, active/resolved state, start time, end time, duration and diagnostic description.
- Added 24-hour, 7-day, 30-day, yearly and all-time ranges.
- Added incident count, active incident count, total downtime and longest incident summaries.
- Included active incidents and refreshed their duration while monitoring continues.
- Improved compatibility with NetWatcher 2.x, 3.x and 4.x outage CSV formats.
- Normalized legacy outage categories and removed duplicate history entries.

## Downloads

- `NetWatcher_Setup_4.0.1.exe`: standard Windows installer
- `NetWatcher_4.0.1_Windows_Portable.zip`: portable package
- Matching `.sha256` files for integrity verification

## Upgrade notes

NetWatcher 4.0.1 uses the same application identity and data locations as 4.0.0, so the installer upgrades the existing installation without deleting settings, measurements, outage history or reports.

The Windows installer is unsigned unless code-signing is configured separately, so Microsoft Defender SmartScreen may display a publisher warning.
