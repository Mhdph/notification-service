package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	js nats.JetStreamContext,
) (*nats.Subscription, error) {
	sub, err := js.QueueSubscribe(
		"ticket.user_mentioned",
		"notification-service",
		c.handleUserMentioned,

		nats.Durable(
			"notification-service-ticket-mentions",
		),

		nats.ManualAck(),

		nats.AckExplicit(),

		nats.DeliverAll(),

		nats.AckWait(
			30*time.Second,
		),

		nats.MaxDeliver(10),
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

	if err := json.Unmarshal(
		msg.Data,
		&event,
	); err != nil {
		fmt.Printf(
			"invalid ticket.user_mentioned event: %v\n",
			err,
		)

		if err := msg.Term(); err != nil {
			fmt.Printf(
				"terminate message: %v\n",
				err,
			)
		}

		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	_, err := c.service.Create(
		ctx,
		notification.CreateInput{
			AppID: event.AppID,

			RecipientID: event.Recipient.ID,

			Type: "ticket.mention",

			Actor: notification.Actor{
				ID: event.Actor.ID,

				Name: event.Actor.Name,
			},

			Resource: notification.Resource{
				Type: "ticket",

				ID: event.Ticket.ID,
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

		if err := msg.Nak(); err != nil {
			fmt.Printf(
				"nak message: %v\n",
				err,
			)
		}

		return
	}

	if err := msg.Ack(); err != nil {
		fmt.Printf(
			"ack message: %v\n",
			err,
		)

		return
	}

	fmt.Printf(
		"notification created from event %s\n",
		event.EventID,
	)
}
