package main

import (
	"bytes"
	"testing"
)

func TestSoftwareSplashScaleKeepsWholeImage(t *testing.T) {
	source := splashBitmap{Width: 2, Height: 2, BGRA: []byte{
		1, 2, 3, 255, 4, 5, 6, 255,
		7, 8, 9, 255, 10, 11, 12, 255,
	}}
	original := append([]byte(nil), source.BGRA...)
	scaled := scaleSplash(source, 4, 4)
	if scaled.Width != 4 || scaled.Height != 4 {
		t.Fatal("incorrect scaled size")
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			i, j := (y*4+x)*4, ((y/2)*2+x/2)*4
			if !bytes.Equal(scaled.BGRA[i:i+4], original[j:j+4]) {
				t.Fatalf("incorrect pixel at %d,%d", x, y)
			}
		}
	}
	if !bytes.Equal(source.BGRA, original) {
		t.Fatal("source artwork was modified")
	}
	if got := scaleSplash(source, 0, 4); len(got.BGRA) != 0 {
		t.Fatal("accepted empty target")
	}
	if got := scaleSplash(splashBitmap{}, 4, 4); len(got.BGRA) != 0 {
		t.Fatal("accepted empty source")
	}
}
