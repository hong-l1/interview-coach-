package server

import (
	"awesomeProject4/backend/api/handler"
	"github.com/gin-gonic/gin"
)

func NewGinEngine(handler *handler.Handler) *gin.Engine {
	engine := gin.New()
	handler.Register(engine)
	return engine
}
