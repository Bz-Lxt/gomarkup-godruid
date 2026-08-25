package timeutil

import "time"

// Beijing is GMT+8, the project default civil timezone.
var Beijing = time.FixedZone("CST", 8*3600)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func InBeijing(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(Beijing)
}

func FormatDisplay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return InBeijing(t).Format("2006-01-02 15:04:05")
}

func FormatISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return InBeijing(t).Format(time.RFC3339)
}
