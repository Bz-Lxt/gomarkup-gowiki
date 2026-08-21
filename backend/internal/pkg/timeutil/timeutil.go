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

func Parse(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}
