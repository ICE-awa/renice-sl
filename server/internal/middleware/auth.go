package middleware

import (
	"github.com/ICE-awa/renice-sl/internal/consts"
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

func extractToken(c *gin.Context) string {
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		return token
	}

	return ""
}
