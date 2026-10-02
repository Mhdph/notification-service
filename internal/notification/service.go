package notification

import (
	"context"
	"time"
)

type CreateInput struct {
	WorkspaceID string
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

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	input CreateInput,
) error {
	n := Notification{
		WorkspaceID: input.WorkspaceID,
		RecipientID: input.RecipientID,

		Type: input.Type,

		Actor: input.Actor,

		Resource: input.Resource,

		Title: input.Title,
		Body:  input.Body,

		Action: input.Action,

		SourceEventID: input.SourceEventID,

		ReadAt: nil,

		CreatedAt: time.Now().UTC(),
	}

	return s.repository.Create(ctx, n)
}

func (s *Service) List(
	ctx context.Context,
	workspaceID string,
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
		workspaceID,
		recipientID,
		limit,
	)
}

func (s *Service) MarkAsRead(
	ctx context.Context,
	workspaceID string,
	recipientID string,
	notificationID string,
) error {
	return s.repository.MarkAsRead(
		ctx,
		workspaceID,
		recipientID,
		notificationID,
	)
}
