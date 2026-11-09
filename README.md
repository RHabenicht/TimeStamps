# TimeStamps — Updated by NightLife

TimeStamps is a Windows 10/11 (64-bit) desktop utility that generates sequential, randomized timestamps. This repository contains the NightLife modernization, the original program credits, and a single-file Windows installer.

**Current installer:** [Download TimeStamps 2.5 Setup (.exe)](https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/TimeStamps_2.5_Setup_Updated_by_NightLife.exe)

[TimeStamps 2.5 release](https://github.com/RHabenicht/TimeStamps/releases/tag/v2.5)

## TimeStamps 2.5

- Displays the user-approved six-second full-screen-artwork splash PNG without an additional icon overlay. Exact decoded pixels match the supplied image.
- Writes a one-file startup diagnostic to `%APPDATA%\TimeStamps\splash_status.txt` and errors to `splash_errors.log`.
- Shows a readable fallback message if the artwork cannot be drawn.
- **The splash is part of application startup, not the installer or uninstaller.**
- Updates About and uninstaller version labels.

## Other TimeStamps 2.5 features

- Six-second NightLife splash screen, displayed **before** the main window is created.
- Purple serpent application icon embedded directly in the application, setup, and uninstaller, rather than depending on an external sidecar icon.
- Accessible Windows interface designed around 1360×768 at 125% display scaling, with large text.
- Setup, Format, and Results screens; sequential timestamp generation and overnight windows; format presets; copy/save output.
- True per-user Windows setup, Start Menu shortcut, optional desktop shortcut, Installed Apps registration, and uninstaller.
- Existing preferences preserved on upgrades. No administrator permissions required for standard installation.

## Install

Download the `.exe` file linked above and run it. The installer is not code-signed. Windows may show a publisher warning. The application runs offline and does not require the Delphi runtime or Go.

If upgrading a previous portable version, the installer keeps a `TimeStamps.ini` located alongside the installer when one is present and no newer settings file exists.

**Verification:** Automated timestamp tests, installer/rollback simulations, and Windows executable checks passed. The **actual published v2.5 setup** also passed a Windows 11 GitHub runner smoke test covering install, upgrade, splash (~4.1 seconds observed), 15 Format/Setup transitions, Generate, and uninstall. See [Windows runtime test report](WINDOWS_RUNTIME_TEST_2026-11-09.md). This does not yet verify Windows 10, the user's physical 1360×768/125%-scale screen, Norton 360 reputation, or mouse-driven visual accessibility.

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

## Preserved project decisions

See [PROJECT_CONTINUITY.md](PROJECT_CONTINUITY.md) for preserved requirements and issue history from the original TimeStamps development chat. The [November 9 Windows 11 runtime report](WINDOWS_RUNTIME_TEST_2026-11-09.md) records the first passed published-installer desktop smoke test and its remaining limitations.

## Release stability

The public **v2.5 installer is pinned**. Routine GitHub builds run Windows tests and upload temporary Actions artifacts; they do **not** overwrite the v2.5 downloadable installer. This avoids changing the published binary every time the source changes. Norton 360 may still evaluate different downloaded files by signature or reputation; keeping a version/filename alone cannot guarantee a permanent antivirus exception.

## Status

The NightLife version is a reconstruction of the original desktop application, not the 1997 Delphi executable. It has now been launched and exercised on a hosted Windows 11 runner, but the specific physical 1360×768 display at 125% scaling and Windows 10 are still unverified. Version 2.5 corrects the outdated uninstaller and About dialog version labels. The splash image pixels match the uploaded reference exactly; the rendering evidence comes from a hosted Windows runner; this is not a guarantee of visibility on every physical PC.
