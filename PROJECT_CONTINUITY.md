# TimeStamps project continuity / handoff
Preserved November 9, 2026, from the previous TimeStamps chat, its QA files, and the checked-in v2.5 GitHub source. This document is intended to remain available after the full prior chat is deleted.

## Project identity and scope

- Product: **TimeStamps**, a Windows desktop utility for generating timestamp sequences, originally authored as a Delphi 3 application in September 1997.
- Original credits: **Scott GrantSmith** (original programming); **Branson Hagerty** (original requirements).
- Modernization credit: **Updated by NightLife**. Keep original authorship visibly distinguished from the modernization credit.
- Target: Windows **10/11**, **64-bit**. The replacement is a Windows Go executable built with Win32 APIs, not an edited/recompiled original Delphi executable; no Delphi runtime needed.
- Repository: https://github.com/RHabenicht/TimeStamps
- Version archived here: **2.5**; released at https://github.com/RHabenicht/TimeStamps/releases/tag/v2.5

## Functionality to retain

- Generate sequential, randomized timestamps constrained by a selected start date, hours, minimum/maximum gaps, and posting window.
- Support overnight posting windows and continuation from a previously generated sequence/last timestamp.
- Setup / Format / Results pages; date and time formatting, custom templates and presets (Classic, Forum, ISO 8601, Date then time); copy selected or all results; save output.
- Persist settings and import any legacy TimeStamps.ini without overwriting a modified user's existing settings.
- Prefer tests of the core behavior, including 576 combinations of posting-window start/end hours, rather than unverified claims about Windows UI behavior.

## User's Windows environment and usability constraints

- Actual reported monitor: LG HDMI, **1360 × 768** desktop resolution with **125%** Windows display scaling.
- The bottom of the older UI was cut off and the user has difficulty reading small text. Controls, headings, footer, page navigation and Generate action must remain visible, usable and large enough for the reported display.
- A−/A+ text-size controls caused instability; **remove/leave absent**, not merely disable; keep fixed, legible **16-point Consolas output**.
- Do not regress Format-page navigation. The previously observed lockup occurred while using the Format screen. v2.3 deferred page changes and preset application via queued Win32 messages instead of synchronous reentrant command handling.
- Main window and four-second splash must fit inside the actual Windows work area, accounting for the taskbar and scaling. Static geometry tests are not proof of interactive display success.

## Branding, splash, and icon decisions

- Show the **four-second** splash **before** the main window appears (not behind it). The current implementation uses a single locked OS thread and pumps messages for splash painting. The checked-in splash image is app/splash_screen.png, with source graphics and decode/tests in app/.
- Approved later visual direction: a **purple neon serpent/crest with no dice**, metallic TimeStamps title, 'Updated by NightLife', collaborator acknowledgments, and the Shadowrun 1st/2nd/3rd Edition caption.
- Historical artwork alternatives have differing collaborator text: the dice-free library reference says **'Blackjack & DerFette'**; an earlier dice-including candidate says **'Blackjack & NathanHawks'**. This is a historical discrepancy, not authorization to change the final checked-in splash; preserve current artwork/credits until user explicitly requests a revision.
- Corrected serpent icon must be **embedded** in TimeStamps.exe, installer and uninstaller; the separate TimeStamps.ico should not be required to show the primary executable icon.
- Do not automatically replace the splash with a newly generated image or change the user's approved design.

## Packaging/install requirements

- Provide a real, single-file, **unsigned Windows setup .exe**, not just a portable ZIP. Installer must work per user **without UAC/admin** for typical operation.
- Default installation directory: `%LOCALAPPDATA%\\Programs\\TimeStamps`; Start Menu group **NightLife**; optional desktop shortcut; Windows Installed Apps registration and uninstaller.
- Preserve `TimeStamps.ini`, existing user changes, settings under `%APPDATA%\\NightLife\\TimeStamps`, and user-created files on install/upgrade/uninstall. Do not install a default INI that overwrites the user's existing one.
- Previous version 2.3 portable QA normalized the top-level directory name to `TimeStamps` instead of versioned directories that install side-by-side.
- Version 2.4 setup design completed static tests, filesystem simulations, rollback tests and Win64 compile tests; no actual Windows desktop/installer execution was verified then.
- Version 2.5 source includes installer and uninstaller, payload verification with an allowlist, temporary staging, and rollback upon failed replacement. Build via `BUILD_RELEASE.ps1` on Windows.
- Windows can warn about the unsigned publisher / SmartScreen. Do not characterize this as signed or Windows-certified.

## Verification (what has and has not been tested)

- v2.3 source/library QA: **17 Go tests**, **576 posting windows**, **12 portable install simulations** reported passing. This does **not** establish physical GUI functionality.
- v2.5 GitHub Actions Windows runner builds application, installer, uninstaller, runs Go tests, checks executables, uploads build artifacts, and publishes release `v2.5`:
  https://github.com/RHabenicht/TimeStamps/actions/runs/42619772441
- **Outstanding validation:** actual interactive Format-page navigation/lockup regression; splash visibility and accurate duration; launching the installed app on Windows 10 and 11; physical 1360 × 768 @125% sizing and accessibility; Start Menu/desktop shortcuts and Installed Apps registration/uninstall via a live desktop session.
- Never present a successful CI build or simulated installer as proof that the above live UI checks passed. If a real Windows test becomes available, add evidence and steps/results to a new report.

## Key source/reference files
- `README.md`, `CHANGELOG.txt`, `ORIGINAL_CREDITS.txt`
- `app/core.go`, `app/core_test.go`, `app/gui_windows.go`, `app/splash_windows.go`, `app/splash_image.go`, `app/splash_screen.png`, `app/icon_source.png`, `app/TimeStamps.ico`, `app/rsrc_windows_amd64.syso`
- `setup/`, `BUILD_RELEASE.ps1`, `.github/workflows/windows-build.yml`
- Library reports: `TimeStamps_QA_2026-11-06.txt`; `TimeStamps_2.4_Installer_Test_Report.txt`
- Library visual reference: `Neon Shadowrun TimeStamps Splash Screen.png` (dice-free approved visual); older variants include `TimeStamps Neon Cyberpunk Splash.png` and `Neon Serpent TimeStamps Splash Screen.png`.

## Historical interpretation
The user's latest request was to publish this **2.5** Windows installer, corrected icon, splash, and source to GitHub and test it, without requiring them to do the development/testing work. On November 9 the project reached a completed, passing Windows GitHub Actions run and a tagged release with three binaries. Keep future TimeStamps work in this repository; do not confuse this project with the distinct **NSRCG II** or **RunnerForge** applications.

Deletion of the prior conversation is the user's decision. This file preserves *retrieved project decisions and references* but is not a verbatim transcript or a complete copy of chat-only attachments.
