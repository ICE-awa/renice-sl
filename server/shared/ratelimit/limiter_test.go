package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	return mr, rdb
}

func rewindBucket(t *testing.T, ctx context.Context, rdb *redis.Client, key string, d time.Duration) {
	t.Helper()

	now, err := rdb.Time(ctx).Result()
	if err != nil {
		t.Fatalf("failed to read redis time: %v", err)
	}
	if err := rdb.HSet(ctx, key, "updated_at", now.Add(-d).UnixMilli()).Err(); err != nil {
		t.Fatalf("failed to rewind bucket timestamp: %v", err)
	}
}

func TestDo_AllowWithinCapacity(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	capacity := 10
	for i := 0; i < capacity; i++ {
		allowed, err := Do(ctx, rdb, "test:basic", capacity, 1, 1)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: should be allowed within capacity", i)
		}
	}
}

func TestDo_RejectOverCapacity(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	capacity := 5
	for i := 0; i < capacity; i++ {
		_, _ = Do(ctx, rdb, "test:over", capacity, 1, 1)
	}

	allowed, err := Do(ctx, rdb, "test:over", capacity, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("should be rejected when over capacity")
	}
}

func TestDo_TokenRefill(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	key := "test:refill"
	capacity := 5
	refillPerSecond := 5

	for i := 0; i < capacity; i++ {
		_, _ = Do(ctx, rdb, key, capacity, refillPerSecond, 1)
	}

	allowed, _ := Do(ctx, rdb, key, capacity, refillPerSecond, 1)
	if allowed {
		t.Fatal("should be empty after exhausting capacity")
	}

	rewindBucket(t, ctx, rdb, key, time.Second)

	allowed, err := Do(ctx, rdb, key, capacity, refillPerSecond, 1)
	if err != nil {
		t.Fatalf("unexpected error after refill: %v", err)
	}
	if !allowed {
		t.Fatal("should be allowed after token refill")
	}
}

func TestDo_RefillCappedAtCapacity(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	key := "test:cap"
	capacity := 3
	refillPerSecond := 10

	_, _ = Do(ctx, rdb, key, capacity, refillPerSecond, 1)
	rewindBucket(t, ctx, rdb, key, 10*time.Second)

	for i := 0; i < capacity; i++ {
		allowed, err := Do(ctx, rdb, key, capacity, refillPerSecond, 1)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("request %d: should be allowed after refill capped at capacity", i)
		}
	}

	allowed, _ := Do(ctx, rdb, key, capacity, refillPerSecond, 1)
	if allowed {
		t.Fatal("should be rejected because refill must not exceed capacity")
	}
}

func TestDo_BurstCost(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	capacity := 10

	allowed, err := Do(ctx, rdb, "test:burst", capacity, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("burst of 10 within capacity 10 should be allowed")
	}

	allowed, _ = Do(ctx, rdb, "test:burst", capacity, 1, 1)
	if allowed {
		t.Fatal("should be rejected after burst exhausted all tokens")
	}
}

func TestDo_CostExceedsCapacity(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	allowed, err := Do(ctx, rdb, "test:bigcost", 5, 1, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("cost exceeding capacity should be rejected")
	}
}

func TestDo_IsolatedKeys(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, _ = Do(ctx, rdb, "test:key1", 3, 1, 1)
	}

	allowed, err := Do(ctx, rdb, "test:key2", 3, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("different keys should have independent token buckets")
	}
}

func TestDo_ConcurrentSafety(t *testing.T) {
	t.Parallel()
	_, rdb := setupMiniredis(t)
	ctx := context.Background()

	capacity := 20
	goroutines := 50

	results := make(chan bool, goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			allowed, err := Do(ctx, rdb, "test:concurrent", capacity, 0, 1)
			if err != nil {
				results <- false
				return
			}
			results <- allowed
		}()
	}

	allowedCount := 0
	for i := 0; i < goroutines; i++ {
		if <-results {
			allowedCount++
		}
	}

	if allowedCount > capacity {
		t.Errorf("allowed %d requests, but capacity is %d", allowedCount, capacity)
	}
}
