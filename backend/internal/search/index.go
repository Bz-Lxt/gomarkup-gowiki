package search

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/lang/cjk"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/google/uuid"

	"gowiki/internal/logger"
	"gowiki/internal/pkg/timeutil"
)

type Doc struct {
	ID        string    `json:"id"`
	SpaceID   string    `json:"spaceId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Engine struct {
	idx      bleve.Index
	analyzer string
}

func Open(path, analyzer string) (*Engine, error) {
	name := AnalyzerGSE
	if strings.EqualFold(analyzer, "cjk") {
		name = AnalyzerCJK
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	idx, err := openOrCreate(path, name)
	if err != nil && name != AnalyzerCJK {
		logger.L().Warn("gse analyzer failed, falling back to cjk", "err", err)
		return Open(path+"-cjk", "cjk")
	}
	if err != nil {
		return nil, err
	}
	return &Engine{idx: idx, analyzer: name}, nil
}

func openOrCreate(path, analyzer string) (bleve.Index, error) {
	idx, err := bleve.Open(path)
	if err == nil {
		return idx, nil
	}
	if err != bleve.ErrorIndexPathDoesNotExist && !isBrokenIndex(path, err) {
		return nil, err
	}
	if isBrokenIndex(path, err) {
		_ = os.RemoveAll(path)
		return nil, nil
	}
	return bleve.New(path, buildMapping(analyzer))
}

func isBrokenIndex(path string, err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "metadata missing") || strings.Contains(msg, "does not exist") {
		return true
	}
	st, e := os.Stat(path)
	return e == nil && st.IsDir()
}

func buildMapping(analyzer string) mapping.IndexMapping {
	im := bleve.NewIndexMapping()
	im.DefaultAnalyzer = analyzer
	doc := bleve.NewDocumentMapping()
	title := bleve.NewTextFieldMapping()
	title.Analyzer = analyzer
	title.Store = true
	content := bleve.NewTextFieldMapping()
	content.Analyzer = analyzer
	content.Store = true
	space := bleve.NewKeywordFieldMapping()
	doc.AddFieldMappingsAt("title", title)
	doc.AddFieldMappingsAt("content", content)
	doc.AddFieldMappingsAt("spaceId", space)
	im.AddDocumentMapping("wiki", doc)
	im.DefaultType = "wiki"
	return im
}

func (e *Engine) Upsert(d Doc) error {
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = timeutil.Now()
	}
	return e.idx.Index(d.ID, d)
}

func (e *Engine) Delete(id uuid.UUID) error {
	return e.idx.Delete(id.String())
}

func (e *Engine) Close() error { return e.idx.Close() }

func (e *Engine) Analyzer() string { return e.analyzer }
