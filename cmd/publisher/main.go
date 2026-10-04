package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	conn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	event := map[string]any{
		"event_id": "evt_test_101",
		"app_id":   "workspace_1",

		"actor": map[string]any{
			"id":   "user_10",
			"name": "Ali",
		},

		"recipient": map[string]any{
			"id": "user_42",
		},

		"ticket": map[string]any{
			"id":    "ticket_392",
			"title": "Connection problem",
		},

		"reply_id": "reply_887",

		"occurred_at": time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	js, err := conn.JetStream()
	if err != nil {
		panic(err)
	}

	ack, err := js.Publish(
		"ticket.user_mentioned",
		data,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"event stored in stream %s sequence %d\n",
		ack.Stream,
		ack.Sequence,
	)

	if err := conn.Flush(); err != nil {
		panic(err)
	}

	fmt.Println("event published")
}
