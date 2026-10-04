package notification

import (
	"context"
	"fmt"
	"time"
)

type OutboxWorker struct {
	repository OutboxRepository
	publisher  EventPublisher

	workerID string

	interval      time.Duration
	leaseDuration time.Duration
	batchSize     int
}

func NewOutboxWorker(
	repository OutboxRepository,
	publisher EventPublisher,
	workerID string,
	interval time.Duration,
	leaseDuration time.Duration,
	batchSize int,
) *OutboxWorker {
	return &OutboxWorker{
		repository: repository,
		publisher:  publisher,

		workerID: workerID,

		interval:      interval,
		leaseDuration: leaseDuration,
		batchSize:     batchSize,
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
	for i := 0; i < w.batchSize; i++ {
		event, err := w.repository.ClaimNext(
			ctx,
			w.workerID,
			w.leaseDuration,
		)

		if err != nil {
			fmt.Printf(
				"outbox: claim event: %v\n",
				err,
			)
			return
		}

		if event == nil {
			return
		}

		if err := w.processEvent(
			ctx,
			*event,
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
		w.workerID,
	); err != nil {
		return fmt.Errorf(
			"mark published: %w",
			err,
		)
	}

	return nil
}
