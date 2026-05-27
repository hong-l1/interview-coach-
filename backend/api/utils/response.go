package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess = 0
	CodeError   = 1
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func newResponse(code int, message string, data any) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func writeJSON(c *gin.Context, status int, resp Response) {
	c.JSON(status, resp)
}

func Success(c *gin.Context, data any) {
	writeJSON(c, http.StatusOK, newResponse(CodeSuccess, "success", data))
}

func SuccessWithMessage(c *gin.Context, message string, data any) {
	writeJSON(c, http.StatusOK, newResponse(CodeSuccess, message, data))
}

func Error(c *gin.Context, status int, message string) {
	writeJSON(c, status, newResponse(CodeError, message, nil))
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func NotImplemented(c *gin.Context, message string) {
	Error(c, http.StatusNotImplemented, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
