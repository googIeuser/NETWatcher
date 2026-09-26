# NetWatcher 4.0.6

NetWatcher 4.0.6 improves recent-history loading and monitoring reliability on Windows.

## Improvements and fixes

- Reads only CSV headers to detect delimiters and skips daily measurement files older than the selected history window.
- Keeps active outages across restarts and closes them when monitoring stops or the connection recovers.
- Applies changed HTTP timeouts to subsequent checks.
- Shows a connection error if the Rust core cannot start, instead of displaying sample data.
- Restores outage and recovery notifications.

## Release assets

- `NetWatcher_Setup_4.0.6.exe`
- `NetWatcher_4.0.6_Windows_Portable.zip`
- Matching `.sha256` checksum files
