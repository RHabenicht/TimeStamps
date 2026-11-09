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
- Main window and six-second splash must fit inside the actual Windows work area, accounting for the taskbar and scaling. Static geometry tests are not proof of interactive display success.

## Branding, splash, and icon decisions

- Show the **six-second** splash **before** the main window appears (not behind it). The current implementation uses a single locked OS thread and pumps messages for splash painting. The checked-in splash image is app/splash_screen.png, with source graphics and decode/tests in app/.
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
- Historical v2.5 GitHub Actions build ran tests and published the initial release; the **current** Windows workflow builds, tests and uploads **Actions artifacts only**, without automatically modifying public release `v2.5`:
  https://github.com/RHabenicht/TimeStamps/actions/runs/42619772441
- **November 9 runtime update:** Published 2.5 installer was executed on a **hosted Windows 11 build 26100** runner: install, Installed Apps registration, shortcuts, upgrade preserving user files, splash-to-main interval (~4.1 seconds observed), 15 live Format/Setup cycles with preset changes, Generate (380 output characters), app exit confirmation and uninstall all PASS. Full evidence: [WINDOWS_RUNTIME_TEST_2026-11-09.md](WINDOWS_RUNTIME_TEST_2026-11-09.md), run https://github.com/RHabenicht/TimeStamps/actions/runs/42624608640. This is actual Windows runner execution, not mere compilation. **Still outstanding:** Windows 10 desktop; the user's physical 1360 × 768 @125%-scale layout/legibility; direct native mouse testing; Norton 360/SmartScreen behavior on the user's machine.
- The dedicated runtime Windows smoke job now verifies many of these UI/installer interactions, but not the physical monitor or Norton. Always distinguish the completed runner tests from tests still outstanding. If a real Windows test becomes available, add evidence and steps/results to a new report.

## Key source/reference files
- `README.md`, `CHANGELOG.txt`, `ORIGINAL_CREDITS.txt`
- `app/core.go`, `app/core_test.go`, `app/gui_windows.go`, `app/splash_windows.go`, `app/splash_image.go`, `app/splash_screen.png`, `app/icon_source.png`, `app/TimeStamps.ico`, `app/rsrc_windows_amd64.syso`
- `setup/`, `BUILD_RELEASE.ps1`, `.github/workflows/windows-build.yml`
- Library reports: `TimeStamps_QA_2026-11-06.txt`; `TimeStamps_2.4_Installer_Test_Report.txt`
- Library visual reference: `Neon Shadowrun TimeStamps Splash Screen.png` (dice-free approved visual); older variants include `TimeStamps Neon Cyberpunk Splash.png` and `Neon Serpent TimeStamps Splash Screen.png`.

## November 9 final update — six-second approved artwork and pinned release

The approved user-supplied full 1672 × 941 purple TimeStamps/Updated by NightLife splash image **already matches** the repository PNG pixel for pixel (decoded RGBA SHA-256 `f1311c993d8286e64151c090799e3733904191fdc00d8fffd2c92fc921c343e1`). The user specifically rejected an additional separately drawn serpent icon; the **entire original PNG, including the built-in large serpent**, is now shown for **six seconds** before main. `splashDurationMs=6000`, with a no-overlay regression test.

The final approved source Windows test: https://github.com/RHabenicht/TimeStamps/actions/runs/42649970586 — PASS (painted full art, ~6.2-second splash, install/upgrade, 15 Format/Setup cycles, generation, uninstall). One-time public asset replacement succeeded at https://github.com/RHabenicht/TimeStamps/actions/runs/42650243890. The published v2.5 installer SHA-256 is `14160d2728a87b2b60160c837f22b037653900e0dcf57e4985c6b35783502ee6`. The prior ~4.1-second timing and old digest elsewhere in this document describe the **historical earlier v2.5 build**, not the current release.

The user emphasized keeping **version 2.5** and avoiding repeated Norton 360 exclusions. The public `v2.5` assets are pinned; routine source commits and builds must not overwrite them. The currently published executable remains unsigned, and any changed binary can require Norton reevaluation regardless of unchanged version/filename.

For Codex transfer, see `AGENTS.md` and `CODEX_HANDOFF_2026-11-09.md`. Remaining unverified: user's physical LG display/125% DPI, Norton 360, SmartScreen, and Windows 10 runtime.

## Historical interpretation
The user's latest request was to publish this **2.5** Windows installer, corrected icon, splash, and source to GitHub and test it, without requiring them to do the development/testing work. On November 9 the project reached a completed Windows GitHub Actions build, a tagged release with three binaries, and a separate PASS on actual Windows runner installer/GUI/uninstaller smoke tests. Those labels were outdated in an earlier v2.5 binary; the currently published v2.5 release has corrected About and uninstall labels. Do not use the early-release SHA or runtime as the current published build. Keep future TimeStamps work in this repository; do not confuse this project with the distinct **NSRCG II** or **RunnerForge** applications.

Deletion of the prior conversation is the user's decision. This file preserves *retrieved project decisions and references* but is not a verbatim transcript or a complete copy of chat-only attachments.

## Codex takeover — source fixes, November 9, 2026

- Initial audit baseline: `6662ff6c7535b369742cd5c906c1364e2994eb71`. Codex reran the existing Windows candidate workflow: [42651653205](https://github.com/RHabenicht/TimeStamps/actions/runs/42651653205), PASS. The complete logs identify **Windows Server 2025, build 26100**; do not describe this as a verified Windows 10/11 desktop test.
- Confirmed core failures: daylight-saving changes shifted local posting hours (New York 09:00–17:00 became 10:00–18:00 on March 8, 2026), and validation rejected documented token-only templates such as `{HH}:{mm}:{ss}`.
- Fixes are on branch `codex/fix-dst-and-template-validation`, [PR #1](https://github.com/RHabenicht/TimeStamps/pull/1). Application fixes: `64b70f4`; corrected cross-process smoke-test helper: `7a7d331`. Local calendar boundaries and clock-hour checks preserve overnight/full-day windows, handle skipped spring hours and permit both repeated autumn hours. Validation and formatting share the token list. Setup now says six seconds; the unused application version constant says 2.5.
- New tests fail against the old core and pass with the fixes: all 576 posting windows across each New York DST transition, exact session boundaries, generation/continuation, skipped/repeated hours, and every documented template token. The timezone data import is test-only; the standalone application gains no runtime dependency.
- The Windows smoke now enters a custom clock-token template, verifies its edit-control text, and validates output format and count. Its first run exposed incorrect use of cross-process `SetWindowText`; the test helper now sends `WM_SETTEXT` with a timeout. Application unit tests and splash rendering passed in that first run, but the overall run failed and must not be cited as a passing end-to-end result.
- A subsequent run reached a pre-existing dialog-initialization race: the smoke read an empty body immediately after the installer dialog HWND appeared. Commit `87554c1` adds a bounded wait for text initialization before assertions. Neither failed run is an end-to-end pass.
- Candidate verification: [42653482015](https://github.com/RHabenicht/TimeStamps/actions/runs/42653482015), source `87554c13e30106d686404a47f59f9031d0c9b434`, **PASS**, complete logs reviewed. Go application/package/installer tests, painted splash pixels, ~6.1-second splash before main, real install/upgrade, 15 Format/Setup transitions, custom-template generation (20 lines, 200 characters), exit and uninstall with user-file preservation all passed. Local application tests, race checks and `go vet` also passed.
- [Candidate artifacts](https://github.com/RHabenicht/TimeStamps/actions/runs/42653482015/artifacts/13691247586) are separate from the release. Candidate installer: 12,562,432 bytes, SHA-256 `1dff5c38f5c62ea55fcfc0ad5c2017c6a7666771ea6e3c38cc7e2349cf7b3905`. Candidate portable EXE SHA-256: `8207595a119788cfb2b8ec8c4d4920c338d37e9ad27b8edb0f7dc93c5f3acfe3`. The continuity-note commit after this run changes documentation only.
- Version remains **2.5**. No release/tag/assets were changed. The exact approved splash image, 6000 ms duration, no-overlay rendering, original credits, fixed 16 pt Consolas, settings format and installer user-data protections remain in place. Physical LG 1360×768/125% display, direct keyboard/mouse accessibility, Windows 10/11 desktop, Norton 360 and SmartScreen remain unverified.

## Physical splash failure report — November 9, 2026

The owner supplied a screenshot showing only a solid purple splash and small TimeStamps/Updated by NightLife text. They confirmed that the normal application opens afterward. This matches the text-only error fallback in `startupSplashWndProc`; it is now an observed physical-machine splash-rendering failure, despite the hosted runner's earlier passing splash test. The screenshot alone does not establish the precise device/driver cause.

Branch `codex/fix-splash-rendering-fallback`, [PR #2](https://github.com/RHabenicht/TimeStamps/pull/2), builds on PR #1. If `StretchDIBits` returns failure, the candidate scales the same embedded image in software and copies its pixels at 1:1 size with `SetDIBitsToDevice`. Text-only fallback remains if both GDI operations fail. `splash_status.txt` now names the renderer, and `splash_errors.log` records failures and recovery attempts. No artwork replacement, crop, overlay, settings change, version bump or release upload is involved.

The Windows-specific regression test injects both stretch-failure return values and compares actual GDI pixels with source-artwork pixels at 1180×664 (the screenshot size) and 976×549 (runner size), including corners and the right-hand artwork. This tests recovery on hosted Windows, not the owner's driver. The owner has not yet run this candidate, so physical-PC recovery remains unverified. The published v2.5 installer remains frozen and does not include this source change.

Candidate verification: [42655436679](https://github.com/RHabenicht/TimeStamps/actions/runs/42655436679), source `7923d111e6fb706dd6a86cd80b0b2558be273405`, **PASS**; complete logs reviewed. Both forced-failure return values recovered with correct actual GDI pixels at both sizes. Full application/package/installer tests, exact artwork hash/no-overlay checks, normal splash pixel inspection, ~6.2-second splash before main, install/upgrade, 15 Format transitions, custom-template generation and uninstall/user-file preservation passed on Windows Server 2025 build 26100. Local application tests and `go vet` passed.

[Candidate artifacts](https://github.com/RHabenicht/TimeStamps/actions/runs/42655436679/artifacts/13691727482): installer 12,563,968 bytes, SHA-256 `6de8ac1c70809a543c670c120e7e8226cd658252ff5a0ef405fede0e28ef71cb`; portable EXE SHA-256 `09296ff1829bbf211bf0632e4e4ab0ef1ca1be757274590f5094c7d0e5db3a5a`. These are development candidates, not the pinned release binaries. The subsequent continuity-note commit changes documentation only.

### Candidate failure confirmed on the owner's PC

The owner subsequently installed that candidate and supplied updated logs. The new `renderer=none` field proves that the alternate-renderer build ran; this was not another launch of the old public release. The completed six-second splash still reported `artworkDrawn=false`. Both `StretchDIBits` and `SetDIBitsToDevice` returned zero scanlines, with no useful Win32 last-error value. Therefore the first fallback is **confirmed insufficient on this PC**, despite passing on the hosted runner. Foreground rejection was also logged but does not by itself explain either transfer failure.

Commit `9c2b0e9` adds a third route after both DIB transfers fail: fill a Windows-owned DIB-section bitmap through its memory pointer, select it into a compatible memory DC, and copy with `BitBlt`. This avoids passing the raw DIB to either failed transfer API. The new Windows test forces both operations to return zero/no error, invokes the production splash WM_PAINT procedure on an actual visible popup, and compares painted corner/interior/right-side pixels against the approved artwork. This improves on the earlier memory-DC-only recovery test. The precise device-specific cause and physical-PC recovery remain unverified.

Verification: [42657869349](https://github.com/RHabenicht/TimeStamps/actions/runs/42657869349), source `9c2b0e96903ac60333c6635b0f359986469ed354`, **PASS**; complete logs reviewed. The production 976×549 popup painted correct approved pixels via BitBlt with both transfers forced to return zero/no error. Full Windows Server 2025 build 26100 application/package tests, exact artwork hash, no-overlay checks, regular splash pixel inspection, ~6.1-second splash, install/upgrade, 15 Format transitions, custom-template output and uninstall/user-file preservation passed. Local application tests and `go vet` passed. This is still not verification on the owner's PC.

Use the **new** [candidate artifact](https://github.com/RHabenicht/TimeStamps/actions/runs/42657869349/artifacts/13692922178), not the earlier artifact that failed on the owner's PC. Installer size: 12,565,504 bytes; SHA-256 `0e1153b39df806bb0ed157852a1ede7281e26de011aaa88a6d91cf44dd13d654`. Portable EXE SHA-256: `ff36b9b18f13ab51222a2f92f28aaa2110a2b602bfd3dfba5c3aa632528fcbe5`. Version remains 2.5; release assets are frozen. The following continuity-note commit changes documentation only.

## Installer PATH failure — November 9, 2026

The owner next supplied a screenshot of the 12,565,504-byte candidate setup reporting that application files were installed but Windows integration failed: `powershell.exe` and `reg.exe` were not found in `%PATH%`. This is a confirmed separate installer defect. The screenshot does not establish whether the newest BitBlt splash route worked or failed afterward. Do not ask the owner to repair their PATH or treat this as an artwork replacement request.

Branch `codex/fix-installer-system-tool-paths`, [PR #3](https://github.com/RHabenicht/TimeStamps/pull/3), builds on PR #2. Commit `9374b24` locates the system directory with `GetSystemDirectoryW` and launches Windows' registry utility and Windows PowerShell through absolute paths, for both install and uninstall. Windows unit tests demonstrate bare helper names fail with empty PATH, then run both helpers successfully through the new resolution; same-named PATH files are ignored. The full installer/upgrade/application/uninstaller smoke now gives every candidate child an explicitly empty PATH. Parent CI tools keep their original environment. A new `TimeStamps-2.5-Setup` Actions artifact contains just the setup executable to simplify downloads.

Verification: [42658831355](https://github.com/RHabenicht/TimeStamps/actions/runs/42658831355), source `9374b24bb475e5d82a7e0a085536e63e9dbdfa60`, **PASS**; complete logs reviewed. Both helper tests passed. Actual candidate install, shortcut creation, Installed Apps registration, upgrade, application launch with ~6.2-second splash, Format transitions/generation and uninstall all passed with empty child PATH on Windows Server 2025 build 26100. Full application/package tests and the forced-double-transfer-failure BitBlt popup test passed; Windows-targeted `go vet ./winutil` passed locally. Physical-PC installation/splash recovery, Windows 10/11 desktop, Norton and SmartScreen remain unverified.

Latest [setup-only candidate ZIP](https://github.com/RHabenicht/TimeStamps/actions/runs/42658831355/artifacts/13692489028) contains just `TimeStamps_2.5_Setup_Updated_by_NightLife.exe`, 12,568,576 bytes; SHA-256 `f415d6e29d04cc71e0fbf4559f3dc1a0a213cf38b66353457589fbe1af7bf560`. It includes all earlier source fixes, including BitBlt recovery. Version remains 2.5; no release/tag/assets were modified. The subsequent continuity-note commit changes documentation only.
