package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type UploadHandler struct{ svc *service.UploadService }

func NewUploadHandler(svc *service.UploadService) *UploadHandler { return &UploadHandler{svc: svc} }

func (h *UploadHandler) Image(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer f.Close()
	url, err := h.svc.Save(fh.Filename, fh.Size, f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, gin.H{"url": url})
}
