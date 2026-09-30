# Build from source

## Requirements

- Windows 10 or Windows 11, x64
- Flutter stable with Windows desktop support
- Dart SDK compatible with `>=3.4.0 <4.0.0`
- Rust stable
- Visual Studio 2022 with **Desktop development with C++**
- Inno Setup 6 for installer creation
- PowerShell

Run the following commands from the repository root.

## Prepare and test

```powershell
.\scripts\prepare-flutter-windows.ps1
.\scripts\test-rust-flutter.ps1
```

The preparation script generates the Windows runner and applies NetWatcher's icon. Generated runner files and build output are not kept in Git.

## Run locally

```powershell
.\scripts\run-rust-flutter.ps1
```

## Build release packages

```powershell
.\scripts\build-stable-release.ps1 -Version "5.1.4"
```

The Flutter and Rust package versions must match the requested release version. Output is written to `dist/`:

```text
NetWatcher_Setup_5.1.4.exe
NetWatcher_Setup_5.1.4.exe.sha256
NetWatcher_5.1.4_Windows_Portable.zip
NetWatcher_5.1.4_Windows_Portable.zip.sha256
```

`scripts/generate-app-icon.py` regenerates the source icons and requires Pillow. `scripts/sign-release.ps1` signs release files using a maintainer's Authenticode certificate and the Windows SDK. Keep certificates and private keys out of the repository.
