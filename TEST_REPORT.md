# NetWatcher 2.2.0 Test Report

Date: 2026-07-20

## Automated tests completed in Linux

- `go test -race ./...`
- `go vet ./...`
- Metrics tests for packet loss, average latency, jitter and quality scoring
- Graph-range normalization and downsampling tests
- Local Access Mode CONNECT-tunnel test using a loopback TCP echo server
- Initial-stream fragmentation data-integrity test
- Outage History and Evidence Report generation tests
- Log-retention cleanup test
- Existing installer state-machine, startup policy, release, target-removal and report-theme tests

## Windows build validation

- Windows AMD64 application cross-build completed successfully
- Windows AMD64 package test binaries compiled successfully with `go test -c`
- Embedded Windows compatibility manifest regenerated for version 2.2.0
- Output: `NetWatcher_Setup_2.2.0.exe`

## Not validated in this Linux environment

The following require a real Windows 10/11 test machine and must be checked before a public release:

- Native Target Manager and Access Mode window interaction
- Windows system-proxy application, restoration and forced-termination recovery
- Behaviour of Chrome, Edge, Firefox and other proxy-aware applications
- Real ISP/DPI compatibility of TLS first-flight fragmentation
- Windows tray menu commands
- Long-running 24-hour graph memory and rendering behaviour
- Upgrade from NetWatcher 2.1.x and uninstall behaviour
- SmartScreen and antivirus false-positive behaviour

Access Mode is intentionally described as experimental. It uses a standard local proxy and does not guarantee access through every DNS, IP, SNI, QUIC or active-DPI filtering system.
