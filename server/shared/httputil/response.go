package httputil

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Data:    data,
		Message: "ok",
	})
}

func OKWithMsg(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Data:    data,
		Message: message,
	})
}

func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, CodeForbidden, msg)
}

func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, CodeNotFound, msg)
}

func InternalServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, CodeInternalServerError, msg)
}

func ServiceUnavailable(c *gin.Context, msg string) {
	Fail(c, http.StatusServiceUnavailable, CodeServiceUnavailable, msg)
}
