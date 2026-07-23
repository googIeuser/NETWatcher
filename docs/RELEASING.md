# Releasing NetWatcher

NetWatcher uses one GitHub Actions workflow named **NetWatcher Stable Release**.

## Normal push

Every push to `main` starts the workflow automatically. Only the **Test Rust and Flutter** job runs. No installer, portable package, tag, or GitHub Release is created.

## Publish a stable release

After testing the application locally:

1. Confirm `flutter_app/pubspec.yaml` and `rust_core/Cargo.toml` use the intended version.
2. Confirm `RELEASE_NOTES_x.y.z.md` exists and `CHANGELOG.md` is updated.
3. Open **Actions → NetWatcher Stable Release**.
4. Select **Run workflow** on `main`.
5. Enter the version without the `v` prefix.
6. The workflow runs tests, builds the installer and portable ZIP, creates checksum files, creates the version tag, and publishes the GitHub Release.

No personal token, repository secret, manual draft, or manual release upload is required.
