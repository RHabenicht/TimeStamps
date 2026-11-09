//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"timestampsinstaller/installer"
	"timestampsinstaller/winutil"
)

//go:embed payload.zip
var payload []byte

func main() {
	title := "TimeStamps 2.5.1 Setup — NightLife"
	dir, err := winutil.InstallDir()
	if err != nil {
		winutil.Message(title, err.Error(), winutil.ErrorIcon)
		return
	}
	verb := "Install"
	if _, err := os.Stat(filepath.Join(dir, "TimeStamps.exe")); err == nil {
		verb = "Upgrade"
	}
	if winutil.Message(title, fmt.Sprintf("%s TimeStamps 2.5.1 for Windows 10/11 (64-bit)?\n\nInstallation folder:\n%s\n\nA Start Menu shortcut and Windows Apps uninstall entry will be created.\nAdministrator privileges are not required.\n\nExisting settings are preserved.\n\nThe four-second splash appears when TimeStamps starts, not during setup.", verb, dir), winutil.ConfirmIcon) != winutil.Yes {
		return
	}
	desktop := winutil.Message(title, "Create a desktop shortcut?", winutil.ConfirmIcon) == winutil.Yes
	if err := installer.Install(payload, dir); err != nil {
		winutil.Message(title, "Installation failed. Previous application files were restored if possible.\n\n"+err.Error(), winutil.ErrorIcon)
		return
	}
	// Upgrade discovery: when launched from a portable folder, retain a legacy
	// INI alongside the installed EXE for the app's existing import routine.
	if self, err := os.Executable(); err == nil {
		oldINI := filepath.Join(filepath.Dir(self), "TimeStamps.ini")
		newINI := filepath.Join(dir, "TimeStamps.ini")
		if _, err := os.Stat(newINI); os.IsNotExist(err) {
			if b, err := os.ReadFile(oldINI); err == nil {
				_ = os.WriteFile(newINI, b, 0600)
			}
		}
	}
	var warnings []string
	if err := winutil.CreateShortcuts(dir, desktop); err != nil {
		warnings = append(warnings, "Shortcuts: "+err.Error())
	}
	if err := winutil.RegisterUninstaller(dir); err != nil {
		warnings = append(warnings, "Windows Apps registration: "+err.Error())
	}
	if len(warnings) > 0 {
		winutil.Message(title, "TimeStamps application files were installed, but Windows integration needs attention:\n\n"+joinWarnings(warnings)+"\n\nYou can run TimeStamps.exe directly from the installation folder.", winutil.ErrorIcon)
		return
	}
	if winutil.Message(title, "TimeStamps 2.5.1 installed successfully.\n\nOpen TimeStamps now?", winutil.ConfirmIcon) == winutil.Yes {
		cmd := exec.Command(filepath.Join(dir, "TimeStamps.exe"))
		cmd.Dir = dir
		if err := cmd.Start(); err != nil {
			winutil.Message(title, err.Error(), winutil.ErrorIcon)
		} else {
			_ = cmd.Process.Release()
		}
	}
}
func joinWarnings(a []string) string {
	v := ""
	for _, x := range a {
		if v != "" {
			v += "\n"
		}
		v += x
	}
	return v
}
