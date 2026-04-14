package mq

import (
	"fmt"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/nats-io/nats.go"
	"log/slog"
	"time"
)

type NatsClient struct {
	Conn      *nats.Conn
	JetStream nats.JetStreamContext
}

func NewNatsClient(cfg config.NatsConfig) (*NatsClient, error) {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(5 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			slog.Warn("NATS Disconnected",
				slog.String("error", err.Error()))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("NATS Reconnected",
				slog.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			slog.Info("NATS Closed")
		}),
	}

	url := fmt.Sprintf("nats://%s:4222")
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("Failed to initialize NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("Failed to initialize JetStream: %w", err)
	}

	slog.Info("NATS initialized",
		slog.String("url", nc.ConnectedUrl()))

	return &NatsClient{
		Conn:      nc,
		JetStream: js,
	}, nil
}

func (c *NatsClient) Close() {
	err := c.Conn.Drain()
	if err != nil {
		slog.Error("Failed to drain NATS connection",
			slog.String("error", err.Error()))
	}
}
