# Codex instructions — TimeStamps

These instructions apply to every change in this repository. Read `CODEX_HANDOFF_2026-11-09.md` and `PROJECT_CONTINUITY.md` before changing code.

## Product and ownership

- This repository is **TimeStamps** (`RHabenicht/TimeStamps`), a native 64-bit Windows 10/11 Go/Win32 rebuild of a 1997 Delphi 3 freeware timestamp generator. It is **not** NSRCG II or RunnerForge.
- Preserve original credits: program by Scott GrantSmith, requirements by Branson Hagerty, modernization **Updated by NightLife**. Retain collaborator credit in the approved splash.
- **Version stays 2.5.** Do not bump to 2.5.1 or rename the installed product, release, or EXE to a new version without explicit owner authorization.

## Protect the published binary and Norton setup

- **Do not edit, republish, overwrite, retag, delete, or replace GitHub release `v2.5` or its three downloadable assets** as a routine development step. The user wants to minimize repeated Norton 360 exclusions. A binary content change can cause Norton to flag an installer even when version and filename do not change.
- Build and test candidates in GitHub Actions artifacts or local build outputs only. The normal `.github/workflows/windows-build.yml` intentionally lacks release-publication permissions or a release-upload step.
- If the owner explicitly authorizes a release replacement, run the complete Windows tests first, disclose that Norton may treat the new binary as a different file, and follow the requested release procedure. Never claim the existing installer was updated when only code was committed.

## Approved splash — exact image, six seconds

- `app/splash_screen.png` is **the exact approved full image**: 1672 × 941 pixels, complete purple/neon `TimeStamps / Updated by NightLife / With Blackjack & DerFette / Shadowrun 1st • 2nd • 3rd Edition / Loading...` design with the large right-hand serpent logo.
- SHA-256 of **decoded 8-bit RGBA pixels in row-major order**: `f1311c993d8286e64151c090799e3733904191fdc00d8fffd2c92fc921c343e1`. This was verified against the image the user supplied in the conversation.
- Paint the **full existing PNG** without cropping, redrawing, regeneration, text edits, added `.ico` overlay, or a second separately drawn serpent. The splash's right-hand logo is already inside the PNG.
- Splash duration is **6000 ms**, before the main window is created. See `app/splash_image.go`, `app/splash_windows.go`, and the image/wiring tests. Use the Windows work area so the art fits the monitor.
- `TimeStamps.ico` is for application/Explorer/window icons; it is **not** extra splash artwork.
- Last launch diagnostic (when launched): `%APPDATA%\TimeStamps\splash_status.txt`, errors `splash_errors.log`.

## Functionality, accessibility, data safety

- Preserve sequential randomized timestamps, configurable gaps and posting windows (including overnight), last-timestamp continuation, setup/format/results, format presets/custom templates, copy/save, and legacy INI migration.
- Preserve legacy `TimeStamps.ini`, settings and unrelated user files through install/upgrade/uninstall. Installer is per-user without elevation in the standard case.
- Physical target: LG HDMI 1360 × 768 at Windows 125% scaling; large usable controls; **16 pt Consolas** result text; no A−/A+ controls. Do not reintroduce synchronous Format-tab page/preset operations that caused lockups.
- Do not introduce dependencies that compromise the standalone offline Windows executable without clear need.

## Tests and truthfulness

- On Windows run `./BUILD_RELEASE.ps1`; it runs Go tests and builds the app, installer and uninstaller. The Windows GitHub Actions workflow also checks PNG rendering and runs `ci/windows-install-smoke-v25.ps1` against the **candidate installer**, not the published release.
- Inspect the complete latest workflow logs, not only a green build indicator. Relevant tests: `app/splash_image_test.go` checks the approved pixel hash and six-second constant; `app/splash_windows_wiring_test.go` prohibits separate-logo painting; `ci/diagnose-splash-v25.ps1` reads painted pixels; the installer smoke performs install, upgrade, six-second splash, Format transitions, generation, and uninstall.
- A hosted Windows runner is **not** equivalent to the user's physical 1360 × 768 / 125% display or a Norton/SmartScreen test. Report these as unverified until actually tested. Do not ask the user to do development tests that Codex can run.
- Work from a branch/PR for further changes; keep GitHub URLs for deliverables rather than sandbox links.
