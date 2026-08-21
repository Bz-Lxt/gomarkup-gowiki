package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/middleware"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type DocumentHandler struct {
	docs *service.DocumentService
	tree *service.TreeService
	wb   *service.WorkbenchService
}

func NewDocumentHandler(docs *service.DocumentService, tree *service.TreeService, wb *service.WorkbenchService) *DocumentHandler {
	return &DocumentHandler{docs: docs, tree: tree, wb: wb}
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var req struct {
		SpaceID  string  `json:"spaceId"`
		ParentID *string `json:"parentId"`
		Title    string  `json:"title"`
		Mode     string  `json:"editorMode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	sid, err := uuid.Parse(req.SpaceID)
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	var pid *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			response.Fail(c, apperr.ErrBadRequest)
			return
		}
		pid = &id
	}
	doc, err := h.docs.Create(middleware.UserID(c), sid, pid, req.Title, req.Mode)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, doc)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	doc, err := h.docs.Get(middleware.UserID(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	fav := h.wb.IsFavorite(middleware.UserID(c), id)
	response.OK(c, gin.H{"document": doc, "favorite": fav})
}

func (h *DocumentHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req struct {
		Title       *string `json:"title"`
		EditorMode  *string `json:"editorMode"`
		ContentMD   *string `json:"contentMd"`
		ContentJSON *string `json:"contentJson"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	doc, err := h.docs.UpdateMeta(middleware.UserID(c), id, req.Title, req.EditorMode, req.ContentMD, req.ContentJSON)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, doc)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.docs.Delete(middleware.UserID(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *DocumentHandler) Tree(c *gin.Context) {
	sid, err := uuid.Parse(c.Query("spaceId"))
	if err != nil {
		response.Fail(c, apperr.ErrBadRequest)
		return
	}
	list, err := h.docs.Tree(sid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *DocumentHandler) Move(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req struct {
		ParentID  *string `json:"parentId"`
		SortOrder *int    `json:"sortOrder"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	in := service.MoveInput{SortOrder: req.SortOrder}
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			response.Fail(c, apperr.ErrBadRequest)
			return
		}
		in.ParentID = &pid
	}
	doc, err := h.tree.Move(id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, doc)
}

func (h *DocumentHandler) Recycle(c *gin.Context) {
	var sid uuid.UUID
	if q := c.Query("spaceId"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			response.Fail(c, apperr.ErrBadRequest)
			return
		}
		sid = id
	}
	list, err := h.docs.Recycle(sid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *DocumentHandler) Restore(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	doc, err := h.docs.Restore(middleware.UserID(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, doc)
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, apperr.ErrBadRequest
	}
	return id, nil
}
