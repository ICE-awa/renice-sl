package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
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
}

type linkService struct {
	repo      repository.LinkRepository
	publisher *event.LinkPublisher
	rdb       *redis.Client
	cfg       *config.LinkConfig
}

func NewLinkService(
	repo repository.LinkRepository,
	publisher *event.LinkPublisher,
	rdb *redis.Client,
	cfg *config.LinkConfig,
) LinkService {
	return &linkService{
		repo:      repo,
		publisher: publisher,
		rdb:       rdb,
		cfg:       cfg,
	}
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
	return err
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

	// 判断对应 code 是否在 redis 中已经缓存
	key := consts.RedisLinkCodeKey + req.Code
	cache, err := s.rdb.Get(c, key).Result()
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
		if err := json.Unmarshal([]byte(cache), &data); err != nil {
			// 尝试解析 若解析失败则判断为缓存失效
			slog.Warn("failed to unmarshal cache",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
			cacheMiss = true
		} else {
			// 若解析成功，则判断快照是否合法，若合法则继续，不合法则直接返回 ErrInvalidLink
			if data.OriginalURL == consts.NullLink ||
				data.Status == "inactive" ||
				(data.ExpiresAt != nil && data.ExpiresAt.Before(time.Now())) {
				return "", consts.ErrInvalidLink
			}
			originalURL = data.OriginalURL
		}
	}

	// 如果缓存没有命中
	if cacheMiss {
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
					slog.Warn("failed to marshal null cache",
						slog.String("code", req.Code),
						slog.String("error", err.Error()))
				} else {
					if err := s.rdb.Set(c, key, linkCache, s.cfg.NullExpires).Err(); err != nil {
						slog.Warn("failed to cache null link in Redis",
							slog.String("code", req.Code),
							slog.String("error", err.Error()))
					}
				}
				return "", consts.ErrInvalidLink
			}
			return "", err
		}

		// 若没有报错 无论如何先写入当前快照到 Redis 中 避免缓存穿透
		originalURL = data.OriginalURL
		linkCache, err := json.Marshal(data)
		if err != nil {
			slog.Warn("failed to marshal cache",
				slog.String("code", req.Code),
				slog.String("error", err.Error()))
		} else {
			if err := s.rdb.Set(c, key, linkCache, s.cfg.Expires).Err(); err != nil {
				slog.Warn("failed to cache original URL in Redis",
					slog.String("code", req.Code),
					slog.String("error", err.Error()))
			}
		}
		// 随后判断其合法性，若不合法则不放行
		if data.Status == "inactive" || (data.ExpiresAt != nil && data.ExpiresAt.Before(time.Now())) {
			return "", consts.ErrInvalidLink
		}
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
