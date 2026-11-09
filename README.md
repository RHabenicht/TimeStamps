# TimeStamps — Updated by NightLife

TimeStamps is a Windows 10/11 (64-bit) desktop utility that generates sequential, randomized timestamps. This repository contains the NightLife modernization, the original program credits, and a single-file Windows installer.

**Current installer:** [TimeStamps 2.5 Setup — Updated by NightLife](releases/TimeStamps_2.5_Setup_Updated_by_NightLife.exe)

## TimeStamps 2.5

- Four-second NightLife splash screen, displayed **before** the main window is created.
- Purple serpent application icon embedded directly in the application, setup, and uninstaller, rather than depending on an external sidecar icon.
- Accessible Windows interface designed around 1360×768 at 125% display scaling, with large text.
- Setup, Format, and Results screens; sequential timestamp generation and overnight windows; format presets; copy/save output.
- True per-user Windows setup, Start Menu shortcut, optional desktop shortcut, Installed Apps registration, and uninstaller.
- Existing preferences preserved on upgrades. No administrator permissions required for standard installation.

## Install

Download the `.exe` file linked above and run it. The installer is not code-signed. Windows may show a publisher warning. The application runs offline and does not require the Delphi runtime or Go.

If upgrading a previous portable version, the installer keeps a `TimeStamps.ini` located alongside the installer when one is present and no newer settings file exists.

**Limit of verification:** Cross-compilation, automated timestamp tests, installation/rollback simulations, and PE icon resource validation passed. Interactive startup and splash visibility have not been verified on a physical Windows 10/11 machine.

## Build from source

`app/` contains the Go application, embedded splash artwork, multi-resolution icon, and Windows resource object. `setup/` contains the Go installer and uninstaller. Each module has its own `go.mod`.

On Windows, with Go installed:

```powershell
cd app
go test ./...
go build -ldflags "-H windowsgui" -o TimeStamps.exe .
cd ..\setup
go test ./...
go build -ldflags "-H windowsgui" -o Uninstall-TimeStamps.exe .\uninstaller
```

Then create `setup/payload.zip` containing the application executable, manifest, ICO, splash PNG, README, CHANGELOG, ORIGINAL_CREDITS, SPLASH_TEST_NOTES, and Uninstall-TimeStamps.exe at the archive root. Run `go build -ldflags "-H windowsgui" -o TimeStamps_Setup.exe .` from `setup/`.

The included `rsrc_windows_amd64.syso` files embed the program icon in compiled executables. `app/build_icon_resource.py` shows how those resources can be regenerated with Python 3 and clang.

GitHub Actions can also build the application and installer and upload the finished artifacts for inspection.

## Original credits

The original Delphi 3 program dates from 1997. Original programming: **Scott GrantSmith**. Requirements: **Branson Hagerty**. The modern update is credited to **NightLife**. See `ORIGINAL_CREDITS.txt` for the original attribution as recovered from the installation files.

## Status

The NightLife version is a reconstruction of the original desktop application. It is not the original 1997 executable, and its Windows GUI should not be considered verified until it is successfully launched and exercised on Windows.
