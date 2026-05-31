package service

import (
	"context"
	"errors"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type LinkEventService interface {
	HandleLinkClicked(context.Context, *dtov1.ClickLinkReq) error
	HandleLinkChecked(context.Context, *dtov1.CheckLinkReq) error
	HandleResolvedDLQMessage(context.Context, int64) error
}

type linkEventService struct {
	linkRepo           repository.LinkRepository
	dlqRepo            repository.DLQRepository
	rdb                *redis.Client
	safeBrowsingClient *util.SafeBrowsingClient
}

func NewLinkEventService(linkRepo repository.LinkRepository,
	dlqRepo repository.DLQRepository,
	rdb *redis.Client,
	safeBrowsingClient *util.SafeBrowsingClient) LinkEventService {
	return &linkEventService{
		linkRepo,
		dlqRepo,
		rdb,
		safeBrowsingClient,
	}
}

func (s *linkEventService) HandleLinkClicked(c context.Context, event *dtov1.ClickLinkReq) error {
	return s.linkRepo.RecordClick(c, event)
}

func (s *linkEventService) HandleLinkChecked(c context.Context, event *dtov1.CheckLinkReq) error {
	// 通过探活 + 跳转链 IP 检查
	final, err := util.CheckURLReachable(c, event.OriginalURL)
	if err != nil {
		if errors.Is(err, util.ErrBlockedHost) {
			_ = s.linkRepo.RecordLinkCheck(c, event.Code, "unsafe")
			return nil
		}
		return err
	}

	status := "safe"
	if s.safeBrowsingClient.IsEnabled() {
		// 对 final URL 做 SafeBrowsing 检查
		safe, err := s.safeBrowsingClient.IsSafe(c, final)
		if err != nil {
			return err
		}

		if !safe {
			status = "unsafe"
		}
	}

	_ = s.linkRepo.RecordLinkCheck(c, event.Code, status)

	// 删除 Redis 缓存的 link，避免缓存不一致
	key := consts.RedisLinkCodeKey + event.Code
	if err := s.rdb.Del(c, key).Err(); err != nil {
		slog.Warn("Failed to delete Redis cache for link code",
			slog.String("code", event.Code),
			slog.String("error", err.Error()))
	}
	return nil
}

func (s *linkEventService) HandleResolvedDLQMessage(c context.Context, messageID int64) error {
	return s.dlqRepo.MarkAsResolved(c, messageID)
}
