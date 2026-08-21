package timeutil_test

import (
	"testing"
	"time"

	"gowiki/internal/pkg/timeutil"
)

func TestFormatParseRoundTrip(t *testing.T) {
	original := time.Date(2026, time.August, 22, 3, 14, 15, 0, time.UTC)

	parsed, err := timeutil.Parse(timeutil.Format(original))
	if err != nil {
		t.Fatalf("parse formatted time: %v", err)
	}
	if !parsed.Equal(original) {
		t.Fatalf("round trip changed instant: original=%s parsed=%s", original.Format(time.RFC3339), parsed.Format(time.RFC3339))
	}
}
