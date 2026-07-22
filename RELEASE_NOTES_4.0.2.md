# NetWatcher 4.0.2

NetWatcher 4.0.2 is the stable Windows patch release for the Rust + Flutter edition.

## Fixes

- Replaces the Flutter default Windows icon with the NetWatcher icon
- Forces Windows resource cache invalidation before local and stable builds
- Verifies the source and runner icon with SHA-256 before compilation
- Restores close-to-tray behavior and the tray context menu
- Shows detailed outage history instead of only a total record count
- Keeps HTML reports, ISP Evidence Reports and Diagnostics ZIP export working

## Downloads

- `NetWatcher_Setup_4.0.2.exe`
- `NetWatcher_4.0.2_Windows_Portable.zip`
- Matching `.sha256` files

## Data compatibility

Settings remain under `%APPDATA%\NetWatcher`.

Measurements, outages and reports remain under:

`%USERPROFILE%\Documents\NetWatcherLogs`

The installer is unsigned unless code signing is configured separately, so Windows may show a publisher warning.
