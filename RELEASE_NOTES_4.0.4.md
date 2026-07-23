# NetWatcher 4.0.4

NetWatcher 4.0.4 restores real latency history and Windows startup controls in the Rust + Flutter desktop application.

## Restored

- Real per-target latency history sourced from locally stored measurements.
- 5-minute, 30-minute, 1-hour and 24-hour graph ranges.
- Start NetWatcher with Windows.
- Start minimized in the notification area.
- Start monitoring automatically.

## Interface improvements

- Uses brighter, easier-to-distinguish latency graph colors.
- Uses clear rounded millisecond steps on the chart axis.
- Makes chart labels, lines and latest sample markers easier to read.
- Replaces the compact latency range dropdown with the same filled History range control used elsewhere.
- Removes the empty divider strip above the first row on the Statistics page.
- Keeps the responsive sidebar and report pages stable at common Windows sizes.

## Reliability and release flow

- Keeps up to 24 hours of graph history and downsamples large graph payloads.
- Prevents widget tests from changing the Windows startup registry.
- Prevents continuous polling from blocking Flutter widget tests.
- Uses one NetWatcher Stable Release workflow for automatic push tests, manual Windows test builds and manual stable publication.
- Publishes the installer, portable ZIP and matching SHA256 files through the workflow.

## Release assets

- `NetWatcher_Setup_4.0.4.exe`
- `NetWatcher_4.0.4_Windows_Portable.zip`
- Matching `.sha256` checksum files
