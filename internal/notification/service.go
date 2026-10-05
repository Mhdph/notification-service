package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreateInput struct {
	AppID       string
	RecipientID string

	Type string

	Actor Actor

	Resource Resource

	Title string
	Body  string

	Action Action

	SourceEventID string
}

type Service struct {
	repository Repository
}

func NewService(
	repository Repository,
) *Service {
	return &Service{
		repository: repository,
	}
}
func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) (Notification, error) {
	now := time.Now().UTC()

	n := Notification{
		ID: bson.NewObjectID(),

		AppID: input.AppID,

		RecipientID: input.RecipientID,

		Type: input.Type,

		Actor: input.Actor,

		Resource: input.Resource,

		Title: input.Title,

		Body: input.Body,

		Action: input.Action,

		SourceEventID: input.SourceEventID,

		ReadAt: nil,

		CreatedAt: now,
	}

	createdEvent := CreatedEvent{
		ID: n.ID.Hex(),

		AppID: n.AppID,

		RecipientID: n.RecipientID,

		Type: n.Type,

		Actor: n.Actor,

		Resource: n.Resource,

		Title: n.Title,

		Body: n.Body,

		Action: n.Action,

		ReadAt: n.ReadAt,

		CreatedAt: n.CreatedAt,
	}

	payload, err :=
		json.Marshal(
			createdEvent,
		)

	if err != nil {
		return Notification{},
			fmt.Errorf(
				"marshal notification created event: %w",
				err,
			)
	}

	outbox := OutboxEvent{
		ID: bson.NewObjectID(),

		EventID: bson.NewObjectID().Hex(),

		Subject: "notification.created",

		Payload: payload,

		CreatedAt: now,

		PublishedAt: nil,
	}

	createdNotification, err :=
		s.repository.CreateWithOutbox(
			ctx,
			n,
			outbox,
		)

	if err != nil {
		return Notification{},
			fmt.Errorf(
				"create notification: %w",
				err,
			)
	}

	return createdNotification, nil
}
func (s *Service) List(
	ctx context.Context,
	AppID string,
	recipientID string,
	limit int64,
) ([]Notification, error) {
	if limit <= 0 {
		limit = 30
	}

	if limit > 100 {
		limit = 100
	}

	return s.repository.ListByRecipient(
		ctx,
		AppID,
		recipientID,
		limit,
	)
}

func (s *Service) MarkAsRead(
	ctx context.Context,
	AppID string,
	recipientID string,
	notificationID string,
) error {
	return s.repository.MarkAsRead(
		ctx,
		AppID,
		recipientID,
		notificationID,
	)
}
func (s *Service) UnreadCount(
	ctx context.Context,
	AppID string,
	recipientID string,
) (int64, error) {
	return s.repository.UnreadCount(
		ctx,
		AppID,
		recipientID,
	)
}

func (s *Service) MarkAllAsRead(
	ctx context.Context,
	AppID string,
	recipientID string,
) error {
	return s.repository.MarkAllAsRead(
		ctx,
		AppID,
		recipientID,
	)
}
