package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultGeneration(t *testing.T) {
	s := defaults()
	s.StartDate = "2051-07-13"
	s.Quantity = 800
	vals, err := Generate(s, nil, bytes.NewReader(bytes.Repeat([]byte{0x31, 0x02, 0x56, 0x00, 0x22, 0x17, 0x07, 0x00}, 4000)))
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 800 {
		t.Fatal("wrong count")
	}
	for i, v := range vals {
		if !allowedAt(v, s) {
			t.Fatalf("outside window %s", v)
		}
		if i > 0 && !v.After(vals[i-1]) {
			t.Fatal("not strictly increasing")
		}
		if i > 0 && v.Sub(vals[i-1]) <= time.Duration(s.MaxGapMinutes)*time.Minute && v.Sub(vals[i-1]) < time.Duration(s.MinGapMinutes)*time.Minute {
			t.Fatalf("too close: %v", v.Sub(vals[i-1]))
		}
	}
}
func TestCrossMidnight(t *testing.T) {
	s := defaults()
	s.Quantity = 200
	s.MinGapMinutes = 100
	s.MaxGapMinutes = 1440
	s.WindowEnd = 3
	vals, err := Generate(s, nil, bytes.NewReader(bytes.Repeat([]byte{2, 4, 6, 8, 10, 12, 14, 16}, 2500)))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vals {
		if !allowedAt(v, s) {
			t.Fatal("bad overnight time", v)
		}
	}
}
func TestUnrestrictedAndContinue(t *testing.T) {
	s := defaults()
	s.RestrictHours = false
	s.Quantity = 20
	r := bytes.NewReader(bytes.Repeat([]byte{1, 1, 1, 1, 1, 1, 1, 1}, 100))
	first, err := Generate(s, nil, r)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(s, &first[len(first)-1], r)
	if err != nil {
		t.Fatal(err)
	}
	if !second[0].After(first[len(first)-1]) {
		t.Fatal("no continuation")
	}
}
func TestEqualGapAndFullDay(t *testing.T) {
	s := defaults()
	s.MinGapMinutes = 30
	s.MaxGapMinutes = 30
	s.WindowStart = 0
	s.WindowEnd = 0
	s.Quantity = 200
	vals, err := Generate(s, nil, bytes.NewReader(bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 1000)))
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(vals); i++ {
		if vals[i].Sub(vals[i-1]) != 30*time.Minute {
			t.Fatal("incorrect equal gap")
		}
	}
}
func TestFormat(t *testing.T) {
	s := defaults()
	d := time.Date(2051, 7, 13, 22, 45, 9, 0, time.UTC)
	if got := Format(d, s); got != "22:45:09/07-13-51" {
		t.Fatal(got)
	}
	s.Template = "({time}/{date}) [{yyyy}-{MM}-{dd}]"
	if got := Format(d, s); got != "(22:45:09/07-13-51) [2051-07-13]" {
		t.Fatal(got)
	}
}
func TestLegacyINI(t *testing.T) {
	sample := `[Settings]
ClearListWhenGenerating=1
KeepFinalTimeStamp=1
StartDate=08/21/2053
ValidateHourRange=1
ValidStartHour=22
ValidEndHour=3
MinBetween=100
MaxBetween=1440
Qty=20
[DateFormatTab]
Month1=1
Day2=1
Year3=1
Month2Digit=1
Day2Digit=1
Year2Digit=1
Hyphen1=1
Hyphen2=1
[TimeFormatTab]
Hours2Digit=1
Minutes2Digit=1
Seconds2Digit=1
TimeFormat24Hour=1
[FormatTab]
InFront="Comment ("
FirstElement=0
Between="/"
LastElement=1
After=")"
`
	p := filepath.Join(t.TempDir(), "TimeStamps.ini")
	if e := os.WriteFile(p, []byte(sample), 0600); e != nil {
		t.Fatal(e)
	}
	s, e := importLegacy(p)
	if e != nil {
		t.Fatal(e)
	}
	if s.StartDate != "2053-08-21" || s.MinGapMinutes != 100 || s.MaxGapMinutes != 1440 || s.WindowEnd != 3 || s.Template != "Comment ({time}/{date})" {
		t.Fatalf("bad import %+v", s)
	}
	if !strings.Contains(Format(time.Date(2053, 8, 21, 22, 1, 2, 0, time.UTC), s), "22:01:02/08-21-53") {
		t.Fatal("bad format")
	}
}
func TestInvalid(t *testing.T) {
	s := defaults()
	s.MaxGapMinutes = 0
	if err := s.validate(); err == nil {
		t.Fatal("accepted invalid gap")
	}
	s = defaults()
	s.Quantity = 10001
	if err := s.validate(); err == nil {
		t.Fatal("accepted invalid quantity")
	}
}

func TestEveryPostingWindow(t *testing.T) {
	for start := 0; start < 24; start++ {
		for end := 0; end < 24; end++ {
			s := defaults()
			s.StartDate = "2051-07-13"
			s.WindowStart = start
			s.WindowEnd = end
			s.MinGapMinutes = 29
			s.MaxGapMinutes = 200
			s.Quantity = 32
			r := bytes.NewReader(bytes.Repeat([]byte{17, 33, 49, 65, 81, 97, 113, 129}, 1024))
			vals, err := Generate(s, nil, r)
			if err != nil {
				t.Fatalf("window %d-%d failed: %v", start, end, err)
			}
			for i, v := range vals {
				if !allowedAt(v, s) {
					t.Fatalf("window %d-%d: %v not permitted", start, end, v)
				}
				if i > 0 && !v.After(vals[i-1]) {
					t.Fatalf("window %d-%d: nonincreasing", start, end)
				}
			}
		}
	}
}
