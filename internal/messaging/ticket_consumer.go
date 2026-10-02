package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"notification-service/internal/notification"

	"github.com/nats-io/nats.go"
)

type TicketConsumer struct {
	service *notification.Service
}

func NewTicketConsumer(
	service *notification.Service,
) *TicketConsumer {
	return &TicketConsumer{
		service: service,
	}
}

func (c *TicketConsumer) Start(
	conn *nats.Conn,
) (*nats.Subscription, error) {
	sub, err := conn.Subscribe(
		"ticket.user_mentioned",
		func(msg *nats.Msg) {
			c.handleUserMentioned(msg)
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"subscribe to ticket.user_mentioned: %w",
			err,
		)
	}

	return sub, nil
}

func (c *TicketConsumer) handleUserMentioned(
	msg *nats.Msg,
) {
	var event TicketUserMentionedEvent

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		fmt.Printf(
			"invalid ticket.user_mentioned event: %v\n",
			err,
		)
		return
	}

	err := c.service.Create(
		context.Background(),
		notification.CreateInput{
			WorkspaceID: event.WorkspaceID,
			RecipientID: event.Recipient.ID,

			Type: "ticket.mention",

			Actor: notification.Actor{
				ID:   event.Actor.ID,
				Name: event.Actor.Name,
			},

			Resource: notification.Resource{
				Type: "ticket",
				ID:   event.Ticket.ID,
			},

			Title: "You were mentioned",

			Body: fmt.Sprintf(
				"%s mentioned you in ticket %s",
				event.Actor.Name,
				event.Ticket.Title,
			),

			Action: notification.Action{
				Type: "open",
				URL: fmt.Sprintf(
					"/tickets/%s?reply=%s",
					event.Ticket.ID,
					event.ReplyID,
				),
			},

			SourceEventID: event.EventID,
		},
	)

	if err != nil {
		fmt.Printf(
			"create notification from ticket mention: %v\n",
			err,
		)
		return
	}

	fmt.Printf(
		"notification created from event %s\n",
		event.EventID,
	)
}
