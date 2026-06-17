package service

import (
	"context"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/metrics"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/nats-io/nats.go"
	"strconv"
)

type DLQService interface {
	GetDLQMessages(context.Context, *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error)
	RetryDLQMessage(context.Context, int64) error
	MarkAsResolved(context.Context, int64) error
}

type dlqService struct {
	repo repository.DLQRepository
	nc   *mq.NatsClient
}

func NewDLQService(repo repository.DLQRepository, nc *mq.NatsClient) DLQService {
	return &dlqService{repo, nc}
}

func (s *dlqService) GetDLQMessages(c context.Context, req *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
	return s.repo.GetDLQMessages(c, req)
}

func (s *dlqService) RetryDLQMessage(c context.Context, id int64) error {
	data, err := s.repo.SetDLQMessageRetrying(c, id)
	if err != nil {
		return err
	}

	msg := nats.NewMsg(data.Subject)
	msg.Data = data.Payload
	msg.Header.Set(consts.DLQMessageIDHeader, strconv.FormatInt(id, 10))

	_, err = s.nc.JetStream.PublishMsg(msg)
	if err != nil {
		return err
	}

	metrics.DLQRetryTotal.WithLabelValues(data.Subject).Inc()
	return nil
}

func (s *dlqService) MarkAsResolved(c context.Context, id int64) error {
	subject, err := s.repo.MarkAsResolved(c, id)
	if err != nil {
		return err
	}

	metrics.DLQResolvedTotal.WithLabelValues(subject).Inc()
	return nil
}
