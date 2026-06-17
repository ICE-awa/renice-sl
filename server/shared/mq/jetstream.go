package mq

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
)

const (
	StreamLinkEvents = "LINK_EVENTS"
	SubjectLinkAll   = "link.*"
	StreamDLQ        = "DLQ"
	SubjectDLQALL    = "dlq.>"
)

func ensureOneStream(client *NatsClient, cfg *nats.StreamConfig) error {
	_, err := client.JetStream.StreamInfo(cfg.Name)
	if err == nil {
		return nil
	}

	if !errors.Is(err, nats.ErrStreamNotFound) {
		return err
	}

	_, err = client.JetStream.AddStream(cfg)
	return err
}

func EnsureStream(client *NatsClient) error {
	streams := []*nats.StreamConfig{
		{
			Name:     StreamLinkEvents,
			Subjects: []string{SubjectLinkAll},
			Storage:  nats.FileStorage,
		},
		{
			Name:     StreamDLQ,
			Subjects: []string{SubjectDLQALL},
			Storage:  nats.FileStorage,
		},
	}

	for _, cfg := range streams {
		if err := ensureOneStream(client, cfg); err != nil {
			return fmt.Errorf("failed to ensure stream %s: %w", cfg.Name, err)
		}
	}

	return nil
}
