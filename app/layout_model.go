package main

// The layout is expressed in 96-DPI-independent units. Windows converts these
// to physical pixels using the current monitor DPI. Only the active page is
// shown. This avoids the fixed-height form used in TimeStamps 2.0.
type uiBox struct{ X, Y, W, H int }
type uiGeometry struct {
	Controls  map[int]uiBox
	Labels    map[string]uiBox
	FooterTop int
}

func makeLayout(w, h int) uiGeometry {
	if w < 800 {
		w = 800
	}
	if h < 465 {
		h = 465
	}
	g := uiGeometry{Controls: make(map[int]uiBox), Labels: make(map[string]uiBox), FooterTop: h - 63}
	c := func(id, x, y, bw, bh int) { g.Controls[id] = uiBox{x, y, bw, bh} }
	l := func(key string, x, y, bw, bh int) { g.Labels[key] = uiBox{x, y, bw, bh} }
	const pad = 34
	mid := w / 2
	colLeft := mid - 56
	colRightX := mid + 22
	colRight := w - colRightX - 34
	if colRight < 250 {
		colRight = 250
	}
	// Navigation and persistent footer (always visible).
	c(201, pad, 88, 142, 38)
	c(202, pad+152, 88, 142, 38)
	c(203, pad+304, 88, 142, 38)
	c(114, w-280, h-54, 244, 43)
	c(119, w-130, 22, 97, 34)
	l("status", pad, h-53, w-345, 40)
	// Page 1: generation. Two independent columns are accessible at 125% DPI.
	l("rules", pad, 150, colLeft, 24)
	l("hoursTitle", colRightX, 150, colRight, 24)
	labelW := 214
	inputX := pad + labelW + 5
	inputW := colLeft - labelW - 5
	if inputW < 110 {
		inputW = 110
	}
	rows := []struct {
		key   string
		id, y int
		title string
	}{
		{"date", 101, 188, ""}, {"qty", 102, 236, ""}, {"min", 103, 284, ""}, {"max", 104, 332, ""},
	}
	for _, row := range rows {
		l(row.key, pad, row.y+5, labelW-6, 29)
		c(row.id, inputX, row.y, inputW, 38)
	}
	c(105, colRightX, 191, colRight, 36)
	l("from", colRightX, 239, 145, 24)
	l("to", colRightX+colRight/2+6, 239, 145, 24)
	hourW := (colRight - 14) / 2
	c(106, colRightX, 265, hourW, 240) // combobox popup height, visible ~38
	c(107, colRightX+hourW+14, 265, hourW, 240)
	c(112, colRightX, 317, colRight, 35)
	c(113, colRightX, 356, colRight, 35)
	// Page 2: output format configuration.
	l("format", pad, 150, colLeft, 24)
	l("timeOptions", colRightX, 150, colRight, 24)
	l("preset", pad, 194, colLeft, 24)
	c(108, pad, 222, colLeft, 240)
	l("template", pad, 288, colLeft, 50)
	c(109, pad, 336, colLeft, 38)
	l("datefmt", colRightX, 194, colRight, 24)
	c(110, colRightX, 222, colRight, 240)
	l("timefmt", colRightX, 288, colRight, 24)
	c(111, colRightX, 316, colRight, 240)
	l("formatHelp", colRightX, 360, colRight, 40)
	// Page 3: output list and actions.
	l("output", pad, 149, w-2*pad, 27)
	toolbarY := 189
	c(115, pad, toolbarY, 157, 41)
	c(116, pad+169, toolbarY, 137, 41)
	c(117, pad+319, toolbarY, 137, 41)
	c(118, pad+469, toolbarY, 125, 41)
	l("resultHint", w-332, toolbarY+7, 298, 28)
	bottom := h - 79
	if bottom < 247 {
		bottom = 247
	}
	l("resultEdit", pad, 243, w-2*pad, bottom-243)
	return g
}
