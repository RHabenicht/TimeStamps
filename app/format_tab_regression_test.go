package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestAllFormatPresetsProduceOutput(t *testing.T) {
	ts := time.Date(2051, 7, 13, 22, 45, 9, 0, time.UTC)
	expected := map[int]string{
		1: "22:45:09/07-13-51",
		2: "(22:45:09/07-13-51)",
		3: "2051-07-13T22:45:09",
		4: "07-13-51 22:45:09",
	}
	for index, want := range expected {
		preset, ok := presetByIndex(index)
		if !ok {
			t.Fatalf("missing preset %d", index)
		}
		s := defaults()
		s.Template = preset.Template
		if preset.DateLayout != "" {
			s.DateLayout = preset.DateLayout
		}
		if preset.TimeLayout != "" {
			s.TimeLayout = preset.TimeLayout
		}
		if err := s.validate(); err != nil {
			t.Fatalf("preset %d validation: %v", index, err)
		}
		if got := Format(ts, s); got != want {
			t.Fatalf("preset %d: got %q want %q", index, got, want)
		}
	}
	for _, bad := range []int{-1, 0, 5, 1000} {
		if _, ok := presetByIndex(bad); ok {
			t.Fatalf("accepted preset %d", bad)
		}
	}
}

func TestFormatTabWin32EventFlowRegression(t *testing.T) {
	// Checks the Windows-only implementation when tests run on Linux.
	data, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	src := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, required := range []string{
		"case ID_PAGE_FORMAT:\n\t\trequestPage(1)",
		"procPostMessage.Call(app.hwnd, WM_APP_SWITCHPAGE",
		"case WM_APP_SWITCHPAGE:",
		"case WM_APP_APPLYPRESET:",
		"procPostMessage.Call(app.hwnd, WM_APP_APPLYPRESET",
		"func applyPreset()",
		"if app.switching || app.updating",
	} {
		if !strings.Contains(src, required) {
			t.Errorf("missing format lockup guard: %s", required)
		}
	}
	if strings.Contains(src, "case ID_PAGE_FORMAT:\n\t\tshowPage(1)") {
		t.Error("Format page is still changed synchronously in WM_COMMAND")
	}
	if strings.Contains(src, "setText(app.c(x.id), title)") {
		t.Error("dynamic nav caption update was reintroduced")
	}
}
