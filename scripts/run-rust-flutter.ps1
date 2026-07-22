param(
    [switch]$Release
)

$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$flutter = Join-Path $repo "flutter_app"
$rust = Join-Path $repo "rust_core"

if (-not (Get-Command cargo -ErrorAction SilentlyContinue)) {
    throw "Cargo was not found. Install Rust with rustup."
}

& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")

$profile = if ($Release) { "release" } else { "debug" }
if ($Release) {
    cargo build --manifest-path (Join-Path $rust "Cargo.toml") --release
} else {
    cargo build --manifest-path (Join-Path $rust "Cargo.toml")
}
$core = Join-Path $rust "target\$profile\netwatcher_core.exe"
if (-not (Test-Path $core)) {
    throw "Rust core binary was not produced: $core"
}
$env:NETWATCHER_CORE_PATH = $core

Push-Location $flutter
try {
    if ($Release) {
        flutter run -d windows --release
    } else {
        flutter run -d windows
    }
} finally {
    Pop-Location
}
