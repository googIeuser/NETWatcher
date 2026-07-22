; NetWatcher 4 stable installer
#ifndef MyAppVersion
  #define MyAppVersion "4.0.2"
#endif
#ifndef SourceDir
  #define SourceDir "..\flutter_app\build\windows\x64\runner\Release"
#endif
#ifndef OutputDir
  #define OutputDir "..\dist"
#endif

#define MyAppName "NetWatcher"
#define MyAppPublisher "NetWatcher Contributors"
#define MyAppExeName "netwatcher.exe"

[Setup]
AppId={{E95B6876-8D42-4F38-90AD-2E5EC83A8C16}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\NetWatcher
DefaultGroupName=NetWatcher
DisableProgramGroupPage=yes
OutputDir={#OutputDir}
OutputBaseFilename=NetWatcher_Setup_{#MyAppVersion}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
SetupIconFile={#SourceDir}\data\flutter_assets\assets\app_icon.ico
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
UninstallDisplayIcon={app}\{#MyAppExeName}
VersionInfoVersion={#MyAppVersion}.0
VersionInfoProductName={#MyAppName}
VersionInfoProductVersion={#MyAppVersion}
VersionInfoCompany={#MyAppPublisher}
VersionInfoDescription=Local connection monitoring and diagnostics utility
CloseApplications=yes
RestartApplications=no

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "turkish"; MessagesFile: "compiler:Languages\Turkish.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\NetWatcher"; Filename: "{app}\{#MyAppExeName}"
Name: "{autodesktop}\NetWatcher"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,NetWatcher}"; Flags: nowait postinstall skipifsilent
