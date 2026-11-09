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
)

var (
	procPeekMessage   = dllUser32.NewProc("PeekMessageW")
	procDestroyWindow = dllUser32.NewProc("DestroyWindow")
	procStretchDIBits = dllGdi32.NewProc("StretchDIBits")
	startupSplash     splashBitmap
	splashInfo        dibHeader
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
				procStretchDIBits.Call(hdc, 0, 0, uintptr(rc.Right), uintptr(rc.Bottom),
					0, 0, uintptr(startupSplash.Width), uintptr(startupSplash.Height),
					uintptr(unsafe.Pointer(&startupSplash.BGRA[0])), uintptr(unsafe.Pointer(&splashInfo)), DIB_RGB_COLORS, SRCCOPY)
			}
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
// The 4-second countdown starts after ShowWindow/UpdateWindow.
// It pumps Win32 messages on the window's own (locked) OS thread to ensure
// painting and interactions remain responsive. The main window is constructed
// only after the splash has closed, so it cannot cover the splash.
func showStartupSplash(instance uintptr, work rect) bool {
	// The caller locks the current OS thread before any HWND is created.
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
	splashInfo = dibHeader{Size: uint32(unsafe.Sizeof(dibHeader{})), Width: int32(bmp.Width), Height: -int32(bmp.Height),
		Planes: 1, BitCount: 32, Compression: 0, SizeImage: uint32(len(bmp.BGRA))}
	x := int(work.Left) + (int(work.Right-work.Left)-w)/2
	y := int(work.Top) + (int(work.Bottom-work.Top)-h)/2
	sh, _, e := procCreateWindowEx.Call(WS_EX_TOPMOST|WS_EX_TOOLWINDOW, uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(u16("TimeStamps — Updated by NightLife"))), WS_POPUP,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0, 0, instance, 0)
	if sh == 0 {
		recordSplashFailure("CreateWindowExW", e)
		return false
	}
	procShowWindow.Call(sh, SW_SHOW)
	procUpdateWindow.Call(sh)
	started := time.Now()
	for time.Since(started) < 4*time.Second {
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
	return true
}
