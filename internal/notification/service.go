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
	notifier   Notifier
}

func NewService(
	repository Repository,
	notifier Notifier,
) *Service {
	return &Service{
		repository: repository,
		notifier:   notifier,
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

	created, err := s.repository.Create(ctx, n)
	if err != nil {
		return err
	}

	if err := s.notifier.Notify(ctx, created); err != nil {
		return err
	}

	return nil
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
