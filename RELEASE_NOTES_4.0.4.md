# NetWatcher 4.0.4

NetWatcher 4.0.4 restores real latency history and the Windows startup controls in the Rust + Flutter desktop application.

## Restored

- Real per-target latency history sourced from locally stored measurements.
- 5-minute, 30-minute, 1-hour and 24-hour graph ranges.
- Readable millisecond and time-axis labels.
- Start NetWatcher with Windows.
- Start minimized in the notification area.
- Start monitoring automatically.

## Reliability

- Keeps up to 24 hours of history in memory and downsamples graph payloads.
- Prevents widget tests from modifying the Windows startup registry.
- Prevents the one-second polling timer from blocking Flutter widget tests.
- Disables obsolete Go CI and release workflows.
- Converts the stable workflow to a manual build-only workflow; GitHub Releases remain user-published.

## Release assets

- `NetWatcher_Setup_4.0.4.exe`
- `NetWatcher_4.0.4_Windows_Portable.zip`
- Matching `.sha256` checksum files
