$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$VersionMatch = Select-String -Path (Join-Path $Root "version.go") -Pattern 'appVersion\s*=\s*"([^"]+)"'
if (-not $VersionMatch) { throw "appVersion was not found." }
$Version = $VersionMatch.Matches[0].Groups[1].Value
$Dist = Join-Path $Root "dist"
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

Push-Location $Root
try {
    go test ./...
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    $Output = Join-Path $Dist "NetWatcher_Setup_$Version.exe"
    go build -trimpath -ldflags "-s -w -H=windowsgui" -o $Output .
    Get-FileHash $Output -Algorithm SHA256 | ForEach-Object { "$($_.Hash.ToLower())  $([IO.Path]::GetFileName($Output))" } | Set-Content "$Output.sha256" -Encoding ascii
    Write-Host "Built $Output"
} finally {
    Pop-Location
}
