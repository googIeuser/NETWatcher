param(
  [switch]$Installer
)
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$project = Join-Path $root "next"
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw "Go is not installed or not in PATH." }
if (-not (Get-Command npm -ErrorAction SilentlyContinue)) { throw "Node.js/npm is not installed or not in PATH." }
if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
  $env:Path += ";$(go env GOPATH)\bin"
}
Push-Location $project
try {
  go mod tidy
  Push-Location frontend
  try { npm ci; npm run build } finally { Pop-Location }
  go test ./internal/...
  if ($Installer) { wails build -nsis } else { wails build }
  Write-Host "Build output: $project\build\bin" -ForegroundColor Green
} finally { Pop-Location }
