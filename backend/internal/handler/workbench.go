package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/middleware"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type WorkbenchHandler struct{ svc *service.WorkbenchService }

func NewWorkbenchHandler(svc *service.WorkbenchService) *WorkbenchHandler {
	return &WorkbenchHandler{svc: svc}
}

func (h *WorkbenchHandler) Home(c *gin.Context) {
	out, err := h.svc.Home(middleware.UserID(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *WorkbenchHandler) ToggleFavorite(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	on, err := h.svc.ToggleFavorite(middleware.UserID(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"favorite": on})
}
