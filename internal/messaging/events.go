package messaging

import "time"

type TicketUserMentionedEvent struct {
	EventID string `json:"event_id"`

	AppID string `json:"app_id"`

	Actor struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"actor"`

	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`

	Ticket struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"ticket"`

	ReplyID string `json:"reply_id"`

	OccurredAt time.Time `json:"occurred_at"`
}
