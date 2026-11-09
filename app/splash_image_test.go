package main

import "testing"

func TestSplashArtworkDecodes(t *testing.T) {
	bmp, err := decodeSplash(splashPNG)
	if err != nil {
		t.Fatal(err)
	}
	if bmp.Width != 1672 || bmp.Height != 941 {
		t.Fatalf("unexpected dimensions: %dx%d", bmp.Width, bmp.Height)
	}
	if len(bmp.BGRA) != bmp.Width*bmp.Height*4 {
		t.Fatal("invalid pixel buffer length")
	}
	if splashDurationMs != 6000 {
		t.Fatal("splash should display for six seconds")
	}
}
func TestSplashFitsAccessibleMonitor(t *testing.T) {
	for _, work := range [][2]int{{1360, 728}, {1360, 768}, {1090, 630}, {1920, 1040}, {800, 600}} {
		w, h := fitSplash(work[0], work[1], 1672, 941)
		if w <= 0 || h <= 0 || w > work[0]-48 || h > work[1]-48 {
			t.Fatalf("splash %dx%d does not fit usable desktop %dx%d", w, h, work[0], work[1])
		}
		original := float64(1672) / 941
		ratio := float64(w) / float64(h)
		if ratio < original-0.01 || ratio > original+0.01 {
			t.Fatalf("aspect ratio not preserved: %f", ratio)
		}
	}
}
