//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"timestampsinstaller/installer"
	"timestampsinstaller/winutil"
)

func main() {
	title := "TimeStamps 2.5 Uninstall"
	dir, err := winutil.InstallDir()
	if err != nil {
		winutil.Message(title, err.Error(), winutil.ErrorIcon)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--worker" {
		// Fixed expected install location, never trust a command-line removal path.
		worker(dir)
		return
	}
	if winutil.Message(title, "Remove TimeStamps from this Windows account?\n\nApplication files and shortcuts will be removed.\n\nYour saved settings in AppData and your legacy TimeStamps.ini will NOT be deleted.", winutil.ConfirmIcon) != winutil.Yes {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		winutil.Message(title, err.Error(), winutil.ErrorIcon)
		return
	}
	// An EXE cannot delete itself while running on Windows. Run a copy from TEMP.
	workerPath := filepath.Join(os.TempDir(), fmt.Sprintf("TimeStamps-Uninstall-%d.exe", os.Getpid()))
	b, err := os.ReadFile(exe)
	if err == nil {
		err = os.WriteFile(workerPath, b, 0600)
	}
	if err != nil {
		winutil.Message(title, "Could not stage uninstaller: "+err.Error(), winutil.ErrorIcon)
		return
	}
	cmd := exec.Command(workerPath, "--worker")
	if err := cmd.Start(); err != nil {
		winutil.Message(title, err.Error(), winutil.ErrorIcon)
		return
	}
	_ = cmd.Process.Release()
}
func worker(dir string) {
	title := "TimeStamps 2.5 Uninstall"
	// Allow the original uninstaller to exit and release its file lock.
	var err error
	for i := 0; i < 20; i++ {
		err = installer.RemoveInstalledFiles(dir)
		if err == nil {
			break
		}
		time.Sleep(350 * time.Millisecond)
	}
	if err != nil {
		winutil.Message(title, "Could not remove TimeStamps while its files are in use.\n\nClose the application, then run Uninstall TimeStamps again.\n\n"+err.Error(), winutil.ErrorIcon)
		return // Keep shortcuts and Windows Apps entry for retry.
	}
	problems := []string{}
	if err := winutil.RemoveShortcuts(dir); err != nil {
		problems = append(problems, err.Error())
	}
	if err := winutil.DeleteUninstallRegistry(); err != nil {
		problems = append(problems, err.Error())
	}
	if len(problems) > 0 {
		winutil.Message(title, "Removal completed with warnings:\n"+strings.Join(problems, "\n")+"\n\nSaved user data was preserved.", winutil.ErrorIcon)
	} else {
		winutil.Message(title, "TimeStamps was removed.\n\nSaved settings were preserved in AppData.", winutil.InfoIcon)
	}
	// TEMP worker is removed by Windows temp cleanup. Not part of the install.
}
