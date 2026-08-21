package diff

import "testing"

func TestCompareInsertDelete(t *testing.T) {
	r := Compare("hello", "hallo")
	if len(r.Char) == 0 {
		t.Fatal("empty char diff")
	}
	var hasDel, hasIns bool
	for _, s := range r.Char {
		if s.Kind == OpDelete && s.Text == "e" {
			hasDel = true
		}
		if s.Kind == OpInsert && s.Text == "a" {
			hasIns = true
		}
	}
	if !hasDel || !hasIns {
		t.Fatalf("unexpected %#v", r.Char)
	}
}

func TestLineDiff(t *testing.T) {
	r := Compare("a\nb\nc", "a\nB\nc")
	joined := ""
	for _, s := range r.Line {
		if s.Kind != OpEqual {
			joined += string(s.Kind) + ":" + s.Text + ";"
		}
	}
	if joined == "" {
		t.Fatal("expected line change")
	}
}
