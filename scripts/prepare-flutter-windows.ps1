$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$flutter = Join-Path $repo "flutter_app"
$icon = Join-Path $flutter "assets\app_icon.ico"

if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
    throw "Flutter was not found. Install Flutter stable and add it to PATH."
}

Push-Location $flutter
try {
    if (-not (Test-Path "windows")) {
        $backup = Join-Path $env:TEMP "netwatcher-flutter-backup-$PID"
        New-Item -ItemType Directory -Force -Path $backup | Out-Null
        Copy-Item "lib" $backup -Recurse
        Copy-Item "test" $backup -Recurse
        Copy-Item "pubspec.yaml" $backup
        Copy-Item "analysis_options.yaml" $backup
        if (Test-Path "assets") {
            Copy-Item "assets" $backup -Recurse
        }
        flutter create --platforms=windows --project-name netwatcher --org com.netwatcher .
        Remove-Item "lib" -Recurse -Force
        Remove-Item "test" -Recurse -Force -ErrorAction SilentlyContinue
        Copy-Item (Join-Path $backup "lib") "." -Recurse
        Copy-Item (Join-Path $backup "test") "." -Recurse
        Copy-Item (Join-Path $backup "pubspec.yaml") "." -Force
        Copy-Item (Join-Path $backup "analysis_options.yaml") "." -Force
        if (Test-Path (Join-Path $backup "assets")) {
            Remove-Item "assets" -Recurse -Force -ErrorAction SilentlyContinue
            Copy-Item (Join-Path $backup "assets") "." -Recurse
        }
        Remove-Item $backup -Recurse -Force
    }
    $runnerIcon = Join-Path $flutter "windows\runner\resources\app_icon.ico"
    if (-not (Test-Path $icon)) {
        throw "NetWatcher icon was not found: $icon"
    }
    if (-not (Test-Path (Split-Path -Parent $runnerIcon))) {
        throw "Flutter Windows runner resources folder was not found."
    }
    Copy-Item $icon $runnerIcon -Force
    flutter pub get
} finally {
    Pop-Location
}
