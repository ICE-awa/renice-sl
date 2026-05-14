package event

import (
	"encoding/json"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/shared/mq"
)

const SubjectLinkClicked = "link.clicked"

type LinkPublisher struct {
	nats *mq.NatsClient
}

func NewLinkPublisher(nats *mq.NatsClient) *LinkPublisher {
	return &LinkPublisher{nats: nats}
}

func (p *LinkPublisher) PublishLinkClicked(event *dtov1.ClickLinkReq) error {

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.nats.JetStream.Publish(SubjectLinkClicked, data)
	return err
}
