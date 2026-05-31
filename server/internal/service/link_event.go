package service

import (
	"context"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/util"
)

type LinkEventService interface {
	HandleLinkClicked(context.Context, *dtov1.ClickLinkReq) error
	HandleLinkChecked(context.Context, *dtov1.CheckLinkReq) error
	HandleResolvedDLQMessage(context.Context, int64) error
}

type linkEventService struct {
	linkRepo repository.LinkRepository
	dlqRepo  repository.DLQRepository
}

func NewLinkEventService(linkRepo repository.LinkRepository, dlqRepo repository.DLQRepository) LinkEventService {
	return &linkEventService{linkRepo, dlqRepo}
}

func (s *linkEventService) HandleLinkClicked(c context.Context, event *dtov1.ClickLinkReq) error {
	return s.linkRepo.RecordClick(c, event)
}

func (s *linkEventService) HandleLinkChecked(c context.Context, event *dtov1.CheckLinkReq) error {
	// 确认 OriginalURL 状态
	_, err := util.NormalizeAndValidateURL(c, event.OriginalURL)
	if err != nil {
		_ = s.linkRepo.RecordLinkCheck(c, event.Code, "unsafe")
		return err
	}

	_ = s.linkRepo.RecordLinkCheck(c, event.Code, "active")
	return nil
}

func (s *linkEventService) HandleResolvedDLQMessage(c context.Context, messageID int64) error {
	return s.dlqRepo.MarkAsResolved(c, messageID)
}
