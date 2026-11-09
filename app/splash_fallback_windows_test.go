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
				})
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
