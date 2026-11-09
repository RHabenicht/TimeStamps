package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"
	_ "time/tzdata" // Tests use the same IANA rules on Linux and Windows runners.
)

func useNewYork(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
	return loc
}

func TestPostingSessionKeepsLocalHoursAcrossDST(t *testing.T) {
	loc := useNewYork(t)
	for _, tc := range []struct {
		date               string
		start, end         int
		wantStart, wantEnd string
	}{
		{"2026-03-08", 9, 17, "2026-03-08 09:00 -0400", "2026-03-08 17:00 -0400"},
		{"2026-11-01", 9, 17, "2026-11-01 09:00 -0500", "2026-11-01 17:00 -0500"},
		{"2026-03-07", 22, 4, "2026-03-07 22:00 -0500", "2026-03-08 04:00 -0400"},
		{"2026-10-31", 22, 4, "2026-10-31 22:00 -0400", "2026-11-01 04:00 -0500"},
		{"2026-03-08", 0, 0, "2026-03-08 00:00 -0500", "2026-03-09 00:00 -0400"},
		{"2026-11-01", 0, 0, "2026-11-01 00:00 -0400", "2026-11-02 00:00 -0500"},
	} {
		t.Run(fmt.Sprintf("%s/%d-%d", tc.date, tc.start, tc.end), func(t *testing.T) {
			day, err := time.ParseInLocation("2006-01-02", tc.date, loc)
			if err != nil {
				t.Fatal(err)
			}
			start, end := nightlySession(day, tc.start, tc.end)
			if start.Format("2006-01-02 15:04 -0700") != tc.wantStart || end.Format("2006-01-02 15:04 -0700") != tc.wantEnd {
				t.Fatalf("got %s — %s; want %s — %s", start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestDSTGenerationAndContinuation(t *testing.T) {
	loc := useNewYork(t)
	for _, date := range []string{"2026-03-08", "2026-11-01"} {
		s := defaults()
		s.StartDate, s.WindowStart, s.WindowEnd, s.Quantity = date, 9, 17, 1
		got, err := Generate(s, nil, bytes.NewReader(make([]byte, 128)))
		if err != nil {
			t.Fatal(err)
		}
		if got[0].Hour() != 9 || got[0].Minute() != 0 {
			t.Fatalf("%s: zero-jitter first posting = %s; want 09:00", date, got[0])
		}
		day, _ := time.ParseInLocation("2006-01-02", date, loc)
		previous := day.AddDate(0, 0, -1)
		y, m, d := previous.Date()
		previous = time.Date(y, m, d, 16, 59, 0, 0, loc)
		s.MinGapMinutes, s.MaxGapMinutes = 1, 1
		got, err = Generate(s, &previous, bytes.NewReader(make([]byte, 128)))
		if err != nil {
			t.Fatal(err)
		}
		if got[0].Format("2006-01-02 15:04") != date+" 09:00" {
			t.Fatalf("continuation: got %s; want %s 09:00", got[0], date)
		}
	}
}

func TestEveryPostingWindowAcrossDST(t *testing.T) {
	useNewYork(t)
	for _, date := range []string{"2026-03-07", "2026-10-31"} {
		for start := 0; start < 24; start++ {
			for end := 0; end < 24; end++ {
				s := defaults()
				s.StartDate, s.WindowStart, s.WindowEnd, s.Quantity = date, start, end, 96
				s.MinGapMinutes, s.MaxGapMinutes = 29, 200
				r := bytes.NewReader(bytes.Repeat([]byte{17, 33, 49, 65, 81, 97, 113, 129}, 1024))
				got, err := Generate(s, nil, r)
				if err != nil {
					t.Fatalf("%s %d-%d: %v", date, start, end, err)
				}
				for i, stamp := range got {
					// Independent wall-clock oracle: do not reuse allowedAt or nightlySession.
					h := stamp.Hour()
					within := start == end || (start < end && h >= start && h < end) || (start > end && (h >= start || h < end))
					if !within {
						t.Fatalf("%s %d-%d: outside window: %s", date, start, end, stamp)
					}
					if i > 0 && !stamp.After(got[i-1]) {
						t.Fatalf("clock did not advance: %s", stamp)
					}
				}
			}
		}
	}
}

func TestSkippedAndRepeatedPostingHours(t *testing.T) {
	loc := useNewYork(t)
	s := defaults()
	s.WindowStart, s.WindowEnd = 2, 3
	before := time.Date(2026, 3, 8, 1, 59, 0, 0, loc)
	if got := nextPermitted(before, s); got.Format("2006-01-02 15:04") != "2026-03-09 02:00" {
		t.Fatalf("nonexistent window should skip to tomorrow: %s", got)
	}
	s.WindowEnd = 4
	if got := nextPermitted(before, s); got.Format("2006-01-02 15:04") != "2026-03-08 03:00" {
		t.Fatalf("nonexistent opening should move forward: %s", got)
	}
	s.WindowStart, s.WindowEnd = 1, 2
	for _, input := range []string{"2026-11-01T01:30:00-04:00", "2026-11-01T01:30:00-05:00"} {
		stamp, _ := time.Parse(time.RFC3339, input)
		stamp = stamp.In(loc)
		if !allowedAt(stamp, s) || !nextPermitted(stamp, s).Equal(stamp) {
			t.Fatalf("both occurrences of 01:30 must be permitted: %s", stamp)
		}
	}
}
