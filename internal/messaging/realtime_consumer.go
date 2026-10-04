package messaging

import (
	"encoding/json"
	"fmt"

	"notification-service/internal/notification"
	"notification-service/internal/realtime"

	"github.com/nats-io/nats.go"
)

type RealtimeConsumer struct {
	hub *realtime.Hub
}

func NewRealtimeConsumer(
	hub *realtime.Hub,
) *RealtimeConsumer {
	return &RealtimeConsumer{
		hub: hub,
	}
}

func (c *RealtimeConsumer) Start(
	conn *nats.Conn,
) (*nats.Subscription, error) {
	sub, err := conn.Subscribe(
		"notification.created",
		func(msg *nats.Msg) {
			c.handleNotificationCreated(msg)
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"subscribe to notification.created: %w",
			err,
		)
	}

	return sub, nil
}

func (c *RealtimeConsumer) handleNotificationCreated(
	msg *nats.Msg,
) {
	var event notification.CreatedEvent

	if err := json.Unmarshal(
		msg.Data,
		&event,
	); err != nil {
		fmt.Printf(
			"invalid notification.created event: %v\n",
			err,
		)
		return
	}

	c.hub.SendToUser(
		event.RecipientID,
		map[string]any{
			"type": "notification.created",
			"data": event,
		},
	)
}
