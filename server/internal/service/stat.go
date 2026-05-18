package service

import (
	"context"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
)

type StatService interface {
	GetLinkStats(context.Context, *dtov1.GetLinkStatReq) ([]*dtov1.LinkStatItem, error)
	GetClickStats(context.Context, *dtov1.GetClickStatReq) ([]*dtov1.ClickStatItem, error)
	GetUserStats(context.Context, *dtov1.GetUserStatReq) ([]*dtov1.UserStatItem, error)
}

type statService struct {
	repo repository.StatRepository
}

func NewStatService(repo repository.StatRepository) StatService {
	return &statService{repo: repo}
}

func (s *statService) GetLinkStats(c context.Context, req *dtov1.GetLinkStatReq) ([]*dtov1.LinkStatItem, error) {
	if req.Bucket == "day" {
		return s.repo.GetLinkDayStat(c, req.Range)
	} else {
		return s.repo.GetLinkHourStat(c, req.Range)
	}
}

func (s *statService) GetClickStats(c context.Context, req *dtov1.GetClickStatReq) ([]*dtov1.ClickStatItem, error) {
	if req.Bucket == "day" {
		return s.repo.GetClickDayStat(c, req.Range)
	} else {
		return s.repo.GetClickHourStat(c, req.Range)
	}
}

func (s *statService) GetUserStats(c context.Context, req *dtov1.GetUserStatReq) ([]*dtov1.UserStatItem, error) {
	if req.Bucket == "day" {
		return s.repo.GetUserDayStat(c, req.Range)
	} else {
		return s.repo.GetUserHourStat(c, req.Range)
	}
}
