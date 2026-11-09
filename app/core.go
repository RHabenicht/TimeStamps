package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const appVersion = "2.3.0"

// Settings are independent of the Windows interface and stored in a user-writable directory.
type Settings struct {
	StartDate     string `json:"start_date"`
	Quantity      int    `json:"quantity"`
	MinGapMinutes int    `json:"minimum_gap_minutes"`
	MaxGapMinutes int    `json:"maximum_gap_minutes"`
	RestrictHours bool   `json:"restrict_hours"`
	WindowStart   int    `json:"window_start_hour"`
	WindowEnd     int    `json:"window_end_hour"`
	Continue      bool   `json:"continue_from_last"`
	Replace       bool   `json:"replace_on_generate"`
	Template      string `json:"template"`
	DateLayout    string `json:"date_layout"`
	TimeLayout    string `json:"time_layout"`
	LastTimestamp string `json:"last_timestamp,omitempty"`
}

func defaults() Settings {
	return Settings{
		StartDate: "2050-01-01", Quantity: 20, MinGapMinutes: 30, MaxGapMinutes: 120,
		RestrictHours: true, WindowStart: 22, WindowEnd: 4, Continue: true, Replace: true,
		Template: "{time}/{date}", DateLayout: "01-02-06", TimeLayout: "15:04:05",
	}
}

func (s Settings) validate() error {
	if _, err := parseDate(s.StartDate); err != nil {
		return fmt.Errorf("start date: %w", err)
	}
	if s.Quantity < 1 || s.Quantity > 10000 {
		return errors.New("quantity must be 1–10,000")
	}
	if s.MinGapMinutes < 1 || s.MaxGapMinutes < 1 || s.MinGapMinutes > s.MaxGapMinutes || s.MaxGapMinutes > 43200 {
		return errors.New("gap must be 1–43,200 minutes, and minimum cannot exceed maximum")
	}
	if s.WindowStart < 0 || s.WindowStart > 23 || s.WindowEnd < 0 || s.WindowEnd > 23 {
		return errors.New("posting hours must be between 0 and 23")
	}
	if !strings.Contains(s.Template, "{time}") && !strings.Contains(s.Template, "{date}") && !strings.Contains(s.Template, "{yyyy}") {
		return errors.New("template must contain a timestamp placeholder, such as {time} or {date}")
	}
	if len(s.Template) > 2000 {
		return errors.New("template is too long")
	}
	return nil
}

func parseDate(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	for _, layout := range []string{"2006-01-02", "1/2/2006", "01/02/2006", "1-2-2006", "01-02-2006"} {
		if d, e := time.ParseInLocation(layout, input, time.Local); e == nil {
			return d, nil
		}
	}
	return time.Time{}, errors.New("use YYYY-MM-DD (for example, 2051-07-13)")
}

func roundSeconds(t time.Time) time.Time { return t.Truncate(time.Second) }

// Uniform randomness inclusive of both endpoints; errors are surfaced instead of using a predictable seed.
func randBetween(r io.Reader, min, max int64) (int64, error) {
	if max < min {
		return 0, errors.New("invalid random range")
	}
	width := uint64(max - min + 1)
	if width == 1 {
		return min, nil
	}
	bound := ^uint64(0) - (^uint64(0) % width)
	for {
		var buf [8]byte
		if _, e := io.ReadFull(r, buf[:]); e != nil {
			return 0, e
		}
		v := binary.LittleEndian.Uint64(buf[:])
		if v < bound {
			return min + int64(v%width), nil
		}
	}
}

// NightlySession is the permitted interval that begins on the given calendar day.
// Start==end denotes an unrestricted 24-hour interval.
func nightlySession(day time.Time, startHour, endHour int) (time.Time, time.Time) {
	y, m, d := day.Date()
	base := time.Date(y, m, d, 0, 0, 0, 0, day.Location())
	start := base.Add(time.Duration(startHour) * time.Hour)
	if startHour == endHour {
		return start, start.Add(24 * time.Hour)
	}
	end := base.Add(time.Duration(endHour) * time.Hour)
	if endHour <= startHour {
		end = end.Add(24 * time.Hour)
	}
	return start, end
}

func allowedAt(t time.Time, s Settings) bool {
	if !s.RestrictHours {
		return true
	}
	// Check today's and yesterday's session to support midnight wrap.
	for i := -1; i <= 0; i++ {
		d := t.AddDate(0, 0, i)
		a, b := nightlySession(d, s.WindowStart, s.WindowEnd)
		if !t.Before(a) && t.Before(b) {
			return true
		}
	}
	return false
}

func nextPermitted(t time.Time, s Settings) time.Time {
	if !s.RestrictHours || allowedAt(t, s) {
		return t
	}
	// Search a nearby interval; no infinite loops, even with crossing-midnight windows.
	var earliest time.Time
	for i := -1; i <= 2; i++ {
		start, end := nightlySession(t.AddDate(0, 0, i), s.WindowStart, s.WindowEnd)
		if !end.After(t) {
			continue
		}
		if start.Before(t) {
			start = t
		}
		if earliest.IsZero() || start.Before(earliest) {
			earliest = start
		}
	}
	return earliest
}

// Generate produces strictly ordered timestamps. The next nightly session may
// cause a gap larger than MaxGapMinutes; this is necessary to honor allowed hours.
func Generate(s Settings, previous *time.Time, r io.Reader) ([]time.Time, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	if r == nil {
		r = rand.Reader
	}
	first, err := parseDate(s.StartDate)
	if err != nil {
		return nil, err
	}
	var last time.Time
	if previous != nil {
		last = roundSeconds(*previous)
	}
	out := make([]time.Time, 0, s.Quantity)
	for i := 0; i < s.Quantity; i++ {
		var current time.Time
		if i == 0 && previous == nil {
			if s.RestrictHours {
				start, _ := nightlySession(first, s.WindowStart, s.WindowEnd)
				// First posting begins within 30 minutes of the first allowed window.
				jitter, e := randBetween(r, 0, 1799)
				if e != nil {
					return nil, e
				}
				current = nextPermitted(start.Add(time.Duration(jitter)*time.Second), s)
			} else {
				jitter, e := randBetween(r, int64(s.MinGapMinutes)*60, int64(s.MaxGapMinutes)*60)
				if e != nil {
					return nil, e
				}
				current = first.Add(time.Duration(jitter) * time.Second)
			}
		} else {
			gap, e := randBetween(r, int64(s.MinGapMinutes)*60, int64(s.MaxGapMinutes)*60)
			if e != nil {
				return nil, e
			}
			candidate := last.Add(time.Duration(gap) * time.Second)
			current = nextPermitted(candidate, s)
			// If the posting window was skipped, spread arrivals into first 30 minutes.
			// Intra-window gap bounds only apply to unskipped sessions.
			if current.After(candidate) {
				jitter, e := randBetween(r, 0, 1799)
				if e != nil {
					return nil, e
				}
				_, end := nightlySession(current, s.WindowStart, s.WindowEnd)
				adjusted := current.Add(time.Duration(jitter) * time.Second)
				if adjusted.Before(end) {
					current = adjusted
				}
			}
		}
		if i != 0 || previous != nil {
			if !current.After(last) {
				return nil, errors.New("generation failed to advance clock")
			}
		}
		out = append(out, current)
		last = current
	}
	return out, nil
}

func Format(t time.Time, s Settings) string {
	date := t.Format(s.DateLayout)
	clock := t.Format(s.TimeLayout)
	vals := map[string]string{
		"{time}": clock, "{date}": date, "{yyyy}": t.Format("2006"), "{yy}": t.Format("06"),
		"{MM}": t.Format("01"), "{dd}": t.Format("02"), "{HH}": t.Format("15"),
		"{mm}": t.Format("04"), "{ss}": t.Format("05"), "{ddd}": t.Format("Mon"),
	}
	replacements := make([]string, 0, len(vals)*2)
	for k, v := range vals {
		replacements = append(replacements, k, v)
	}
	return strings.NewReplacer(replacements...).Replace(s.Template)
}

func FormatList(a []time.Time, s Settings) string {
	var b strings.Builder
	for _, t := range a {
		b.WriteString(Format(t, s))
		b.WriteString("\r\n")
	}
	return b.String()
}

func configPath() (string, error) {
	dir, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(dir, "NightLife", "TimeStamps", "settings.json"), nil
}

func SaveSettings(s Settings) error {
	p, e := configPath()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	bs, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	tmp := p + ".tmp"
	if e = os.WriteFile(tmp, bs, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, p)
}

func LoadSettings(legacyPath string) Settings {
	s := defaults()
	p, e := configPath()
	if e == nil {
		if bs, err := os.ReadFile(p); err == nil {
			if json.Unmarshal(bs, &s) == nil && s.validate() == nil {
				return s
			}
		}
	}
	if legacyPath != "" {
		if migrated, err := importLegacy(legacyPath); err == nil {
			s = migrated
		}
	}
	return s
}

// importLegacy reads old Delphi-3 INI settings without modifying the user's file.
func importLegacy(path string) (Settings, error) {
	s := defaults()
	f, e := os.Open(path)
	if e != nil {
		return s, e
	}
	defer f.Close()
	values := map[string]string{}
	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok {
			values[section+"."+strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
	}
	if e := scanner.Err(); e != nil {
		return s, e
	}
	get := func(section, key string) string { return values[section+"."+key] }
	num := func(section, key string, fallback int) int {
		v := get(section, key)
		if n, e := strconv.Atoi(v); e == nil {
			return n
		}
		return fallback
	}
	yes := func(section, key string, fallback bool) bool {
		v := get(section, key)
		if v == "1" || strings.EqualFold(v, "true") {
			return true
		}
		if v == "0" || strings.EqualFold(v, "false") {
			return false
		}
		return fallback
	}
	date := get("settings", "startdate")
	if date != "" {
		if d, e := parseDate(date); e == nil {
			s.StartDate = d.Format("2006-01-02")
		}
	}
	s.MinGapMinutes = num("settings", "minbetween", s.MinGapMinutes)
	s.MaxGapMinutes = num("settings", "maxbetween", s.MaxGapMinutes)
	s.Quantity = num("settings", "qty", s.Quantity)
	s.WindowStart = num("settings", "validstarthour", s.WindowStart)
	s.WindowEnd = num("settings", "validendhour", s.WindowEnd)
	s.RestrictHours = yes("settings", "validatehourrange", s.RestrictHours)
	s.Continue = yes("settings", "keepfinaltimestamp", s.Continue)
	s.Replace = yes("settings", "clearlistwhengenerating", s.Replace)
	ftd := func(key string) bool { return yes("dateformattab", strings.ToLower(key), false) }
	// Respect the date ordering and separators from the Delphi dialog.
	component := func(pos int) string {
		suffix := strconv.Itoa(pos)
		switch {
		case ftd("day" + suffix):
			if ftd("day1digit") {
				return "2"
			}
			return "02"
		case ftd("month" + suffix):
			if ftd("monthfull") {
				return "January"
			}
			if ftd("monthabbreviated") {
				return "Jan"
			}
			if ftd("month1digit") {
				return "1"
			}
			return "01"
		default:
			if ftd("year4digit") {
				return "2006"
			}
			return "06"
		}
	}
	sep := func(pos int) string {
		suffix := strconv.Itoa(pos)
		for _, x := range []struct{ key, val string }{{"slash", "/"}, {"period", "."}, {"comma", ", "}, {"space", " "}, {"hyphen", "-"}} {
			if ftd(x.key + suffix) {
				return x.val
			}
		}
		return "-"
	}
	s.DateLayout = component(1) + sep(1) + component(2) + sep(2) + component(3)
	if ftd("dowabbreviated") {
		s.DateLayout = "Mon, " + s.DateLayout
	}
	if ftd("dowfull") {
		s.DateLayout = "Monday, " + s.DateLayout
	}
	ft := func(key string) bool { return yes("timeformattab", strings.ToLower(key), false) }
	hh, mm, ss := "15", "04", "05"
	if ft("timeformatampm") {
		hh = "03"
	}
	if ft("timeformatmilitary") {
		s.TimeLayout = "150405"
	} else {
		if ft("hours1digit") {
			if ft("timeformatampm") {
				hh = "3"
			}
			if !ft("timeformatampm") {
				hh = "15"
			}
		}
		if ft("minutes1digit") {
			mm = "4"
		}
		s.TimeLayout = hh + ":" + mm
		if !ft("secondsnone") {
			if ft("seconds1digit") {
				ss = "5"
			}
			s.TimeLayout += ":" + ss
		}
		if ft("timeformatampm") {
			s.TimeLayout += " PM"
		}
	}
	unquote := func(v string) string {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, "\"")
		// Original Delphi INI was ANSI; map the common copyright-style section marker.
		return strings.ReplaceAll(v, string([]byte{0xa4}), "¤")
	}
	before := unquote(get("formattab", "infront"))
	middle := unquote(get("formattab", "between"))
	after := unquote(get("formattab", "after"))
	// The default INI uses a custom, themed prefix, so preserve it.
	first, last := "{time}", "{date}"
	if num("formattab", "firstelement", 0) == 1 {
		first = "{date}"
	}
	if num("formattab", "lastelement", 1) == 0 {
		last = "{time}"
	}
	if get("formattab", "infront") != "" {
		s.Template = before + first + middle + last + after
	}
	if e := s.validate(); e != nil {
		return defaults(), fmt.Errorf("legacy config invalid: %w", e)
	}
	return s, nil
}
