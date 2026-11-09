# TimeStamps 2.3

**Portable timestamp generator for 64-bit Windows 10 and Windows 11**

Originally developed as a Delphi 3 application in 1997 and reconstructed as a native Win32 desktop application. Windows modernization: **Updated by NightLife**.

## What it does

- Generates sequential timestamps with randomized intervals between entries.
- Supports posting-hour restrictions, including schedules that cross midnight.
- Provides date/time presets and custom token templates.
- Copies selected or all results to the clipboard and saves plain-text output.
- Can continue from the most recently generated timestamp between sessions.
- Imports settings from the original `TimeStamps.ini` without modifying that file.
- Stores new settings at `%APPDATA%\\NightLife\\TimeStamps\\settings.json`.

## Version 2.3 changes (November 6, 2026)

- Reworked Setup / Format / Results navigation to queue page changes after the originating Windows control event.
- Reworked Format preset selection to avoid nested Win32 event handling.
- Removed unstable A+ / A− font-size controls; Results uses fixed 16-point Consolas.
- Preserved existing presets, custom formatting, clipboard operations, and save behavior.
- Targeted a 1360 × 768 display at 125% Windows scaling.

## Running the application

The original 2.3 distribution contains `TimeStamps.exe`, `TimeStamps.exe.manifest`, `TimeStamps.ico`, and supporting documentation. Extract all distribution files into one folder and run `TimeStamps.exe`. No installer, administrator rights, .NET, Python, or network connection is required.

**Distribution note:** the Windows executable and ZIP archive are **not yet uploaded to this repository**. The documentation here does not constitute an installable build.

## Formatting tokens

| Token | Meaning |
|---|---|
| `{time}` | Time in selected style |
| `{date}` | Date in selected style |
| `{yyyy}`, `{yy}` | Four-/two-digit year |
| `{MM}`, `{dd}` | Month / day |
| `{HH}`, `{mm}`, `{ss}` | Hours / minutes / seconds |
| `{ddd}` | Abbreviated weekday |

## Verification status

Automated Go unit tests, format-preset checks, layout geometry assertions, and Windows 64-bit cross-compilation were reported as passing for this build. **The executable has not been interactively tested on a Windows 10/11 system in the supplied test report.** The Format-tab lockup fix remains a candidate pending runtime verification.

## Original credits and distribution terms

- **Original program (1997):** Scott GrantSmith
- **Original requirements:** Branson Hagerty
- **Windows 10/11 modernization:** Updated by NightLife

The original author released the program and source as freeware, permitting use and modification but prohibiting claims of original authorship or charging for distribution. Preserve the original credits and distribute without charge.

See `ORIGINAL_CREDITS.txt` for the original author's statement.

**Important:** The supplied TimeStamps 2.3 ZIP contains an executable, configuration file and documentation, **not the modernized source code**. This repository must not be treated as a source release unless that source is separately added.
