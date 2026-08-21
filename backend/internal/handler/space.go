package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/middleware"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type SpaceHandler struct{ svc *service.SpaceService }

func NewSpaceHandler(svc *service.SpaceService) *SpaceHandler { return &SpaceHandler{svc: svc} }

func (h *SpaceHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *SpaceHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	sp, err := h.svc.Create(middleware.UserID(c), req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, sp)
}

func (h *SpaceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	sp, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, sp)
}

func (h *SpaceHandler) Rename(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	sp, err := h.svc.Rename(id, req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, sp)
}
