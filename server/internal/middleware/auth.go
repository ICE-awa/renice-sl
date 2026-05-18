package middleware

import (
	"github.com/ICE-awa/renice-sl/internal/consts"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthRequired(cfg *config.JwtConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    consts.CodeUnauthorized,
				"message": "invalid token",
			})
			return
		}

		claims, err := util.ParseAccessToken(cfg, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    consts.CodeUnauthorized,
				"message": "invalid or expired token",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func AdminRequired(repo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		user, err := repo.FindUserByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    consts.CodeInternalServerError,
				"message": "Server Temporarily Unavailable",
			})
			return
		}

		if user.Role != consts.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    consts.CodeForbidden,
				"message": "You do not have permission to access this resource.",
			})
			return
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		return token
	}

	return ""
}
