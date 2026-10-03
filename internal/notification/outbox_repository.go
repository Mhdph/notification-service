package notification

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OutboxRepository interface {
	FindUnpublished(
		ctx context.Context,
		limit int64,
	) ([]OutboxEvent, error)

	MarkPublished(
		ctx context.Context,
		id bson.ObjectID,
	) error
}
