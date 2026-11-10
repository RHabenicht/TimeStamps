//go:build windows

package winutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSystemToolsWorkWithoutPATH(t *testing.T) {
	t.Setenv("PATH", "")
	for _, name := range []string{"reg.exe", "powershell.exe"} {
		if _, err := exec.LookPath(name); err == nil {
			t.Fatalf("test does not reproduce missing helper: %s is still on PATH", name)
		}
	}
	cmd, err := systemCommand("reg.exe", "/?")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(cmd.Path) {
		t.Fatalf("helper is not absolute: %s", cmd.Path)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("absolute reg.exe failed with empty PATH: %v: %s", err, output)
	}
	marker := filepath.Join(t.TempDir(), "shortcut helper 測試.txt")
	if err := runPowerShell(`[IO.File]::WriteAllText($env:TS_TEST_MARKER,'helper ran')`, "TS_TEST_MARKER="+marker); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "helper ran" {
		t.Fatalf("PowerShell did not run with empty PATH: data=%q err=%v", data, err)
	}
	t.Log("reg.exe and Windows PowerShell both ran successfully with empty PATH")
}

func TestSystemToolsIgnoreSameNamedPATHFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	for _, name := range []string{"reg.exe", "powershell.exe"} {
		fake := filepath.Join(dir, name)
		if err := os.WriteFile(fake, []byte("not a Windows executable"), 0600); err != nil {
			t.Fatal(err)
		}
		cmd, err := systemCommand(name)
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsAbs(cmd.Path) || cmd.Path == fake {
			t.Fatalf("selected PATH helper: %s", cmd.Path)
		}
		if _, err := os.Stat(cmd.Path); err != nil {
			t.Fatalf("system helper unavailable: %s: %v", cmd.Path, err)
		}
	}
}
