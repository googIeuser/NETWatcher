# NetWatcher 2.1.0 Test Report

Date: 2026-07-20

## Completed in the Linux build environment

- `gofmt` completed for all Go source files.
- `go test -race ./...` passed.
- `go vet ./...` passed.
- Windows AMD64 cross-build completed with `CGO_ENABLED=0`.
- The generated setup file was identified as a PE32+ Windows GUI executable.
- Semantic-version comparison tests passed.
- CSV statistics aggregation tests passed.
- HTML statistics-page generation test passed.
- ZIP log-export test passed and produced a non-empty archive.
- Existing installer state, startup policy, and wizard model tests passed.

## Not executable in this environment

The Linux environment cannot perform real Windows interaction tests for tray notifications, taskbar behavior, startup registration, SmartScreen, Authenticode signatures, or visual layout. These must be checked on clean Windows 10 and Windows 11 virtual machines before marking the GitHub release as stable.
