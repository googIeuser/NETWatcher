# Contributing

Thank you for helping improve NetWatcher.

1. Search existing issues before opening a new one.
2. Keep pull requests focused on one change.
3. Run `go test -race ./...` before submitting.
4. Confirm the Windows AMD64 build succeeds:

```powershell
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build ./...
```

5. Do not commit certificates, private keys, generated EXE files, personal logs, or real IP-address evidence.

User-facing text is currently English. Changes that affect installer, tray, startup, or UI-thread behavior should include a clear manual Windows test plan.
