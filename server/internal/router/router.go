package router

import (
	"github.com/ICE-awa/renice-sl/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(h *handler.Handlers) *gin.Engine {
	r := gin.New()

	api := r.Group("/api")
	api.GET("/healthz", h.HealthH.Healthz)
	api.GET("/readyz", h.HealthH.Readyz)

	return r
}
