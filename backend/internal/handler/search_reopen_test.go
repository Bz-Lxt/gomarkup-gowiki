package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"gowiki/internal/handler"
	"gowiki/internal/search"
)

func TestSearchSurvivesIndexReopen(t *testing.T) {
	indexPath := filepath.Join(t.TempDir(), "wiki-index")
	eng, err := search.Open(indexPath, "cjk")
	if err != nil {
		t.Fatalf("create index: %v", err)
	}
	if err := eng.Upsert(search.Doc{
		ID: "doc-after-restart", SpaceID: "space-1",
		Title: "重启校验文档", Content: "这份内容应在服务重启后继续可检索",
	}); err != nil {
		t.Fatalf("index document: %v", err)
	}
	if err := eng.Close(); err != nil {
		t.Fatalf("close initial index: %v", err)
	}

	reopened, err := search.Open(indexPath, "cjk")
	if err != nil {
		t.Fatalf("reopen index: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/search", handler.NewSearchHandler(reopened).Query)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/search?q=重启校验", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("search after reopen returned HTTP %d: %s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Data struct {
			Hits []search.Hit `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode search response: %v", err)
	}
	if len(body.Data.Hits) != 1 || body.Data.Hits[0].ID != "doc-after-restart" {
		t.Fatalf("search after reopen lost indexed document: %#v", body.Data.Hits)
	}
	if err := reopened.Close(); err != nil {
		t.Fatalf("close reopened index: %v", err)
	}
}
