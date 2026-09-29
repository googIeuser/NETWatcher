# Contributing

Thank you for helping improve NetWatcher.

1. Search existing issues before opening a new one.
2. Keep pull requests focused on one change.
3. Run the Rust tests and Flutter checks before submitting:

```powershell
.\scripts\test-rust-flutter.ps1
```

4. Confirm the Windows x64 build succeeds with `.\scripts\build-stable-release.ps1 -Version "5.1.3"`.
5. Do not commit certificates, private keys, generated EXE files, personal logs, or real IP-address evidence.

User-facing text is currently English. Changes that affect installer, tray, startup, or UI-thread behavior should include a clear manual Windows test plan.
