@echo off
setlocal
cd /d "%~dp0"
powershell.exe -NoProfile -ExecutionPolicy Bypass -Command ^
  "$ErrorActionPreference='Stop'; Remove-Item '.\flutter_app\build\windows' -Recurse -Force -ErrorAction SilentlyContinue; .\scripts\prepare-flutter-windows.ps1; .\scripts\run-rust-flutter.ps1"
if errorlevel 1 (
  echo.
  echo NetWatcher icon rebuild failed.
  pause
  exit /b 1
)
endlocal
