package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"notification-service/internal/notification"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OutboxRepository struct {
	collection *mongo.Collection
}

func NewOutboxRepository(
	db *mongo.Database,
) *OutboxRepository {
	return &OutboxRepository{
		collection: db.Collection("notification_outbox"),
	}
}

func (r *OutboxRepository) FindUnpublished(
	ctx context.Context,
	limit int64,
) ([]notification.OutboxEvent, error) {
	filter := bson.M{
		"published_at": nil,
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: 1},
		}).
		SetLimit(limit)

	cursor, err := r.collection.Find(
		ctx,
		filter,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find unpublished outbox events: %w",
			err,
		)
	}

	defer cursor.Close(ctx)

	var events []notification.OutboxEvent

	if err := cursor.All(ctx, &events); err != nil {
		return nil, fmt.Errorf(
			"decode outbox events: %w",
			err,
		)
	}

	return events, nil
}

func (r *OutboxRepository) MarkPublished(
	ctx context.Context,
	id bson.ObjectID,
	workerID string,
) error {
	now := time.Now().UTC()

	filter := bson.M{
		"_id":          id,
		"claimed_by":   workerID,
		"published_at": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"published_at": now,
		},
		"$unset": bson.M{
			"claimed_by":    "",
			"claimed_until": "",
		},
	}

	result, err :=
		r.collection.UpdateOne(
			ctx,
			filter,
			update,
		)

	if err != nil {
		return fmt.Errorf(
			"mark outbox event as published: %w",
			err,
		)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf(
			"outbox event is no longer owned by worker %s",
			workerID,
		)
	}

	return nil
}

func (r *OutboxRepository) EnsureIndexes(
	ctx context.Context,
) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   "published_at",
					Value: 1,
				},
				{
					Key:   "claimed_until",
					Value: 1,
				},
				{
					Key:   "created_at",
					Value: 1,
				},
			},
		},
		{
			Keys: bson.D{
				{
					Key:   "event_id",
					Value: 1,
				},
			},
			Options: options.Index().
				SetUnique(true),
		},
	}

	_, err :=
		r.collection.Indexes().CreateMany(
			ctx,
			indexes,
		)

	if err != nil {
		return fmt.Errorf(
			"create outbox indexes: %w",
			err,
		)
	}

	return nil
}
func (r *OutboxRepository) ClaimNext(
	ctx context.Context,
	workerID string,
	leaseDuration time.Duration,
) (*notification.OutboxEvent, error) {
	now := time.Now().UTC()

	claimedUntil := now.Add(
		leaseDuration,
	)

	filter := bson.M{
		"published_at": nil,

		"$or": bson.A{
			bson.M{
				"claimed_until": nil,
			},
			bson.M{
				"claimed_until": bson.M{
					"$lte": now,
				},
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"claimed_by":    workerID,
			"claimed_until": claimedUntil,
		},
	}

	opts := options.FindOneAndUpdate().
		SetSort(
			bson.D{
				{
					Key:   "created_at",
					Value: 1,
				},
			},
		).
		SetReturnDocument(
			options.After,
		)

	var event notification.OutboxEvent

	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&event)

	if errors.Is(
		err,
		mongo.ErrNoDocuments,
	) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf(
			"claim next outbox event: %w",
			err,
		)
	}

	return &event, nil
}
