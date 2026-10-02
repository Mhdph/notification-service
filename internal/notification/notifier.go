package notification

import "context"

type Notifier interface {
	Notify(
		ctx context.Context,
		notification Notification,
	) error
}
