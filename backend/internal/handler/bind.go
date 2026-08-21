package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/response"
)

func bindJSON(c *gin.Context, dest any) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		response.Fail(c, apperr.New(apperr.CodeValidation, 400, "请求体不合法"))
		return false
	}
	return true
}
