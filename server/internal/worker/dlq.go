package worker

import (
	"context"
	"encoding/json"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/repository"
	"github.com/ICE-awa/renice-sl/shared/mq"
	"github.com/nats-io/nats.go"
	"log/slog"
	"time"
)

type DLQWorker struct {
	natsClient *mq.NatsClient
	repo       repository.DLQRepository
}

func NewDLQWorker(natsClient *mq.NatsClient, repo repository.DLQRepository) *DLQWorker {
	return &DLQWorker{natsClient, repo}
}

type MaxDeliverAdvisory struct {
	Stream    string `json:"stream"`
	Consumer  string `json:"consumer"`
	StreamSeq uint64 `json:"stream_seq"`
}

func (w *DLQWorker) StartDLQWorker() error {
	_, err := w.natsClient.Conn.Subscribe(
		consts.NATSMaxDeliverAdvisory,
		func(msg *nats.Msg) {
			var adv MaxDeliverAdvisory
			if err := json.Unmarshal(msg.Data, &adv); err != nil {
				return
			}

			raw, err := w.natsClient.JetStream.GetMsg(adv.Stream, adv.StreamSeq)
			if err != nil {
				return
			}

			dlq := dtov1.DLQMessage{
				SourceStream:   adv.Stream,
				SourceConsumer: adv.Consumer,
				StreamSeq:      adv.StreamSeq,
				Subject:        raw.Subject,
				Payload:        raw.Data,
				Reason:         "max_deliveries",
				FailedAt:       time.Now(),
			}

			data, err := json.Marshal(dlq)
			if err != nil {
				return
			}

			_, err = w.natsClient.JetStream.Publish("dlq."+raw.Subject, data)
			if err != nil {
				slog.Warn("Failed to publish DLQ message to NATS",
					slog.String("error", err.Error()),
					slog.Any("dlq", dlq))
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = w.repo.RecordDLQMessage(ctx, &dlq)
			if err != nil {
				slog.Warn("Failed to Record DLQ message to database",
					slog.String("error", err.Error()),
					slog.Any("dlq", dlq))
			}
		},
	)

	return err
}
