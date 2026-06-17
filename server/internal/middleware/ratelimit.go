package middleware

import (
	"github.com/ICE-awa/renice-sl/internal/consts"
	"github.com/ICE-awa/renice-sl/shared/ratelimit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strconv"
)

func RateLimitByIP(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := consts.RedisRateLimit + "ip:" + ip

		ok, err := ratelimit.Do(c, rdb, key, 20, 5, 1)
		if err != nil {
			slog.Warn("[ratelimit] Rate Limit Server Error",
				slog.String("error", err.Error()))
			c.Next()
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    consts.CodeRateLimitExceeded,
				"message": "You have reached the rate limit. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

func RateLimitByUserID(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		key := consts.RedisRateLimit + "user:" + strconv.FormatInt(userID, 10)

		ok, err := ratelimit.Do(c, rdb, key, 20, 10, 1)
		if err != nil {
			slog.Warn("[ratelimit] Rate Limit Server Error",
				slog.String("error", err.Error()))
			c.Next()
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    consts.CodeRateLimitExceeded,
				"message": "You have reached the rate limit. Please try again later.",
			})
			return
		}
		c.Next()
	}
}
