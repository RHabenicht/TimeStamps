# TimeStamps → Codex transfer

**Prepared:** 2026-11-09  
**GitHub:** https://github.com/RHabenicht/TimeStamps  
**Branch / source baseline before transfer:** `main` at `2d65021c3cdff123626f8e0ef03809d8d3a4e19f` (before this documentation-only commit).  
**Current product/release:** **TimeStamps 2.5**, not 2.5.1.  
**Project separation:** This is unrelated to NSRCG II and RunnerForge.

## Assignment for Codex

Take over future **source maintenance, Windows testing, bug investigation and packaging verification** for TimeStamps. Read `AGENTS.md`, this handoff, `PROJECT_CONTINUITY.md`, and `README.md`. **Do not publish or replace a release or change the version** merely because you can build a better candidate. Start by auditing the current repo/source, establishing a reproducible tested baseline and identifying any remaining real issues. Report findings with evidence.

## Current published release — freeze this until explicitly authorized

- Release: https://github.com/RHabenicht/TimeStamps/releases/tag/v2.5
- Installer: https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/TimeStamps_2.5_Setup_Updated_by_NightLife.exe
- Portable executable: https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/TimeStamps.exe
- Uninstaller: https://github.com/RHabenicht/TimeStamps/releases/download/v2.5/Uninstall-TimeStamps.exe
- Published assets were updated **2026-11-09 16:00:01 UTC** and subsequently left pinned.
- Current release installer: **12,562,944 bytes**, SHA-256 `14160d2728a87b2b60160c837f22b037653900e0dcf57e4985c6b35783502ee6`.
- Current portable EXE: **6,279,168 bytes**, SHA-256 `dcc12cda65131870c926ef945ff63b66be2b3e671977cfe7523c9f225b3990a3`.
- Current uninstaller: **3,387,392 bytes**, SHA-256 `7386aa342f00a251e8f5ce0ea353428e42071d4f460290ba907f1fbcaa457866`.

**Why freeze:** The user uses Norton 360 and specifically requested staying at **2.5 to avoid repeatedly changing Norton exclusions**. A changed executable may be identified as new even under the same version or filename. Earlier development replaced the v2.5 asset multiple times; **do not repeat this unintentionally**. The standard Windows CI workflow builds/uploads **Actions artifacts only**, with read-only repository content permission. Publishing requires explicit owner request and deliberate release action. New development should not alter these public assets.

## Precise approved splash

The user supplied an image showing the **complete** splash design:
- Purple neon sci-fi frame/background, metallic white `TimeStamps`, `UPDATED BY NightLife`, `WITH Blackjack & DerFette`, `SHADOWRUN 1ST • 2ND • 3RD EDITION`, `Loading...`, and the large purple serpent logo on the **right**.
- The checked-in `app/splash_screen.png` is **already pixel-identical** to that image at 1672 × 941. Exact decoded RGBA SHA-256: `f1311c993d8286e64151c090799e3733904191fdc00d8fffd2c92fc921c343e1`.
- **Six seconds**, not the historical four-second setting: `const splashDurationMs = 6000` in `app/splash_image.go`. Start main window only after splash has finished.
- User reported a splash with **no logo** on their computer in a prior installer. A mistaken remedy added a separate serpent application icon onto the splash. The user clarified that the **full PNG itself** is the intended logo. The extra icon overlay has now been **removed**; the whole PNG is drawn alone using Win32/GDI.
- Do not swap the full splash for the `.ico` file, crop the big right-side serpent, add a second overlay, or regenerate the image.
- `app/splash_image_test.go` checks the exact approved pixel fingerprint; `app/splash_windows_wiring_test.go` checks that no extra icon painting is reintroduced.
- Latest reported splash diagnostics path: `%APPDATA%\TimeStamps\splash_status.txt`; failures: `splash_errors.log`. Actual display on the user's machine after the final replacement is **not yet independently confirmed**.

## Verified source and end-to-end testing

- Windows tests of the **exact final six-second full-artwork source**: https://github.com/RHabenicht/TimeStamps/actions/runs/42649970586 — **PASS**.
  - Full embedded 1672 × 941 PNG rendered in Win32 splash, painted pixels sampled successfully.
  - Installed app: splash detected before main, **~6.2 seconds** measured.
  - Real installer and upgrade; preserved legacy INI and unrelated user files.
  - Windows registration/shortcuts, 15 Format/Setup transitions, preset changes, timestamp generation (380 characters), unsaved output close, uninstaller: PASS.
- One-time **final corrected release-asset upload**: https://github.com/RHabenicht/TimeStamps/actions/runs/42650243890 — **PASS**; source from tested run above.
- Subsequent Windows build workflow at the source baseline before transfer: https://github.com/RHabenicht/TimeStamps/actions/runs/42650358165 — **PASS**.
- Historical earlier test report `WINDOWS_RUNTIME_TEST_2026-11-09.md` documents a *different earlier v2.5 binary* with ~4.1 seconds and an old SHA. **Do not mistake its installer checksum/timing for the current published v2.5 asset.** The SHA-256 identifiers above describe the current published binaries.
- Tests ran on hosted Windows with OS 10.0.26100 (Windows 11-class runner) and Win32 API automation. **Not established**: the actual user's monitor/125% DPI, direct mouse testing, Windows 10, Norton 360 local detection, and real SmartScreen behavior. Avoid promising that physical splash issue is resolved until confirmed.

## Critical regression constraints

1. **Version 2.5** in UI, installer, uninstaller, Installed Apps; four-part Win32 manifest identity `2.5.0.0` (a three-part version previously caused a Windows side-by-side launch failure).
2. No user-data loss. Legacy `TimeStamps.ini` stays untouched. Settings saved in `%APPDATA%\NightLife\TimeStamps\settings.json`; user-generated files persist on upgrade/uninstall.
3. Interface fits physical LG HDMI **1360 × 768, 125%** scaling. Fixed **16-point Consolas output**; do not restore A−/A+ controls.
4. Format screen must not hang: deferred Win32 message handling for navigation and preset application, with reentrancy safeguards.
5. Preserve original 1997 freeware credits: Scott GrantSmith (program), Branson Hagerty (requirements), modern work credited `Updated by NightLife`. Splash includes collaborator credit `Blackjack & DerFette`.
6. Standalone offline Go/Win32 Windows executable, per-user installer, signed status **unsigned**.
7. Don't confuse installed executable icon resources with the full-screen splash image.

## Source map and reproduction

| Need | File |
| --- | --- |
| Current rules for Codex | `AGENTS.md`; this file |
| Full earlier continuity | `PROJECT_CONTINUITY.md` |
| Splash image | `app/splash_screen.png` |
| Image decoder, sizing, duration | `app/splash_image.go` |
| Splash Win32 paint, foreground, diagnostics | `app/splash_windows.go` |
| Artwork fingerprint and duration tests | `app/splash_image_test.go` |
| Prevent icon overlay and verify startup ordering | `app/splash_windows_wiring_test.go` |
| Main UI / Format navigation | `app/gui_windows.go`; `app/format_tab_regression_test.go` |
| Core timestamp algorithm and tests | `app/core.go`; `app/core_test.go` |
| Manifest version regression | `app/manifest_version_test.go`; `app/TimeStamps.exe.manifest` |
| Installer and user-file preservation | `setup/`; `setup/installer/core.go`; `setup/package_test.go` |
| Windows build orchestration | `BUILD_RELEASE.ps1` |
| Build + graphical/installer test workflow | `.github/workflows/windows-build.yml` |
| Actual splash pixel inspection | `ci/diagnose-splash-v25.ps1` |
| Install, live GUI, generate, uninstall | `ci/windows-install-smoke-v25.ps1` |

Run `./BUILD_RELEASE.ps1` on a Windows host or trigger the existing Windows Actions workflow on a development branch. The build produces `dist/TimeStamps.exe`, `dist/Uninstall-TimeStamps.exe` and `dist/TimeStamps_2.5_Setup_Updated_by_NightLife.exe`; the default workflow **does not upload these to the public release**. Inspect workflow logs and check package payload identity, application launch, six-second splash pixels and time, Format behavior, output, uninstall, settings preservation.

## First Codex actions

1. Read `AGENTS.md`, this handoff, `PROJECT_CONTINUITY.md`, latest code/tests and GitHub Actions results.
2. Confirm no unrequested release or version change; record SHA-256 of any development artifacts separately from current published v2.5 hashes.
3. Reproduce the full Windows CI run. If a bug is found, fix it on a branch, add focused regression coverage and report the evidence. Do **not** alter public v2.5 artifacts without an explicit owner decision.
4. Audit any remaining physical-DPI/splash weaknesses, but distinguish hypothetical problems from reproduced problems.
5. Report a concise **tested/untested** summary, source commit, check results and artifact access on GitHub. No sandbox-only download links.
