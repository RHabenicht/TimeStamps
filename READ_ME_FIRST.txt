TimeStamps 2.3 — Updated by NightLife
===================================
Stability corrections for Windows 10/11:
A+ / A− controls were removed because they were reported to cause a crash.
The Results text is fixed at 16 points for readability.

Format tab navigation and preset changes now run outside the originating
Windows control event to prevent nested message processing and lockups.
These are candidate fixes pending verification on a Windows 10/11 machine.

Extract all files before running TimeStamps.exe.
The app is portable; no installer or administrator access is required.

For 1360 × 768 at 125% Windows scaling: use the Setup, Format and Results
pages. All main controls are intended to fit within the usable work area.

This build passed automated tests and Windows cross-compilation, but has not
been interactively tested on Windows 10/11 in this environment.
