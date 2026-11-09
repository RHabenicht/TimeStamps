TIMESTAMPS 2.5 — UPDATED BY NIGHTLIFE
=====================================
Portable Windows 10/11 application (64-bit).

ORIGINAL CREDITS
- Original 1997 Delphi 3 application: Scott GrantSmith
- Original program requirements: Branson Hagerty
- Windows 10/11 reconstruction and modernization: Updated by NightLife

QUICK START
1. Extract the entire ZIP archive into any writable folder.
2. Double-click TimeStamps.exe.
3. Choose a start date, number of stamps, and minimum/maximum separation.
4. Optionally restrict generated times to an overnight posting window.
5. Click Generate timestamps.
6. Copy selected text, Copy all, or Save .txt.

The app does not require Delphi, Python, .NET or an internet connection.
It does not require administrator privileges.

OLD SETTINGS
An existing TimeStamps.ini from your 1997 program is supported if it is
placed next to TimeStamps.exe. The portable ZIP deliberately does not include
a default INI, so upgrades cannot overwrite your settings.
When launching the modern app for the first time, it reads that INI file and
imports your stored settings, including the original time and date style and
custom comment prefix. The original INI is never modified.

New settings are saved in:
%APPDATA%\NightLife\TimeStamps\settings.json

If you want factory defaults instead, rename or remove TimeStamps.ini BEFORE
first launch. To reset after first launch, close the program and rename the
settings.json file in the user settings folder.

FORMAT TEMPLATES
  {time}       Time using selected time style
  {date}       Date using selected date style
  {yyyy}       Four-digit year
  {yy}         Two-digit year
  {MM}         Two-digit month
  {dd}         Two-digit day
  {HH}         24-hour hour
  {mm}         Minutes
  {ss}         Seconds
  {ddd}        Abbreviated day of week

Examples:
  {time}/{date}                22:45:09/07-13-51
  ({time}/{date})              (22:45:09/07-13-51)
  {date}T{time}                2051-07-13T22:45:09 (with ISO date style)

TIME WINDOW BEHAVIOR
A posting window such as 22:00 to 04:00 spans midnight. The first generated
stamp is placed in the first half hour of the requested evening window;
later stamps advance by a randomly selected separation. When the next stamp
would fall outside allowed hours, it moves to the next permitted session.
The overnight waiting period is exempt from the configured maximum gap;
otherwise both requirements cannot simultaneously be met. This corrects an
edge case in the 1997 generator.

Selecting the same start and end hour enables a continuous 24-hour window.
Unchecking the restriction removes the posting-hours limit.

CONTINUATION AND SAFETY
- Continue from last generated timestamp continues across Generate clicks
  and saves the last timestamp to user preferences for later sessions.
- Changing the start date starts a new sequence.
- Clear earlier results on Generate controls replace versus append.
- The app warns before clearing or closing with unsaved generated output.
- Save .txt uses UTF-8 text with Windows CRLF line breaks.
- Copy selected uses normal text selection in the output field.

PORTABILITY
All executable dependencies are built into TimeStamps.exe. The optional
external manifest requests modern Windows common-control styling; keep it
beside the EXE. TimeStamps.ico supplies the application window icon.


STARTUP SPLASH (2.5)
- The approved purple serpent artwork (without dice) and embedded serpent logo icon are included in TimeStamps.exe. The logo is drawn explicitly on top of the artwork.
- The splash is shown once on application startup for six seconds.
- The main window is created only after the splash closes.
- The splash fits entirely inside the usable Windows work area, including
  1360x768 at 125% display scaling with a normal Windows taskbar.
- If the image or splash window cannot initialize, the program skips the splash
  and opens normally. Existing timestamp generation and formats are unchanged.
- The original 1997 credits remain in ORIGINAL_CREDITS.txt. The splash artwork
  uses the exact wording approved by the user, including Blackjack & DerFette.

FORMAT TAB STABILITY FIX (2.3)
- Format tab navigation now posts its view change to the Windows message queue.
- Date/time format preset edits are also applied after the selection event.
- Re-entrant command notifications are suppressed while controls update.
- Format templates and original saved configurations are retained.
- If the Format tab still stalls, close the program and report exactly which
  control was used; a native Windows GUI reproduction is still necessary.

TESTING / LIMITATIONS
- Compiled as a 64-bit Windows GUI executable on Linux using Go 1.23.2.
- Timestamp engine passed unit tests for overnight windows, sequential
  ordering, equal interval limits, continuation, formatting and INI import.
- Windows graphical UI has NOT been executed on a Windows 10/11 machine.
  Verify layout, clipboard, save dialog, date picker and compatibility there.
- This program is a new native Win32 GUI, not a binary patch to the 1997 EXE.

LICENSE / ATTRIBUTION
The original author explicitly released his program and source as freeware,
allowing modifications, but forbidding claims of authorship of the original or
charging for distribution. This package preserves that attribution and is
provided without charge. See ORIGINAL_CREDITS.txt.
