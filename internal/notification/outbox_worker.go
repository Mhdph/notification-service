package notification

import (
	"context"
	"fmt"
	"time"
)

type OutboxWorker struct {
	repository OutboxRepository
	publisher  EventPublisher

	interval  time.Duration
	batchSize int64
}

func NewOutboxWorker(
	repository OutboxRepository,
	publisher EventPublisher,
	interval time.Duration,
	batchSize int64,
) *OutboxWorker {
	return &OutboxWorker{
		repository: repository,
		publisher:  publisher,
		interval:   interval,
		batchSize:  batchSize,
	}
}

func (w *OutboxWorker) Run(
	ctx context.Context,
) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	fmt.Println("Outbox worker started")

	// Don't wait for the first tick.
	w.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Outbox worker stopped")
			return

		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(
	ctx context.Context,
) {
	events, err := w.repository.FindUnpublished(
		ctx,
		w.batchSize,
	)

	if err != nil {
		fmt.Printf(
			"outbox: find unpublished events: %v\n",
			err,
		)
		return
	}

	for _, event := range events {
		if err := w.processEvent(
			ctx,
			event,
		); err != nil {
			fmt.Printf(
				"outbox: process event %s: %v\n",
				event.EventID,
				err,
			)
		}
	}
}

func (w *OutboxWorker) processEvent(
	ctx context.Context,
	event OutboxEvent,
) error {
	if err := w.publisher.Publish(
		ctx,
		event.Subject,
		event.Payload,
	); err != nil {
		return fmt.Errorf(
			"publish: %w",
			err,
		)
	}

	if err := w.repository.MarkPublished(
		ctx,
		event.ID,
	); err != nil {
		return fmt.Errorf(
			"mark published: %w",
			err,
		)
	}

	return nil
}
