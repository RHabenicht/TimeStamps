package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"testing"
	"timestampsinstaller/installer"
)

//go:embed payload.zip
var releasePayload []byte

func TestReleaseInstallerPayload(t *testing.T) {
	if err := installer.VerifyPayload(releasePayload); err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(releasePayload), int64(len(releasePayload)))
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "Install", "TimeStamps")
	if err := installer.Install(releasePayload, root); err != nil {
		t.Fatal(err)
	}
	for _, f := range r.File {
		h, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		expected, e := io.ReadAll(h)
		_ = h.Close()
		if e != nil {
			t.Fatal(e)
		}
		actual, e := os.ReadFile(filepath.Join(root, f.Name))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(actual, expected) {
			t.Errorf("%s differs from embedded payload", f.Name)
		}
	}
	// Emulate user settings; verify upgrade, then removal, leaves them untouched.
	_ = os.WriteFile(filepath.Join(root, "TimeStamps.ini"), []byte("legacy_user_preferences"), 0600)
	if err := installer.Install(releasePayload, root); err != nil {
		t.Fatal(err)
	}
	if err := installer.RemoveInstalledFiles(root); err != nil {
		t.Fatal(err)
	}
	b, e := os.ReadFile(filepath.Join(root, "TimeStamps.ini"))
	if e != nil || string(b) != "legacy_user_preferences" {
		t.Fatal("settings overwritten or removed")
	}
}
