package crdt

import (
	"math/rand"
	"testing"
)

func TestTwoSitePingPong(t *testing.T) {
	a := NewDoc(1)
	b := NewDoc(2)
	for i := 0; i < 40; i++ {
		src, dst := a, b
		if i%2 == 1 {
			src, dst = b, a
		}
		text := []rune(src.Text())
		var ops []Op
		if len(text) == 0 || i%3 != 0 {
			idx := 0
			if len(text) > 0 {
				idx = i % (len(text) + 1)
			}
			ops, _ = src.LocalInsert(idx, string(rune('A'+i%26)))
		} else {
			ops, _ = src.LocalDelete(i%len(text), 1)
		}
		for _, op := range ops {
			if err := dst.Apply(op); err != nil {
				t.Fatal(err)
			}
		}
	}
	if a.Text() != b.Text() {
		t.Fatalf("%q vs %q", a.Text(), b.Text())
	}
}

func TestManyShortWords(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	words := []string{"树", "检索", "版本", "回滚", "锁"}
	d := NewDoc(3)
	var all []Op
	for i := 0; i < 30; i++ {
		ops, _ := d.LocalInsert(rng.Intn(len([]rune(d.Text()))+1), words[i%len(words)])
		all = append(all, ops...)
	}
	e := NewDoc(4)
	for i := len(all) - 1; i >= 0; i-- {
		if err := e.Apply(all[i]); err != nil {
			t.Fatal(err)
		}
	}
	if e.Text() != d.Text() {
		t.Fatalf("%q vs %q", e.Text(), d.Text())
	}
}
