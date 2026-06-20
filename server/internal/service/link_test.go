package service

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/ICE-awa/renice-sl/shared/cache"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

// --- Mock Repository ---

type mockLinkRepo struct {
	getCacheByCodeFn func(ctx context.Context, code string) (*dtov1.LinkCache, error)
	callCount        atomic.Int64
}

func (m *mockLinkRepo) GetLinkCacheByCode(ctx context.Context, code string) (*dtov1.LinkCache, error) {
	m.callCount.Add(1)
	if m.getCacheByCodeFn != nil {
		return m.getCacheByCodeFn(ctx, code)
	}
	return nil, pgx.ErrNoRows
}

// 以下方法仅满足接口，不在 Redirect 测试中使用
func (m *mockLinkRepo) CreateLink(context.Context, *dtov1.CreateLinkReq) (int64, error) {
	return 0, nil
}
func (m *mockLinkRepo) GetLinks(context.Context, *dtov1.GetLinksReq) (*dtov1.GetLinksResp, error) {
	return nil, nil
}
func (m *mockLinkRepo) UpdateLink(context.Context, *dtov1.UpdateLinkReq) (string, error) {
	return "", nil
}
func (m *mockLinkRepo) GetLinkByID(context.Context, int64, int64) (*model.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) DeleteLink(context.Context, *dtov1.DeleteLinkReq) (string, error) {
	return "", nil
}
func (m *mockLinkRepo) CheckCodeConflict(context.Context, string) (bool, error) { return false, nil }
func (m *mockLinkRepo) RecordClick(context.Context, *dtov1.ClickLinkReq) error  { return nil }
func (m *mockLinkRepo) GetViewCountByUserID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *mockLinkRepo) GetLinkCountByUserID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (m *mockLinkRepo) GetAllLinkCodes(context.Context) ([]string, error) { return nil, nil }
func (m *mockLinkRepo) RecordLinkCheck(ctx context.Context, code string, status string) error {
	return nil
}

// --- Helper ---

func newTestService(t *testing.T, repo *mockLinkRepo) (*linkService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	bloom := cache.NewBloomFilter(10000, 0.01)
	cfg := &config.LinkConfig{
		Expires:     5 * time.Minute,
		NullExpires: 1 * time.Minute,
	}
	svc := &linkService{
		repo:  repo,
		rdb:   rdb,
		cfg:   cfg,
		bloom: bloom,
	}
	return svc, mr
}

// --- Tests ---

// Bloom Filter 拦截：code 不在 bloom 中，直接返回 not found
func TestRedirect_BloomFilterReject(t *testing.T) {
	t.Parallel()
	repo := &mockLinkRepo{}
	svc, _ := newTestService(t, repo)

	// 不添加任何 code 到 bloom
	req := &dtov1.ClickLinkReq{Code: "notexist"}
	_, err := svc.Redirect(context.Background(), req)
	if err == nil || err.Error() != consts.ErrLinkNotFound.Error() {
		t.Fatalf("expected ErrLinkNotFound, got %v", err)
	}

	// repo 不应被调用
	if repo.callCount.Load() != 0 {
		t.Fatal("repo should not be called when bloom filter rejects")
	}
}

// Redis 缓存命中：直接返回，不查 DB
func TestRedirect_RedisCacheHit(t *testing.T) {
	t.Parallel()
	repo := &mockLinkRepo{}
	svc, mr := newTestService(t, repo)

	code := "cached1"
	svc.bloom.Add(code)

	// 预热 Redis 缓存
	linkCache := &dtov1.LinkCache{
		OriginalURL:  "https://example.com",
		Status:       "active",
		SafetyStatus: "safe",
		ExpiresAt:    nil,
	}
	data, _ := json.Marshal(linkCache)
	mr.Set(consts.RedisLinkCodeKey+code, string(data))

	req := &dtov1.ClickLinkReq{Code: code, SkipStats: true}
	url, err := svc.Redirect(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com" {
		t.Fatalf("expected https://example.com, got %s", url)
	}
	if repo.callCount.Load() != 0 {
		t.Fatal("repo should not be called on cache hit")
	}
}

// 缓存未命中 → 回源 DB → 写缓存
func TestRedirect_CacheMiss_FallbackToDB(t *testing.T) {
	t.Parallel()
	repo := &mockLinkRepo{
		getCacheByCodeFn: func(ctx context.Context, code string) (*dtov1.LinkCache, error) {
			return &dtov1.LinkCache{
				OriginalURL:  "https://db-result.com",
				Status:       "active",
				SafetyStatus: "safe",
			}, nil
		},
	}
	svc, mr := newTestService(t, repo)

	code := "miss1"
	svc.bloom.Add(code)

	req := &dtov1.ClickLinkReq{Code: code, SkipStats: true}
	url, err := svc.Redirect(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://db-result.com" {
		t.Fatalf("expected https://db-result.com, got %s", url)
	}
	if repo.callCount.Load() != 1 {
		t.Fatalf("expected exactly 1 DB call, got %d", repo.callCount.Load())
	}

	// 验证缓存已写入 Redis
	val, err := mr.Get(consts.RedisLinkCodeKey + code)
	if err != nil {
		t.Fatalf("cache should be written to Redis: %v", err)
	}
	if val == "" {
		t.Fatal("cached value should not be empty")
	}
}

// 空对象缓存：code 在 bloom 中但 DB 不存在 → 缓存 NullLink
func TestRedirect_NullObjectCache(t *testing.T) {
	t.Parallel()
	repo := &mockLinkRepo{
		getCacheByCodeFn: func(ctx context.Context, code string) (*dtov1.LinkCache, error) {
			return nil, pgx.ErrNoRows
		},
	}
	svc, mr := newTestService(t, repo)

	code := "ghost1"
	svc.bloom.Add(code) // bloom 误判放行

	req := &dtov1.ClickLinkReq{Code: code, SkipStats: true}
	_, err := svc.Redirect(context.Background(), req)
	if err == nil || err.Error() != consts.ErrLinkNotFound.Error() {
		t.Fatalf("expected ErrLinkNotFound, got %v", err)
	}

	// 验证 Redis 中写入了空对象
	val, redisErr := mr.Get(consts.RedisLinkCodeKey + code)
	if redisErr != nil {
		t.Fatalf("null object should be cached: %v", redisErr)
	}
	var cached dtov1.LinkCache
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		t.Fatalf("failed to unmarshal cached null: %v", err)
	}
	if cached.OriginalURL != consts.NullLink {
		t.Fatalf("expected NullLink marker, got %s", cached.OriginalURL)
	}
}

// Singleflight 合并：100 个并发请求同一个 code，只触发 1 次 DB 查询
func TestRedirect_SingleflightDedup(t *testing.T) {
	t.Parallel()

	repo := &mockLinkRepo{
		getCacheByCodeFn: func(ctx context.Context, code string) (*dtov1.LinkCache, error) {
			// 模拟 DB 查询延迟
			time.Sleep(50 * time.Millisecond)
			return &dtov1.LinkCache{
				OriginalURL:  "https://singleflight.com",
				Status:       "active",
				SafetyStatus: "safe",
			}, nil
		},
	}
	svc, _ := newTestService(t, repo)

	code := "sf1"
	svc.bloom.Add(code)

	goroutines := 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make([]error, goroutines)
	urls := make([]string, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			req := &dtov1.ClickLinkReq{Code: code, SkipStats: true}
			url, err := svc.Redirect(context.Background(), req)
			errs[idx] = err
			urls[idx] = url
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, err)
		}
		if urls[i] != "https://singleflight.com" {
			t.Fatalf("goroutine %d: expected https://singleflight.com, got %s", i, urls[i])
		}
	}

	// 核心断言：DB 只被调用了 1 次（singleflight 合并了所有并发请求）
	dbCalls := repo.callCount.Load()
	if dbCalls != 1 {
		t.Fatalf("expected exactly 1 DB call (singleflight dedup), got %d", dbCalls)
	}
	t.Logf("singleflight: %d goroutines, %d DB call(s)", goroutines, dbCalls)
}

// validateLinkCache: 各种无效状态被正确拒绝
func TestValidateLinkCache(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cache   *dtov1.LinkCache
		wantErr error
	}{
		{
			name:    "NullLink",
			cache:   &dtov1.LinkCache{OriginalURL: consts.NullLink},
			wantErr: consts.ErrLinkNotFound,
		},
		{
			name:    "inactive",
			cache:   &dtov1.LinkCache{OriginalURL: "https://x.com", Status: "inactive"},
			wantErr: consts.ErrLinkInactive,
		},
		{
			name:    "unsafe",
			cache:   &dtov1.LinkCache{OriginalURL: "https://x.com", Status: "active", SafetyStatus: "unsafe"},
			wantErr: consts.ErrLinkUnsafe,
		},
		{
			name:    "pending",
			cache:   &dtov1.LinkCache{OriginalURL: "https://x.com", Status: "active", SafetyStatus: "pending"},
			wantErr: consts.ErrLinkPending,
		},
		{
			name:    "unknown",
			cache:   &dtov1.LinkCache{OriginalURL: "https://x.com", Status: "active", SafetyStatus: "unknown"},
			wantErr: consts.ErrLinkUnknown,
		},
		{
			name: "expired",
			cache: &dtov1.LinkCache{
				OriginalURL:  "https://x.com",
				Status:       "active",
				SafetyStatus: "safe",
				ExpiresAt:    timePtr(time.Now().Add(-1 * time.Hour)),
			},
			wantErr: consts.ErrLinkExpired,
		},
		{
			name: "valid",
			cache: &dtov1.LinkCache{
				OriginalURL:  "https://valid.com",
				Status:       "active",
				SafetyStatus: "safe",
				ExpiresAt:    timePtr(time.Now().Add(24 * time.Hour)),
			},
			wantErr: nil,
		},
	}

	svc := &linkService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := svc.validateLinkCache(tt.cache)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
