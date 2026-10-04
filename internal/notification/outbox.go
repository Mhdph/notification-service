package notification

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OutboxEvent struct {
	ID bson.ObjectID `bson:"_id,omitempty"`

	EventID string `bson:"event_id"`

	Subject string `bson:"subject"`

	Payload []byte `bson:"payload"`

	CreatedAt time.Time `bson:"created_at"`

	PublishedAt *time.Time `bson:"published_at,omitempty"`
	ClaimedBy   string     `bson:"claimed_by,omitempty"`

	ClaimedUntil *time.Time `bson:"claimed_until,omitempty"`
}

type CreatedEvent struct {
	ID string `json:"id"`

	WorkspaceID string `json:"workspace_id"`
	RecipientID string `json:"recipient_id"`

	Type string `json:"type"`

	Actor Actor `json:"actor"`

	Resource Resource `json:"resource"`

	Title string `json:"title"`
	Body  string `json:"body"`

	Action Action `json:"action"`

	ReadAt *time.Time `json:"read_at"`

	CreatedAt time.Time `json:"created_at"`
}
