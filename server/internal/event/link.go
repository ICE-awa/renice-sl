package event

import (
	"encoding/json"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/metrics"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"time"
)

const (
	SubjectLinkClicked = "link.clicked"
	SubjectLinkChecked = "link.checked"
)

type LinkPublisher struct {
	nats *mq.NatsClient
}

func NewLinkPublisher(nats *mq.NatsClient) *LinkPublisher {
	return &LinkPublisher{nats: nats}
}

func (p *LinkPublisher) PublishLinkClicked(event *dtov1.ClickLinkReq) error {
	start := time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.nats.JetStream.Publish(SubjectLinkClicked, data)
	if err != nil {
		return err
	}

	metrics.MQPublishTotal.WithLabelValues(SubjectLinkClicked).Inc()
	metrics.MQPublishDurationSeconds.WithLabelValues(SubjectLinkClicked).Observe(time.Since(start).Seconds())
	return nil
}

func (p *LinkPublisher) PublishLinkChecked(event *dtov1.CheckLinkReq) error {
	start := time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.nats.JetStream.Publish(SubjectLinkChecked, data)
	if err != nil {
		return err
	}

	metrics.MQPublishTotal.WithLabelValues(SubjectLinkChecked).Inc()
	metrics.MQPublishDurationSeconds.WithLabelValues(SubjectLinkChecked).Observe(time.Since(start).Seconds())
	return nil
}
