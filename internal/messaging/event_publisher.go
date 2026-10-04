package messaging

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

type EventPublisher struct {
	conn *nats.Conn
}

func NewEventPublisher(
	conn *nats.Conn,
) *EventPublisher {
	return &EventPublisher{
		conn: conn,
	}
}

func (p *EventPublisher) Publish(
	ctx context.Context,
	subject string,
	payload []byte,
) error {
	if err := p.conn.Publish(
		subject,
		payload,
	); err != nil {
		return fmt.Errorf(
			"publish event to %s: %w",
			subject,
			err,
		)
	}

	if err := p.conn.FlushWithContext(
		ctx,
	); err != nil {
		return fmt.Errorf(
			"flush event to %s: %w",
			subject,
			err,
		)
	}

	return nil
}
