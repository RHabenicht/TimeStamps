package installer

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func payloadForTest(t *testing.T) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, name := range InstalledFiles {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.Write([]byte("payload " + name))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func TestFreshInstallAndUninstall(t *testing.T) {
	p := payloadForTest(t)
	root := t.TempDir()
	dest := filepath.Join(root, "Programs", "TimeStamps")
	if err := Install(p, dest); err != nil {
		t.Fatal(err)
	}
	for _, name := range InstalledFiles {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dest, "TimeStamps.ini"), []byte("user original settings"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := RemoveInstalledFiles(dest); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dest, "TimeStamps.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "user original settings" {
		t.Fatal("legacy INI changed")
	}
	for _, name := range InstalledFiles {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatalf("%s remained", name)
		}
	}
}
func TestUpgradePreservesSettingsAndOtherFiles(t *testing.T) {
	p := payloadForTest(t)
	dir := filepath.Join(t.TempDir(), "TimeStamps")
	if err := Install(p, dir); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("MY DATA"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "TimeStamps.ini"), []byte("OLD INI"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "TimeStamps.exe"), []byte("old exe"), 0644)
	if err := Install(p, dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "TimeStamps.exe"))
	if string(b) != "payload TimeStamps.exe" {
		t.Fatal("exe not upgraded")
	}
	b, _ = os.ReadFile(filepath.Join(dir, "TimeStamps.ini"))
	if string(b) != "OLD INI" {
		t.Fatal("legacy INI lost")
	}
	b, _ = os.ReadFile(filepath.Join(dir, "notes.txt"))
	if string(b) != "MY DATA" {
		t.Fatal("user data lost")
	}
}
func TestRejectUnexpectedOrMissingPayload(t *testing.T) {
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	f, _ := w.Create("../bad.exe")
	_, _ = f.Write([]byte("x"))
	_ = w.Close()
	if VerifyPayload(out.Bytes()) == nil {
		t.Fatal("accepted path traversal")
	}
	out.Reset()
	w = zip.NewWriter(&out)
	f, _ = w.Create("TimeStamps.exe")
	_, _ = f.Write([]byte("x"))
	_ = w.Close()
	if VerifyPayload(out.Bytes()) == nil {
		t.Fatal("accepted incomplete payload")
	}
}
func TestUnicodeDirectoryAndRepeatInstall(t *testing.T) {
	p := payloadForTest(t)
	d := filepath.Join(t.TempDir(), "Program Files 測試", "TimeStamps")
	for i := 0; i < 3; i++ {
		if err := Install(p, d); err != nil {
			t.Fatal(err)
		}
	}
}
func TestBlockedTargetLeavesOriginalIntact(t *testing.T) {
	p := payloadForTest(t)
	d := filepath.Join(t.TempDir(), "TimeStamps")
	_ = os.MkdirAll(d, 0755)
	// Force a failure when processing a later file; ensure earlier files roll back.
	_ = os.WriteFile(filepath.Join(d, "TimeStamps.exe"), []byte("original"), 0644)
	_ = os.Mkdir(filepath.Join(d, "README.txt"), 0755)
	if err := Install(p, d); err == nil {
		t.Fatal("expected failed install")
	}
	b, err := os.ReadFile(filepath.Join(d, "TimeStamps.exe"))
	if err != nil || string(b) != "original" {
		t.Fatal("rollback did not restore original executable")
	}
}
