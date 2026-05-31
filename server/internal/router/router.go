package router

import (
	"github.com/ICE-awa/renice-sl/internal/handler"
	"github.com/ICE-awa/renice-sl/internal/middleware"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Setup(
	h *handler.Handlers,
	cfg *config.JwtConfig,
	rdb *redis.Client,
	userRepo repository.UserRepository,
) *gin.Engine {
	r := gin.New()

	public := r.Group("/api")
	{
		v1 := public.Group("/v1")
		{
			// auth
			v1.POST("/auth/register", middleware.RateLimitByIP(rdb), h.AuthHV1.Register)
			v1.POST("/auth/login", middleware.RateLimitByIP(rdb), h.AuthHV1.Login)
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
			v1.POST("/link", middleware.RateLimitByUserID(rdb), h.LinkHV1.CreateLink)
			v1.PUT("/link/:id", h.LinkHV1.UpdateLink)
			v1.DELETE("/link/:id", h.LinkHV1.DeleteLink)

			// stats
			v1.GET("/stats", h.LinkHV1.GetStats)

			// admin
			admin := v1.Group("/admin")
			admin.Use(middleware.AdminRequired(userRepo))
			// admin stats
			admin.GET("/stats/link", h.StatHV1.GetLinkStats)
			admin.GET("/stats/user", h.StatHV1.GetUserStats)
			admin.GET("/stats/click", h.StatHV1.GetClickStats)
			// admin dlq
			admin.GET("/dlq", h.DLQHV1.GetDLQMessages)
			admin.POST("/dlq/retry/:id", h.DLQHV1.RetryDLQMessage)
			admin.POST("/dlq/resolve/:id", h.DLQHV1.MarkAsResolved)
		}
	}

	health := r.Group("/api")
	{
		health.GET("/healthz", h.HealthH.Healthz)
		health.GET("/readyz", h.HealthH.Readyz)
	}

	return r
}
