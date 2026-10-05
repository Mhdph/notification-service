package repository

import (
	"context"
	"errors"
	"fmt"
	"notification-service/internal/notification"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotificationRepository struct {
	database   *mongo.Database
	collection *mongo.Collection
}

func NewNotificationRepository(
	db *mongo.Database,
) *NotificationRepository {
	return &NotificationRepository{
		database:   db,
		collection: db.Collection("notifications"),
	}
}

var errAlreadyProcessed = errors.New("notification event already processed")

func (r *NotificationRepository) CreateWithOutbox(
	ctx context.Context,
	n notification.Notification,
	outbox notification.OutboxEvent,
) (notification.Notification, error) {
	session, err :=
		r.database.Client().StartSession()

	if err != nil {
		return notification.Notification{},
			fmt.Errorf(
				"start mongo session: %w",
				err,
			)
	}

	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(
			txCtx context.Context,
		) (any, error) {
			_, err := r.collection.InsertOne(
				txCtx,
				n,
			)

			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return nil, errAlreadyProcessed
				}

				return nil, fmt.Errorf(
					"insert notification: %w",
					err,
				)
			}

			outboxCollection :=
				r.database.Collection(
					"notification_outbox",
				)

			_, err = outboxCollection.InsertOne(
				txCtx,
				outbox,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"insert outbox event: %w",
					err,
				)
			}

			return nil, nil
		},
	)

	if errors.Is(
		err,
		errAlreadyProcessed,
	) {
		var existing notification.Notification

		findErr :=
			r.collection.FindOne(
				ctx,
				bson.M{
					"source_event_id": n.SourceEventID,

					"recipient_id": n.RecipientID,
				},
			).Decode(
				&existing,
			)

		if findErr != nil {
			return notification.Notification{},
				fmt.Errorf(
					"find existing notification after duplicate: %w",
					findErr,
				)
		}

		return existing, nil
	}
	if err != nil {
		return notification.Notification{},
			fmt.Errorf(
				"create notification transaction: %w",
				err,
			)
	}

	return n, nil
}
func (r *NotificationRepository) EnsureIndexes(
	ctx context.Context,
) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "app_id", Value: 1},
				{Key: "recipient_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "app_id", Value: 1},
				{Key: "recipient_id", Value: 1},
				{Key: "read_at", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "source_event_id", Value: 1},
				{Key: "recipient_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(
		ctx,
		indexes,
	)

	if err != nil {
		return fmt.Errorf(
			"create notification indexes: %w",
			err,
		)
	}

	return nil
}

func (r *NotificationRepository) ListByRecipient(
	ctx context.Context,
	appID string,
	recipientID string,
	limit int64,
) ([]notification.Notification, error) {
	filter := bson.M{
		"app_id":       appID,
		"recipient_id": recipientID,
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
		}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []notification.Notification

	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("decode notifications: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(
	ctx context.Context,
	appID string,
	recipientID string,
	notificationID string,
) error {
	id, err := bson.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification id: %w", err)
	}

	filter := bson.M{
		"_id":          id,
		"app_id":       appID,
		"recipient_id": recipientID,
	}

	update := bson.M{
		"$set": bson.M{
			"read_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	if result.MatchedCount == 0 {
		return notification.ErrNotFound
	}

	return nil
}

func (r *NotificationRepository) UnreadCount(
	ctx context.Context,
	appID string,
	recipientID string,
) (int64, error) {
	filter := bson.M{
		"app_id":       appID,
		"recipient_id": recipientID,
		"read_at":      nil,
	}

	count, err := r.collection.CountDocuments(
		ctx,
		filter,
	)

	if err != nil {
		return 0, fmt.Errorf(
			"count unread notifications: %w",
			err,
		)
	}

	return count, nil
}

func (r *NotificationRepository) MarkAllAsRead(
	ctx context.Context,
	appID string,
	recipientID string,
) error {
	now := time.Now().UTC()

	filter := bson.M{
		"app_id":       appID,
		"recipient_id": recipientID,
		"read_at":      nil,
	}

	update := bson.M{
		"$set": bson.M{
			"read_at": now,
		},
	}

	_, err := r.collection.UpdateMany(
		ctx,
		filter,
		update,
	)

	if err != nil {
		return fmt.Errorf(
			"mark all notifications as read: %w",
			err,
		)
	}

	return nil
}
