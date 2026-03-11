!include "MUI2.nsh"

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
  
  File "build\bin\ollama英特尔.exe"
  
  CreateDirectory "$INSTDIR\ollama-bin"
  SetOutPath "$INSTDIR\ollama-bin"
  
  File "ollama-bin\ollama.exe"
  File "ollama-bin\ollama-lib.exe"
  File "ollama-bin\dnnl.dll"
  File "ollama-bin\ggml-base.dll"
  File "ollama-bin\ggml-cpu.dll"
  File "ollama-bin\ggml-sycl.dll"
  File "ollama-bin\ggml.dll"
  File "ollama-bin\libc++.dll"
  File "ollama-bin\libhwloc-15.dll"
  File "ollama-bin\libiomp5md.dll"
  File "ollama-bin\libmmd.dll"
  File "ollama-bin\llama.dll"
  File "ollama-bin\llava_shared.dll"
  File "ollama-bin\mkl_core.2.dll"
  File "ollama-bin\mkl_sycl_blas.5.dll"
  File "ollama-bin\mkl_tbb_thread.2.dll"
  File "ollama-bin\mtmd_shared.dll"
  File "ollama-bin\svml_dispmd.dll"
  File "ollama-bin\sycl8.dll"
  File "ollama-bin\tbb12.dll"
  File "ollama-bin\tbbbind.dll"
  File "ollama-bin\tbbbind_2_0.dll"
  File "ollama-bin\tbbbind_2_5.dll"
  File "ollama-bin\tbbmalloc.dll"
  File "ollama-bin\tbbmalloc_proxy.dll"
  File "ollama-bin\tcm.dll"
  File "ollama-bin\umf.dll"
  File "ollama-bin\ur_adapter_level_zero.dll"
  File "ollama-bin\ur_adapter_opencl.dll"
  File "ollama-bin\ur_loader.dll"
  File "ollama-bin\ur_win_proxy_loader.dll"
  
  WriteRegStr HKLM "Software\ollama-intel" "Install_Dir" "$INSTDIR"
  
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "DisplayName" "ollama intel"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel" "NoRepair" 1
  
  WriteUninstaller "$INSTDIR\uninstall.exe"
  
  CreateDirectory "$SMPROGRAMS\ollama intel"
  CreateShortcut "$SMPROGRAMS\ollama intel\ollama intel.lnk" "$INSTDIR\ollama英特尔.exe" "" "$INSTDIR\ollama英特尔.exe" 0
  CreateShortcut "$SMPROGRAMS\ollama intel\Uninstall.lnk" "$INSTDIR\uninstall.exe" "" "$INSTDIR\uninstall.exe" 0
  
  CreateShortcut "$DESKTOP\ollama intel.lnk" "$INSTDIR\ollama英特尔.exe" "" "$INSTDIR\ollama英特尔.exe" 0
  
SectionEnd

Section "Uninstall"
  
  Delete "$INSTDIR\ollama英特尔.exe"
  
  Delete "$INSTDIR\ollama-bin\ollama.exe"
  Delete "$INSTDIR\ollama-bin\ollama-lib.exe"
  Delete "$INSTDIR\ollama-bin\dnnl.dll"
  Delete "$INSTDIR\ollama-bin\ggml-base.dll"
  Delete "$INSTDIR\ollama-bin\ggml-cpu.dll"
  Delete "$INSTDIR\ollama-bin\ggml-sycl.dll"
  Delete "$INSTDIR\ollama-bin\ggml.dll"
  Delete "$INSTDIR\ollama-bin\libc++.dll"
  Delete "$INSTDIR\ollama-bin\libhwloc-15.dll"
  Delete "$INSTDIR\ollama-bin\libiomp5md.dll"
  Delete "$INSTDIR\ollama-bin\libmmd.dll"
  Delete "$INSTDIR\ollama-bin\llama.dll"
  Delete "$INSTDIR\ollama-bin\llava_shared.dll"
  Delete "$INSTDIR\ollama-bin\mkl_core.2.dll"
  Delete "$INSTDIR\ollama-bin\mkl_sycl_blas.5.dll"
  Delete "$INSTDIR\ollama-bin\mkl_tbb_thread.2.dll"
  Delete "$INSTDIR\ollama-bin\mtmd_shared.dll"
  Delete "$INSTDIR\ollama-bin\svml_dispmd.dll"
  Delete "$INSTDIR\ollama-bin\sycl8.dll"
  Delete "$INSTDIR\ollama-bin\tbb12.dll"
  Delete "$INSTDIR\ollama-bin\tbbbind.dll"
  Delete "$INSTDIR\ollama-bin\tbbbind_2_0.dll"
  Delete "$INSTDIR\ollama-bin\tbbbind_2_5.dll"
  Delete "$INSTDIR\ollama-bin\tbbmalloc.dll"
  Delete "$INSTDIR\ollama-bin\tbbmalloc_proxy.dll"
  Delete "$INSTDIR\ollama-bin\tcm.dll"
  Delete "$INSTDIR\ollama-bin\umf.dll"
  Delete "$INSTDIR\ollama-bin\ur_adapter_level_zero.dll"
  Delete "$INSTDIR\ollama-bin\ur_adapter_opencl.dll"
  Delete "$INSTDIR\ollama-bin\ur_loader.dll"
  Delete "$INSTDIR\ollama-bin\ur_win_proxy_loader.dll"
  
  RMDir "$INSTDIR\ollama-bin"
  
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"
  
  Delete "$SMPROGRAMS\ollama intel\*.*"
  RMDir "$SMPROGRAMS\ollama intel"
  
  Delete "$DESKTOP\ollama intel.lnk"
  
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\ollama-intel"
  DeleteRegKey HKLM "Software\ollama-intel"
  
SectionEnd