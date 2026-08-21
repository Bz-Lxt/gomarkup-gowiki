package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/response"
	"gowiki/internal/search"
)

type SearchHandler struct{ eng *search.Engine }

func NewSearchHandler(eng *search.Engine) *SearchHandler { return &SearchHandler{eng: eng} }

func (h *SearchHandler) Query(c *gin.Context) {
	q := c.Query("q")
	hits, err := h.eng.Search(q, 20)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"hits": hits, "analyzer": h.eng.Analyzer()})
}
