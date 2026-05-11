package service

import (
	"context"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/util"
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
}

func NewLinkService(repo repository.LinkRepository, publisher *event.LinkPublisher) LinkService {
	return &linkService{repo: repo, publisher: publisher}
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
	return s.repo.UpdateLink(c, req)
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
	return s.repo.DeleteLink(c, req)
}

func (s *linkService) Redirect(c context.Context, req *dtov1.ClickLinkReq) (string, error) {
	originalURL, err := s.repo.GetOriginalURLByCode(c, req.Code)
	if err != nil {
		return "", err
	}

	if err := s.repo.RecordClick(c, req); err != nil {
		return "", err
	}

	//if err := s.publisher.PublishLinkClicked(req); err != nil {
	//	slog.Warn("failed to publish link clicked event",
	//		slog.String("code", req.code),
	//		slog.String("error", err.Error()))
	//}

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
