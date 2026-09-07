package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"
)

type Worker struct {
	repository Repository
	provider   DeliveryProvider
	logger     *slog.Logger
	workerID   string
	batchSize  int
}

func NewWorker(repository Repository, provider DeliveryProvider, logger *slog.Logger, workerID string) *Worker {
	return &Worker{repository: repository, provider: provider, logger: logger, workerID: workerID, batchSize: 20}
}

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	items, err := w.repository.ClaimBatch(ctx, w.workerID, w.batchSize)
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		if err := w.provider.Send(ctx, item); err != nil {
			dead := item.AttemptCount >= item.MaxAttempts
			delay := time.Duration(math.Min(math.Pow(2, float64(item.AttemptCount))*30, 3600)) * time.Second
			if markErr := w.repository.MarkFailed(ctx, item.ID, err.Error(), time.Now().UTC().Add(delay), dead); markErr != nil {
				return len(items), fmt.Errorf("delivery failed: %v; marking failure also failed: %w", err, markErr)
			}
			w.logger.Warn("notification delivery failed", "id", item.ID, "attempt", item.AttemptCount, "dead", dead, "error", err)
			continue
		}
		if err := w.repository.MarkDelivered(ctx, item.ID); err != nil {
			return len(items), err
		}
		w.logger.Info("notification delivered", "id", item.ID, "topic", item.Topic, "channel", item.Channel)
	}
	return len(items), nil
}

func (w *Worker) Run(ctx context.Context, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		if _, err := w.RunOnce(ctx); err != nil {
			w.logger.Error("notification worker iteration failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
