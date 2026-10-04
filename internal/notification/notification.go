package notification

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Actor struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

type Resource struct {
	Type string `bson:"type" json:"type"`
	ID   string `bson:"id" json:"id"`
}

type Action struct {
	Type string `bson:"type" json:"type"`
	URL  string `bson:"url" json:"url"`
}

type Notification struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id"`

	AppID       string `bson:"app_id" json:"app_id"`
	RecipientID string `bson:"recipient_id" json:"recipient_id"`

	Type string `bson:"type" json:"type"`

	Actor Actor `bson:"actor" json:"actor"`

	Resource Resource `bson:"resource" json:"resource"`

	Title string `bson:"title" json:"title"`
	Body  string `bson:"body" json:"body"`

	Action Action `bson:"action" json:"action"`

	SourceEventID string `bson:"source_event_id" json:"source_event_id"`

	ReadAt    *time.Time `bson:"read_at,omitempty" json:"read_at"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
}
