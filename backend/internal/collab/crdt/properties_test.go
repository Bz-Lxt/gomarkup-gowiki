package crdt

import (
	"math/rand"
	"testing"
)

func TestInterleaveInsertDelete(t *testing.T) {
	a := NewDoc(1)
	_, _ = a.LocalInsert(0, "知识库协同编辑")
	_, _ = a.LocalDelete(2, 2)
	_, _ = a.LocalInsert(2, "Wiki")
	if a.Text() != "知识Wiki同编辑" {
		t.Fatalf("%q", a.Text())
	}
}

func TestReplicaPairwiseShuffle(t *testing.T) {
	src := NewDoc(1)
	var ops []Op
	for _, s := range []string{"Go", "语雀", "Notion"} {
		part, _ := src.LocalInsert(len([]rune(src.Text())), s)
		ops = append(ops, part...)
	}
	del, _ := src.LocalDelete(2, 2)
	ops = append(ops, del...)

	for seed := int64(1); seed <= 8; seed++ {
		rng := rand.New(rand.NewSource(seed))
		perm := append([]Op{}, ops...)
		rng.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
		d := NewDoc(uint64(20 + seed))
		for _, op := range perm {
			if err := d.Apply(op); err != nil {
				t.Fatal(err)
			}
		}
		if d.Text() != src.Text() {
			t.Fatalf("seed %d %q vs %q", seed, d.Text(), src.Text())
		}
	}
}

func TestEmptyAndUnicode(t *testing.T) {
	d := NewDoc(7)
	if d.Text() != "" {
		t.Fatal("empty")
	}
	_, err := d.LocalInsert(0, "🙂协同")
	if err != nil {
		t.Fatal(err)
	}
	if []rune(d.Text())[0] != '🙂' {
		t.Fatalf("%q", d.Text())
	}
}

func TestCloneIndependence(t *testing.T) {
	a := NewDoc(1)
	_, _ = a.LocalInsert(0, "abc")
	b := Clone(a, 2)
	_, _ = b.LocalInsert(3, "d")
	if a.Text() == b.Text() {
		t.Fatal("clone should diverge after local edit")
	}
	if a.Text() != "abc" {
		t.Fatalf("src mutated %q", a.Text())
	}
}
