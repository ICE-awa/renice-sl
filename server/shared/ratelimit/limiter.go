package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
)

//go:embed scripts/token_bucket.lua
var tokenBucketLua string

var tokenBucketScript = redis.NewScript(tokenBucketLua)

func Do(c context.Context,
	rdb *redis.Client,
	key string,
	capacity int,
	refillPerSecond int,
	token int) (bool, error) {
	res, err := tokenBucketScript.Run(
		c,
		rdb,
		[]string{key},
		capacity,
		refillPerSecond,
		token,
	).Result()
	if err != nil {
		return false, err
	}

	allowed := res.(int64) == 1
	return allowed, nil
}
