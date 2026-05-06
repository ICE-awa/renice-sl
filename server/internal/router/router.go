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
			v1.POST("/auth/refresh", h.AuthHV1.Refresh)

			// link
			v1.GET("/s/:code", h.LinkHV1.Redirect)
		}
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired(cfg))
	{
		v1 := protected.Group("/v1")
		{
			// auth
			v1.POST("/auth/logout", h.AuthHV1.Logout)
			v1.GET("/auth/me", h.AuthHV1.Me)

			// link
			v1.GET("/links", h.LinkHV1.GetLinks)
			v1.GET("/link/:id", h.LinkHV1.GetLinkByID)
			v1.POST("/link", h.LinkHV1.CreateLink)
			v1.PUT("/link/:id", h.LinkHV1.UpdateLink)
			v1.DELETE("/link/:id", h.LinkHV1.DeleteLink)
		}
	}

	health := r.Group("/api")
	{
		health.GET("/healthz", h.HealthH.Healthz)
		health.GET("/readyz", h.HealthH.Readyz)
	}

	return r
}
