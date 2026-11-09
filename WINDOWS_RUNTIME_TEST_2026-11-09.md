# TimeStamps 2.5 — historical initial published installer Windows runtime test
**Historical evidence notice:** This report describes an **earlier v2.5 binary** with a four-second splash. It is not the currently published installer. The final six-second original-artwork build was tested in [Windows run 42649970586](https://github.com/RHabenicht/TimeStamps/actions/runs/42649970586) and published in [release run 42650243890](https://github.com/RHabenicht/TimeStamps/actions/runs/42650243890). Current published v2.5 installer SHA-256: `14160d2728a87b2b60160c837f22b037653900e0dcf57e4985c6b35783502ee6`.

**Date:** November 9, 2026
**Result:** PASS — published release installer, actual Windows runtime interaction
**GitHub Actions:** https://github.com/RHabenicht/TimeStamps/actions/runs/42624608640
**Executable under test:** `TimeStamps_2.5_Setup_Updated_by_NightLife.exe` from the unchanged public v2.5 release (not freshly compiled for this run).
**Setup file size:** 12,558,848 bytes
**SHA-256:** `9467B83DA01B165EC04455DE8ACAAB5E59146129F1D1963BC39132A77B3E00B3`
**Host:** Windows runner, reported `Microsoft Windows NT 10.0.26100.0` (Windows 11 24H2 build 26100); process session ID 2. Windows 10 was not exercised.

## Observed results (not just compiler/unit test assertions)

- PASS — published setup downloaded and launched; installer confirmation dialogs answered through Win32.
- PASS — actual per-user installation created application, splash image, icon, manifest and uninstaller.
- PASS — Windows Start Menu shortcut created and HKCU Installed Apps registration reports version 2.5.
- PASS — real setup run for upgrade; existing legacy `TimeStamps.ini` sentinel and unrelated user file retained; requested desktop shortcut created.
- PASS — installed application started; visible splash-window class observed before main-window class; about **4.1 seconds** elapsed between observing the splash and observing the main window.
- PASS — 15 alternating Format/Setup navigation cycles, including selection and notification of the Format preset, were handled without a hung Win32 message; relevant controls became visible and hidden as expected.
- PASS — invoked Generate, found nonempty timestamp output (**380 characters**) in the Windows EDIT control, and observed Results controls.
- PASS — closed app and accepted its unsaved-output confirmation.
- PASS — launched actual uninstaller, answered its confirmation and completion prompts; application executable, Start Menu and desktop shortcuts, and Installed Apps entry were removed; user INI and unrelated user file retained.

## Scope and limits

The test uses a hosted Windows runner with programmatic Win32 window discovery and message dispatch; it validates actual process execution and OS integration. It is **not** a screenshot-based visual design review, a mouse-driven accessibility audit, or testing on the user's Windows machine. Tested version should not be mistaken for a signed/certified product.

**Still unverified:** Windows 10 desktop compatibility; actual 1360×768 monitor at 125% scaling; high-DPI appearance/legibility and potential physical clipping; native mouse-driven navigation in the user's setup; Windows Defender/Norton 360 classification on the user's own device; SmartScreen reputation of unsigned binaries; real-world install/uninstall outside the hosted runner.

**Known cosmetic discrepancy in the tested 2.5 executable:** uninstaller dialog title still says "TimeStamps 2.4 Uninstall"; About dialog text still refers to TimeStamps 2.3. These labels should be corrected in a subsequent release rather than describing the existing public v2.5 binary as having already been changed.

## Reproduction

Source of the executable automated test: `ci/windows-install-smoke.ps1`.
Workflow: `.github/workflows/windows-install-smoke.yml`.
The build/unit-test workflow remains a separate verification track.

The Windows runtime output in the linked job logs includes every PASS line, SHA-256 digest and the measured splash delay. A passing runtime smoke test is stronger than compilation alone, but is not proof of every possible GUI state.
