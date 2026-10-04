package notification

import "context"

type Repository interface {
	CreateWithOutbox(
		ctx context.Context,
		notification Notification,
		outbox OutboxEvent,
	) (Notification, error)

	ListByRecipient(
		ctx context.Context,
		appID string,
		recipientID string,
		limit int64,
	) ([]Notification, error)

	MarkAsRead(
		ctx context.Context,
		appID string,
		recipientID string,
		notificationID string,
	) error

	UnreadCount(
		ctx context.Context,
		appID string,
		recipientID string,
	) (int64, error)

	MarkAllAsRead(
		ctx context.Context,
		appID string,
		recipientID string,
	) error
}
