package cache

import (
	"context"
	"fmt"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.ADDR(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(c).Err(); err != nil {
		return nil, fmt.Errorf("ping redis error: %w", err)
	}

	return client, nil
}
