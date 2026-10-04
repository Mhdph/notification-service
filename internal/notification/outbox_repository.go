package notification

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OutboxRepository interface {
	ClaimNext(
		ctx context.Context,
		workerID string,
		leaseDuration time.Duration,
	) (*OutboxEvent, error)

	MarkPublished(
		ctx context.Context,
		id bson.ObjectID,
		workerID string,
	) error
}
