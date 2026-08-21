package lock

import (
	"testing"
	"time"
)

func TestReleaseAllOnlyDropsDepartingUsersLocks(t *testing.T) {
	store := New(time.Minute)
	if _, ok := store.Acquire("doc", "intro", "alice", "Alice"); !ok {
		t.Fatal("Alice should acquire intro")
	}
	if _, ok := store.Acquire("doc", "summary", "alice", "Alice"); !ok {
		t.Fatal("Alice should acquire summary")
	}
	if _, ok := store.Acquire("doc", "details", "bob", "Bob"); !ok {
		t.Fatal("Bob should acquire details")
	}

	dropped := store.ReleaseAll("doc", "alice")
	if len(dropped) != 2 {
		t.Fatalf("released %d locks for departing user, want 2", len(dropped))
	}

	remaining := store.List("doc")
	if len(remaining) != 1 {
		t.Fatalf("remaining locks = %d, want 1", len(remaining))
	}
	if remaining[0].HolderID != "bob" || remaining[0].ParagraphID != "details" {
		t.Fatalf("remaining lock = %+v, want Bob's details lock", remaining[0])
	}
}
