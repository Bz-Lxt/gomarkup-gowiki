package search

import (
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/go-ego/gse"
)

const (
	AnalyzerGSE  = "gowiki_gse"
	TokenizerGSE = "gowiki_gse"
	AnalyzerCJK  = "cjk"
)

var (
	segOnce sync.Once
	seg     gse.Segmenter
	segErr  error
)

func segmenter() (*gse.Segmenter, error) {
	segOnce.Do(func() {
		segErr = seg.LoadDict()
	})
	if segErr != nil {
		return nil, segErr
	}
	return &seg, nil
}

type gseTokenizer struct {
	seg *gse.Segmenter
}

func (t *gseTokenizer) Tokenize(input []byte) analysis.TokenStream {
	src := string(input)
	parts := t.seg.CutSearch(src, true)
	stream := make(analysis.TokenStream, 0, len(parts))
	pos := 0
	cursor := 0
	termBuf := make([]byte, 0, 32)
	for _, p := range parts {
		if p == "" {
			continue
		}
		idx := strings.Index(src[cursor:], p)
		if idx < 0 {
			idx = 0
		}
		start := cursor + idx
		end := start + len(p)
		termBuf = append(termBuf[:0], p...)
		stream = append(stream, &analysis.Token{
			Term:     termBuf,
			Start:    start,
			End:      end,
			Position: pos + 1,
		})
		pos++
		cursor = end
	}
	return stream
}

func tokenizerConstructor(_ map[string]interface{}, _ *registry.Cache) (analysis.Tokenizer, error) {
	s, err := segmenter()
	if err != nil {
		return nil, err
	}
	return &gseTokenizer{seg: s}, nil
}

func analyzerConstructor(_ map[string]interface{}, cache *registry.Cache) (analysis.Analyzer, error) {
	tk, err := cache.TokenizerNamed(TokenizerGSE)
	if err != nil {
		return nil, err
	}
	lf, err := cache.TokenFilterNamed(lowercase.Name)
	if err != nil {
		return nil, err
	}
	return &analysis.DefaultAnalyzer{
		Tokenizer:    tk,
		TokenFilters: []analysis.TokenFilter{lf},
	}, nil
}

func init() {
	_ = registry.RegisterTokenizer(TokenizerGSE, tokenizerConstructor)
	_ = registry.RegisterAnalyzer(AnalyzerGSE, analyzerConstructor)
}
