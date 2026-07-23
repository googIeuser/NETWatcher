# Releasing NetWatcher

## Normal development flow

1. Copy the prepared source files into the repository.
2. Review the changes in GitHub Desktop.
3. Commit and push to `main`.
4. Confirm **NetWatcher CI** finishes successfully.

The CI workflow runs automatically for relevant pushes and pull requests. It performs the Rust and Flutter checks but does not create downloadable application packages or a GitHub Release.

## Create a Windows test build

1. Open **Actions → NetWatcher Test Build**.
2. Select **Run workflow** on `main`.
3. Enter the same version used in `flutter_app/pubspec.yaml` and `rust_core/Cargo.toml`.
4. Download the `NetWatcher-vx.y.z-Windows-TEST` artifact.
5. Test the installer and portable package on Windows.

The test workflow never creates a tag or GitHub Release.

## Publish the stable release

After the CI and Windows test build are successful:

1. Confirm `RELEASE_NOTES_x.y.z.md` exists and `CHANGELOG.md` is updated.
2. Open **Actions → NetWatcher Stable Release**.
3. Select **Run workflow** on `main`.
4. Enter the stable version without the `v` prefix.
5. The workflow runs the complete test suite again, builds the installer and portable package, creates the `vx.y.z` tag, and publishes or updates the GitHub Release.

No personal access token, custom repository secret, manual draft, or manual asset upload is required. The workflow uses GitHub's built-in token with `contents: write`.

## Optional Authenticode signing

Run `scripts/sign-release.ps1` with a publicly trusted Windows code-signing certificate before publication when signing is configured. Never commit a PFX file or password.

## Release verification

Verify each downloaded asset against its `.sha256` file:

```powershell
(Get-FileHash .\NetWatcher_Setup_4.0.4.exe -Algorithm SHA256).Hash
```

Before publication, test installation, upgrade, uninstall, Windows startup, startup-to-tray, minimize-to-tray, monitoring controls, every graph range, outage history, reports, and portable launch on clean Windows 10 and Windows 11 systems.
