package mq

import "github.com/nats-io/nats.go"

const (
	StreamLinkEvents = "LINK_EVENTS"
	SubjectLinkAll   = "link.*"
)

func EnsureStream(client *NatsClient) error {
	_, err := client.JetStream.StreamInfo(StreamLinkEvents)
	if err == nil {
		return nil
	}

	_, err = client.JetStream.AddStream(&nats.StreamConfig{
		Name:     StreamLinkEvents,
		Subjects: []string{SubjectLinkAll},
		Storage:  nats.MemoryStorage,
	})

	return err
}
