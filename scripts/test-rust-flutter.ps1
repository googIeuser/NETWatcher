$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot

& (Join-Path $PSScriptRoot "prepare-flutter-windows.ps1")
cargo test --manifest-path (Join-Path $repo "rust_core\Cargo.toml")
if ($LASTEXITCODE -ne 0) { throw "Rust tests failed." }

Push-Location (Join-Path $repo "flutter_app")
try {
    flutter analyze
    if ($LASTEXITCODE -ne 0) { throw "Flutter analysis failed." }
    flutter test
    if ($LASTEXITCODE -ne 0) { throw "Flutter tests failed." }
} finally {
    Pop-Location
}
