package search_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"gowiki/internal/search"
)

func TestOpenRecoversIncompleteIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wiki.bleve")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	eng, err := search.Open(path, "cjk")
	if err != nil {
		t.Fatalf("open incomplete index: %v", err)
	}
	id := uuid.New().String()
	if err := eng.Upsert(search.Doc{ID: id, Title: "恢复索引", Content: "协作文档"}); err != nil {
		t.Fatalf("index document after recovery: %v", err)
	}
	hits, err := eng.Search("恢复", 10)
	if err != nil {
		t.Fatalf("search recovered index: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != id {
		t.Fatalf("search after recovery returned %#v", hits)
	}
	if err := eng.Close(); err != nil {
		t.Fatalf("close recovered index: %v", err)
	}
}
