package repository

import (
	"testing"

	"github.com/google/uuid"
)

func TestPathJoinAndCycle(t *testing.T) {
	root := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	child := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	rp := PathJoin("/", root)
	cp := PathJoin(rp, child)
	if rp != "/11111111-1111-1111-1111-111111111111/" {
		t.Fatalf("root path %s", rp)
	}
	if !IsAncestorPath(rp, cp) {
		t.Fatal("child path should be under root")
	}
	if IsAncestorPath(cp, rp) {
		t.Fatal("root should not be under child")
	}
	if !IsAncestorPath(rp, rp) {
		t.Fatal("node is ancestor of itself via prefix")
	}
}
