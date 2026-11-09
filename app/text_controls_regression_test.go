package main

import (
	"os"
	"strings"
	"testing"
)

func TestResultTextZoomRemoved(t *testing.T) {
	g := makeLayout(1040, 515)
	for _, id := range []int{204, 205} {
		if _, ok := g.Controls[id]; ok {
			t.Fatalf("removed text resize control %d remains in layout", id)
		}
	}
	if _, ok := g.Labels["resultHint"]; !ok {
		t.Fatal("fixed text-size notice missing")
	}
	b, ok := g.Labels["resultHint"]
	if !ok || b.X+b.W > 1040-9 {
		t.Fatalf("text notice off-screen: %+v", b)
	}
}
func TestDynamicResultFontCodeRemoved(t *testing.T) {
	b, err := os.ReadFile("gui_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, old := range []string{"changeResultFont(", "ID_TEXT_MINUS", "ID_TEXT_PLUS", `"A+"`, `"A−"`} {
		if strings.Contains(s, old) {
			t.Errorf("unstable font control remains: %q", old)
		}
	}
	if !strings.Contains(s, `newFont(16, 400, "Consolas")`) {
		t.Error("fixed 16pt font not installed")
	}
}
