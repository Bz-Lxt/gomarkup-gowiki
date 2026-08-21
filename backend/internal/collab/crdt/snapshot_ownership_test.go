package crdt_test

import (
	"testing"

	"gowiki/internal/collab/crdt"
)

func TestLoadedSnapshotIsIndependentFromInput(t *testing.T) {
	source := crdt.NewDoc(10)
	if _, err := source.LocalInsert(0, "协作快照"); err != nil {
		t.Fatalf("insert source text: %v", err)
	}

	snapshot := source.Snapshot()
	restored := crdt.NewDoc(20)
	restored.LoadSnapshot(snapshot)
	want := restored.Text()

	for i := range snapshot.Atoms {
		snapshot.Atoms[i].Value = "替"
		snapshot.Atoms[i].Deleted = true
	}

	if got := restored.Text(); got != want {
		t.Fatalf("loaded document changed after the input snapshot was reused: got %q, want %q", got, want)
	}
}
