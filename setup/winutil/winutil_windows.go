//go:build windows

package winutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	ConfirmIcon = 0x4 | 0x20 // MB_YESNO | MB_ICONQUESTION
	InfoIcon    = 0x40
	ErrorIcon   = 0x10
	Yes         = 6
)

var procMessageBox = syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")

func Message(title, body string, flags uintptr) int {
	p1, _ := syscall.UTF16PtrFromString(body)
	p2, _ := syscall.UTF16PtrFromString(title)
	v, _, _ := procMessageBox.Call(0, uintptr(unsafe.Pointer(p1)), uintptr(unsafe.Pointer(p2)), flags)
	return int(v)
}
func InstallDir() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not defined")
	}
	return filepath.Join(local, "Programs", "TimeStamps"), nil
}

const UninstallRegistry = `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\NightLife-TimeStamps`

func regAdd(name, value, typ string) error {
	cmd := exec.Command("reg.exe", "ADD", UninstallRegistry, "/v", name, "/t", typ, "/d", value, "/f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("registry %s: %v: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}
func RegisterUninstaller(dir string) error {
	values := [][3]string{
		{"DisplayName", "TimeStamps — Updated by NightLife", "REG_SZ"},
		{"DisplayVersion", "2.5.1", "REG_SZ"},
		{"Publisher", "NightLife", "REG_SZ"},
		{"InstallLocation", dir, "REG_SZ"},
		{"DisplayIcon", filepath.Join(dir, "TimeStamps.exe") + ",0", "REG_SZ"},
		{"UninstallString", `"` + filepath.Join(dir, "Uninstall-TimeStamps.exe") + `"`, "REG_SZ"},
		{"NoModify", "1", "REG_DWORD"},
		{"NoRepair", "1", "REG_DWORD"},
		{"EstimatedSize", "6855", "REG_DWORD"},
	}
	for _, v := range values {
		if err := regAdd(v[0], v[1], v[2]); err != nil {
			return err
		}
	}
	return nil
}
func DeleteUninstallRegistry() error {
	out, err := exec.Command("reg.exe", "DELETE", UninstallRegistry, "/f").CombinedOutput()
	if err != nil && !strings.Contains(strings.ToLower(string(out)), "unable to find") {
		return fmt.Errorf("remove Apps entry: %v: %s", err, string(out))
	}
	return nil
}
func runPowerShell(script string, envvars ...string) error {
	cmd := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.Env = append(os.Environ(), envvars...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Windows shortcut operation: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Uses actual Desktop and Programs known-folder paths so OneDrive folder
// redirection and non-English user names are handled by Windows.
func CreateShortcuts(dir string, desktop bool) error {
	script := `$ErrorActionPreference='Stop'
$programs=[Environment]::GetFolderPath('Programs')
$folder=Join-Path $programs 'NightLife'
[void][IO.Directory]::CreateDirectory($folder)
$shell=New-Object -ComObject WScript.Shell
$app=$env:TS_APP
$uninstall=$env:TS_UNINSTALL
$link=$shell.CreateShortcut((Join-Path $folder 'TimeStamps.lnk'))
$link.TargetPath=$app
$link.WorkingDirectory=(Split-Path -Parent $app)
$link.Description='TimeStamps 2.5.1 - Updated by NightLife'
$link.IconLocation=$env:TS_ICON
$link.Save()
$link=$shell.CreateShortcut((Join-Path $folder 'Uninstall TimeStamps.lnk'))
$link.TargetPath=$uninstall
$link.WorkingDirectory=(Split-Path -Parent $uninstall)
$link.Description='Uninstall TimeStamps'
$link.Save()
if ($env:TS_DESKTOP -eq 'yes') {
  $desktop=[Environment]::GetFolderPath('Desktop')
  $link=$shell.CreateShortcut((Join-Path $desktop 'TimeStamps.lnk'))
  $link.TargetPath=$app
  $link.WorkingDirectory=(Split-Path -Parent $app)
  $link.Description='TimeStamps 2.5.1 - Updated by NightLife'
  $link.IconLocation=$env:TS_ICON
  $link.Save()
}`
	opt := "no"
	if desktop {
		opt = "yes"
	}
	return runPowerShell(script, "TS_APP="+filepath.Join(dir, "TimeStamps.exe"), "TS_UNINSTALL="+filepath.Join(dir, "Uninstall-TimeStamps.exe"), "TS_ICON="+filepath.Join(dir, "TimeStamps.ico"), "TS_DESKTOP="+opt)
}
func RemoveShortcuts(dir string) error {
	script := `$ErrorActionPreference='Stop'
$shell=New-Object -ComObject WScript.Shell
$programs=[Environment]::GetFolderPath('Programs')
$folder=Join-Path $programs 'NightLife'
$app=$env:TS_APP
$uninstall=$env:TS_UNINSTALL
foreach ($pair in @(
  [pscustomobject]@{Link=(Join-Path $folder 'TimeStamps.lnk');Target=$app},
  [pscustomobject]@{Link=(Join-Path $folder 'Uninstall TimeStamps.lnk');Target=$uninstall},
  [pscustomobject]@{Link=(Join-Path ([Environment]::GetFolderPath('Desktop')) 'TimeStamps.lnk');Target=$app}
)) {
  $link=$pair.Link;$target=$pair.Target
  if (Test-Path -LiteralPath $link) {
   try { $resolved=$shell.CreateShortcut($link).TargetPath
     if ([string]::Equals($resolved,$target,[StringComparison]::OrdinalIgnoreCase)) { Remove-Item -LiteralPath $link -Force }
   } catch {}
  }
}
if (Test-Path -LiteralPath $folder) {
  if (@(Get-ChildItem -LiteralPath $folder -Force).Count -eq 0) {Remove-Item -LiteralPath $folder}
}`
	return runPowerShell(script, "TS_APP="+filepath.Join(dir, "TimeStamps.exe"), "TS_UNINSTALL="+filepath.Join(dir, "Uninstall-TimeStamps.exe"))
}
