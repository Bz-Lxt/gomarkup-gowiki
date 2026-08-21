package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
	"gowiki/internal/service"
)

const (
	CtxUserID = "uid"
	CtxName   = "uname"
	CtxEmail  = "uemail"
)

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			token = c.Query("token")
		}
		if token == "" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := auth.Parse(token)
		if err != nil || claims.TokenType == "refresh" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		if _, err := uuid.Parse(claims.UserID); err != nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxName, claims.DisplayName)
		c.Set(CtxEmail, claims.Email)
		c.Next()
	}
}

func UserID(c *gin.Context) uuid.UUID {
	id, _ := uuid.Parse(c.GetString(CtxUserID))
	return id
}
