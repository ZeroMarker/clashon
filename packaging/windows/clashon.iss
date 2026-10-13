; Inno Setup script. Per-user install, no admin rights needed.
#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif

[Setup]
AppId={{6F1C2A5E-3B7D-4C8A-9E21-0D4B5A7C9E13}
AppName=ClashOn
AppVersion={#AppVersion}
AppPublisher=ClashOn
AppPublisherURL=https://github.com/ZeroMarker/clashon
DefaultDirName={localappdata}\Programs\ClashOn
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputBaseFilename=ClashOn-setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
UninstallDisplayIcon={app}\clashon-gui.exe
CloseApplications=yes
LicenseFile=..\..\LICENSE

[Languages]
Name: "en"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; Flags: unchecked

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\ClashOn"; Filename: "{app}\clashon-gui.exe"
Name: "{autodesktop}\ClashOn"; Filename: "{app}\clashon-gui.exe"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Classes\clash"; ValueType: string; ValueData: "URL:clash protocol"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\clash"; ValueType: string; ValueName: "URL Protocol"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\clash\shell\open\command"; ValueType: string; ValueData: """{app}\clashon-gui.exe"" ""%1"""

[Run]
Filename: "{app}\clashon-gui.exe"; Description: "{cm:LaunchProgram,ClashOn}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
; Stop the background core and restore the system proxy before files go away.
Filename: "{app}\clashon.exe"; Parameters: "sysproxy off"; Flags: runhidden; RunOnceId: "sysproxy"
Filename: "{app}\clashon.exe"; Parameters: "down"; Flags: runhidden; RunOnceId: "down"
Filename: "{app}\clashon.exe"; Parameters: "desktop autostart off"; Flags: runhidden; RunOnceId: "autostart"
