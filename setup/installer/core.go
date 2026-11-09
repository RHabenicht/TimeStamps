package installer

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// InstalledFiles is an allowlist. Uninstallation removes only these files,
// never user-created files or the legacy TimeStamps.ini.
var InstalledFiles = []string{
	"TimeStamps.exe", "TimeStamps.exe.manifest", "TimeStamps.ico",
	"splash_screen.png", "README.txt", "CHANGELOG.txt", "ORIGINAL_CREDITS.txt",
	"SPLASH_TEST_NOTES.txt", "Uninstall-TimeStamps.exe",
}

func allowed(name string) bool {
	for _, x := range InstalledFiles {
		if name == x {
			return true
		}
	}
	return false
}

func VerifyPayload(payload []byte) error {
	r, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, f := range r.File {
		if !allowed(f.Name) || f.FileInfo().IsDir() || strings.ContainsAny(f.Name, `/\\`) {
			return fmt.Errorf("unexpected installation file %q", f.Name)
		}
		if seen[f.Name] {
			return fmt.Errorf("duplicate installation file %q", f.Name)
		}
		if f.UncompressedSize64 > 100<<20 {
			return errors.New("oversized application file")
		}
		seen[f.Name] = true
		h, e := f.Open()
		if e != nil {
			return e
		}
		_, e = io.Copy(io.Discard, h)
		closeErr := h.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
	}
	for _, x := range InstalledFiles {
		if !seen[x] {
			return fmt.Errorf("missing required installation file %q", x)
		}
	}
	return nil
}
func token() string { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }

// Install installs into a fixed per-user directory. It stages all files and
// rolls back changes if copying an application file fails. User content and
// settings are never part of this transaction.
func Install(payload []byte, dest string) error {
	if err := VerifyPayload(payload); err != nil {
		return err
	}
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp(parent, ".timestamps-setup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	staged := filepath.Join(scratch, "new")
	backups := filepath.Join(scratch, "old")
	if err := os.MkdirAll(staged, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(backups, 0755); err != nil {
		return err
	}
	r, _ := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	for _, f := range r.File {
		h, e := f.Open()
		if e != nil {
			return e
		}
		target := filepath.Join(staged, f.Name)
		out, e := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			h.Close()
			return e
		}
		_, e = io.Copy(out, h)
		c1 := out.Close()
		c2 := h.Close()
		if e != nil {
			return e
		}
		if c1 != nil {
			return c1
		}
		if c2 != nil {
			return c2
		}
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	var backed, placed []string
	rollback := func(err error) error {
		for i := len(placed) - 1; i >= 0; i-- {
			_ = os.Remove(filepath.Join(dest, placed[i]))
		}
		for i := len(backed) - 1; i >= 0; i-- {
			_ = os.Rename(filepath.Join(backups, backed[i]), filepath.Join(dest, backed[i]))
		}
		return err
	}
	for _, name := range InstalledFiles {
		old := filepath.Join(dest, name)
		if _, err := os.Lstat(old); err == nil {
			info, e := os.Lstat(old)
			if e != nil {
				return rollback(e)
			}
			if !info.Mode().IsRegular() {
				return rollback(fmt.Errorf("installation target %q is not a regular file", name))
			}
			if err := os.Rename(old, filepath.Join(backups, name)); err != nil {
				return rollback(fmt.Errorf("close TimeStamps and retry (%s): %w", name, err))
			}
			backed = append(backed, name)
		} else if !os.IsNotExist(err) {
			return rollback(err)
		}
		if err := os.Rename(filepath.Join(staged, name), old); err != nil {
			return rollback(err)
		}
		placed = append(placed, name)
	}
	return nil
}

// RemoveInstalledFiles only removes files listed in InstalledFiles. Other
// files in the application directory, including TimeStamps.ini, remain.
func RemoveInstalledFiles(dest string) error {
	var issues []string
	for _, name := range InstalledFiles {
		err := os.Remove(filepath.Join(dest, name))
		if err != nil && !os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("%s: %v", name, err))
		}
	}
	// If there is user content or a legacy INI in the directory, preserve it.
	_ = os.Remove(dest)
	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}
