package main

// Presets are independent of the Windows UI so output can be tested without
// opening a Win32 window. Index 0 is the current/custom setting.
type formatPreset struct {
	Template   string
	DateLayout string
	TimeLayout string
}

func presetByIndex(index int) (formatPreset, bool) {
	switch index {
	case 1:
		return formatPreset{Template: "{time}/{date}"}, true
	case 2:
		return formatPreset{Template: "({time}/{date})"}, true
	case 3:
		return formatPreset{Template: "{date}T{time}", DateLayout: "2006-01-02", TimeLayout: "15:04:05"}, true
	case 4:
		return formatPreset{Template: "{date} {time}"}, true
	default:
		return formatPreset{}, false
	}
}
