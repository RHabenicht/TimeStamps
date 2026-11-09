# Test the actually published v2.5 setup on a disposable Windows GitHub runner.
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Runtime.InteropServices;
public static class TimeStampsWin32 {
    public delegate bool EnumWindowHandler(IntPtr hwnd, IntPtr state);
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowHandler handler, IntPtr state);
    [DllImport("user32.dll")] public static extern bool EnumChildWindows(IntPtr parent, EnumWindowHandler handler, IntPtr state);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] static extern int GetClassName(IntPtr h, StringBuilder s, int capacity);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] static extern int GetWindowText(IntPtr h, StringBuilder s, int capacity);
    [DllImport("user32.dll")] static extern uint GetWindowThreadProcessId(IntPtr h, out uint processId);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr h);
    [DllImport("user32.dll")] public static extern IntPtr GetDlgItem(IntPtr parent, int id);
    [DllImport("user32.dll")] public static extern int GetDlgCtrlID(IntPtr h);
    [DllImport("user32.dll")] public static extern bool PostMessage(IntPtr h, uint message, IntPtr w, IntPtr l);
    [DllImport("user32.dll", CharSet=CharSet.Unicode)] static extern IntPtr SendMessage(IntPtr h, uint message, IntPtr w, StringBuilder text);
    [DllImport("user32.dll", SetLastError=true)]
    static extern IntPtr SendMessageTimeout(IntPtr h, uint message, IntPtr w, IntPtr l, uint flags, uint timeoutMs, out IntPtr result);
    public static string Class(IntPtr h) {
        var b=new StringBuilder(128); GetClassName(h,b,b.Capacity); return b.ToString();
    }
    public static string Title(IntPtr h) {
        var b=new StringBuilder(512); GetWindowText(h,b,b.Capacity); return b.ToString();
    }
    public static IntPtr FindTop(string windowClass, string titlePrefix, int pid) {
        IntPtr found=IntPtr.Zero;
        EnumWindows((h,_) => {
            if (pid > 0) {uint p; GetWindowThreadProcessId(h,out p); if (p!=(uint)pid) return true;}
            if (windowClass!="" && Class(h)!=windowClass) return true;
            if (titlePrefix!="" && !Title(h).StartsWith(titlePrefix,StringComparison.Ordinal)) return true;
            found=h; return false;
        },IntPtr.Zero);
        return found;
    }
    public static IntPtr FindOutputEdit(IntPtr parent) {
        IntPtr found=IntPtr.Zero;
        EnumChildWindows(parent,(h,_) => {
            if(Class(h).Equals("Edit",StringComparison.OrdinalIgnoreCase) && GetDlgCtrlID(h)==0) {found=h;return false;}
            return true;
        },IntPtr.Zero);
        return found;
    }
    public static string Text(IntPtr h) {
        if(h==IntPtr.Zero) return "";
        var b = new StringBuilder(131072);
        SendMessage(h,0x000D,(IntPtr)b.Capacity,b); // WM_GETTEXT, Windows-managed marshalling
        return b.ToString();
    }
    public static string DialogBody(IntPtr parent) {
        var b=new StringBuilder();
        EnumChildWindows(parent,(h,_) => {
            if(Class(h).Equals("Static",StringComparison.OrdinalIgnoreCase)) b.Append(Text(h));
            return true;
        },IntPtr.Zero);
        return b.ToString();
    }
    public static bool SendSafe(IntPtr h, uint msg, long w, long l, uint timeoutMs) {
        IntPtr result;
        return SendMessageTimeout(h,msg,(IntPtr)w,(IntPtr)l,2,timeoutMs,out result)!=IntPtr.Zero;
    }
}
'@
function Note([string]$s) { Write-Host "PASS: $s" }
function Assert([bool]$ok, [string]$s) { if (-not $ok) { throw "FAIL: $s" } }
function Wait-Top([string]$cl, [string]$title, [int]$targetProcessId, [int]$timeoutSeconds=20) {
  $deadline=[DateTime]::UtcNow.AddSeconds($timeoutSeconds)
  do {
    $h=[TimeStampsWin32]::FindTop($cl,$title,$targetProcessId)
    if($h -ne [IntPtr]::Zero) {return $h}
    Start-Sleep -Milliseconds 150
  } while ([DateTime]::UtcNow -lt $deadline)
  throw "Timed out waiting for window class='$cl', title='$title', pid=$targetProcessId"
}
function Choice([IntPtr]$h, [int]$id) {
  Assert ([TimeStampsWin32]::PostMessage($h,0x0111,[IntPtr]$id,[IntPtr]::Zero)) "could not post selection $id to dialog"
  Start-Sleep -Milliseconds 500
}
function Run-Installer([string]$setupExe, [bool]$makeDesktop) {
  $p=Start-Process -FilePath $setupExe -PassThru
  try {
    $dlg=Wait-Top '#32770' 'TimeStamps 2.5 Setup' $p.Id 30
    Assert ([TimeStampsWin32]::DialogBody($dlg) -match 'Windows 10/11') 'initial installer prompt unexpected'
    Choice $dlg 6
    $dlg=Wait-Top '#32770' 'TimeStamps 2.5 Setup' $p.Id 45
    Assert ([TimeStampsWin32]::DialogBody($dlg) -match 'desktop shortcut') 'desktop shortcut prompt unexpected'
    Choice $dlg $(if($makeDesktop){6}else{7})
    $dlg=Wait-Top '#32770' 'TimeStamps 2.5 Setup' $p.Id 60
    $message=[TimeStampsWin32]::DialogBody($dlg)
    Assert ($message -match 'installed successfully') "setup reported a failure or unexpected result: $message"
    Choice $dlg 7 # Do not auto-start; smoke test launches the app directly.
    Assert ($p.WaitForExit(15000)) 'installer did not exit'
    Assert ($p.ExitCode -eq 0) "installer exit code $($p.ExitCode)"
  } finally {
    if(-not $p.HasExited){Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue}
  }
}
function Send-Command([IntPtr]$h, [int]$id) {
  Assert ([TimeStampsWin32]::SendSafe($h,0x0111,$id,0,2500)) "WM_COMMAND $id hung or failed"
}
function Wait-Visible([IntPtr]$h, [bool]$expected, [string]$label) {
  $deadline=[DateTime]::UtcNow.AddSeconds(5)
  do {
    if([TimeStampsWin32]::IsWindowVisible($h) -eq $expected) {return}
    Start-Sleep -Milliseconds 100
  } while ([DateTime]::UtcNow -lt $deadline)
  throw "FAIL: $label visibility expected=$expected"
}
$release='https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/TimeStamps_2.5_Setup_Updated_by_NightLife.exe'
$setupPath=Join-Path $env:RUNNER_TEMP 'TimeStamps_2.5_Setup_Updated_by_NightLife.exe'
$dest=Join-Path $env:LOCALAPPDATA 'Programs\TimeStamps'
$registry='HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\NightLife-TimeStamps'
$programs=[Environment]::GetFolderPath('Programs')
$startShortcut=Join-Path $programs 'NightLife\TimeStamps.lnk'
$desktopShortcut=Join-Path ([Environment]::GetFolderPath('Desktop')) 'TimeStamps.lnk'
$appProcess=$null
Write-Host 'Testing locally built 2.5 installer candidate before publication...'
Copy-Item -LiteralPath (Join-Path $env:GITHUB_WORKSPACE 'dist/TimeStamps_2.5_Setup_Updated_by_NightLife.exe') -Destination $setupPath
$hash=(Get-FileHash $setupPath -Algorithm SHA256).Hash
$bytes=(Get-Item $setupPath).Length
Assert ($bytes -gt 1000000) 'release download size unexpected'
Write-Host "Released installer SHA256: $hash, bytes: $bytes"
Write-Host "Windows: $([Environment]::OSVersion.VersionString)"
Write-Host "Session: $([System.Diagnostics.Process]::GetCurrentProcess().SessionId)"
try {
  Assert (-not (Test-Path -LiteralPath (Join-Path $dest 'TimeStamps.exe'))) 'runner already has TimeStamps installed'
  Run-Installer $setupPath $false
  foreach($file in @('TimeStamps.exe','TimeStamps.exe.manifest','TimeStamps.ico','splash_screen.png','Uninstall-TimeStamps.exe')) {
    Assert (Test-Path -LiteralPath (Join-Path $dest $file)) "installed file missing: $file"
  }
  Note 'released setup installed executable, icon, splash, manifest and uninstaller'
  Assert (Test-Path -LiteralPath $startShortcut) 'Start Menu shortcut missing'
  Assert (Test-Path -LiteralPath $registry) 'Installed Apps registry entry missing'
  $uninstallReg=Get-ItemProperty -LiteralPath $registry
  Assert ($uninstallReg.DisplayVersion -eq '2.5') 'Installed Apps version mismatch'
  Note 'Start Menu shortcut and version 2.5 Installed Apps registration created'
  $legacyINI=Join-Path $dest 'TimeStamps.ini'
  $userFile=Join-Path $dest 'user_notes_do_not_delete.txt'
  [IO.File]::WriteAllText($legacyINI, '[TimeStamps]'+[Environment]::NewLine+'CustomMarker=retain-on-upgrade'+[Environment]::NewLine)
  [IO.File]::WriteAllText($userFile,'user-created data retained')
  Run-Installer $setupPath $true
  Assert ((Get-Content -LiteralPath $legacyINI -Raw).Contains('retain-on-upgrade')) 'legacy INI lost during upgrade'
  Assert ((Get-Content -LiteralPath $userFile -Raw) -eq 'user-created data retained') 'unrelated file changed on upgrade'
  Assert (Test-Path -LiteralPath $desktopShortcut) 'requested desktop shortcut missing'
  Note 'real setup upgrade preserves legacy INI and unrelated files; creates desktop shortcut'
  $appProcess=Start-Process -FilePath (Join-Path $dest 'TimeStamps.exe') -PassThru -WorkingDirectory $dest
  $splash=Wait-Top 'NightLifeTimeStampsSplashV25' '' $appProcess.Id 20
  Assert ([TimeStampsWin32]::IsWindowVisible($splash)) 'splash exists but is not visible'
  $splashSeen=[DateTime]::UtcNow
  $main=Wait-Top 'NightLifeTimeStampsV2' '' $appProcess.Id 25
  Wait-Visible $main $true 'main window'
  $delay=([DateTime]::UtcNow-$splashSeen).TotalSeconds
  Assert ($delay -ge 2.5 -and $delay -le 14) "splash-to-main timing outside expected range: $delay"
  Assert ([TimeStampsWin32]::IsWindowVisible($main)) 'main window invisible'
  Note ("real application shows splash before main window (~{0:N1}s observed)" -f $delay)
  $preset=[TimeStampsWin32]::GetDlgItem($main,108)
  $qty=[TimeStampsWin32]::GetDlgItem($main,102)
  $copy=[TimeStampsWin32]::GetDlgItem($main,116)
  Assert ($preset -ne [IntPtr]::Zero -and $qty -ne [IntPtr]::Zero -and $copy -ne [IntPtr]::Zero) 'expected Win32 controls missing'
  for($i=0;$i -lt 15;$i++){
    Send-Command $main 202 # Format
    Wait-Visible $preset $true 'Format preset'
    Wait-Visible $qty $false 'Setup quantity hidden on Format'
    Assert ([TimeStampsWin32]::SendSafe($preset,0x014E,1,0,1500)) 'preset selection failed'
    $notify=108 -bor (1 -shl 16) # CBN_SELCHANGE
    Assert ([TimeStampsWin32]::SendSafe($main,0x0111,$notify,$preset.ToInt64(),1500)) 'preset change hung'
    Start-Sleep -Milliseconds 60
    Send-Command $main 201 # Setup
    Wait-Visible $qty $true 'Setup quantity'
    Wait-Visible $preset $false 'Format preset hidden on Setup'
  }
  Note '15 live Format/Setup transitions and preset changes responded without a hung Win32 message'
  Send-Command $main 114 # Generate
  $output=[TimeStampsWin32]::FindOutputEdit($main)
  Assert ($output -ne [IntPtr]::Zero) 'output EDIT control not found'
  $generated=[TimeStampsWin32]::Text($output)
  Assert ($generated.Trim().Length -gt 0) 'Generate did not produce timestamp output'
  Wait-Visible $copy $true 'Results copy action'
  Note ("Generate produced {0} characters of timestamp text" -f $generated.Length)
  Assert ([TimeStampsWin32]::PostMessage($main,0x0010,[IntPtr]::Zero,[IntPtr]::Zero)) 'could not close application'
  $confirm=Wait-Top '#32770' 'TimeStamps — unsaved output' $appProcess.Id 10
  Choice $confirm 6
  Assert ($appProcess.WaitForExit(12000)) 'application did not exit on confirmation'
  Note 'application exits after unsaved output confirmation'
  $appProcess=$null
  $uninstall=Start-Process -FilePath (Join-Path $dest 'Uninstall-TimeStamps.exe') -PassThru
  try {
    $dlg=Wait-Top '#32770' 'TimeStamps 2.5 Uninstall' $uninstall.Id 20
    Assert ([TimeStampsWin32]::DialogBody($dlg) -match 'Remove TimeStamps') 'unexpected uninstall prompt'
    Choice $dlg 6
    Assert ($uninstall.WaitForExit(10000)) 'uninstaller parent did not exit'
    $workerDialog=Wait-Top '#32770' 'TimeStamps 2.5 Uninstall' 0 30
    $result=[TimeStampsWin32]::DialogBody($workerDialog)
    Assert ($result -match 'was removed') "uninstaller failed: $result"
    Choice $workerDialog 1
  } finally {
    if(-not $uninstall.HasExited){Stop-Process -Id $uninstall.Id -Force -ErrorAction SilentlyContinue}
  }
  Start-Sleep -Seconds 2
  Assert (-not (Test-Path -LiteralPath (Join-Path $dest 'TimeStamps.exe'))) 'uninstaller left executable'
  Assert (-not (Test-Path -LiteralPath $registry)) 'uninstaller left Installed Apps registration'
  Assert (-not (Test-Path -LiteralPath $startShortcut)) 'uninstaller left Start Menu shortcut'
  Assert (-not (Test-Path -LiteralPath $desktopShortcut)) 'uninstaller left desktop shortcut'
  Assert ((Get-Content -LiteralPath $legacyINI -Raw).Contains('retain-on-upgrade')) 'uninstaller deleted legacy INI'
  Assert ((Get-Content -LiteralPath $userFile -Raw) -eq 'user-created data retained') 'uninstaller deleted user file'
  Note 'real uninstaller removed executable and integration while preserving user files'
  Write-Host 'RUNTIME SMOKE RESULT: PASS (2.5 candidate installer; Windows runner)'
} finally {
  if($appProcess -and -not $appProcess.HasExited){Stop-Process -Id $appProcess.Id -Force -ErrorAction SilentlyContinue}
}
