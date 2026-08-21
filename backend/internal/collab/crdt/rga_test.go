package crdt

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func applyAll(t *testing.T, d *Doc, ops []Op) {
	t.Helper()
	for _, op := range ops {
		if err := d.Apply(op); err != nil {
			t.Fatalf("apply %+v: %v", op, err)
		}
	}
}

func TestLocalInsertDelete(t *testing.T) {
	d := NewDoc(1)
	ops, err := d.LocalInsert(0, "你好世界")
	if err != nil {
		t.Fatal(err)
	}
	if d.Text() != "你好世界" {
		t.Fatalf("got %q", d.Text())
	}
	if len(ops) != 4 {
		t.Fatalf("ops=%d", len(ops))
	}
	if _, err := d.LocalDelete(2, 2); err != nil {
		t.Fatal(err)
	}
	if d.Text() != "你好" {
		t.Fatalf("got %q", d.Text())
	}
}

func TestIdempotentApply(t *testing.T) {
	a := NewDoc(1)
	ops, _ := a.LocalInsert(0, "abc")
	b := NewDoc(2)
	applyAll(t, b, ops)
	applyAll(t, b, ops)
	applyAll(t, b, ops)
	if a.Text() != b.Text() || b.Text() != "abc" {
		t.Fatalf("a=%q b=%q", a.Text(), b.Text())
	}
}

func TestCommutativity(t *testing.T) {
	origin := NewDoc(1)
	ops, _ := origin.LocalInsert(0, "协作")
	more, _ := origin.LocalInsert(2, "Wiki")
	ops = append(ops, more...)

	perms := [][]Op{append([]Op{}, ops...), nil}
	rev := append([]Op{}, ops...)
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	perms[1] = rev

	var texts []string
	for i, p := range perms {
		d := NewDoc(uint64(10 + i))
		applyAll(t, d, p)
		if d.PendingCount() != 0 {
			t.Fatalf("pending leftover on perm %d: %d", i, d.PendingCount())
		}
		texts = append(texts, d.Text())
	}
	if texts[0] != texts[1] {
		t.Fatalf("not commutative: %q vs %q", texts[0], texts[1])
	}
}

func TestConcurrentInsertSamePosition(t *testing.T) {
	baseOps := func() []Op {
		d := NewDoc(1)
		ops, _ := d.LocalInsert(0, "AB")
		return ops
	}
	base := baseOps()

	left := NewDoc(2)
	applyAll(t, left, base)
	lops, _ := left.LocalInsert(1, "X")

	right := NewDoc(3)
	applyAll(t, right, base)
	rops, _ := right.LocalInsert(1, "Y")

	mergeL := NewDoc(8)
	applyAll(t, mergeL, base)
	applyAll(t, mergeL, lops)
	applyAll(t, mergeL, rops)

	mergeR := NewDoc(9)
	applyAll(t, mergeR, base)
	applyAll(t, mergeR, rops)
	applyAll(t, mergeR, lops)

	if mergeL.Text() != mergeR.Text() {
		t.Fatalf("diverge %q vs %q", mergeL.Text(), mergeR.Text())
	}
	got := mergeL.Text()
	if got != "AXYB" && got != "AYXB" {
		t.Fatalf("unexpected merge %q", got)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	d := NewDoc(4)
	_, _ = d.LocalInsert(0, "快照")
	_, _ = d.LocalDelete(1, 1)
	raw, err := d.Snapshot().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseSnapshot(raw)
	if err != nil {
		t.Fatal(err)
	}
	n := NewDoc(4)
	n.LoadSnapshot(parsed)
	if n.Text() != d.Text() {
		t.Fatalf("got %q want %q", n.Text(), d.Text())
	}
}

func TestGCKeepsText(t *testing.T) {
	d := NewDoc(1)
	_, _ = d.LocalInsert(0, "abcdef")
	_, _ = d.LocalDelete(1, 4)
	before := d.Text()
	_ = d.GC()
	if d.Text() != before || before != "af" {
		t.Fatalf("text=%q after gc atoms=%d", d.Text(), d.AtomCount())
	}
}

func TestConvergenceRandom(t *testing.T) {
	const replicas = 10
	const rounds = 200
	rng := rand.New(rand.NewSource(42))

	docs := make([]*Doc, replicas)
	for i := 0; i < replicas; i++ {
		docs[i] = NewDoc(uint64(i + 1))
	}
	var all []Op
	for r := 0; r < rounds; r++ {
		src := docs[rng.Intn(replicas)]
		text := []rune(src.Text())
		var ops []Op
		if len(text) == 0 || rng.Intn(3) != 0 {
			idx := 0
			if len(text) > 0 {
				idx = rng.Intn(len(text) + 1)
			}
			s := string(rune('a' + rng.Intn(26)))
			ops, _ = src.LocalInsert(idx, s)
		} else {
			idx := rng.Intn(len(text))
			ops, _ = src.LocalDelete(idx, 1)
		}
		all = append(all, ops...)
	}

	// Every replica applies a different permutation of the full op log.
	want := ""
	for i := 0; i < replicas; i++ {
		perm := append([]Op{}, all...)
		rng.Shuffle(len(perm), func(a, b int) { perm[a], perm[b] = perm[b], perm[a] })
		d := NewDoc(uint64(100 + i))
		applyAll(t, d, perm)
		if d.PendingCount() != 0 {
			t.Fatalf("replica %d pending=%d", i, d.PendingCount())
		}
		if i == 0 {
			want = d.Text()
			continue
		}
		if d.Text() != want {
			t.Fatalf("replica %d text=%q want=%q", i, d.Text(), want)
		}
	}
}

func TestMalformedOpRejected(t *testing.T) {
	d := NewDoc(1)
	err := d.Apply(Op{Type: "boom", ID: ID{Site: 1, Clock: 1}})
	if err == nil {
		t.Fatal("expected error")
	}
	err = d.Apply(Op{Type: OpInsert, ID: StartID, After: StartID, Value: "x"})
	if err == nil {
		t.Fatal("expected start id reject")
	}
}

func TestDeterministicChildOrder(t *testing.T) {
	// Two inserts after Start with known IDs — higher ID sits closer to parent.
	d := NewDoc(1)
	ops := []Op{
		{Type: OpInsert, ID: ID{Site: 1, Clock: 1}, After: StartID, Value: "a"},
		{Type: OpInsert, ID: ID{Site: 2, Clock: 2}, After: StartID, Value: "b"},
	}
	applyAll(t, d, ops)
	if d.Text() != "ba" {
		t.Fatalf("got %q", d.Text())
	}
	rev := []Op{ops[1], ops[0]}
	e := NewDoc(3)
	applyAll(t, e, rev)
	if e.Text() != d.Text() {
		t.Fatalf("%q vs %q", e.Text(), d.Text())
	}
}

func TestSortHelperUsed(t *testing.T) {
	ids := []ID{{1, 3}, {2, 1}, {1, 1}}
	sort.Slice(ids, func(i, j int) bool { return ids[i].Less(ids[j]) })
	if ids[0].Clock != 1 || ids[2].Clock != 3 {
		t.Fatalf("%v", ids)
	}
	_ = fmt.Sprintf("%v", ids)
}
