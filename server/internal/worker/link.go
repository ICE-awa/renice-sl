package worker

import (
	"context"
	"encoding/json"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/internal/metrics"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/nats-io/nats.go"
	"log/slog"
	"strconv"
	"time"
)

const (
	QueueAnalyticsWorker = "analytics-worker"
	DurableLinkClicked   = "link-clicked-worker"
	DurableLinkChecked   = "link-checked-worker"
)

type LinkWorker struct {
	svc    service.LinkEventService
	client *mq.NatsClient
}

func NewLinkWorker(svc service.LinkEventService, client *mq.NatsClient) *LinkWorker {
	return &LinkWorker{svc: svc, client: client}
}

func (w *LinkWorker) StartLinkClickWorker() error {
	_, err := w.client.JetStream.QueueSubscribe(
		event.SubjectLinkClicked,
		QueueAnalyticsWorker,
		func(msg *nats.Msg) {
			start := time.Now()
			defer func() {
				metrics.WorkerProcessDurationSeconds.WithLabelValues(event.SubjectLinkClicked).Observe(time.Since(start).Seconds())
			}()
			var e *dtov1.ClickLinkReq
			dlqMessageID := msg.Header.Get(consts.DLQMessageIDHeader)

			if err := json.Unmarshal(msg.Data, &e); err != nil {
				_ = msg.Term()
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := w.svc.HandleLinkClicked(ctx, e); err != nil {
				_ = msg.NakWithDelay(3 * time.Second)
				return
			}

			_ = msg.Ack()

			if dlqMessageID != "" {
				id, _ := strconv.ParseInt(dlqMessageID, 10, 64)
				err := w.svc.HandleResolvedDLQMessage(ctx, id)
				if err != nil {
					slog.Warn("Failed to mark DLQ message as resolved",
						slog.String("error", err.Error()),
						slog.String("message_id", dlqMessageID))
				}
			}
		},
		nats.Durable(DurableLinkClicked),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
		nats.MaxDeliver(3),
	)

	return err
}

func (w *LinkWorker) StartLinkCheckWorker() error {
	_, err := w.client.JetStream.QueueSubscribe(
		event.SubjectLinkChecked,
		QueueAnalyticsWorker,
		func(msg *nats.Msg) {
			start := time.Now()
			defer func() {
				metrics.WorkerProcessDurationSeconds.WithLabelValues(event.SubjectLinkChecked).Observe(time.Since(start).Seconds())
			}()
			var e *dtov1.CheckLinkReq
			dlqMessageID := msg.Header.Get(consts.DLQMessageIDHeader)

			if err := json.Unmarshal(msg.Data, &e); err != nil {
				_ = msg.Term()
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := w.svc.HandleLinkChecked(ctx, e); err != nil {
				_ = msg.NakWithDelay(3 * time.Second)
				return
			}

			_ = msg.Ack()

			if dlqMessageID != "" {
				id, _ := strconv.ParseInt(dlqMessageID, 10, 64)
				err := w.svc.HandleResolvedDLQMessage(ctx, id)
				if err != nil {
					slog.Warn("Failed to mark DLQ message as resolved",
						slog.String("error", err.Error()),
						slog.String("message_id", dlqMessageID))
				}
			}
		},
		nats.Durable(DurableLinkChecked),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
		nats.MaxDeliver(3),
	)

	return err
}
