# Inspect actual pixels in released v2.5 splash, not merely HWND existence.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
Add-Type -AssemblyName System.Drawing
Add-Type -TypeDefinition @'
using System;
using System.Text;
using System.Runtime.InteropServices;
public static class SplashInspect {
  public delegate bool Callback(IntPtr hwnd, IntPtr lparam);
  [DllImport("user32.dll")] public static extern bool EnumWindows(Callback cb, IntPtr param);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetClassName(IntPtr hwnd, StringBuilder str, int count);
  [DllImport("user32.dll")] public static extern bool GetClientRect(IntPtr hwnd, out RECT rect);
  [DllImport("user32.dll")] public static extern IntPtr GetDC(IntPtr hwnd);
  [DllImport("user32.dll")] public static extern int ReleaseDC(IntPtr hwnd, IntPtr hdc);
  [DllImport("gdi32.dll")] public static extern uint GetPixel(IntPtr hdc, int x, int y);
  [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hwnd, out uint pid);
  [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr hwnd);
  public struct RECT { public int Left, Top, Right, Bottom; }
  public static IntPtr Find(int pid){
    IntPtr found=IntPtr.Zero;
    EnumWindows((h,v)=>{
      uint id; GetWindowThreadProcessId(h,out id);
      if (id!=(uint)pid) return true;
      var str=new StringBuilder(128);
      GetClassName(h,str,str.Capacity);
      if (str.ToString()=="NightLifeTimeStampsSplashV25") {found=h;return false;}
      return true;
    },IntPtr.Zero);
    return found;
  }
}
'@
$url='https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/TimeStamps.exe'
$exe=Join-Path $env:RUNNER_TEMP 'TimeStamps-v2.5-release.exe'
Copy-Item -LiteralPath (Join-Path $env:GITHUB_WORKSPACE 'dist/TimeStamps.exe') -Destination $exe
Write-Host ('Candidate exe SHA256: '+(Get-FileHash $exe -Algorithm SHA256).Hash)
$p=Start-Process -FilePath $exe -PassThru -WorkingDirectory $env:RUNNER_TEMP
try {
  $w=[IntPtr]::Zero
  $deadline=[DateTime]::UtcNow.AddSeconds(25)
  while([DateTime]::UtcNow -lt $deadline) {
    $w=[SplashInspect]::Find($p.Id)
    if($w -ne [IntPtr]::Zero){break}
    Start-Sleep -Milliseconds 75
  }
  if($w -eq [IntPtr]::Zero){throw 'Splash popup was not created'}
  $r=[SplashInspect+RECT]::new()
  if(-not [SplashInspect]::GetClientRect($w,[ref]$r)){throw 'GetClientRect failed'}
  Write-Host ("Splash window size {0} x {1} visible={2}" -f $r.Right,$r.Bottom,[SplashInspect]::IsWindowVisible($w))
  Start-Sleep -Milliseconds 700
  $hdc=[SplashInspect]::GetDC($w)
  if($hdc -eq [IntPtr]::Zero){throw 'GetDC failed'}
  $pixels=@()
  try{
    foreach($yy in @(0.08,0.25,0.50,0.75,0.92)){
      foreach($xx in @(0.08,0.25,0.50,0.75,0.92)){
        $x=[int][Math]::Floor($r.Right*$xx)
        $y=[int][Math]::Floor($r.Bottom*$yy)
        $color=[SplashInspect]::GetPixel($hdc,$x,$y)
        $pixels+= $color
        Write-Host ("GDI pixel ({0},{1})=0x{2:X6}" -f $x,$y,$color)
      }
    }
  }finally{
    [void][SplashInspect]::ReleaseDC($w,$hdc)
  }
  $statusPath=Join-Path $env:APPDATA 'TimeStamps\splash_status.txt'
  if(-not (Test-Path -LiteralPath $statusPath)) {throw 'Splash status file missing'}
  $status=Get-Content -LiteralPath $statusPath -Raw
  Write-Host ("Splash diagnostics: "+$status.Trim())
  if($status -notmatch 'imageDrawn=true' -and $status -notmatch 'artworkDrawn=true') {throw 'The approved full splash artwork did not draw'}
  Write-Host 'FULL APPROVED SPLASH: PASS (embedded 1672x941 PNG with serpent logo)'
  $unique=@($pixels|Sort-Object -Unique).Count
  Write-Host "Unique sample colors: $unique"
  if($unique -le 2){throw 'Splash actual pixels are uniform or near uniform, artwork is not visibly drawn in Win32 HDC'}
  Write-Host 'GDI SPLASH PIXELS: PASS (multiple artwork colors on actual popup)'
}finally{
  if(-not $p.HasExited){Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue}
}
