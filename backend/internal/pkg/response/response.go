package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/timeutil"
)

type Body struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:      "OK",
		Message:   "ok",
		Data:      data,
		Timestamp: timeutil.Format(timeutil.Now()),
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{
		Code:      "CREATED",
		Message:   "created",
		Data:      data,
		Timestamp: timeutil.Format(timeutil.Now()),
	})
}

func Fail(c *gin.Context, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		c.JSON(ae.HTTP, Body{
			Code:      string(ae.Code),
			Message:   ae.Message,
			Timestamp: timeutil.Format(timeutil.Now()),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Body{
		Code:      string(apperr.CodeInternal),
		Message:   "服务器内部错误",
		Timestamp: timeutil.Format(timeutil.Now()),
	})
}
