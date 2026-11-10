TIMESTAMPS 2.5 — UPDATED BY NIGHTLIFE
=====================================
Windows 10/11 desktop application (64-bit).

QUICK START
Installer: run TimeStamps_2.5_Setup_Updated_by_NightLife.exe and follow the prompts.
Portable copy: extract the ZIP to a writable folder and run TimeStamps.exe.

1. Choose a start date, number of timestamps, and minimum/maximum spacing.
2. Set an overnight posting window if needed.
3. Click Generate timestamps.
4. Copy selected text, copy all results, or save them as a .txt file.

TimeStamps runs offline and includes its executable dependencies. Standard per-user installation does not require administrator access.

SETTINGS
On first launch, TimeStamps imports settings from TimeStamps.ini when that file is beside TimeStamps.exe. It leaves the original INI unchanged.

New preferences are saved here:
%APPDATA%\NightLife\TimeStamps\settings.json

To start with factory defaults, rename or remove TimeStamps.ini before the first launch. To reset preferences later, close TimeStamps and rename settings.json.

FORMAT TEMPLATES
  {time}       Time using the selected time style
  {date}       Date using the selected date style
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

POSTING WINDOWS
A window such as 22:00 to 04:00 spans midnight. The first timestamp falls within the first half hour of the evening window. Each later timestamp advances by a randomly selected interval. If the next timestamp would fall outside the allowed hours, it moves to the next permitted session. The overnight wait does not count toward the maximum gap; otherwise the time window and gap limits could conflict.

Selecting the same start and end hour allows timestamps throughout the day. Clear the posting-hours restriction to remove the time window.

SEQUENCES AND OUTPUT
- Continue from last generated timestamp carries the sequence across Generate clicks and saves the last timestamp for later sessions.
- Changing the start date begins a new sequence.
- Clear earlier results on Generate controls whether new output replaces or appends to the current results.
- TimeStamps warns before clearing or closing with unsaved output.
- Save .txt writes UTF-8 text with Windows CRLF line endings.
- Copy selected uses the normal text selection in the output field.

WINDOW ICON AND SPLASH
TimeStamps.ico supplies the application window icon. Keep the optional external manifest beside TimeStamps.exe for modern Windows common-control styling.

The approved full splash image appears for six seconds before the main window opens. The PNG already contains the large serpent logo; TimeStamps draws the full image without an icon overlay. If the splash cannot initialize, the app opens normally.

Splash diagnostics:
%APPDATA%\TimeStamps\splash_status.txt
%APPDATA%\TimeStamps\splash_errors.log

TEST STATUS
The published v2.5 installer passed a hosted Windows 11 test covering splash rendering, installation, upgrade, 15 Format/Setup transitions, timestamp generation, and uninstall. The runner measured about 6.2 seconds for the splash. See:
https://github.com/RHabenicht/TimeStamps/actions/runs/42649970586

Windows 10, the user's physical 1360 x 768 display at 125% scaling, and Norton 360/SmartScreen behavior on the user's computer remain unverified. The installer is unsigned, so Windows may show a publisher or reputation warning.

CREDITS
Original 1997 program: Scott GrantSmith
Original requirements: Branson Hagerty
Modernization: Updated by NightLife

The original freeware terms and full attribution appear in ORIGINAL_CREDITS.txt.
