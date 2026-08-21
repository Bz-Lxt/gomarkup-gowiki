package diff

import "testing"

func TestEmptySides(t *testing.T) {
	r := Compare("", "abc")
	if len(r.Char) == 0 {
		t.Fatal("expected insert")
	}
	r = Compare("abc", "")
	if len(r.Char) == 0 {
		t.Fatal("expected delete")
	}
	r = Compare("", "")
	if len(r.Line)+len(r.Char) != 0 {
		t.Fatalf("%+v", r)
	}
}

func TestIdentical(t *testing.T) {
	r := Compare("same\nline", "same\nline")
	for _, s := range r.Line {
		if s.Kind != OpEqual {
			t.Fatalf("%+v", s)
		}
	}
}

func TestChineseLines(t *testing.T) {
	r := Compare("协同算法\n稳定", "协同引擎\n稳定")
	var saw bool
	for _, s := range r.Line {
		if s.Kind != OpEqual {
			saw = true
		}
	}
	if !saw {
		t.Fatal("expected change")
	}
}
