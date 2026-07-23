# Releasing NetWatcher

NetWatcher uses one GitHub Actions workflow named **NetWatcher Stable Release**.

## Normal push

Every push to `main` starts the workflow automatically. Only the Rust and Flutter test job runs. It does not create a package, tag or GitHub Release.

## Create a Windows test package

1. Open **Actions → NetWatcher Stable Release**.
2. Select **Run workflow** on `main`.
3. Choose `test-build`.
4. Enter the version used by Flutter and Rust.
5. Download the `NetWatcher-vx.y.z-Windows-TEST` artifact.
6. Test both the installer and portable ZIP locally.

The test-build mode never creates a tag or GitHub Release.

## Publish the stable release

After the test package is approved:

1. Confirm `RELEASE_NOTES_x.y.z.md` exists and `CHANGELOG.md` is updated.
2. Open **Actions → NetWatcher Stable Release**.
3. Select **Run workflow** on `main`.
4. Choose `stable-release` and enter the version.
5. The workflow runs tests again, builds the Windows files, creates the version tag and publishes the GitHub Release.

No personal token, repository secret, manual draft or manual asset upload is required.
