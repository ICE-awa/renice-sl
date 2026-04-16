package router

import (
	"github.com/ICE-awa/renice-sl/internal/handler"
	"github.com/ICE-awa/renice-sl/internal/middleware"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/gin-gonic/gin"
)

func Setup(h *handler.Handlers, cfg *config.JwtConfig) *gin.Engine {
	r := gin.New()

	public := r.Group("/api")
	{
		v1 := public.Group("/v1")
		{
			// auth
			v1.POST("/auth/register", h.AuthHV1.Register)
			v1.POST("/auth/login", h.AuthHV1.Login)
		}
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired(cfg))
	{
		v1 := protected.Group("/v1")
		{
			// auth
			v1.POST("/auth/refresh", h.AuthHV1.Refresh)
			v1.POST("/auth/logout", h.AuthHV1.Logout)
		}
	}

	health := r.Group("/api")
	{
		health.GET("/healthz", h.HealthH.Healthz)
		health.GET("/readyz", h.HealthH.Readyz)
	}

	return r
}
