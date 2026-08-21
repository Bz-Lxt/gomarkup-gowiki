package service

import (
	"testing"

	"github.com/google/uuid"

	"gowiki/internal/model"
	"gowiki/internal/repository"
)

func TestWouldCycle(t *testing.T) {
	root := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	child := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	rp := repository.PathJoin("/", root)
	cp := repository.PathJoin(rp, child)
	if !WouldCycle(rp, cp) {
		t.Fatal("moving root under child is a cycle")
	}
	if WouldCycle(cp, rp) {
		t.Fatal("moving child under root is fine")
	}
}

func TestSortTree(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	docs := []model.Document{
		{ID: b, ParentID: &a, Title: "child", SortOrder: 0},
		{ID: a, Title: "root", SortOrder: 0},
	}
	got := SortTree(docs)
	if len(got) != 2 || got[0].ID != a || got[1].ID != b {
		t.Fatalf("%+v", got)
	}
	ids := CollectIDs(got)
	if len(ids) != 2 {
		t.Fatal(ids)
	}
}
