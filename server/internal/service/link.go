package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/cache"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"log/slog"
	"time"
)

type LinkService interface {
	CreateLink(context.Context, *dtov1.CreateLinkReq) error
	GetLinks(context.Context, *dtov1.GetLinksReq) ([]*dtov1.LinkItem, error)
	UpdateLink(context.Context, *dtov1.UpdateLinkReq) error
	GetLinkByID(context.Context, int64, int64) (*dtov1.LinkItem, error)
	DeleteLink(context.Context, *dtov1.DeleteLinkReq) error
	Redirect(context.Context, *dtov1.ClickLinkReq) (string, error)
	GetStats(context.Context, int64) (*dtov1.GetStatsResponse, error)
	InitBloomFilter() error
}

type linkService struct {
	repo           repository.LinkRepository
	publisher      *event.LinkPublisher
	rdb            *redis.Client
	cfg            *config.LinkConfig
	bloom          *cache.BloomFilter
	cacheFillGroup singleflight.Group
}

func NewLinkService(
	repo repository.LinkRepository,
	publisher *event.LinkPublisher,
	rdb *redis.Client,
	cfg *config.LinkConfig,
	bloom *cache.BloomFilter,
) LinkService {
	return &linkService{
		repo:      repo,
		publisher: publisher,
		rdb:       rdb,
		cfg:       cfg,
		bloom:     bloom,
	}
}

func (s *linkService) validateLinkCache(data *dtov1.LinkCache) bool {
	if data.OriginalURL == consts.NullLink ||
		data.Status == "inactive" ||
		(data.ExpiresAt != nil && data.ExpiresAt.Before(time.Now())) {
		return false
	}
	return true
}

func (s *linkService) cacheLink(c context.Context, req *dtov1.ClickLinkReq) (*dtov1.LinkCache, error) {
	c, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()

	value, err, _ := s.cacheFillGroup.Do(req.Code, func() (any, error) {
		// 再查询一次 Redis，避免其他请求已经缓存
		key := consts.RedisLinkCodeKey + req.Code
		cacheVal, err := s.rdb.Get(c, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			slog.Warn("failed to get original URL from Redis",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
		}
		cacheMiss := err != nil

		// 把 cacheVal unmarshal 成 LinkCache
		if !cacheMiss {
			var data dtov1.LinkCache
			if err := json.Unmarshal([]byte(cacheVal), &data); err != nil {
				slog.Warn("failed to unmarshal cache in cacheFillGroup",
					slog.String("code", req.Code),
					slog.String("error", err.Error()))
			} else {
				if !s.validateLinkCache(&data) {
					return nil, consts.ErrInvalidLink
				}
				return &data, nil
			}
		}

		data, err := s.repo.GetLinkCacheByCode(c, req.Code)
		if err != nil {
			// 若为空链接
			if errors.Is(err, pgx.ErrNoRows) {
				nullLink := &dtov1.LinkCache{
					OriginalURL: consts.NullLink,
					Status:      "",
					ExpiresAt:   nil,
				}
				// 尝试将此空链接缓存入 Redis 中
				linkCache, err := json.Marshal(nullLink)
				if err != nil {
					slog.Warn("failed to marshal null cache in cacheFillGroup",
						slog.String("code", req.Code),
						slog.String("error", err.Error()))
				} else {
					if err := s.rdb.Set(c, key, linkCache, s.cfg.NullExpires).Err(); err != nil {
						slog.Warn("failed to cache null link in Redis in cacheFillGroup",
							slog.String("code", req.Code),
							slog.String("error", err.Error()))
					}
				}
				return nil, consts.ErrInvalidLink
			}
			return nil, err
		}

		// 若没有报错 无论如何先写入当前快照到 Redis 中 避免缓存穿透
		linkCache, err := json.Marshal(data)
		if err != nil {
			slog.Warn("failed to marshal cache in cacheFillGroup",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
		} else {
			if err := s.rdb.Set(c, key, linkCache, s.cfg.Expires).Err(); err != nil {
				slog.Warn("failed to cache original URL in Redis in cacheFillGroup",
					slog.String("code", req.Code),
					slog.String("error", err.Error()))
			}
		}

		// 随后判断其合法性，若不合法则不放行
		if !s.validateLinkCache(data) {
			return nil, consts.ErrInvalidLink
		}

		// 若合法则返回 pg 回源链接
		return data, nil
	})

	if err != nil {
		return nil, err
	}

	data, ok := value.(*dtov1.LinkCache)
	if !ok {
		return nil, consts.ErrInvalidLink
	}

	return data, nil
}

func (s *linkService) CreateLink(c context.Context, req *dtov1.CreateLinkReq) error {
	for i := 0; i < 5; i++ {
		code, err := util.RandomCode(6)
		if err != nil {
			return err
		}
		exists, err := s.repo.CheckCodeConflict(c, code)
		if err != nil {
			return err
		}
		if !exists {
			req.Code = code
			break
		}
	}
	if req.Code == "" {
		return consts.ErrFailedToGenerateCode
	}

	_, err := s.repo.CreateLink(c, req)
	if err != nil {
		return err
	}

	// 将 code 添加至布隆过滤器中
	s.bloom.Add(req.Code)
	return nil
}

func (s *linkService) GetLinks(c context.Context, req *dtov1.GetLinksReq) ([]*dtov1.LinkItem, error) {
	return s.repo.GetLinks(c, req)
}

func (s *linkService) UpdateLink(c context.Context, req *dtov1.UpdateLinkReq) error {
	code, err := s.repo.UpdateLink(c, req)
	if err != nil {
		return err
	}

	key := consts.RedisLinkCodeKey + code
	if err := s.rdb.Del(c, key).Err(); err != nil {
		slog.Warn("failed to delete cache in Redis after link update",
			slog.String("code", code),
			slog.String("error", err.Error()))
	}
	return nil
}

func (s *linkService) GetLinkByID(c context.Context, id int64, userID int64) (*dtov1.LinkItem, error) {
	link, err := s.repo.GetLinkByID(c, id, userID)
	if err != nil {
		return nil, err
	}

	resp := &dtov1.LinkItem{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		Code:        link.Code,
		ViewCount:   link.ViewCount,
		Status:      link.Status,
		CreatedAt:   link.CreatedAt,
		UpdatedAt:   link.UpdatedAt,
		ExpiresAt:   link.ExpiresAt,
	}
	return resp, nil
}

func (s *linkService) DeleteLink(c context.Context, req *dtov1.DeleteLinkReq) error {
	code, err := s.repo.DeleteLink(c, req)
	if err != nil {
		return err
	}

	key := consts.RedisLinkCodeKey + code
	if err := s.rdb.Del(c, key).Err(); err != nil {
		slog.Warn("failed to delete cache in Redis after link deletion",
			slog.String("code", code),
			slog.String("error", err.Error()))
	}
	return nil
}

func (s *linkService) Redirect(c context.Context, req *dtov1.ClickLinkReq) (string, error) {
	// 判断 code 是否一定不存在
	if !s.bloom.MayContain(req.Code) {
		return "", consts.ErrInvalidLink
	}

	// 如果 code 可能存在 则继续往下走
	// 判断对应 code 是否在 redis 中已经缓存
	key := consts.RedisLinkCodeKey + req.Code
	cacheVal, err := s.rdb.Get(c, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		slog.Warn("failed to get original URL from Redis",
			slog.String("code", req.Code),
			slog.String("error", err.Error()))
	}
	cacheMiss := err != nil

	// 若缓存过了
	var originalURL string
	if !cacheMiss {
		var data dtov1.LinkCache
		if err := json.Unmarshal([]byte(cacheVal), &data); err != nil {
			// 尝试解析 若解析失败则判断为缓存失效
			slog.Warn("failed to unmarshal cache",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
			cacheMiss = true
		} else {
			// 若解析成功，则判断快照是否合法，若合法则继续，不合法则直接返回 ErrInvalidLink
			if !s.validateLinkCache(&data) {
				return "", consts.ErrInvalidLink
			}
			originalURL = data.OriginalURL
		}
	}

	// 如果缓存没有命中
	if cacheMiss {
		data, err := s.cacheLink(c, req)
		if err != nil {
			return "", err
		}

		originalURL = data.OriginalURL
	}

	// 避免为浏览器 prefetch/prerender 写入浏览记录
	if !req.SkipStats {
		req.EventID = uuid.NewString()
		if err := s.publisher.PublishLinkClicked(req); err != nil {
			slog.Warn("failed to publish link clicked event",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
		}
	}

	return originalURL, nil
}

func (s *linkService) GetStats(c context.Context, userID int64) (*dtov1.GetStatsResponse, error) {
	viewCount, err := s.repo.GetViewCountByUserID(c, userID)
	if err != nil {
		return nil, err
	}

	linkCount, err := s.repo.GetLinkCountByUserID(c, userID)
	if err != nil {
		return nil, err
	}

	resp := &dtov1.GetStatsResponse{
		ViewCount: viewCount,
		LinkCount: linkCount,
	}

	return resp, nil
}

func (s *linkService) InitBloomFilter() error {
	c := context.Background()
	codes, err := s.repo.GetAllLinkCodes(c)
	if err != nil {
		return err
	}

	for _, code := range codes {
		s.bloom.Add(code)
	}

	return nil
}
