package utils

import (
	"github.com/gin-gonic/gin"
)

func SetupSSEResponse(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeaderNow()
}
func SendSSEvent(c *gin.Context, name string, payload any) {
	c.SSEvent(name, payload)
	c.Writer.Flush()
}
