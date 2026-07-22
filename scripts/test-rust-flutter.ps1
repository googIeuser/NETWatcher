$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot

& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")
cargo test --manifest-path (Join-Path $repo "rust_core\Cargo.toml")

Push-Location (Join-Path $repo "flutter_app")
try {
    flutter analyze
    flutter test
} finally {
    Pop-Location
}
