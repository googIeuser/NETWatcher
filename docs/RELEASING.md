# Releasing NetWatcher

## Normal release

1. Update `appVersion` in `main.go` and `CHANGELOG.md`.
2. Commit and push the changes.
3. Create and push a matching tag:

```powershell
git tag -a v2.1.0 -m "NetWatcher 2.1.0"
git push origin v2.1.0
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
(Get-FileHash .\NetWatcher_Setup_2.1.0.exe -Algorithm SHA256).Hash
```

Test installation, upgrade, uninstall, startup-to-tray, minimize-to-tray, outage notification, statistics generation, ZIP export, and update checking on a clean Windows 10/11 virtual machine before announcing the release.
