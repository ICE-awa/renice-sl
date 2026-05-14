package service

import (
	"context"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
)

type LinkEventService interface {
	HandleLinkClicked(context.Context, *dtov1.ClickLinkReq) error
}

type linkEventService struct {
	repo repository.LinkRepository
}

func NewLinkEventService(repo repository.LinkRepository) LinkEventService {
	return &linkEventService{repo: repo}
}

func (s *linkEventService) HandleLinkClicked(c context.Context, event *dtov1.ClickLinkReq) error {
	return s.repo.RecordClick(c, event)
}
