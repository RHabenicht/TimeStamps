$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$App = Join-Path $Root 'app'
$Setup = Join-Path $Root 'setup'
$Out = Join-Path $Root 'dist'
New-Item -ItemType Directory -Force -Path $Out | Out-Null
Push-Location $App
try {
  go test -v ./...
  if ($LASTEXITCODE -ne 0) { throw 'TimeStamps Go unit tests failed' }
  go build -ldflags '-H windowsgui' -o (Join-Path $Out 'TimeStamps.exe') .
  if ($LASTEXITCODE -ne 0) { throw 'TimeStamps application build failed' }
} finally { Pop-Location }
Push-Location $Setup
try {
  go build -ldflags '-H windowsgui' -o (Join-Path $Out 'Uninstall-TimeStamps.exe') ./uninstaller
  if ($LASTEXITCODE -ne 0) { throw 'Uninstaller build failed' }
  $Payload = Join-Path $Setup 'payload.zip'
  if (Test-Path $Payload) {Remove-Item $Payload -Force}
  Add-Type -AssemblyName System.IO.Compression
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $zip=[IO.Compression.ZipFile]::Open($Payload,[IO.Compression.ZipArchiveMode]::Create)
  try {
    foreach ($name in @('TimeStamps.exe','TimeStamps.exe.manifest','TimeStamps.ico','splash_screen.png','README.txt','CHANGELOG.txt','ORIGINAL_CREDITS.txt','SPLASH_TEST_NOTES.txt','Uninstall-TimeStamps.exe')) {
      if ($name -in @('TimeStamps.exe','Uninstall-TimeStamps.exe')) { $path=Join-Path $Out $name }
      else { $path=Join-Path $App $name }
      [IO.Compression.ZipFileExtensions]::CreateEntryFromFile($zip,$path,$name,[IO.Compression.CompressionLevel]::Optimal) | Out-Null
    }
  } finally { $zip.Dispose() }
  go test ./...
  if ($LASTEXITCODE -ne 0) { throw 'Installer Go unit tests failed' }
  go build -ldflags '-H windowsgui' -o (Join-Path $Out 'TimeStamps_2.5_Setup_Updated_by_NightLife.exe') .
  if ($LASTEXITCODE -ne 0) { throw 'Installer build failed' }
} finally { Pop-Location }
Write-Host "Finished builds under $Out"
