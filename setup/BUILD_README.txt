TimeStamps 2.5 Windows Native Installer Sources

The package contains Go source for the single-file, per-user Windows setup,
its independent uninstaller, and the embedded payload.zip built from the
approved TimeStamps 2.5 release (including splash). The app's own source is
in TimeStamps_2.5_Source_Updated_by_NightLife.zip (separate artifact).

Requirements: Go 1.23 or later. No external Go modules.

Build uninstaller before regenerating the payload:
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
    -ldflags='-H windowsgui -s -w' -o payload/Uninstall-TimeStamps.exe ./uninstaller

Create payload.zip from the nine payload/ files, preserving basenames at ZIP root.
Build installer:
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
    -ldflags='-H windowsgui -s -w' -o TimeStamps_Setup.exe .

Test cross-platform installer filesystem logic:
  go test ./...

Do not assume successful Windows execution from Linux compilation alone.

Application and setup both require the embedded rsrc_windows_amd64.syso in their respective Go package directories.
