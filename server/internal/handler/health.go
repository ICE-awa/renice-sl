package handler

import (
	"context"
	"github.com/ICE-awa/renice-sl/shared/httputil"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"time"
)

type HealthHandler struct {
	db         *pgxpool.Pool
	rdb        *redis.Client
	natsClient *mq.NatsClient
}

func NewHealthHandler(
	db *pgxpool.Pool,
	rdb *redis.Client,
	natsClient *mq.NatsClient,
) *HealthHandler {
	return &HealthHandler{
		db:         db,
		rdb:        rdb,
		natsClient: natsClient,
	}
}

// GET /api/healthz
func (h *HealthHandler) Healthz(c *gin.Context) {
	httputil.OK(c, nil)
}

// GET /api/readyz
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		httputil.ServiceUnavailable(c, "Failed to connect to database")
		return
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		httputil.ServiceUnavailable(c, "Failed to connect to Redis")
		return
	}

	if h.natsClient.Conn.Status() != nats.CONNECTED {
		httputil.ServiceUnavailable(c, "Failed to connect to NATS")
		return
	}

	httputil.OK(c, nil)
}
