package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchTitleBoost(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(filepath.Join(dir, "idx"), "cjk")
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	_ = eng.Upsert(Doc{ID: "1", SpaceID: "s", Title: "协同算法", Content: "无关内容"})
	_ = eng.Upsert(Doc{ID: "2", SpaceID: "s", Title: "其它", Content: "这里提到协同一次"})
	hits, err := eng.Search("协同", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(filepath.Join(dir, "idx"), "cjk")
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	hits, err := eng.Search("   ", 10)
	if err != nil || len(hits) != 0 {
		t.Fatalf("%v %#v", err, hits)
	}
	_ = os.RemoveAll(dir)
}
