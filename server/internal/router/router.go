package router

import (
	"github.com/ICE-awa/renice-sl/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(h *handler.Handlers) *gin.Engine {
	r := gin.New()

	r.Group("/api")
	r.GET("/healthz", h.HealthH.Healthz)
	r.GET("/readyz", h.HealthH.Readyz)

	return r
}
