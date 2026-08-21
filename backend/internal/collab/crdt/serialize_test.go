package crdt

import "testing"

func TestSerializeOpsRoundTrip(t *testing.T) {
	d := NewDoc(1)
	ops, _ := d.LocalInsert(0, "序列化")
	raw, err := MarshalOps(ops)
	if err != nil {
		t.Fatal(err)
	}
	back, err := UnmarshalOps(raw)
	if err != nil {
		t.Fatal(err)
	}
	e := NewDoc(2)
	if err := e.ApplyAll(back); err != nil {
		t.Fatal(err)
	}
	if e.Text() != d.Text() {
		t.Fatalf("%q %q", e.Text(), d.Text())
	}
	st := d.Stats()
	if st.Visible != 3 || st.Site != 1 {
		t.Fatalf("%+v", st)
	}
}
