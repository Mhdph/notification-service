package repository

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/notification"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(
	db *mongo.Database,
) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) Create(
	ctx context.Context,
	n notification.Notification,
) error {
	result, err := r.collection.InsertOne(ctx, n)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}

	fmt.Printf("Created notification: %v\n", result.InsertedID)

	return nil
}

func (r *NotificationRepository) EnsureIndexes(
	ctx context.Context,
) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "workspace_id", Value: 1},
				{Key: "recipient_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "source_event_id", Value: 1},
				{Key: "recipient_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("create notification indexes: %w", err)
	}

	return nil
}

func (r *NotificationRepository) ListByRecipient(
	ctx context.Context,
	workspaceID string,
	recipientID string,
	limit int64,
) ([]notification.Notification, error) {
	filter := bson.M{
		"workspace_id": workspaceID,
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
	workspaceID string,
	recipientID string,
	notificationID string,
) error {
	id, err := bson.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification id: %w", err)
	}

	filter := bson.M{
		"_id":          id,
		"workspace_id": workspaceID,
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
