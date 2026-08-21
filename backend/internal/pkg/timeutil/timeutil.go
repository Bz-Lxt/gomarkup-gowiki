package timeutil

import "time"

// Beijing is GMT+8. All business timestamps use this zone.
var Beijing = time.FixedZone("CST", 8*60*60)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func Format(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(Beijing).Format("2006-01-02 15:04:05")
}

// Parse is the inverse of Format: it interprets the naive wall-clock string
// as Beijing time, so that Format(t) -> Parse -> t preserves the instant.
// An empty string (the Format output of the zero time) yields the zero time.
func Parse(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", s, Beijing)
}
