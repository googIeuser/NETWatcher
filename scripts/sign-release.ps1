param(
    [Parameter(Mandatory=$true)][string]$File,
    [Parameter(Mandatory=$true)][string]$CertificatePath,
    [Parameter(Mandatory=$true)][string]$CertificatePassword
)
$ErrorActionPreference = "Stop"
$SignTool = Get-ChildItem "${env:ProgramFiles(x86)}\Windows Kits\10\bin" -Filter signtool.exe -Recurse |
    Sort-Object FullName -Descending |
    Select-Object -First 1
if (-not $SignTool) { throw "signtool.exe was not found. Install the Windows SDK." }
& $SignTool.FullName sign /fd SHA256 /td SHA256 /tr http://timestamp.digicert.com /f $CertificatePath /p $CertificatePassword $File
if ($LASTEXITCODE -ne 0) { throw "Authenticode signing failed." }
& $SignTool.FullName verify /pa /v $File
