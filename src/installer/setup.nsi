Unicode true
!include "MUI2.nsh"
Name "Yangi alifbo konvertori"
OutFile "Yangi_alifbo_konvertori_setup.exe"
InstallDir "$LOCALAPPDATA\Programs\Yangi alifbo konvertori"
RequestExecutionLevel user
SetCompressor /SOLID lzma
VIProductVersion "1.7.0.0"
VIAddVersionKey "ProductName" "Yangi alifbo konvertori"
VIAddVersionKey "FileDescription" "Yangi alifbo konvertori o'rnatuvchisi"
VIAddVersionKey "FileVersion" "1.7.0"
VIAddVersionKey "ProductVersion" "1.7.0"
VIAddVersionKey "LegalCopyright" "(c) 2026 Abduxoliqov Otabek"
VIAddVersionKey "CompanyName" "Abduxoliqov Otabek"
BrandingText "Abduxoliqov Otabek · 2026 · (33) 071-60-70"
!define MUI_ICON "icon.ico"
!define MUI_UNICON "icon.ico"
!define MUI_FINISHPAGE_RUN "$INSTDIR\Yangi_alifbo_konvertori.exe"
!define UNKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\YangiAlifboKonvertori"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "Uzbek"
!insertmacro MUI_LANGUAGE "Russian"
!insertmacro MUI_LANGUAGE "English"

Function .onInstSuccess
  IfSilent 0 +2
  Exec '"$INSTDIR\Yangi_alifbo_konvertori.exe"'
FunctionEnd

Section "Install"
  nsExec::Exec 'taskkill /F /IM Yangi_alifbo_konvertori.exe'
  Sleep 500
  SetOutPath "$INSTDIR"
  File "Yangi_alifbo_konvertori.exe"
  File "icon.ico"
  File "LICENSE-uz_UZ-2026-GPL3.txt"
  File "LICENSE-freq-GPL3.txt"
  File "LICENSE-IBM-Plex-OFL.txt"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateShortCut "$DESKTOP\Yangi alifbo konvertori.lnk" "$INSTDIR\Yangi_alifbo_konvertori.exe" "" "$INSTDIR\icon.ico"
  CreateDirectory "$SMPROGRAMS\Yangi alifbo konvertori"
  CreateShortCut "$SMPROGRAMS\Yangi alifbo konvertori\Yangi alifbo konvertori.lnk" "$INSTDIR\Yangi_alifbo_konvertori.exe" "" "$INSTDIR\icon.ico"
  CreateShortCut "$SMPROGRAMS\Yangi alifbo konvertori\O'chirish.lnk" "$INSTDIR\Uninstall.exe"
  WriteRegStr HKCU "${UNKEY}" "DisplayName" "Yangi alifbo konvertori"
  WriteRegStr HKCU "${UNKEY}" "DisplayIcon" "$INSTDIR\icon.ico"
  WriteRegStr HKCU "${UNKEY}" "DisplayVersion" "1.7.0"
  WriteRegStr HKCU "${UNKEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${UNKEY}" "Publisher" "Abduxoliqov Otabek"
  WriteRegStr HKCU "${UNKEY}" "HelpTelephone" "(33) 071-60-70"
  WriteRegStr HKCU "${UNKEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
  WriteRegDWORD HKCU "${UNKEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNKEY}" "NoRepair" 1
SectionEnd

Section "Uninstall"
  nsExec::Exec 'taskkill /F /IM Yangi_alifbo_konvertori.exe'
  Sleep 500
  Delete "$DESKTOP\Yangi alifbo konvertori.lnk"
  RMDir /r "$SMPROGRAMS\Yangi alifbo konvertori"
  Delete "$INSTDIR\Yangi_alifbo_konvertori.exe"
  Delete "$INSTDIR\icon.ico"
  Delete "$INSTDIR\LICENSE-uz_UZ-2026-GPL3.txt"
  Delete "$INSTDIR\LICENSE-freq-GPL3.txt"
  Delete "$INSTDIR\LICENSE-IBM-Plex-OFL.txt"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
  RMDir /r "$LOCALAPPDATA\YangiAlifbo"
  DeleteRegKey HKCU "${UNKEY}"
SectionEnd
