$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$gitignore = Join-Path $repo ".gitignore"
$rules = @(
    "flutter_app/.dart_tool/",
    "flutter_app/build/",
    "flutter_app/.flutter-plugins",
    "flutter_app/.flutter-plugins-dependencies",
    "flutter_app/windows/flutter/ephemeral/",
    "rust_core/target/",
    "dist/"
)

if (-not (Test-Path $gitignore)) {
    New-Item -ItemType File -Path $gitignore | Out-Null
}
$current = Get-Content $gitignore -ErrorAction SilentlyContinue
foreach ($rule in $rules) {
    if ($current -notcontains $rule) {
        Add-Content -Path $gitignore -Value $rule
    }
}
Write-Host "Rust/Flutter build output rules added to .gitignore."
