package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/middleware"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req loginReq
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Register(req.Email, req.Password, req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, out)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AuthHandler) Me(c *gin.Context) {
	response.OK(c, gin.H{
		"id":    c.GetString(middleware.CtxUserID),
		"name":  c.GetString(middleware.CtxName),
		"email": c.GetString(middleware.CtxEmail),
	})
}
