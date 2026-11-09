//go:build windows

package main

import (
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// Exercise the failure path on real Windows GDI surfaces, even when the hosted
// runner's driver accepts StretchDIBits. No environment switch ships in the app.
func TestSplashSoftwareFallbackPaintsApprovedPixels(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv("APPDATA", t.TempDir())
	source, err := decodeSplash(splashPNG)
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{1180, 664}, {976, 549}} {
		for _, rejected := range []uintptr{0, uintptr(^uint32(0))} {
			scaled := scaleSplash(source, size[0], size[1])
			header := splashHeader(scaled)
			dc, _, dcErr := dllGdi32.NewProc("CreateCompatibleDC").Call(0)
			if dc == 0 {
				t.Fatalf("CreateCompatibleDC: %v", dcErr)
			}
			var pixels uintptr
			bitmap, _, bitmapErr := dllGdi32.NewProc("CreateDIBSection").Call(dc,
				uintptr(unsafe.Pointer(&header)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&pixels)), 0, 0)
			if bitmap == 0 {
				dllGdi32.NewProc("DeleteDC").Call(dc)
				t.Fatalf("CreateDIBSection: %v", bitmapErr)
			}
			old, _, _ := procSelectObject.Call(dc, bitmap)
			func() {
				defer func() {
					procSelectObject.Call(dc, old)
					procDeleteObject.Call(bitmap)
					dllGdi32.NewProc("DeleteDC").Call(dc)
				}()
				called := false
				renderer, err := drawSplashArtwork(dc, source, scaled, func(...uintptr) (uintptr, uintptr, error) {
					called = true
					return rejected, 0, syscall.Errno(87)
				}, procSetDIBitsToDevice.Call)
				if err != nil || !called || !strings.HasPrefix(renderer, "SetDIBitsToDevice") {
					t.Fatalf("forced StretchDIBits failure: renderer=%q, err=%v", renderer, err)
				}
				// Compare actual GDI pixels with the corresponding original artwork
				// pixels, including the right-hand serpent and all four corners.
				for _, point := range [][2]int{{0, 0}, {size[0] - 1, 0}, {0, size[1] - 1}, {size[0] - 1, size[1] - 1},
					{size[0] / 4, size[1] / 4}, {size[0] / 2, size[1] / 2}, {size[0] * 3 / 4, size[1] / 2}, {size[0] * 9 / 10, size[1] * 3 / 4}} {
					x, y := point[0], point[1]
					from := ((y*source.Height/size[1])*source.Width + x*source.Width/size[0]) * 4
					want := uintptr(source.BGRA[from+2]) | uintptr(source.BGRA[from+1])<<8 | uintptr(source.BGRA[from])<<16
					got, _, _ := dllGdi32.NewProc("GetPixel").Call(dc, uintptr(x), uintptr(y))
					if got != want {
						t.Fatalf("%dx%d pixel %d,%d: got %#x, want %#x", size[0], size[1], x, y, got, want)
					}
				}
				t.Logf("forced StretchDIBits result %#x: %dx%d approved artwork pixels painted with %s", rejected, size[0], size[1], renderer)
			}()
		}
	}
}

// Reproduce the owner's exact failure combination in the real WM_PAINT path,
// rather than only testing a memory DC. The production app has no test switch.
func TestSplashPopupRecoversWhenBothDIBTransfersFail(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv("APPDATA", t.TempDir())
	procSetProcessDPIAware.Call()
	source, err := decodeSplash(splashPNG)
	if err != nil {
		t.Fatal(err)
	}
	oldSource, oldScaled := startupSplash, startupSplashScaled
	oldStretch, oldCopy := splashStretchCall, splashDirectCopyCall
	oldPainted, oldDrawn, oldRenderer := startupSplashPainted, startupSplashDrawn, startupSplashRenderer
	defer func() {
		startupSplash, startupSplashScaled = oldSource, oldScaled
		splashStretchCall, splashDirectCopyCall = oldStretch, oldCopy
		startupSplashPainted, startupSplashDrawn, startupSplashRenderer = oldPainted, oldDrawn, oldRenderer
	}()
	stretchCalls, copyCalls := 0, 0
	splashStretchCall = func(...uintptr) (uintptr, uintptr, error) { stretchCalls++; return 0, 0, syscall.Errno(0) }
	splashDirectCopyCall = func(...uintptr) (uintptr, uintptr, error) { copyCalls++; return 0, 0, syscall.Errno(0) }
	var work rect
	if ok, _, err := procSystemParametersInfo.Call(48, 0, uintptr(unsafe.Pointer(&work)), 0); ok == 0 {
		t.Fatalf("work area: %v", err)
	}
	w, h := fitSplash(int(work.Right-work.Left), int(work.Bottom-work.Top), source.Width, source.Height)
	startupSplash, startupSplashScaled = source, scaleSplash(source, w, h)
	startupSplashPainted, startupSplashDrawn = false, false
	module, _, _ := procGetModuleHandle.Call(0)
	name := u16("TimeStampsSplashRecoveryTest")
	wc := wndclassex{Size: uint32(unsafe.Sizeof(wndclassex{})), WndProc: syscall.NewCallback(startupSplashWndProc), Instance: module, ClassName: name}
	if ok, _, err := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc))); ok == 0 {
		t.Fatalf("RegisterClassExW: %v", err)
	}
	defer dllUser32.NewProc("UnregisterClassW").Call(uintptr(unsafe.Pointer(name)), module)
	window, _, err := procCreateWindowEx.Call(WS_EX_TOPMOST|WS_EX_TOOLWINDOW, uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(u16("TimeStamps splash recovery test"))), WS_POPUP,
		uintptr(work.Left+24), uintptr(work.Top+24), uintptr(w), uintptr(h), 0, 0, module, 0)
	if window == 0 {
		t.Fatalf("CreateWindowExW: %v", err)
	}
	defer procDestroyWindow.Call(window)
	procShowWindow.Call(window, SW_SHOW)
	procInvalidateSplash.Call(window, 0, 0)
	procUpdateWindow.Call(window)
	if !startupSplashPainted || !startupSplashDrawn || !strings.HasPrefix(startupSplashRenderer, "BitBlt") || stretchCalls == 0 || copyCalls == 0 {
		t.Fatalf("popup did not recover: painted=%t drawn=%t renderer=%s calls=%d/%d", startupSplashPainted, startupSplashDrawn, startupSplashRenderer, stretchCalls, copyCalls)
	}
	dc, _, err := dllUser32.NewProc("GetDC").Call(window)
	if dc == 0 {
		t.Fatalf("GetDC: %v", err)
	}
	defer dllUser32.NewProc("ReleaseDC").Call(window, dc)
	for _, point := range [][2]int{{0, 0}, {w - 1, 0}, {0, h - 1}, {w - 1, h - 1}, {w / 4, h / 4}, {w / 2, h / 2}, {w * 3 / 4, h / 2}, {w * 9 / 10, h * 3 / 4}} {
		x, y := point[0], point[1]
		from := ((y*source.Height/h)*source.Width + x*source.Width/w) * 4
		want := uintptr(source.BGRA[from+2]) | uintptr(source.BGRA[from+1])<<8 | uintptr(source.BGRA[from])<<16
		got, _, _ := dllGdi32.NewProc("GetPixel").Call(dc, uintptr(x), uintptr(y))
		if got != want {
			t.Fatalf("popup pixel %d,%d: got %#x want %#x", x, y, got, want)
		}
	}
	t.Logf("actual %dx%d popup: both DIB transfers returned zero/no error; approved artwork painted by %s", w, h, startupSplashRenderer)
}
