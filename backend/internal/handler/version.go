package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/middleware"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type VersionHandler struct{ svc *service.VersionService }

func NewVersionHandler(svc *service.VersionService) *VersionHandler {
	return &VersionHandler{svc: svc}
}

func (h *VersionHandler) List(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	list, err := h.svc.List(id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *VersionHandler) Save(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	var req struct {
		Label string `json:"label"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	v, err := h.svc.SaveNamed(middleware.UserID(c), id, req.Label)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, v)
}

func (h *VersionHandler) Diff(c *gin.Context) {
	left, err := uuid.Parse(c.Query("left"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	right := c.Query("right")
	if right == "" || right == "current" {
		res, err := h.svc.DiffAgainstCurrent(left)
		if err != nil {
			response.Fail(c, err)
			return
		}
		response.OK(c, res)
		return
	}
	rid, err := uuid.Parse(right)
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	res, err := h.svc.Diff(left, rid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

func (h *VersionHandler) Rollback(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	v, err := h.svc.Rollback(middleware.UserID(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, v)
}
