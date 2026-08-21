package search_test

import (
	"path/filepath"
	"testing"

	"gowiki/internal/search"
)

func TestSearchKeepsEveryTokenFromDocument(t *testing.T) {
	eng, err := search.Open(filepath.Join(t.TempDir(), "idx"), "gse")
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()

	if err := eng.Upsert(search.Doc{
		ID:      "multi-token",
		SpaceID: "space-1",
		Title:   "release notes",
		Content: "alpha bravo",
	}); err != nil {
		t.Fatal(err)
	}

	searchFor := func(query string) []search.Hit {
		t.Helper()
		hits, err := eng.Search(query, 10)
		if err != nil {
			t.Fatal(err)
		}
		return hits
	}

	last := searchFor("bravo")
	if len(last) != 1 || last[0].ID != "multi-token" {
		t.Fatalf("searching the last token: got %#v, want document %q", last, "multi-token")
	}
	first := searchFor("alpha")
	if len(first) != 1 || first[0].ID != "multi-token" {
		t.Fatalf("searching the first token: got %#v, want document %q", first, "multi-token")
	}
}
