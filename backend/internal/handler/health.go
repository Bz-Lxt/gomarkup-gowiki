package handler

import (
	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/response"
	"gowiki/internal/pkg/timeutil"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "ok",
		"time":   timeutil.Format(timeutil.Now()),
		"tz":     "Asia/Shanghai",
	})
}
