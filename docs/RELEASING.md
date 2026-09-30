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

1. Add the current release notes at `docs/releases/x.y.z.md` and update `CHANGELOG.md`.
2. Update the Flutter, Rust, installer and documentation versions, then commit and push to `main`.
3. Create and push the annotated `vx.y.z` tag on that commit using the maintainer's Git identity.
4. Create a draft GitHub Release for that tag, using `docs/releases/x.y.z.md` as its description.
5. Open **Actions → NetWatcher Stable Release**, select **Run workflow** on `main`, and choose `stable-release` with the matching version.
6. The workflow runs tests, builds and verifies the Windows packages, checks the existing tag, uploads assets to the draft and publishes it as the latest stable release.

The workflow uses GitHub's built-in token. It requires the maintainer's tag and draft release to exist before publication.

Keep only the current release-note file in the repository. Previous release descriptions remain available in GitHub Releases, and the version history is preserved in `CHANGELOG.md`.
