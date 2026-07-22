# Releasing NetWatcher

## Normal release

1. Update `appVersion` in `version.go` and add the release notes to `CHANGELOG.md`.
2. Commit and push the changes.
3. Either run **Actions → Release → Run workflow** with the matching version, or create and push a matching tag:

```powershell
git tag -a v2.2.2 -m "NetWatcher 2.2.2"
git push origin v2.2.2
```

The release workflow builds a repository-aware setup executable. The application uses that repository identifier for GitHub update checks.

## Optional Authenticode signing

Create repository Actions secrets:

- `WINDOWS_CERTIFICATE_BASE64`: Base64-encoded PFX file
- `WINDOWS_CERTIFICATE_PASSWORD`: PFX password

The workflow skips signing when either secret is missing. Never commit a PFX file or password.

A publicly trusted code-signing certificate is required to show a verified publisher and improve SmartScreen reputation. Signing support cannot replace a certificate.

## Release verification

Download the EXE and `.sha256` asset from the GitHub Release, then verify:

```powershell
(Get-FileHash .\NetWatcher_Setup_2.2.2.exe -Algorithm SHA256).Hash
```

Before announcing a release, test installation, upgrade, uninstall, startup-to-tray, minimize-to-tray, target editing, ping/TCP/HTTPS checks, graph ranges, outage notifications, Statistics, Outage History, Evidence Report, ZIP export, and update checking on clean Windows 10 and Windows 11 virtual machines.

