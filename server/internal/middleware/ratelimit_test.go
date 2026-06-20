package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ICE-awa/renice-sl/internal/consts"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func newRateLimitRedis(t *testing.T) *redis.Client {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
	})
	return rdb
}

func TestRateLimitByIPAllowsWithinCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/limited", RateLimitByIP(newRateLimitRedis(t)), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestRateLimitByUserIDRejectsOverCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rdb := newRateLimitRedis(t)
	r := gin.New()
	r.GET("/limited",
		func(c *gin.Context) {
			c.Set("user_id", int64(42))
			c.Next()
		},
		RateLimitByUserID(rdb),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected status 204, got %d", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if int(body["code"].(float64)) != consts.CodeRateLimitExceeded {
		t.Fatalf("expected rate limit code, got %#v", body["code"])
	}
}
