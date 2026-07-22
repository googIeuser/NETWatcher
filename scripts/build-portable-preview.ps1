$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$flutter = Join-Path $repo "flutter_app"
$rust = Join-Path $repo "rust_core"
$dist = Join-Path $repo "dist"
$version = "4.0.1-preview"

& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")
& (Join-Path $PSScriptRoot "test-rust-flutter.ps1")

cargo build --manifest-path (Join-Path $rust "Cargo.toml") --release
Push-Location $flutter
try {
    flutter build windows --release
} finally {
    Pop-Location
}

$releaseDir = Join-Path $flutter "build\windows\x64\runner\Release"
$core = Join-Path $rust "target\release\netwatcher_core.exe"
if (-not (Test-Path $releaseDir)) { throw "Flutter release folder was not produced." }
if (-not (Test-Path $core)) { throw "Rust release core was not produced." }
Copy-Item $core (Join-Path $releaseDir "netwatcher_core.exe") -Force

New-Item -ItemType Directory -Force -Path $dist | Out-Null
$zip = Join-Path $dist "NetWatcher_${version}_Windows_Portable.zip"
Remove-Item $zip -Force -ErrorAction SilentlyContinue
Compress-Archive -Path (Join-Path $releaseDir "*") -DestinationPath $zip -CompressionLevel Optimal
$hash = (Get-FileHash $zip -Algorithm SHA256).Hash.ToLower()
"$hash  $(Split-Path $zip -Leaf)" | Set-Content "$zip.sha256" -Encoding ascii
Write-Host "Portable preview created: $zip"
