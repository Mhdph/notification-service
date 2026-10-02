package messaging

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

func Connect(url string) (*nats.Conn, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	return conn, nil
}

func JetStream(
	conn *nats.Conn,
) (nats.JetStreamContext, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	return js, nil
}

func EnsureTicketStream(
	js nats.JetStreamContext,
) error {
	_, err := js.StreamInfo("TICKETS")

	if err == nil {
		return nil
	}

	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("get TICKETS stream: %w", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "TICKETS",
		Subjects: []string{"ticket.>"},
		Storage:  nats.FileStorage,
	})

	if err != nil {
		return fmt.Errorf("create TICKETS stream: %w", err)
	}

	return nil
}
