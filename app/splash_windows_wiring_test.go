package main

import (
	"os"
	"strings"
	"testing"
)

func TestSplashStartsBeforeMainWindow(t *testing.T) {
	g, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	s, err := os.ReadFile("splash_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	gui, splash := string(g), string(s)
	lock := strings.Index(gui, "runtime.LockOSThread()")
	show := strings.Index(gui, "showStartupSplash(module, work)")
	create := strings.Index(gui, "h, _, err := procCreateWindowEx.Call(")
	if lock < 0 || show < 0 || create < 0 || !(lock < show && show < create) {
		t.Fatal("main must lock OS thread and show splash before main HWND")
	}
	for _, need := range []string{"procShowWindow.Call(sh, SW_SHOW)", "procUpdateWindow.Call(sh)", "procPeekMessage.Call", "time.Since(started) < 4*time.Second", "procStretchDIBits.Call", "WS_EX_TOPMOST|WS_EX_TOOLWINDOW"} {
		if !strings.Contains(splash, need) {
			t.Errorf("splash missing %q", need)
		}
	}
}
