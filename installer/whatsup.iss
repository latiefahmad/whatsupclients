; Inno Setup script for the Windows installer (see docs/releasing.md).
;
;   iscc /DAppVersion=0.10.0 /DExe=..\bin\dist\WhatsUpClients-windows-amd64.exe installer\whatsup.iss
;
; It installs for the current user only, into %LocalAppData%\Programs,
; so neither installing nor the app's own updates (internal/update, which
; replaces WhatsUpClients.exe in place) ask for admin rights.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef Exe
  #define Exe "..\bin\dist\WhatsUpClients-windows-amd64.exe"
#endif
#ifndef OutDir
  #define OutDir "..\bin\dist"
#endif

[Setup]
; Never change AppId: it's how a new installer finds the installed app.
AppId={{6B0F3B52-9C1E-4D7A-A3E5-2F7C8D41B9E6}
AppName=WhatsUp Clients
AppVersion={#AppVersion}
AppPublisher=latiefahmad
AppPublisherURL=https://github.com/latiefahmad/whatsupclients
AppSupportURL=https://github.com/latiefahmad/whatsupclients/issues
AppUpdatesURL=https://github.com/latiefahmad/whatsupclients/releases
DefaultDirName={autopf}\WhatsUpClients
DefaultGroupName=WhatsUpClients
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
UninstallDisplayIcon={app}\WhatsUpClients.exe
UninstallDisplayName=WhatsUp Clients
OutputDir={#OutDir}
OutputBaseFilename=WhatsUpClients-Setup
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
SetupIconFile=whatsup.ico
; The app may be running in the notification area, without a window.
CloseApplications=force
RestartApplications=no

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "{#Exe}"; DestDir: "{app}"; DestName: "WhatsUpClients.exe"; Flags: ignoreversion

[Icons]
; The app's notifications use this AppUserModelID (ui.appID).
Name: "{autoprograms}\WhatsUp Clients"; Filename: "{app}\WhatsUpClients.exe"; AppUserModelID: "WhatsUpClients.Desktop"
Name: "{autodesktop}\WhatsUp Clients"; Filename: "{app}\WhatsUpClients.exe"; AppUserModelID: "WhatsUpClients.Desktop"; Tasks: desktopicon

[Run]
Filename: "{app}\WhatsUpClients.exe"; Description: "{cm:LaunchProgram,WhatsUp Clients}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
; Quit the app first, even while it's only in the notification area.
Filename: "{sys}\taskkill.exe"; Parameters: "/f /im WhatsUpClients.exe"; Flags: runhidden; RunOnceId: "QuitApp"

[UninstallDelete]
; Left by an update (internal/update).
Type: files; Name: "{app}\WhatsUpClients.exe.old"
Type: files; Name: "{app}\WhatsUpClients.exe.new"

[Registry]
; What the app writes for itself: notifications (internal/notify) and
; starting at login (internal/desktop). Removed on uninstall only.
Root: HKCU; Subkey: "Software\Classes\AppUserModelId\WhatsUpClients.Desktop"; ValueType: none; Flags: uninsdeletekey dontcreatekey
Root: HKCU; Subkey: "Software\Classes\CLSID\{{EC56F81C-5D4A-4D8C-8044-761DDA20EDC8}"; ValueType: none; Flags: uninsdeletekey dontcreatekey
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueName: "WhatsUpClients"; ValueType: none; Flags: uninsdeletevalue dontcreatekey

[Code]
// Chats and the linked account stay unless the user says otherwise, so
// installing again doesn't mean linking again.
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  Data: String;
begin
  if CurUninstallStep <> usPostUninstall then
    Exit;
  Data := ExpandConstant('{userappdata}\WhatsUpClients');
  if not DirExists(Data) then
    Data := ExpandConstant('{userappdata}\WazzapClients');
  if not DirExists(Data) or UninstallSilent then
    Exit;
  if MsgBox('Also delete your chats and log WhatsUp Clients out of WhatsApp on this computer?' + #13#10#13#10 +
            'Keep them to pick up where you left off if you install it again.',
            mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES then
    DelTree(Data, True, True, True);
end;
