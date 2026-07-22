$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$flutter = Join-Path $repo "flutter_app"
$icon = Join-Path $flutter "assets\app_icon.ico"
$runnerMarker = Join-Path $flutter "windows\runner\Runner.rc"
$runnerIcon = Join-Path $flutter "windows\runner\resources\app_icon.ico"
$hashMarker = Join-Path $flutter "windows\runner\resources\.netwatcher-icon.sha256"

if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
    throw "Flutter was not found. Install Flutter stable and add it to PATH."
}
if (-not (Test-Path $icon)) {
    throw "NetWatcher icon was not found: $icon"
}

Push-Location $flutter
try {
    # A partially existing windows folder is not a valid Windows runner.
    if (-not (Test-Path $runnerMarker)) {
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
        if ($LASTEXITCODE -ne 0) {
            throw "Flutter Windows runner creation failed."
        }

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

    $runnerResourceDir = Split-Path -Parent $runnerIcon
    New-Item -ItemType Directory -Force -Path $runnerResourceDir | Out-Null

    $sourceHash = (Get-FileHash $icon -Algorithm SHA256).Hash.ToLower()
    $storedHash = ""
    if (Test-Path $hashMarker) {
        $storedHash = (Get-Content $hashMarker -Raw).Trim().ToLower()
    }

    $destinationHash = ""
    if (Test-Path $runnerIcon) {
        $destinationHash = (Get-FileHash $runnerIcon -Algorithm SHA256).Hash.ToLower()
    }

    # The old package could copy the correct icon without invalidating Ninja's
    # compiled Windows resource. Missing marker therefore also forces rebuild.
    $iconChanged = (
        $destinationHash -ne $sourceHash -or
        $storedHash -ne $sourceHash
    )

    Copy-Item $icon $runnerIcon -Force
    (Get-Item $runnerIcon).LastWriteTimeUtc = [DateTime]::UtcNow
    Set-Content -Path $hashMarker -Value $sourceHash -Encoding ascii

    $copiedHash = (Get-FileHash $runnerIcon -Algorithm SHA256).Hash.ToLower()
    if ($copiedHash -ne $sourceHash) {
        throw "Windows runner icon verification failed."
    }

    if ($iconChanged) {
        $windowsBuild = Join-Path $flutter "build\windows"
        if (Test-Path $windowsBuild) {
            Remove-Item $windowsBuild -Recurse -Force
        }
        Write-Host "NetWatcher icon changed; cached Windows resources were removed."
    }

    flutter pub get
    if ($LASTEXITCODE -ne 0) {
        throw "flutter pub get failed."
    }
} finally {
    Pop-Location
}
