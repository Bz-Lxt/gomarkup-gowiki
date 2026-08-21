package search

import (
	"errors"
	"fmt"
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

// errIndexUnusable is returned when the underlying bleve index is nil.
// It is a last line of defense so that a misconfigured Engine fails its
// operations with an error instead of dereferencing a nil pointer.
var errIndexUnusable = errors.New("search index is not available")

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
	if idx == nil {
		return nil, fmt.Errorf("search index at %q is not usable: %w", path, errIndexUnusable)
	}
	return &Engine{idx: idx, analyzer: name}, nil
}

func openOrCreate(path, analyzer string) (bleve.Index, error) {
	idx, err := bleve.Open(path)
	if err == nil {
		return idx, nil
	}
	if !errors.Is(err, bleve.ErrorIndexPathDoesNotExist) && !isBrokenIndex(path, err) {
		return nil, err
	}
	// Broken or partially written index directory (e.g. metadata missing
	// after an interrupted shutdown). Tear it down and rebuild a fresh one so
	// that self-recovery yields a usable index instead of a nil handle that
	// crashes on the first write.
	if isBrokenIndex(path, err) {
		logger.L().Warn("removing broken bleve index, recreating", "path", path, "err", err)
		if rmErr := os.RemoveAll(path); rmErr != nil {
			return nil, fmt.Errorf("remove broken index %q: %w", path, rmErr)
		}
	}
	return bleve.New(path, buildMapping(analyzer))
}

func isBrokenIndex(path string, err error) bool {
	if err == nil {
		return false
	}
	// Metadata missing/corrupt means the directory exists but the index was not
	// fully written (e.g. interrupted shutdown): worth tearing down & recreating.
	if errors.Is(err, bleve.ErrorIndexMetaMissing) || errors.Is(err, bleve.ErrorIndexMetaCorrupt) {
		return true
	}
	// A non-empty index directory that still failed to open is likely a torn
	// store; the first-run case (path absent entirely) is handled separately
	// via ErrorIndexPathDoesNotExist and must NOT be treated as broken.
	msg := err.Error()
	if strings.Contains(msg, "metadata missing") || strings.Contains(msg, "metadata corrupt") {
		return true
	}
	st, e := os.Stat(path)
	if e != nil || !st.IsDir() {
		return false
	}
	entries, e := os.ReadDir(path)
	return e == nil && len(entries) > 0
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
	if e.idx == nil {
		return errIndexUnusable
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = timeutil.Now()
	}
	return e.idx.Index(d.ID, d)
}

func (e *Engine) Delete(id uuid.UUID) error {
	if e.idx == nil {
		return errIndexUnusable
	}
	return e.idx.Delete(id.String())
}

func (e *Engine) Close() error {
	if e.idx == nil {
		return nil
	}
	return e.idx.Close()
}

func (e *Engine) Analyzer() string { return e.analyzer }
