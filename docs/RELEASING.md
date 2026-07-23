# Releasing NetWatcher

## Prepare version 4.x

1. Update the version in `flutter_app/pubspec.yaml` and `rust_core/Cargo.toml`.
2. Add `RELEASE_NOTES_x.y.z.md` and update `CHANGELOG.md`.
3. Push the changes and confirm **Rust Flutter Preview** finishes successfully.
4. In **Actions → NetWatcher Stable Build**, run the workflow with the same version.
5. Download the stable Actions artifact containing the installer, portable ZIP and checksum files.
6. Open **Releases → Draft a new release**, create tag `vx.y.z`, target `main`, attach the stable files and use the matching release-notes file.
7. Mark the release as latest and publish it manually.

Creating a tag does not publish or modify a GitHub Release automatically.

## Optional Authenticode signing

Run `scripts/sign-release.ps1` with a publicly trusted Windows code-signing certificate before uploading the installer. Never commit a PFX file or password.

## Release verification

Verify each downloaded asset against its `.sha256` file:

```powershell
(Get-FileHash .\NetWatcher_Setup_4.0.4.exe -Algorithm SHA256).Hash
```

Before publishing, test installation, upgrade, uninstall, Windows startup, startup-to-tray, minimize-to-tray, monitoring controls, all graph ranges, outage history, reports and portable launch on clean Windows 10 and Windows 11 systems.
