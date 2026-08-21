package search

import (
	"github.com/blevesearch/bleve/v2"
)

type Hit struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Fragment  string  `json:"fragment"`
	Score     float64 `json:"score"`
	SpaceID   string  `json:"spaceId"`
}

func (e *Engine) Search(q string, size int) ([]Hit, error) {
	if e.idx == nil {
		return nil, errIndexUnusable
	}
	if size <= 0 || size > 50 {
		size = 20
	}
	q = trimQuery(q)
	if q == "" {
		return []Hit{}, nil
	}
	titleQ := bleve.NewMatchQuery(q)
	titleQ.SetField("title")
	titleQ.SetBoost(3)
	contentQ := bleve.NewMatchQuery(q)
	contentQ.SetField("content")
	dis := bleve.NewDisjunctionQuery(titleQ, contentQ)
	req := bleve.NewSearchRequestOptions(dis, size, 0, false)
	req.Fields = []string{"title", "content", "spaceId"}
	req.Highlight = bleve.NewHighlight()
	req.Highlight.AddField("content")
	req.Highlight.AddField("title")
	res, err := e.idx.Search(req)
	if err != nil {
		return nil, err
	}
	out := make([]Hit, 0, len(res.Hits))
	for _, h := range res.Hits {
		hit := Hit{ID: h.ID, Score: h.Score}
		if v, ok := h.Fields["title"].(string); ok {
			hit.Title = v
		}
		if v, ok := h.Fields["spaceId"].(string); ok {
			hit.SpaceID = v
		}
		hit.Fragment = firstFragment(h.Fragments)
		if hit.Fragment == "" {
			if v, ok := h.Fields["content"].(string); ok {
				hit.Fragment = clip(v, 160)
			}
		}
		out = append(out, hit)
	}
	return out, nil
}

func firstFragment(frags map[string][]string) string {
	for _, key := range []string{"content", "title"} {
		if arr := frags[key]; len(arr) > 0 {
			return arr[0]
		}
	}
	return ""
}

func clip(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n]) + "…"
}

func trimQuery(q string) string {
	for len(q) > 0 && (q[0] == ' ' || q[0] == '\n' || q[0] == '\t') {
		q = q[1:]
	}
	return q
}
