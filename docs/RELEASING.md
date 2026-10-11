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
3. Open **Actions → NetWatcher Stable Release**, select **Run workflow** on `main`, and choose `stable-release` with the matching version.
4. The workflow runs tests, builds and verifies the Windows packages, creates the `vx.y.z` tag on that commit if it does not exist, and publishes the release with the packages as the latest stable release.

New releases are published with the `RELEASE_TOKEN` repository secret, so they appear under the maintainer's account. Create it once as a fine-grained personal access token for this repository with **Contents: Read and write** permission, and add it under **Settings → Secrets and variables → Actions**. Without the secret, the workflow can only publish a draft release that the maintainer created beforehand.

Keep only the current release-note file in the repository. The version history is preserved in `CHANGELOG.md`.
