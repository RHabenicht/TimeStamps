package main

import "testing"

// Real target: 1360x768 display / 125% scale, less Windows chrome and taskbar.
// Verify the form remains usable at typical client sizes and does not force
// controls outside the visible client area.
func TestAccessibilityLayoutFitsAt125Percent(t *testing.T) {
	for _, v := range []struct{ width, height int }{
		{1040, 515}, // user's likely client size at 1360x768 125%
		{1024, 500},
		{920, 475},
		{830, 465}, // fallback minimum logical client size
		{1160, 610},
	} {
		g := makeLayout(v.width, v.height)
		check := func(label string, b uiBox) {
			t.Helper()
			if b.X < 0 || b.Y < 0 || b.W <= 0 || b.H <= 0 {
				t.Fatalf("%dx%d invalid %s: %+v", v.width, v.height, label, b)
			}
			if b.X+b.W > v.width-9 {
				t.Fatalf("%dx%d %s extends off right: %+v", v.width, v.height, label, b)
			}
			visibleHeight := b.H
			if b.H > 100 {
				visibleHeight = 38
			} // COMBOBOX popup height
			if b.Y+visibleHeight > v.height-7 {
				t.Fatalf("%dx%d %s extends below window: %+v", v.width, v.height, label, b)
			}
		}
		for id, b := range g.Controls {
			check("control", b)
			_ = id
		}
		for name, b := range g.Labels {
			check(name, b)
		}
		if g.Controls[113].Y+g.Controls[113].H >= g.FooterTop {
			t.Errorf("%dx%d posting options obstruct footer", v.width, v.height)
		}
		if g.Labels["formatHelp"].Y+g.Labels["formatHelp"].H >= g.FooterTop {
			t.Errorf("%dx%d format help obstructs footer", v.width, v.height)
		}
		r := g.Labels["resultEdit"]
		if r.H < 110 {
			t.Fatalf("output too small at %dx%d: %+v", v.width, v.height, r)
		}
	}
}

func TestAccessibilityControlsDoNotOverlapAcrossColumns(t *testing.T) {
	for _, w := range []int{830, 920, 1040} {
		g := makeLayout(w, 480)
		for _, id := range []int{101, 102, 103, 104} {
			left := g.Controls[id]
			if left.X+left.W >= g.Controls[105].X {
				t.Fatalf("width %d: input %d overlaps posting settings", w, id)
			}
		}
		// Footer controls must remain on-screen without covering the status.
		if g.Labels["status"].X+g.Labels["status"].W >= g.Controls[114].X {
			t.Fatalf("width %d: status text collides with Generate", w)
		}
	}
}
