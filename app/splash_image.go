package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
)

// Embedded splash artwork approved for TimeStamps 2.4.
//
//go:embed splash_screen.png
var splashPNG []byte

type splashBitmap struct {
	Width, Height int
	BGRA          []byte
}

// decodeSplash produces top-down, tightly packed 32-bit BGRA pixels for
// GDI's BI_RGB DIB drawing. An opaque background avoids alpha blending issues.
func decodeSplash(data []byte) (splashBitmap, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return splashBitmap{}, fmt.Errorf("decode splash PNG: %w", err)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || w > 8192 || h > 8192 {
		return splashBitmap{}, fmt.Errorf("invalid splash dimensions %dx%d", w, h)
	}
	out := splashBitmap{Width: w, Height: h, BGRA: make([]byte, w*h*4)}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bb, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			i := (y*w + x) * 4
			out.BGRA[i] = byte(bb >> 8)
			out.BGRA[i+1] = byte(g >> 8)
			out.BGRA[i+2] = byte(r >> 8)
			out.BGRA[i+3] = 0xff
		}
	}
	return out, nil
}

func fitSplash(workW, workH, imageW, imageH int) (int, int) {
	if workW <= 0 || workH <= 0 || imageW <= 0 || imageH <= 0 {
		return 0, 0
	}
	maxW, maxH := workW-48, workH-48
	if maxW > 1180 {
		maxW = 1180
	}
	if maxH > 664 {
		maxH = 664
	}
	if maxW < 1 || maxH < 1 {
		return 0, 0
	}
	w := maxW
	h := w * imageH / imageW
	if h > maxH {
		h = maxH
		w = h * imageW / imageH
	}
	if w < 1 || h < 1 {
		return 0, 0
	}
	return w, h
}

const splashDurationMs = 4000
