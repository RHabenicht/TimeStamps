//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

const (
	WS_POPUP         = 0x80000000
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_TOPMOST    = 0x00000008
	PM_REMOVE        = 0x0001
	DIB_RGB_COLORS   = 0
	SRCCOPY          = 0x00CC0020
	DI_NORMAL        = 0x0003
)

var (
	procPeekMessage   = dllUser32.NewProc("PeekMessageW")
	procDestroyWindow = dllUser32.NewProc("DestroyWindow")
	procStretchDIBits = dllGdi32.NewProc("StretchDIBits")
	procSetWindowPos = dllUser32.NewProc("SetWindowPos")
	procSetForegroundWindow = dllUser32.NewProc("SetForegroundWindow")
	procBringWindowToTop = dllUser32.NewProc("BringWindowToTop")
	procInvalidateSplash = dllUser32.NewProc("InvalidateRect")
	procIsWindowVisible = dllUser32.NewProc("IsWindowVisible")
	procDrawIconEx     = dllUser32.NewProc("DrawIconEx")
	procDestroyIcon    = dllUser32.NewProc("DestroyIcon")
	startupSplash     splashBitmap
	splashInfo        dibHeader
	startupSplashPainted bool
	startupSplashDrawn bool
	startupLogoIcon uintptr
	startupLogoDrawn bool
)

type dibHeader struct {
	Size                             uint32
	Width, Height                    int32
	Planes, BitCount                 uint16
	Compression, SizeImage           uint32
	XpixelsPerMeter, YpixelsPerMeter int32
	ColorsUsed, ColorsImportant      uint32
}

// Diagnostics are only written when the splash cannot be created.
// This provides a concrete error report in case a specific Windows
// configuration does not support the popup, rather than silently skipping it.
func recordSplashFailure(step string, err error) {
	dir, e := os.UserConfigDir()
	if e != nil {
		return
	}
	path := filepath.Join(dir, "TimeStamps")
	if os.MkdirAll(path, 0700) != nil {
		return
	}
	content := fmt.Sprintf("%s: %s: %v\n", time.Now().Format(time.RFC3339), step, err)
	f, e := os.OpenFile(filepath.Join(path, "splash_errors.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e == nil {
		_, _ = f.WriteString(content)
		_ = f.Close()
	}
}

// Store the latest splash status to diagnose machines that fail to show it.
func writeSplashStatus(stage string) {
	dir, e := os.UserConfigDir()
	if e != nil { return }
	path := filepath.Join(dir, "TimeStamps")
	if os.MkdirAll(path, 0700) != nil { return }
	_ = os.WriteFile(filepath.Join(path, "splash_status.txt"),
		[]byte(fmt.Sprintf("%s\n%s\n", time.Now().Format(time.RFC3339), stage)), 0600)
}

// The PE contains RT_GROUP_ICON ID 1. The installed .ico is the fallback
// for machines with stricter executable-resource loading.
func loadStartupLogo(instance uintptr) uintptr {
	icon, _, _ := procLoadImage.Call(instance, uintptr(1), IMAGE_ICON, 160, 160, 0)
	if icon != 0 {
		return icon
	}
	exe, err := os.Executable()
	if err != nil {
		recordSplashFailure("LoadStartupLogo", err)
		return 0
	}
	ico := filepath.Join(filepath.Dir(exe), "TimeStamps.ico")
	icon, _, _ = procLoadImage.Call(0, uintptr(unsafe.Pointer(u16(ico))), IMAGE_ICON, 160, 160, LR_LOADFROMFILE)
	if icon == 0 {
		recordSplashFailure("LoadStartupLogo", fmt.Errorf("embedded icon 1 and %q unavailable", ico))
	}
	return icon
}

// The art is drawn first. The explicit logo is a separate paint operation,
// so a background image without its logo cannot silently pass validation.
func drawStartupLogo(hdc uintptr, client rect) {
	if startupLogoIcon == 0 {
		return
	}
	width := int(client.Right - client.Left)
	height := int(client.Bottom - client.Top)
	size := height / 3
	if size > 144 { size = 144 }
	if size < 48 { size = 48 }
	margin := size / 6
	if margin < 14 { margin = 14 }
	x := width - size - 2*margin
	y := (height - size) / 2
	if x < margin { x = margin }
	if y < margin { y = margin }
	backing := rect{Left:int32(x-margin/2), Top:int32(y-margin/2), Right:int32(x+size+margin/2), Bottom:int32(y+size+margin/2)}
	b, _, _ := procCreateSolidBrush.Call(rgb(23, 8, 37))
	if b != 0 {
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&backing)), b)
		procDeleteObject.Call(b)
	}
	ok, _, drawErr := procDrawIconEx.Call(hdc, uintptr(x), uintptr(y), startupLogoIcon, uintptr(size), uintptr(size), 0, 0, DI_NORMAL)
	if ok == 0 {
		recordSplashFailure("DrawIconEx", drawErr)
	} else {
		startupLogoDrawn = true
	}
}

func startupSplashWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_ERASEBKGND:
		return 1
	case WM_PAINT:
		var ps paintstruct
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		if hdc != 0 {
			var rc rect
			procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
			if len(startupSplash.BGRA) > 0 {
				n, _, drawErr := procStretchDIBits.Call(hdc, 0, 0, uintptr(rc.Right), uintptr(rc.Bottom),
					0, 0, uintptr(startupSplash.Width), uintptr(startupSplash.Height),
					uintptr(unsafe.Pointer(&startupSplash.BGRA[0])), uintptr(unsafe.Pointer(&splashInfo)), DIB_RGB_COLORS, SRCCOPY)
				if n == 0 || int32(n) == -1 {
					recordSplashFailure("StretchDIBits", fmt.Errorf("scanlines=%d: %v", int32(n), drawErr))
					// Display a readable message if the graphics driver rejects the DIB.
					b, _, _ := procCreateSolidBrush.Call(rgb(28, 8, 50))
					procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), b)
					procDeleteObject.Call(b)
					procSetBkMode.Call(hdc, TRANSPARENT)
					procSetTextColor.Call(hdc, rgb(255, 255, 255))
					caption := u16("TimeStamps — Updated by NightLife")
					procDrawText.Call(hdc, uintptr(unsafe.Pointer(caption)), uintptr(^uint32(0)),
						uintptr(unsafe.Pointer(&rc)), DT_SINGLELINE|DT_VCENTER)
				} else {
					startupSplashDrawn = true
				}
			}
			drawStartupLogo(hdc, rc)
			startupSplashPainted = true
			procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		}
		return 0
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	}
	r, _, _ := procDefWindowProc.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

// showStartupSplash displays the splash BEFORE the main window is created.
// The 6-second countdown starts after ShowWindow/UpdateWindow.
// It pumps Win32 messages on the window's own (locked) OS thread to ensure
// painting and interactions remain responsive. The main window is constructed
// only after the splash has closed, so it cannot cover the splash.
func showStartupSplash(instance uintptr, work rect) bool {
	// The caller locks the current OS thread before any HWND is created.
	writeSplashStatus("Preparing TimeStamps 2.5 startup splash")
	bmp, err := decodeSplash(splashPNG)
	if err != nil {
		recordSplashFailure("decodeSplash", err)
		return false
	}
	w, h := fitSplash(int(work.Right-work.Left), int(work.Bottom-work.Top), bmp.Width, bmp.Height)
	if w == 0 || h == 0 {
		recordSplashFailure("fitSplash", fmt.Errorf("invalid work area %v", work))
		return false
	}
	name := u16("NightLifeTimeStampsSplashV25")
	wc := wndclassex{Size: uint32(unsafe.Sizeof(wndclassex{})), WndProc: syscall.NewCallback(startupSplashWndProc), Instance: instance, ClassName: name}
	cursor, _, _ := procLoadCursor.Call(0, 32512)
	wc.Cursor = cursor
	if ok, _, e := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); ok == 0 {
		recordSplashFailure("RegisterClassExW", e)
		return false
	}
	startupSplash = bmp
	startupLogoIcon = loadStartupLogo(instance)
	defer func() {
		if startupLogoIcon != 0 {
			procDestroyIcon.Call(startupLogoIcon)
			startupLogoIcon = 0
		}
	}()
	splashInfo = dibHeader{Size: uint32(unsafe.Sizeof(dibHeader{})), Width: int32(bmp.Width), Height: -int32(bmp.Height),
		Planes: 1, BitCount: 32, Compression: 0, SizeImage: uint32(len(bmp.BGRA))}
	x := int(work.Left) + (int(work.Right-work.Left)-w)/2
	y := int(work.Top) + (int(work.Bottom-work.Top)-h)/2
	startupSplashPainted = false
	startupSplashDrawn = false
	startupLogoDrawn = false
	sh, _, e := procCreateWindowEx.Call(WS_EX_TOPMOST|WS_EX_TOOLWINDOW, uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(u16("TimeStamps — Updated by NightLife"))), WS_POPUP,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0, 0, instance, 0)
	if sh == 0 {
		recordSplashFailure("CreateWindowExW", e)
		return false
	}
	procShowWindow.Call(sh, SW_SHOW)
	// Topmost does not necessarily imply foreground: request both.
	if ok, _, posErr := procSetWindowPos.Call(sh, ^uintptr(0), 0, 0, 0, 0, 0x0001|0x0002|0x0040); ok == 0 {
		recordSplashFailure("SetWindowPos", posErr)
	}
	procBringWindowToTop.Call(sh)
	if ok, _, _ := procSetForegroundWindow.Call(sh); ok == 0 {
		recordSplashFailure("SetForegroundWindow", fmt.Errorf("foreground request was rejected"))
	}
	procInvalidateSplash.Call(sh, 0, 0)
	procUpdateWindow.Call(sh)
	if !startupSplashPainted {
		recordSplashFailure("UpdateWindow", fmt.Errorf("WM_PAINT not delivered synchronously"))
	}
	visible, _, _ := procIsWindowVisible.Call(sh)
	writeSplashStatus(fmt.Sprintf("Created at %d,%d size %dx%d; visible=%t; painted=%t; imageDrawn=%t; logoLoaded=%t; logoDrawn=%t",
		x, y, w, h, visible != 0, startupSplashPainted, startupSplashDrawn, startupLogoIcon != 0, startupLogoDrawn))
	started := time.Now()
	for time.Since(started) < time.Duration(splashDurationMs)*time.Millisecond {
		// Drain queued messages so the splash actually repaints on Windows.
		var m msg
		for {
			ok, _, _ := procPeekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, PM_REMOVE)
			if ok == 0 {
				break
			}
			if m.Message == 0x0012 { // WM_QUIT: do not swallow application exit requests
				procDestroyWindow.Call(sh)
				procPostQuitMessage.Call(m.WParam)
				return true
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
		}
		time.Sleep(12 * time.Millisecond)
	}
	procDestroyWindow.Call(sh)
	writeSplashStatus(fmt.Sprintf("Completed six-second startup splash; visible=%t; WM_PAINT=%t; artworkDrawn=%t; logoLoaded=%t; logoDrawn=%t", visible != 0, startupSplashPainted, startupSplashDrawn, startupLogoIcon != 0, startupLogoDrawn))
	return true
}
