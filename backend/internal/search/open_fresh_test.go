package search_test

import (
	"path/filepath"
	"testing"

	"gowiki/internal/search"
)

func TestOpenCreatesMissingIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "indexes", "wiki")
	eng, err := search.Open(path, "cjk")
	if err != nil {
		t.Fatalf("open a fresh index: %v", err)
	}
	defer eng.Close()

	if err := eng.Upsert(search.Doc{ID: "first", Title: "first page"}); err != nil {
		t.Fatalf("write the fresh index: %v", err)
	}
	hits, err := eng.Search("first", 10)
	if err != nil {
		t.Fatalf("search the fresh index: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != "first" {
		t.Fatalf("fresh index did not retain its first document: %#v", hits)
	}
}
