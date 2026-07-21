# NetWatcher 3.0.0 RC1 validation report

Preparation environment: Linux container with Go 1.23.2 and Node.js 22.16.0. Target application: Windows amd64.

## Passed

- `npm ci --offline` from a clean `frontend/node_modules` directory
- `npm run build` (TypeScript and Vite production build)
- `go vet ./internal/...`
- `go test ./internal/...`
- `go test -race ./internal/config ./internal/monitor ./internal/reports ./internal/statistics ./internal/storage`
- Legacy NetWatcher 2.x semicolon/BOM sample-log import test
- Legacy and early-preview outage-schema import tests
- Report HTML escaping and Turkish outage-category test
- Diagnostics ZIP content test
- Windows/amd64 compile check of the complete application source against a Wails v2 API-compatible compile stub
- JSON and GitHub Actions YAML parsing

## Not performed in this environment

- Real Wails/WebView2 application launch
- Native Windows ping output validation under every Windows display language
- Tray icon and tray menu interaction
- Windows startup registry behavior
- Windows toast/balloon notification delivery
- Sleep/wake and network-adapter transitions
- NSIS executable generation, installation, upgrade and uninstall
- Code signing and SmartScreen reputation

These Windows-specific checks must pass before RC1 replaces NetWatcher 2.2.7 as the stable release.
