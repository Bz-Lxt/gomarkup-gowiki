package timeutil

import "testing"

func TestBeijingOffset(t *testing.T) {
	_, off := Now().Zone()
	if off != 8*3600 {
		t.Fatalf("offset=%d", off)
	}
	s := Format(Now())
	if len(s) != 19 {
		t.Fatalf("format %q", s)
	}
}
