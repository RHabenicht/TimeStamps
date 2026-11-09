//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	procPeekMessage                          = dllUser32.NewProc("PeekMessageW")
	procDestroyWindow                        = dllUser32.NewProc("DestroyWindow")
	procStretchDIBits                        = dllGdi32.NewProc("StretchDIBits")
	procSetDIBitsToDevice                    = dllGdi32.NewProc("SetDIBitsToDevice")
	procCreateCompatibleDC                   = dllGdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection                     = dllGdi32.NewProc("CreateDIBSection")
	procDeleteDC                             = dllGdi32.NewProc("DeleteDC")
	procBitBlt                               = dllGdi32.NewProc("BitBlt")
	splashStretchCall       splashRasterCall = procStretchDIBits.Call
	splashDirectCopyCall    splashRasterCall = procSetDIBitsToDevice.Call
	procSetWindowPos                         = dllUser32.NewProc("SetWindowPos")
	procSetForegroundWindow                  = dllUser32.NewProc("SetForegroundWindow")
	procBringWindowToTop                     = dllUser32.NewProc("BringWindowToTop")
	procInvalidateSplash                     = dllUser32.NewProc("InvalidateRect")
	procIsWindowVisible                      = dllUser32.NewProc("IsWindowVisible")
	startupSplash           splashBitmap
	startupSplashScaled     splashBitmap
	startupSplashRenderer   string
	startupSplashPainted    bool
	startupSplashDrawn      bool
)

type dibHeader struct {
	Size                             uint32
	Width, Height                    int32
	Planes, BitCount                 uint16
	Compression, SizeImage           uint32
	XpixelsPerMeter, YpixelsPerMeter int32
	ColorsUsed, ColorsImportant      uint32
}

type splashRasterCall func(...uintptr) (uintptr, uintptr, error)

func splashHeader(bmp splashBitmap) dibHeader {
	return dibHeader{Size: uint32(unsafe.Sizeof(dibHeader{})), Width: int32(bmp.Width), Height: -int32(bmp.Height),
		Planes: 1, BitCount: 32, SizeImage: uint32(len(bmp.BGRA))}
}

// Try native scaling first. If a driver rejects it, copy software-scaled pixels
// at 1:1 size using a separate GDI operation. If both DIB transfer APIs fail,
// copy through a software-filled Windows bitmap instead. Tests inject both
// failures and verify the actual pixels painted by the production popup.
func drawSplashArtwork(hdc uintptr, source, scaled splashBitmap, stretch, directCopy splashRasterCall) (string, error) {
	if hdc == 0 || len(source.BGRA) == 0 || len(scaled.BGRA) == 0 {
		return "none", fmt.Errorf("invalid splash drawing surface or pixels")
	}
	header := splashHeader(source)
	// The injectable native call takes uintptr arguments. Pin its buffers and
	// header explicitly so stack growth or GC cannot invalidate their addresses.
	var pinned runtime.Pinner
	pinned.Pin(&header)
	pinned.Pin(&source.BGRA[0])
	pinned.Pin(&scaled.BGRA[0])
	defer pinned.Unpin()
	n, _, stretchErr := stretch(hdc, 0, 0, uintptr(scaled.Width), uintptr(scaled.Height),
		0, 0, uintptr(source.Width), uintptr(source.Height), uintptr(unsafe.Pointer(&source.BGRA[0])),
		uintptr(unsafe.Pointer(&header)), DIB_RGB_COLORS, SRCCOPY)
	if n != 0 && int32(n) != -1 {
		return "StretchDIBits", nil
	}
	recordSplashFailure("StretchDIBits; trying software-scaled artwork", fmt.Errorf("scanlines=%d: %v", int32(n), stretchErr))
	stretchLines := n
	header = splashHeader(scaled)
	n, _, copyErr := directCopy(hdc, 0, 0, uintptr(scaled.Width), uintptr(scaled.Height),
		0, 0, 0, uintptr(scaled.Height), uintptr(unsafe.Pointer(&scaled.BGRA[0])),
		uintptr(unsafe.Pointer(&header)), DIB_RGB_COLORS)
	if n == uintptr(scaled.Height) {
		return "SetDIBitsToDevice (software scaled)", nil
	}
	recordSplashFailure("SetDIBitsToDevice; trying memory bitmap", fmt.Errorf("scanlines=%d/%d: %v", n, scaled.Height, copyErr))
	if err := copySplashBitmap(hdc, scaled); err == nil {
		return "BitBlt (software-filled DIB section)", nil
	} else {
		return "none", fmt.Errorf("StretchDIBits scanlines=0x%x (%v); SetDIBitsToDevice scanlines=%d/%d (%v); bitmap copy: %w",
			stretchLines, stretchErr, n, scaled.Height, copyErr, err)
	}
}

// Fill a Windows-owned memory bitmap directly, then copy it to the window.
// Unlike the two DIB transfer APIs above, BitBlt receives no raw DIB buffer.
func copySplashBitmap(hdc uintptr, scaled splashBitmap) error {
	dc, _, err := procCreateCompatibleDC.Call(hdc)
	if dc == 0 {
		return fmt.Errorf("CreateCompatibleDC: %v", err)
	}
	defer procDeleteDC.Call(dc)
	header := splashHeader(scaled)
	var bits unsafe.Pointer
	bitmap, _, err := procCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&header)), DIB_RGB_COLORS,
		uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bitmap == 0 {
		return fmt.Errorf("CreateDIBSection: %v", err)
	}
	defer procDeleteObject.Call(bitmap)
	if bits == nil {
		return fmt.Errorf("CreateDIBSection returned no pixels")
	}
	copy(unsafe.Slice((*byte)(bits), len(scaled.BGRA)), scaled.BGRA)
	old, _, err := procSelectObject.Call(dc, bitmap)
	if old == 0 || old == ^uintptr(0) {
		return fmt.Errorf("SelectObject: %v", err)
	}
	defer procSelectObject.Call(dc, old)
	if ok, _, err := procBitBlt.Call(hdc, 0, 0, uintptr(scaled.Width), uintptr(scaled.Height), dc, 0, 0, SRCCOPY); ok == 0 {
		return fmt.Errorf("BitBlt: %v", err)
	}
	return nil
}

// Record creation/rendering failures, including recovery through software
// scaling, so a report can distinguish a driver failure from missing artwork.
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
	if e != nil {
		return
	}
	path := filepath.Join(dir, "TimeStamps")
	if os.MkdirAll(path, 0700) != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(path, "splash_status.txt"),
		[]byte(fmt.Sprintf("%s\n%s\n", time.Now().Format(time.RFC3339), stage)), 0600)
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
				if startupSplashScaled.Width != int(rc.Right) || startupSplashScaled.Height != int(rc.Bottom) {
					startupSplashScaled = scaleSplash(startupSplash, int(rc.Right), int(rc.Bottom))
				}
				var drawErr error
				startupSplashRenderer, drawErr = drawSplashArtwork(hdc, startupSplash, startupSplashScaled, splashStretchCall, splashDirectCopyCall)
				startupSplashDrawn = drawErr == nil
				if drawErr != nil {
					recordSplashFailure("drawSplashArtwork", drawErr)
					// Display a readable message if the graphics driver rejects the DIB.
					b, _, _ := procCreateSolidBrush.Call(rgb(28, 8, 50))
					procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), b)
					procDeleteObject.Call(b)
					procSetBkMode.Call(hdc, TRANSPARENT)
					procSetTextColor.Call(hdc, rgb(255, 255, 255))
					caption := u16("TimeStamps — Updated by NightLife")
					procDrawText.Call(hdc, uintptr(unsafe.Pointer(caption)), uintptr(^uint32(0)),
						uintptr(unsafe.Pointer(&rc)), DT_SINGLELINE|DT_VCENTER)
				}
			}
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
	startupSplashScaled = scaleSplash(bmp, w, h)
	startupSplashRenderer = "none"
	x := int(work.Left) + (int(work.Right-work.Left)-w)/2
	y := int(work.Top) + (int(work.Bottom-work.Top)-h)/2
	startupSplashPainted = false
	startupSplashDrawn = false
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
	writeSplashStatus(fmt.Sprintf("Created at %d,%d size %dx%d; visible=%t; painted=%t; imageDrawn=%t; renderer=%s",
		x, y, w, h, visible != 0, startupSplashPainted, startupSplashDrawn, startupSplashRenderer))
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
	writeSplashStatus(fmt.Sprintf("Completed six-second startup splash; visible=%t; WM_PAINT=%t; artworkDrawn=%t; renderer=%s", visible != 0, startupSplashPainted, startupSplashDrawn, startupSplashRenderer))
	return true
}
