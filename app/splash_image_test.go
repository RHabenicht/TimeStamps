package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

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

func TestFullSplashExactlyMatchesApprovedNightLifeArtwork(t *testing.T) {
	// The original approved full splash (including the large right-side serpent
	// logo) is already embedded. Its pixels match the user-uploaded 1672x941 PNG.
	// Hash decoded pixels so PNG encoding or metadata changes do not cause failure.
	const approvedRGBA = "f1311c993d8286e64151c090799e3733904191fdc00d8fffd2c92fc921c343e1"
	bmp, err := decodeSplash(splashPNG)
	if err != nil { t.Fatal(err) }
	rgba := make([]byte, len(bmp.BGRA))
	for i := 0; i < len(rgba); i += 4 {
		rgba[i+0] = bmp.BGRA[i+2]
		rgba[i+1] = bmp.BGRA[i+1]
		rgba[i+2] = bmp.BGRA[i+0]
		rgba[i+3] = bmp.BGRA[i+3]
	}
	digest := sha256.Sum256(rgba)
	if got := hex.EncodeToString(digest[:]); got != approvedRGBA {
		t.Fatalf("embedded splash does not match the approved full image pixels: got %s", got)
	}
}
