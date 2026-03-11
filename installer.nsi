!include "MUI2.nsh"
!include "nsDialogs.nsh"

Name "ollama intel"
OutFile "build\bin\ollama-intel-installer.exe"
InstallDir "$PROGRAMFILES64\ollama-intel"
InstallDirRegKey HKLM "Software\ollama-intel" "Install_Dir"
RequestExecutionLevel admin

!define MUI_ABORTWARNING
!define MUI_ICON "build\windows\icon.ico"
!define MUI_UNICON "build\windows\icon.ico"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "ollama intel" SecInstall
  SectionIn RO
  
  SetOutPath "$INSTDIR"
  
  ; Copy main executable
  File "build\bin\ollama-intel.exe"
  
  ; Copy ollama-bin.zip
  File "build\ollama-bin.zip"
  
  ; Extract ollama-bin.zip using PowerShell
  DetailPrint "Extracting ollama-bin.zip..."
  nsExec::ExecToLog 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Expand-Archive -Path \"$INSTDIR\ollama-bin.zip\" -DestinationPath \"$INSTDIR\ollama-bin\" -Force"'
  
  ; Delete the zip file after extraction
  Delete "$INSTDIR\ollama-bin.zip"
  
  ; Write registry keys
  WriteRegStr HKLM "Software\ollama-intel" "Install_Dir" "$INSTDIR"
  
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "DisplayName" "ollama intel"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "NoRepair" 1
  
  WriteUninstaller "$INSTDIR\uninstall.exe"
  
  ; Create start menu shortcuts
  CreateDirectory "$SMPROGRAMS\ollama intel"
  CreateShortcut "$SMPROGRAMS\ollama intel\ollama intel.lnk" "$INSTDIR\ollama-intel.exe" "" "$INSTDIR\ollama-intel.exe" 0
  CreateShortcut "$SMPROGRAMS\ollama intel\Uninstall.lnk" "$INSTDIR\uninstall.exe" "" "$INSTDIR\uninstall.exe" 0
  
  ; Create desktop shortcut
  CreateShortcut "$DESKTOP\ollama intel.lnk" "$INSTDIR\ollama-intel.exe" "" "$INSTDIR\ollama-intel.exe" 0
  
SectionEnd

Section "Uninstall"
  
  ; Delete main executable
  Delete "$INSTDIR\ollama-intel.exe"
  
  ; Delete ollama-bin directory
  RMDir /r "$INSTDIR\ollama-bin"
  
  ; Delete uninstaller
  Delete "$INSTDIR\uninstall.exe"
  
  ; Remove installation directory
  RMDir "$INSTDIR"
  
  ; Delete start menu shortcuts
  Delete "$SMPROGRAMS\ollama intel\*.*"
  RMDir "$SMPROGRAMS\ollama intel"
  
  ; Delete desktop shortcut
  Delete "$DESKTOP\ollama intel.lnk"
  
  ; Delete registry keys
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel"
  DeleteRegKey HKLM "Software\ollama-intel"
  
SectionEnd