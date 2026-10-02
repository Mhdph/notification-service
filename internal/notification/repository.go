package notification

import "context"

type Repository interface {
	Create(
		ctx context.Context,
		notification Notification,
	) (Notification, error)

	ListByRecipient(
		ctx context.Context,
		workspaceID string,
		recipientID string,
		limit int64,
	) ([]Notification, error)

	MarkAsRead(
		ctx context.Context,
		workspaceID string,
		recipientID string,
		notificationID string,
	) error
}
