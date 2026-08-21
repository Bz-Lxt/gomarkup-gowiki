package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"gowiki/internal/logger"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.L().Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ms", time.Since(start).Milliseconds(),
		)
	}
}

func Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rec any) {
		logger.L().Error("panic recovered", "err", rec, "path", c.Request.URL.Path)
		c.AbortWithStatusJSON(500, gin.H{"code": "INTERNAL", "message": "服务器内部错误"})
	})
}
