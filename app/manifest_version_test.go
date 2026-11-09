package main

import (
    "os"
    "regexp"
    "testing"
)

func TestWindowsManifestUsesFourPartVersion(t *testing.T) {
    data, err := os.ReadFile("TimeStamps.exe.manifest")
    if err != nil { t.Fatal(err) }
    // Windows side-by-side manifests require four numeric components.
    re := regexp.MustCompile(`<assemblyIdentity name="NightLife.TimeStamps" type="win32" version="([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)"`)
    m := re.FindSubmatch(data)
    if m == nil { t.Fatal("application manifest lacks a four-part version") }
    if string(m[1]) != "2.5.0.0" { t.Fatalf("application version must remain 2.5, got %s", m[1]) }
}
