package notification

import "time"

type Notification struct {
	ID string

	WorkspaceID string
	RecipientID string
	ActorID     string

	Type string

	Title string
	Body  string

	ResourceType string
	ResourceID   string

	ActionURL string

	ReadAt    *time.Time
	CreatedAt time.Time
}
