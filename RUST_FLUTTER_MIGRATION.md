# NetWatcher Rust + Flutter migration

This workspace is the in-place replacement architecture for NetWatcher.

- `flutter_app/`: responsive Windows desktop user interface
- `rust_core/`: monitoring engine and JSON-line process bridge
- `scripts/run-rust-flutter.ps1`: builds the Rust core and launches Flutter
- Existing NetWatcher 3.0.0 Wails files remain untouched during validation.

## Current phase

Phase 1 provides:

- Responsive Dashboard, Statistics, Outages, Reports, Targets and Settings pages
- Layouts that reflow instead of clipping at narrow widths
- Widget tests at multiple Windows sizes and with long labels
- Rust models compatible with NetWatcher 3 settings
- `%APPDATA%\NetWatcher\settings.json` load/save compatibility
- Ping, TCP, HTTP and HTTPS checks
- Background Start/Stop monitoring
- JSON-line bridge between Flutter and Rust
- Mock fallback so the Flutter UI can be tested before the Rust binary is available

This is a development preview, not the NetWatcher 4 stable release.

## Windows prerequisites

- Flutter stable with Windows desktop enabled
- Rust stable (`rustup`, `cargo`)
- Visual Studio with Desktop development with C++

## Run

From the repository root:

```powershell
.\scripts\run-rust-flutter.ps1
```

The script creates the missing generated Windows runner, builds `rust_core`, sets
`NETWATCHER_CORE_PATH`, and launches the Flutter desktop app.

## Test

```powershell
.\scripts\test-rust-flutter.ps1
```


## Phase 3 reports

- Persistent measurement CSV logs
- Confirmed outage CSV history
- Printable HTML connection report
- 1/7/30-day ISP Evidence Report
- Diagnostics ZIP with summaries and raw logs
- Report file/folder opening through the Rust process bridge
