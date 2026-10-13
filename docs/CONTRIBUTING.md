# Contributing

Thank you for helping improve NetWatcher.

Please follow the [code of conduct](CODE_OF_CONDUCT.md). Project documentation, user-facing text, issues, and pull requests should be written in English.

## Choose the right channel

- Installation or usage questions: [support guide](SUPPORT.md) and [help form](https://github.com/googIeuser/NETWatcher/issues/new?template=help.yml).
- Reproducible defects or focused proposals: [issue forms](https://github.com/googIeuser/NETWatcher/issues/new/choose).
- Security vulnerabilities: the private channel in [SECURITY.md](SECURITY.md), not a public issue or pull request.

Before a substantial change, describe the user problem and proposed behavior in an issue. Small corrections can go straight to a pull request. Include compatibility, privacy, monitoring, and performance implications when relevant.

## Development and verification

Follow [BUILDING.md](BUILDING.md) for the Windows x64 toolchain and repository setup. Run commands from the repository root.

1. Search existing issues before opening a new one.
2. Keep pull requests focused on one change.
3. For code changes, run the Rust tests and Flutter checks before submitting:

```powershell
.\scripts\test-rust-flutter.ps1
```

4. For code or packaging changes, confirm the Windows x64 build succeeds with `.\scripts\build-stable-release.ps1 -Version "5.1.6"`; use the current matching package version when it changes.
5. Do not commit certificates, private keys, generated EXE files, personal logs, or real IP-address evidence.

User-facing text is currently English. Changes that affect installer, tray, startup, or UI-thread behavior should include a clear manual Windows test plan.

## Pull requests and review

Use the pull request template to describe the problem, resulting behavior, related issue, and verification commands/results. State which checks were not run and why. Documentation-only changes need link/template validation rather than a local application rebuild; the repository's configured CI may still run automatically.

Keep generated build output and signing material out of Git. Use synthetic test data and remove identifying details from screenshots and logs. Maintainer feedback may request a smaller scope or additional evidence. Contributions are distributed under the project's [MIT license](../LICENSE). See [RELEASING.md](RELEASING.md) for the separate maintainer release process.
