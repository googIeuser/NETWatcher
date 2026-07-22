param(
    [Parameter(Mandatory = $false)]
    [string]$Version = "4.0.3"
)

$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$flutter = Join-Path $repo "flutter_app"
$rust = Join-Path $repo "rust_core"
$dist = Join-Path $repo "dist"
$installerScript = Join-Path $repo "installer\NetWatcher.iss"

if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    throw "Stable version must use x.y.z format: $Version"
}

$cargoText = Get-Content (Join-Path $rust "Cargo.toml") -Raw
$cargoMatch = [regex]::Match($cargoText, '(?m)^version\s*=\s*"([^"]+)"')
if (-not $cargoMatch.Success) {
    throw "Rust version was not found."
}
if ($cargoMatch.Groups[1].Value -ne $Version) {
    throw "Rust version $($cargoMatch.Groups[1].Value) does not match requested version $Version."
}

$pubspecText = Get-Content (Join-Path $flutter "pubspec.yaml") -Raw
$flutterMatch = [regex]::Match($pubspecText, '(?m)^version:\s*([0-9]+\.[0-9]+\.[0-9]+)(?:\+\d+)?\s*$')
if (-not $flutterMatch.Success) {
    throw "Flutter stable version was not found."
}
if ($flutterMatch.Groups[1].Value -ne $Version) {
    throw "Flutter version $($flutterMatch.Groups[1].Value) does not match requested version $Version."
}

& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")
& (Join-Path $PSScriptRoot "test-rust-flutter.ps1")

cargo build --manifest-path (Join-Path $rust "Cargo.toml") --release
if ($LASTEXITCODE -ne 0) { throw "Rust release build failed." }

$windowsBuildCache = Join-Path $flutter "build\windows"
if (Test-Path $windowsBuildCache) {
    Remove-Item $windowsBuildCache -Recurse -Force
}
& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")

Push-Location $flutter
try {
    flutter build windows --release
    if ($LASTEXITCODE -ne 0) { throw "Flutter Windows release build failed." }
} finally {
    Pop-Location
}

$releaseDir = Join-Path $flutter "build\windows\x64\runner\Release"
$core = Join-Path $rust "target\release\netwatcher_core.exe"
$app = Join-Path $releaseDir "netwatcher.exe"
$releaseIcon = Join-Path $releaseDir "NetWatcher_4.0.3.ico"
if (-not (Test-Path $releaseDir)) { throw "Flutter release folder was not produced." }
if (-not (Test-Path $core)) { throw "Rust release core was not produced." }
if (-not (Test-Path $app)) { throw "Flutter executable was not produced: $app" }

Copy-Item (Join-Path $flutter "assets\app_icon.ico") $releaseIcon -Force

$sourceIconHash = (Get-FileHash (Join-Path $flutter "assets\app_icon.ico") -Algorithm SHA256).Hash
$releaseIconHash = (Get-FileHash $releaseIcon -Algorithm SHA256).Hash
if ($sourceIconHash -ne $releaseIconHash) {
    throw "Versioned release icon verification failed."
}

Copy-Item $core (Join-Path $releaseDir "netwatcher_core.exe") -Force
Set-Content -Path (Join-Path $releaseDir "VERSION.txt") -Value $Version -Encoding ascii

New-Item -ItemType Directory -Force -Path $dist | Out-Null
Get-ChildItem $dist -File -ErrorAction SilentlyContinue | Remove-Item -Force

$portable = Join-Path $dist "NetWatcher_${Version}_Windows_Portable.zip"
Compress-Archive -Path (Join-Path $releaseDir "*") -DestinationPath $portable -CompressionLevel Optimal

$innoCandidates = @(
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe"
)
$iscc = $innoCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $iscc) {
    throw "Inno Setup 6 was not found."
}

$sourceDefine = "/DSourceDir=$releaseDir"
$outputDefine = "/DOutputDir=$dist"
$versionDefine = "/DMyAppVersion=$Version"
& $iscc $versionDefine $sourceDefine $outputDefine $installerScript
if ($LASTEXITCODE -ne 0) { throw "Installer compilation failed." }

$installer = Join-Path $dist "NetWatcher_Setup_${Version}.exe"
if (-not (Test-Path $installer)) {
    throw "Installer was not produced: $installer"
}

foreach ($asset in @($portable, $installer)) {
    $hash = (Get-FileHash $asset -Algorithm SHA256).Hash.ToLower()
    "$hash  $(Split-Path $asset -Leaf)" |
        Set-Content "$asset.sha256" -Encoding ascii
}

Write-Host ""
Write-Host "Stable assets created:"
Write-Host "  $installer"
Write-Host "  $portable"
