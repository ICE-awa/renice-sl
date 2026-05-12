package worker

import (
	"context"
	"encoding/json"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/nats-io/nats.go"
	"time"
)

const (
	QueueAnalyticsWorker = "analytics-worker"
	DurableLinkClicked   = "link-clicked-worker"
)

type LinkClickWorker struct {
	svc service.LinkEventService
}

func NewLinkClickWorker(svc service.LinkEventService) *LinkClickWorker {
	return &LinkClickWorker{svc: svc}
}

func (w *LinkClickWorker) StartLinkClickWorker(client *mq.NatsClient) error {
	_, err := client.JetStream.QueueSubscribe(
		event.SubjectLinkClicked,
		QueueAnalyticsWorker,
		func(msg *nats.Msg) {
			var e *dtov1.ClickLinkReq

			if err := json.Unmarshal(msg.Data, &e); err != nil {
				_ = msg.Term()
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := w.svc.HandleLinkClicked(ctx, e); err != nil {
				_ = msg.Nak()
				return
			}

			_ = msg.Ack()
		},
		nats.Durable(DurableLinkClicked),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
	)

	return err
}
