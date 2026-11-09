//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_CLIPCHILDREN     = 0x02000000
	WS_BORDER           = 0x00800000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_EX_CLIENTEDGE    = 0x00000200
	ES_LEFT             = 0x0
	ES_MULTILINE        = 0x4
	ES_AUTOVSCROLL      = 0x0040
	ES_AUTOHSCROLL      = 0x0080
	ES_READONLY         = 0x0800
	ES_NUMBER           = 0x2000
	CBS_DROPDOWNLIST    = 0x3
	BS_AUTOCHECKBOX     = 0x3
	WM_CREATE           = 0x0001
	WM_DESTROY          = 0x0002
	WM_SIZE             = 0x0005
	WM_GETMINMAXINFO    = 0x0024
	WM_PAINT            = 0x000F
	WM_CLOSE            = 0x0010
	WM_COMMAND          = 0x0111
	// Defer view changes until after the originating Win32 command returns.
	WM_APP_SWITCHPAGE   = 0x8031
	WM_APP_APPLYPRESET  = 0x8032
	WM_CTLCOLOREDIT     = 0x0133
	WM_CTLCOLORSTATIC   = 0x0138
	WM_ERASEBKGND       = 0x0014
	WM_SETFONT          = 0x0030
	WM_SETICON          = 0x0080
	WM_COPY             = 0x0301
	WM_GETTEXT          = 0x000D
	BM_GETCHECK         = 0x00F0
	BM_SETCHECK         = 0x00F1
	BST_CHECKED         = 1
	CB_ADDSTRING        = 0x0143
	CB_GETCURSEL        = 0x0147
	CB_SETCURSEL        = 0x014E
	EM_SETSEL           = 0x00B1
	EM_REPLACESEL       = 0x00C2
	EM_LIMITTEXT        = 0x00C5
	EM_SCROLLCARET      = 0x00B7
	EN_CHANGE           = 0x0300
	CBN_SELCHANGE       = 1
	SW_SHOW             = 5
	SW_HIDE             = 0
	CW_USEDEFAULT       = 0x80000000
	DT_SINGLELINE       = 0x20
	DT_VCENTER          = 0x4
	DT_LEFT             = 0
	TRANSPARENT         = 1
	PS_SOLID            = 0
	IMAGE_ICON          = 1
	LR_LOADFROMFILE     = 0x10
	ICON_SMALL          = 0
	ICON_BIG            = 1
	CF_UNICODETEXT      = 13
	GMEM_MOVEABLE       = 0x2
	DTM_FIRST           = 0x1000
	DTM_GETSYSTEMTIME   = DTM_FIRST + 1
	DTM_SETSYSTEMTIME   = DTM_FIRST + 2
	DTM_SETFORMATW      = DTM_FIRST + 50
	GDT_VALID           = 0
	OFN_OVERWRITEPROMPT = 0x2
	OFN_PATHMUSTEXIST   = 0x800
	MB_OK               = 0
	MB_ICONERROR        = 0x10
	MB_ICONINFORMATION  = 0x40
	MB_YESNO            = 0x4
	IDYES               = 6
	ID_DATE             = 101
	ID_QTY              = 102
	ID_MIN              = 103
	ID_MAX              = 104
	ID_HOURS            = 105
	ID_START            = 106
	ID_END              = 107
	ID_PRESET           = 108
	ID_TEMPLATE         = 109
	ID_DATELAYOUT       = 110
	ID_TIMELAYOUT       = 111
	ID_CONTINUE         = 112
	ID_REPLACE          = 113
	ID_GENERATE         = 114
	ID_COPYSEL          = 115
	ID_COPYALL          = 116
	ID_SAVE             = 117
	ID_CLEAR            = 118
	ID_ABOUT            = 119
	ID_PAGE_SETUP       = 201
	ID_PAGE_FORMAT      = 202
	ID_PAGE_RESULTS     = 203
)

var (
	dllUser32                = syscall.NewLazyDLL("user32.dll")
	dllKernel32              = syscall.NewLazyDLL("kernel32.dll")
	dllGdi32                 = syscall.NewLazyDLL("gdi32.dll")
	dllComctl32              = syscall.NewLazyDLL("comctl32.dll")
	dllComdlg32              = syscall.NewLazyDLL("comdlg32.dll")
	procRegisterClassEx      = dllUser32.NewProc("RegisterClassExW")
	procCreateWindowEx       = dllUser32.NewProc("CreateWindowExW")
	procDefWindowProc        = dllUser32.NewProc("DefWindowProcW")
	procShowWindow           = dllUser32.NewProc("ShowWindow")
	procUpdateWindow         = dllUser32.NewProc("UpdateWindow")
	procGetMessage           = dllUser32.NewProc("GetMessageW")
	procTranslateMessage     = dllUser32.NewProc("TranslateMessage")
	procDispatchMessage      = dllUser32.NewProc("DispatchMessageW")
	procPostQuitMessage      = dllUser32.NewProc("PostQuitMessage")
	procSendMessage          = dllUser32.NewProc("SendMessageW")
	procPostMessage          = dllUser32.NewProc("PostMessageW")
	procSetWindowText        = dllUser32.NewProc("SetWindowTextW")
	procGetWindowText        = dllUser32.NewProc("GetWindowTextW")
	procGetWindowTextLength  = dllUser32.NewProc("GetWindowTextLengthW")
	procMoveWindow           = dllUser32.NewProc("MoveWindow")
	procInvalidateRect       = dllUser32.NewProc("InvalidateRect")
	procGetClientRect        = dllUser32.NewProc("GetClientRect")
	procBeginPaint           = dllUser32.NewProc("BeginPaint")
	procEndPaint             = dllUser32.NewProc("EndPaint")
	procFillRect             = dllUser32.NewProc("FillRect")
	procDrawText             = dllUser32.NewProc("DrawTextW")
	procSetBkMode            = dllGdi32.NewProc("SetBkMode")
	procSetTextColor         = dllGdi32.NewProc("SetTextColor")
	procCreateSolidBrush     = dllGdi32.NewProc("CreateSolidBrush")
	procCreateRoundRectRgn   = dllGdi32.NewProc("CreateRoundRectRgn")
	procCreateFont           = dllGdi32.NewProc("CreateFontW")
	procSelectObject         = dllGdi32.NewProc("SelectObject")
	procDeleteObject         = dllGdi32.NewProc("DeleteObject")
	procSetWindowRgn         = dllUser32.NewProc("SetWindowRgn")
	procGetDpiForWindow      = dllUser32.NewProc("GetDpiForWindow")
	procLoadImage            = dllUser32.NewProc("LoadImageW")
	procLoadCursor           = dllUser32.NewProc("LoadCursorW")
	procGetModuleHandle      = dllKernel32.NewProc("GetModuleHandleW")
	procOpenClipboard        = dllUser32.NewProc("OpenClipboard")
	procEmptyClipboard       = dllUser32.NewProc("EmptyClipboard")
	procSetClipboardData     = dllUser32.NewProc("SetClipboardData")
	procCloseClipboard       = dllUser32.NewProc("CloseClipboard")
	procGlobalAlloc          = dllKernel32.NewProc("GlobalAlloc")
	procGlobalLock           = dllKernel32.NewProc("GlobalLock")
	procGlobalUnlock         = dllKernel32.NewProc("GlobalUnlock")
	procGlobalFree           = dllKernel32.NewProc("GlobalFree")
	procGetSaveFileName      = dllComdlg32.NewProc("GetSaveFileNameW")
	procMessageBox           = dllUser32.NewProc("MessageBoxW")
	procSetProcessDPIAware   = dllUser32.NewProc("SetProcessDPIAware")
	procSystemParametersInfo = dllUser32.NewProc("SystemParametersInfoW")
	procInitCommonControls   = dllComctl32.NewProc("InitCommonControlsEx")
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type paintstruct struct {
	Hdc         uintptr
	Erase       int32
	RcPaint     rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}
type msg struct {
	Hwnd           uintptr
	Message        uint32
	_              uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	LPrivate       uint32
	_pad           uint32
}
type wndclassex struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSmall                          uintptr
}
type systemtime struct{ Year, Month, DayOfWeek, Day, Hour, Minute, Second, Milliseconds uint16 }
type minmaxinfo struct {
	Reserved, MaxSize, MaxPosition, MinTrackSize, MaxTrackSize point
}
type initcommon struct{ Size, DwICC uint32 }
type openfilename struct {
	Size                         uint32
	_pad                         uint32
	Owner, Instance              uintptr
	Filter, CustomFilter         *uint16
	MaxCustomFilter, FilterIndex uint32
	File                         *uint16
	MaxFile                      uint32
	FileTitle                    *uint16
	MaxFileTitle                 uint32
	InitialDir, Title            *uint16
	Flags                        uint32
	FileOffset, FileExtension    uint16
	DefaultExt                   *uint16
	CustData, Hook, TemplateName uintptr
	Reserved                     uintptr
	ReservedDWORD, FlagsEx       uint32
}

func u16(s string) *uint16 { return syscall.StringToUTF16Ptr(s) }
func send(h uintptr, m uint32, w, l uintptr) uintptr {
	r, _, _ := procSendMessage.Call(h, uintptr(m), w, l)
	return r
}
func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }
func scale(n int) int32        { return int32(n) * int32(app.dpi) / 96 }
func move(h uintptr, x, y, w, ht int) {
	procMoveWindow.Call(h, uintptr(scale(x)), uintptr(scale(y)), uintptr(scale(w)), uintptr(scale(ht)), 1)
}
func setText(h uintptr, s string) { procSetWindowText.Call(h, uintptr(unsafe.Pointer(u16(s)))) }
func getText(h uintptr) string {
	n, _, _ := procGetWindowTextLength.Call(h)
	if n > 2000000 {
		return ""
	}
	buf := make([]uint16, n+2)
	got, _, _ := procGetWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:got])
}
func selected(h uintptr) int       { return int(send(h, CB_GETCURSEL, 0, 0)) }
func setSelected(h uintptr, n int) { send(h, CB_SETCURSEL, uintptr(n), 0) }
func checked(h uintptr) bool       { return send(h, BM_GETCHECK, 0, 0) == BST_CHECKED }
func setChecked(h uintptr, b bool) {
	n := uintptr(0)
	if b {
		n = BST_CHECKED
	}
	send(h, BM_SETCHECK, n, 0)
}
func addCombo(h uintptr, values ...string) {
	for _, v := range values {
		send(h, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(u16(v))))
	}
}
func info(msg string) {
	procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(u16(msg))), uintptr(unsafe.Pointer(u16("TimeStamps"))), MB_OK|MB_ICONINFORMATION)
}
func warn(msg string) {
	procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(u16(msg))), uintptr(unsafe.Pointer(u16("TimeStamps - input error"))), MB_OK|MB_ICONERROR)
}

type UI struct {
	hwnd, instance, whiteBrush, lightBrush, headerFont, titleFont, bodyFont, semiFont, monoFont uintptr
	dpi                                                                                         int
	width, height                                                                               int
	current                                                                                     Settings
	last                                                                                        *time.Time
	controls                                                                                    map[int]uintptr
	labels                                                                                      map[string]uintptr
	status                                                                                      string
	dirty                                                                                       bool
	outputCount                                                                                 int
	updating                                                                                    bool
	page                                                                                        int
	pageReady, switching, navigationQueued, presetQueued                                        bool
	pendingPage                                                                                 int
}

var app UI

func (ui *UI) c(id int) uintptr { return ui.controls[id] }
func control(class, title string, id int, style, exStyle uint32) uintptr {
	h, _, _ := procCreateWindowEx.Call(uintptr(exStyle), uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(title))), uintptr(style|WS_CHILD|WS_VISIBLE), 0, 0, 1, 1, app.hwnd, uintptr(id), app.instance, 0)
	if h == 0 {
		panic(fmt.Sprintf("CreateWindowEx %s (%d) failed", class, id))
	}
	if id > 0 {
		app.controls[id] = h
	}
	send(h, WM_SETFONT, app.bodyFont, 1)
	return h
}
func label(k, s string, bold bool) uintptr {
	h := control("STATIC", s, 0, 0, 0)
	if bold {
		send(h, WM_SETFONT, app.semiFont, 1)
	}
	app.labels[k] = h
	return h
}
func edit(id int, txt string, style uint32) uintptr {
	h := control("EDIT", txt, id, WS_TABSTOP|WS_BORDER|ES_AUTOHSCROLL|style, 0)
	return h
}
func btn(id int, txt string) uintptr { return control("BUTTON", txt, id, WS_TABSTOP, 0) }
func check(id int, txt string) uintptr {
	return control("BUTTON", txt, id, WS_TABSTOP|BS_AUTOCHECKBOX, 0)
}
func comb(id int) uintptr {
	return control("COMBOBOX", "", id, WS_TABSTOP|WS_VSCROLL|CBS_DROPDOWNLIST, 0)
}
func newFont(points, weight int, face string) uintptr {
	height := -int32(points) * int32(app.dpi) / 72
	r, _, _ := procCreateFont.Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(u16(face))))
	return r
}
func setDPI() {
	app.dpi = 96
	if err := procGetDpiForWindow.Find(); err == nil {
		if d, _, _ := procGetDpiForWindow.Call(app.hwnd); d >= 96 && d <= 384 {
			app.dpi = int(d)
		}
	}
}
func createUI() {
	app.controls = make(map[int]uintptr)
	app.labels = make(map[string]uintptr)
	setDPI()
	app.bodyFont = newFont(13, 400, "Segoe UI")
	app.semiFont = newFont(13, 600, "Segoe UI")
	app.titleFont = newFont(23, 700, "Segoe UI")
	app.headerFont = newFont(11, 400, "Segoe UI")
	// Fixed 16-point output font; avoid recreating/deleting active GDI fonts at runtime.
	app.monoFont = newFont(16, 400, "Consolas")
	app.whiteBrush, _, _ = procCreateSolidBrush.Call(rgb(255, 255, 255))
	app.lightBrush, _, _ = procCreateSolidBrush.Call(rgb(225, 236, 245))
	btn(ID_PAGE_SETUP, "SETUP")
	btn(ID_PAGE_FORMAT, "FORMAT")
	btn(ID_PAGE_RESULTS, "RESULTS")
	// Main settings card
	label("rules", "TIMESTAMP SETTINGS", true)
	label("hoursTitle", "POSTING HOURS & OPTIONS", true)
	date := control("SysDateTimePick32", "", ID_DATE, WS_TABSTOP|WS_BORDER, 0)
	send(date, DTM_SETFORMATW, 0, uintptr(unsafe.Pointer(u16("yyyy-MM-dd"))))
	label("date", "Start date", false)
	label("qty", "Timestamps", false)
	label("min", "Minimum gap (minutes)", false)
	label("max", "Maximum gap (minutes)", false)
	edit(ID_QTY, strconv.Itoa(app.current.Quantity), ES_NUMBER)
	edit(ID_MIN, strconv.Itoa(app.current.MinGapMinutes), ES_NUMBER)
	edit(ID_MAX, strconv.Itoa(app.current.MaxGapMinutes), ES_NUMBER)
	check(ID_HOURS, "Restrict to posting hours")
	label("from", "From", false)
	label("to", "To", false)
	st := comb(ID_START)
	ed := comb(ID_END)
	for i := 0; i < 24; i++ {
		v := fmt.Sprintf("%02d:00", i)
		addCombo(st, v)
		addCombo(ed, v)
	}
	label("format", "OUTPUT FORMAT", true)
	label("timeOptions", "DATE & TIME DISPLAY", true)
	label("preset", "Format", false)
	pre := comb(ID_PRESET)
	addCombo(pre, "Custom / imported", "Classic (time/date)", "Forum (time/date)", "ISO 8601", "Date then time")
	label("template", "Template — {time}, {date} and other tokens", false)
	edit(ID_TEMPLATE, app.current.Template, 0)
	label("datefmt", "Date style", false)
	label("timefmt", "Time style", false)
	label("formatHelp", "Change formats without losing results.\r\nUse the Results tab to copy or save.", false)
	dc := comb(ID_DATELAYOUT)
	tc := comb(ID_TIMELAYOUT)
	addCombo(dc, "Custom / imported", "MM-DD-YY", "MM-DD-YYYY", "YYYY-MM-DD", "DD/MM/YYYY", "Day, Month DD YYYY")
	addCombo(tc, "Custom / imported", "24-hour with seconds", "24-hour, no seconds", "Military (HHMMSS)", "12-hour AM/PM")
	check(ID_CONTINUE, "Continue from last generated timestamp")
	check(ID_REPLACE, "Clear earlier results on Generate")
	btn(ID_GENERATE, "Generate timestamps")
	// Output card
	label("output", "GENERATED TIMESTAMPS", true)
	label("resultHint", "Output text: 16 pt (fixed for stability)", false)
	btn(ID_COPYSEL, "Copy selected")
	btn(ID_COPYALL, "Copy all")
	btn(ID_SAVE, "Save .txt")
	btn(ID_CLEAR, "Clear")
	output := control("EDIT", "", 0, WS_BORDER|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY|WS_TABSTOP, WS_EX_CLIENTEDGE)
	app.labels["resultEdit"] = output
	send(output, EM_LIMITTEXT, 4*1024*1024, 0)
	send(output, WM_SETFONT, app.monoFont, 1)
	label("status", "Ready — choose a date and generate a sequence.", false)
	btn(ID_ABOUT, "About")
	app.page = 0
	// Read values from previous configuration.
	setSelected(st, app.current.WindowStart)
	setSelected(ed, app.current.WindowEnd)
	setChecked(app.c(ID_HOURS), app.current.RestrictHours)
	setChecked(app.c(ID_CONTINUE), app.current.Continue)
	setChecked(app.c(ID_REPLACE), app.current.Replace)
	setSelected(pre, 0)
	setSelected(dc, dateFormatIndex(app.current.DateLayout))
	setSelected(tc, timeFormatIndex(app.current.TimeLayout))
	d, e := parseDate(app.current.StartDate)
	if e != nil {
		d = time.Date(2050, 1, 1, 0, 0, 0, 0, time.Local)
	}
	setDate(date, d)
	setHoursEnabled()
	app.status = "Ready — choose a date and generate a sequence."
	showPage(0)
	procInvalidateRect.Call(app.hwnd, 0, 1)
}

// Keep each page isolated so all controls fit even at 1360x768/125%.
// Switching occurs only after the initiating button command returns. This
// prevents focus, visibility and control notifications from nesting inside it.
func requestPage(page int) {
	if page < 0 || page > 2 || app.switching {
		return
	}
	app.pendingPage = page
	if app.navigationQueued {
		return
	}
	app.navigationQueued = true
	if ok, _, _ := procPostMessage.Call(app.hwnd, WM_APP_SWITCHPAGE, 0, 0); ok == 0 {
		// Keep the current page usable if the Windows message queue is unavailable.
		app.navigationQueued = false
	}
}

func setGroupVisible(i int, visible bool) {
	groups := [][]int{
		{ID_DATE, ID_QTY, ID_MIN, ID_MAX, ID_HOURS, ID_START, ID_END, ID_CONTINUE, ID_REPLACE},
		{ID_PRESET, ID_TEMPLATE, ID_DATELAYOUT, ID_TIMELAYOUT},
		{ID_COPYSEL, ID_COPYALL, ID_SAVE, ID_CLEAR},
	}
	labelGroups := [][]string{
		{"rules", "hoursTitle", "date", "qty", "min", "max", "from", "to"},
		{"format", "timeOptions", "preset", "template", "datefmt", "timefmt", "formatHelp"},
		{"output", "resultEdit", "resultHint"},
	}
	visibility := uintptr(SW_HIDE)
	if visible {
		visibility = SW_SHOW
	}
	for _, id := range groups[i] {
		if h := app.c(id); h != 0 {
			procShowWindow.Call(h, visibility)
		}
	}
	for _, key := range labelGroups[i] {
		if h := app.labels[key]; h != 0 {
			procShowWindow.Call(h, visibility)
		}
	}
}

func showPage(page int) {
	if page < 0 || page > 2 || app.switching {
		return
	}
	if app.pageReady && app.page == page {
		return
	}
	app.switching = true
	defer func() { app.switching = false }()
	if !app.pageReady {
		// Every child starts visible. Hide the other two pages once.
		for i := 0; i < 3; i++ {
			if i != page {
				setGroupVisible(i, false)
			}
		}
		app.pageReady = true
	} else {
		// Only touch controls that actually need changing, reducing Win32
		// visibility/focus notifications during each page transition.
		setGroupVisible(app.page, false)
		setGroupVisible(page, true)
	}
	app.page = page
	procInvalidateRect.Call(app.hwnd, 0, 1)
}

func dateFormatIndex(t string) int {
	switch t {
	case "01-02-06":
		return 1
	case "01-02-2006":
		return 2
	case "2006-01-02":
		return 3
	case "02/01/2006":
		return 4
	case "Mon, January 02 2006":
		return 5
	}
	return 0
}
func timeFormatIndex(t string) int {
	switch t {
	case "15:04:05":
		return 1
	case "15:04":
		return 2
	case "150405":
		return 3
	case "03:04:05 PM":
		return 4
	}
	return 0
}
func setDate(h uintptr, dt time.Time) {
	y, m, d := dt.Date()
	v := systemtime{Year: uint16(y), Month: uint16(m), Day: uint16(d)}
	send(h, DTM_SETSYSTEMTIME, GDT_VALID, uintptr(unsafe.Pointer(&v)))
}
func getDate(h uintptr) (string, error) {
	var st systemtime
	if send(h, DTM_GETSYSTEMTIME, 0, uintptr(unsafe.Pointer(&st))) != GDT_VALID {
		return "", fmt.Errorf("select a valid date")
	}
	return fmt.Sprintf("%04d-%02d-%02d", st.Year, st.Month, st.Day), nil
}
func setHoursEnabled() {
	enable := checked(app.c(ID_HOURS))
	p := dllUser32.NewProc("EnableWindow")
	n := uintptr(0)
	if enable {
		n = 1
	}
	p.Call(app.c(ID_START), n)
	p.Call(app.c(ID_END), n)
}
func syncSettings() (Settings, error) {
	s := app.current
	var e error
	s.StartDate, e = getDate(app.c(ID_DATE))
	if e != nil {
		return s, e
	}
	toint := func(id int, what string) (int, error) {
		n, err := strconv.Atoi(strings.TrimSpace(getText(app.c(id))))
		if err != nil {
			return 0, fmt.Errorf("%s must be a number", what)
		}
		return n, nil
	}
	if s.Quantity, e = toint(ID_QTY, "Quantity"); e != nil {
		return s, e
	}
	if s.MinGapMinutes, e = toint(ID_MIN, "Minimum gap"); e != nil {
		return s, e
	}
	if s.MaxGapMinutes, e = toint(ID_MAX, "Maximum gap"); e != nil {
		return s, e
	}
	s.RestrictHours = checked(app.c(ID_HOURS))
	s.WindowStart = selected(app.c(ID_START))
	s.WindowEnd = selected(app.c(ID_END))
	s.Continue = checked(app.c(ID_CONTINUE))
	s.Replace = checked(app.c(ID_REPLACE))
	s.Template = getText(app.c(ID_TEMPLATE))
	switch selected(app.c(ID_DATELAYOUT)) {
	case 1:
		s.DateLayout = "01-02-06"
	case 2:
		s.DateLayout = "01-02-2006"
	case 3:
		s.DateLayout = "2006-01-02"
	case 4:
		s.DateLayout = "02/01/2006"
	case 5:
		s.DateLayout = "Mon, January 02 2006"
	}
	switch selected(app.c(ID_TIMELAYOUT)) {
	case 1:
		s.TimeLayout = "15:04:05"
	case 2:
		s.TimeLayout = "15:04"
	case 3:
		s.TimeLayout = "150405"
	case 4:
		s.TimeLayout = "03:04:05 PM"
	}
	return s, s.validate()
}
func status(s string) { app.status = s; setText(app.labels["status"], s) }
func generate() {
	s, err := syncSettings()
	if err != nil {
		warn(err.Error())
		return
	}
	if s.StartDate != app.current.StartDate {
		app.last = nil
	}
	var prev *time.Time
	if s.Continue && app.last != nil {
		prev = app.last
	}
	generated, err := Generate(s, prev, nil)
	if err != nil {
		warn(err.Error())
		return
	}
	newText := FormatList(generated, s)
	output := app.labels["resultEdit"]
	if s.Replace && app.dirty && getText(output) != "" {
		r, _, _ := procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(u16("Replace the current unsaved timestamp list?"))), uintptr(unsafe.Pointer(u16("Replace timestamps"))), MB_YESNO|MB_ICONINFORMATION)
		if r != IDYES {
			return
		}
	}
	if s.Replace {
		setText(output, newText)
		app.outputCount = len(generated)
	} else {
		old := getText(output)
		if old != "" && !strings.HasSuffix(old, "\n") {
			old += "\r\n"
		}
		setText(output, old+newText)
		app.outputCount += len(generated)
	}
	app.last = &generated[len(generated)-1]
	app.dirty = true
	s.LastTimestamp = app.last.Format(time.RFC3339)
	app.current = s
	if err := SaveSettings(s); err != nil {
		status("Generated, but settings could not be saved: " + err.Error())
	} else {
		status(fmt.Sprintf("%d new timestamps  ·  %d in list  ·  last: %s", len(generated), app.outputCount, app.last.Format("2006-01-02 15:04:05")))
	}
	send(output, EM_SETSEL, uintptr(len([]rune(getText(output)))), uintptr(len([]rune(getText(output)))))
	send(output, EM_SCROLLCARET, 0, 0)
	showPage(2)
}
func copyAll() {
	text := getText(app.labels["resultEdit"])
	if text == "" {
		status("Nothing to copy.")
		return
	}
	if err := setClipboard(text); err != nil {
		warn("Clipboard failed: " + err.Error())
		return
	}
	status("Copied all timestamps to clipboard.")
}
func setClipboard(txt string) error {
	utf16 := syscall.StringToUTF16(txt)
	size := uintptr(len(utf16) * 2)
	if r, _, _ := procOpenClipboard.Call(app.hwnd); r == 0 {
		return fmt.Errorf("could not open clipboard")
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	h, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if h == 0 {
		return fmt.Errorf("memory allocation failed")
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("memory lock failed")
	}
	dest := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(utf16))
	copy(dest, utf16)
	procGlobalUnlock.Call(h)
	if r, _, _ := procSetClipboardData.Call(CF_UNICODETEXT, h); r == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("could not set clipboard data")
	}
	return nil
}
func saveFile() {
	output := getText(app.labels["resultEdit"])
	if output == "" {
		status("Generate timestamps before saving.")
		return
	}
	buf := make([]uint16, 2048)
	copy(buf, syscall.StringToUTF16("timestamps.txt"))
	filters := syscall.StringToUTF16("Text files (*.txt)\x00*.txt\x00All files (*.*)\x00*.*\x00\x00")
	ofn := openfilename{Size: uint32(unsafe.Sizeof(openfilename{})), Owner: app.hwnd, File: &buf[0], MaxFile: uint32(len(buf)), Filter: &filters[0], FilterIndex: 1, Title: u16("Save timestamp list"), DefaultExt: u16("txt"), Flags: OFN_OVERWRITEPROMPT | OFN_PATHMUSTEXIST}
	r, _, _ := procGetSaveFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return
	} // The user cancelled.
	filename := syscall.UTF16ToString(buf)
	// CRLF and UTF-8 text are readable by current Windows Notepad and text editors.
	if err := os.WriteFile(filename, []byte(output), 0644); err != nil {
		warn("Could not save file:\n" + err.Error())
		return
	}
	app.dirty = false
	status("Saved: " + filename)
}
func showAbout() {
	info("TimeStamps 2.5.1 — accessible Windows 10/11 edition\n\nUpdated by NightLife\n\nOriginal program: Scott GrantSmith (1997)\nOriginal requirements: Branson Hagerty\n\nRebuilt from the original Delphi 3 freeware source.\nFreeware. Not for sale.\n\nCreates sequential randomized timestamps for fictional documents, message logs and Shadowrun handouts.")
}
func handler(cmd int) {
	switch cmd {
	case ID_GENERATE:
		generate()
	case ID_PAGE_SETUP:
		requestPage(0)
	case ID_PAGE_FORMAT:
		requestPage(1)
	case ID_PAGE_RESULTS:
		requestPage(2)
	case ID_COPYSEL:
		send(app.labels["resultEdit"], WM_COPY, 0, 0)
		status("Selected text copied, if any text was selected.")
	case ID_COPYALL:
		copyAll()
	case ID_SAVE:
		saveFile()
	case ID_CLEAR:
		if app.dirty && getText(app.labels["resultEdit"]) != "" {
			r, _, _ := procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(u16("Discard the generated list?"))), uintptr(unsafe.Pointer(u16("Clear timestamps"))), MB_YESNO|MB_ICONINFORMATION)
			if r != IDYES {
				return
			}
		}
		setText(app.labels["resultEdit"], "")
		app.dirty = false
		app.outputCount = 0
		status("Output cleared. Continuation remains available.")
	case ID_ABOUT:
		showAbout()
	}
}

// Selecting a format preset must not mutate the template EDIT control inside
// the COMBOBOX selection-change notification. Defer those edits until the
// originating CBN_SELCHANGE message has finished processing.
func requestPresetUpdate() {
	if app.presetQueued || app.updating || app.switching {
		return
	}
	app.presetQueued = true
	if ok, _, _ := procPostMessage.Call(app.hwnd, WM_APP_APPLYPRESET, 0, 0); ok == 0 {
		app.presetQueued = false
	}
}

func applyPreset() {
	if app.switching || app.updating || !app.pageReady || app.page != 1 {
		return
	}
	app.updating = true
	defer func() { app.updating = false }()
	preset, ok := presetByIndex(selected(app.c(ID_PRESET)))
	if !ok {
		return
	}
	setText(app.c(ID_TEMPLATE), preset.Template)
	if preset.DateLayout != "" {
		setSelected(app.c(ID_DATELAYOUT), dateFormatIndex(preset.DateLayout))
	}
	if preset.TimeLayout != "" {
		setSelected(app.c(ID_TIMELAYOUT), timeFormatIndex(preset.TimeLayout))
	}
}

func onCommand(w uintptr) {
	if app.switching || app.updating {
		return
	}
	id := int(w & 0xFFFF)
	code := int((w >> 16) & 0xFFFF)
	switch id {
	case ID_HOURS:
		if code == 0 {
			setHoursEnabled()
		}
	case ID_PRESET, ID_DATELAYOUT, ID_TIMELAYOUT, ID_START, ID_END:
		if id == ID_PRESET && code == CBN_SELCHANGE {
			requestPresetUpdate()
		}
	case ID_TEMPLATE:
		if code == EN_CHANGE && !app.updating {
			setSelected(app.c(ID_PRESET), 0)
		}
	default:
		if code == 0 {
			handler(id)
		}
	}
}
func onClose() {
	if app.dirty && getText(app.labels["resultEdit"]) != "" {
		r, _, _ := procMessageBox.Call(app.hwnd, uintptr(unsafe.Pointer(u16("Unsaved timestamps will be lost. Exit anyway?"))), uintptr(unsafe.Pointer(u16("TimeStamps — unsaved output"))), MB_YESNO|MB_ICONINFORMATION)
		if r != IDYES {
			return
		}
	}
	if s, err := syncSettings(); err == nil {
		if s.StartDate != app.current.StartDate { // A manually changed date starts a new sequence.
			s.LastTimestamp = ""
		} else if app.last != nil {
			s.LastTimestamp = app.last.Format(time.RFC3339)
		}
		_ = SaveSettings(s)
	}
	dllUser32.NewProc("DestroyWindow").Call(app.hwnd)
}
func cleanup() {
	for _, h := range []uintptr{app.bodyFont, app.semiFont, app.titleFont, app.headerFont, app.monoFont, app.whiteBrush, app.lightBrush} {
		if h != 0 {
			procDeleteObject.Call(h)
		}
	}
}
func layout(width, height int) {
	if len(app.labels) == 0 {
		return
	}
	app.width, app.height = width, height
	g := makeLayout(width, height)
	for id, b := range g.Controls {
		move(app.c(id), b.X, b.Y, b.W, b.H)
	}
	for key, b := range g.Labels {
		move(app.labels[key], b.X, b.Y, b.W, b.H)
	}
}

func drawRect(hdc uintptr, r rect, col uintptr) {
	b, _, _ := procCreateSolidBrush.Call(col)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	procDeleteObject.Call(b)
}
func roundPanel(hdc uintptr, r rect, color uintptr) {
	// A region plus a solid fill avoids legacy bevels and is compatible with GDI.
	rg, _, _ := procCreateRoundRectRgn.Call(uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom), uintptr(scale(18)), uintptr(scale(18)))
	brush, _, _ := procCreateSolidBrush.Call(color)
	procFillRgn := dllGdi32.NewProc("FillRgn")
	procFillRgn.Call(hdc, rg, brush)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(rg)
}
func drawText(hdc uintptr, txt string, r rect, font uintptr, color uintptr) {
	old, _, _ := procSelectObject.Call(hdc, font)
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, color)
	procDrawText.Call(hdc, uintptr(unsafe.Pointer(u16(txt))), uintptr(^uint32(0)), uintptr(unsafe.Pointer(&r)), DT_SINGLELINE|DT_LEFT|DT_VCENTER)
	procSelectObject.Call(hdc, old)
}
func paint(h uintptr) {
	var ps paintstruct
	hdc, _, _ := procBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer procEndPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	var client rect
	procGetClientRect.Call(h, uintptr(unsafe.Pointer(&client)))
	W, H := client.Right, client.Bottom
	drawRect(hdc, rect{0, 0, W, H}, rgb(240, 245, 250))
	drawRect(hdc, rect{0, 0, W, scale(77)}, rgb(25, 38, 57))
	drawRect(hdc, rect{0, scale(74), W, scale(77)}, rgb(43, 169, 157))
	titleRect := rect{scale(34), scale(4), W - scale(160), scale(45)}
	drawText(hdc, "TIMESTAMPS", titleRect, app.titleFont, rgb(255, 255, 255))
	subRect := rect{scale(35), scale(40), W - scale(180), scale(67)}
	drawText(hdc, "Updated by NightLife  •  Windows 10/11", subRect, app.headerFont, rgb(222, 235, 245))
	panelBottom := H - scale(72)
	if panelBottom > scale(139) {
		roundPanel(hdc, rect{scale(17), scale(137), W - scale(17), panelBottom}, rgb(255, 255, 255))
	}
	// Persistent footer with the Generate action; never hidden by page switching.
	drawRect(hdc, rect{0, H - scale(63), W, H}, rgb(225, 236, 245))
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CREATE:
		app.hwnd = hwnd
		createUI()
		return 0
	case WM_GETMINMAXINFO:
		m := (*minmaxinfo)(unsafe.Pointer(lParam))
		m.MinTrackSize = point{X: scale(830), Y: scale(530)}
		return 0
	case WM_SIZE:
		wi := int(uint16(lParam & 0xFFFF))
		he := int(uint16((lParam >> 16) & 0xFFFF))
		if app.dpi <= 0 {
			app.dpi = 96
		}
		layout(wi*96/app.dpi, he*96/app.dpi)
		procInvalidateRect.Call(hwnd, 0, 1)
		return 0
	case WM_COMMAND:
		onCommand(wParam)
		return 0
	case WM_APP_SWITCHPAGE:
		if app.navigationQueued {
			app.navigationQueued = false
			showPage(app.pendingPage)
		}
		return 0
	case WM_APP_APPLYPRESET:
		if app.presetQueued {
			app.presetQueued = false
			applyPreset()
		}
		return 0
	case WM_CTLCOLOREDIT:
		procSetTextColor.Call(wParam, rgb(25, 38, 57))
		return app.whiteBrush
	case WM_CTLCOLORSTATIC:
		procSetBkMode.Call(wParam, TRANSPARENT)
		procSetTextColor.Call(wParam, rgb(35, 50, 69))
		if lParam == app.labels["status"] {
			return app.lightBrush
		}
		return app.whiteBrush
	case WM_ERASEBKGND:
		return 1
	case WM_PAINT:
		paint(hwnd)
		return 0
	case WM_CLOSE:
		onClose()
		return 0
	case WM_DESTROY:
		cleanup()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func main() {
	runtime.LockOSThread() // Win32 HWNDs and message pumps must use the same OS thread.
	procSetProcessDPIAware.Call()
	icc := initcommon{Size: uint32(unsafe.Sizeof(initcommon{})), DwICC: 0x00000100 | 0x00000004}
	procInitCommonControls.Call(uintptr(unsafe.Pointer(&icc)))
	exe, _ := os.Executable()
	legacy := filepath.Join(filepath.Dir(exe), "TimeStamps.ini")
	app.current = LoadSettings(legacy)
	if app.current.LastTimestamp != "" {
		if last, err := time.Parse(time.RFC3339, app.current.LastTimestamp); err == nil {
			local := last.In(time.Local)
			app.last = &local
		}
	}
	app.dpi = 96
	module, _, _ := procGetModuleHandle.Call(0)
	app.instance = module
	class := u16("NightLifeTimeStampsV2")
	wc := wndclassex{Size: uint32(unsafe.Sizeof(wndclassex{})), Style: 3, WndProc: syscall.NewCallback(wndProc), Instance: module, ClassName: class}
	cursor, _, _ := procLoadCursor.Call(0, 32512)
	wc.Cursor = cursor
	if r, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic(fmt.Sprintf("register window class: %s", err))
	}
	// Fit the screen work area, including taskbar, at the current Windows scale.
	var work rect
	if ok, _, _ := procSystemParametersInfo.Call(48, 0, uintptr(unsafe.Pointer(&work)), 0); ok == 0 || work.Right <= work.Left || work.Bottom <= work.Top {
		work = rect{0, 0, 1280, 720}
	}
	showStartupSplash(module, work) // First display, then create the main window.
	workW := int(work.Right - work.Left)
	workH := int(work.Bottom - work.Top)
	winW, winH := 1300, 690
	if workW-28 < winW {
		winW = workW - 28
	}
	if workH-28 < winH {
		winH = workH - 28
	}
	if winW < 640 {
		winW = 640
	}
	if winH < 480 {
		winH = 480
	}
	x0 := int(work.Left) + (workW-winW)/2
	y0 := int(work.Top) + (workH-winH)/2
	h, _, err := procCreateWindowEx.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(u16("TimeStamps — Updated by NightLife"))), WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN, uintptr(x0), uintptr(y0), uintptr(winW), uintptr(winH), 0, 0, module, 0)
	if h == 0 {
		panic(fmt.Sprintf("create application window: %s", err))
	}
	// Icon file is optional; no additional system component is required.
	iconPath := filepath.Join(filepath.Dir(exe), "TimeStamps.ico")
	if _, err := os.Stat(iconPath); err == nil {
		icon, _, _ := procLoadImage.Call(0, uintptr(unsafe.Pointer(u16(iconPath))), IMAGE_ICON, 48, 48, LR_LOADFROMFILE)
		if icon != 0 {
			send(h, WM_SETICON, ICON_BIG, icon)
			send(h, WM_SETICON, ICON_SMALL, icon)
		}
	}
	procShowWindow.Call(h, SW_SHOW)
	procUpdateWindow.Call(h)
	var m msg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}
