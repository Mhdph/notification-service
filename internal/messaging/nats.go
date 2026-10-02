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
